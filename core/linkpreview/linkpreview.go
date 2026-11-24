// Package linkpreview reads a page's Open Graph tags (og:title,
// og:description, og:image, og:site_name) and fetches its picture, for the
// preview of a link in a message.
//
// The links come from other people, so the fetcher reaches only public
// addresses: loopback, private, link-local and other special networks are
// refused on every connection, redirects included. Pages and pictures are
// read up to a size limit, with a timeout, and nothing is sent but a GET
// without cookies.
package linkpreview

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

const (
	maxPage    = 1 << 20 // the head of a page is enough
	maxImage   = 3 << 20
	maxText    = 300 // description, in runes
	timeout    = 10 * time.Second
	cacheSize  = 200
	userAgent  = "Mozilla/5.0 (compatible; MegaChat link preview)"
	acceptHTML = "text/html,application/xhtml+xml;q=0.9,*/*;q=0.1"
)

// ErrNoPreview means the page has nothing to show (not HTML, no title).
var ErrNoPreview = errors.New("linkpreview: nothing to show")

// Preview is what a link shows.
type Preview struct {
	URL         string // the page after redirects
	SiteName    string
	Title       string
	Description string
	Image       []byte // nil: none or it did not load
	ImageMIME   string
}

// Fetcher fetches previews and keeps the latest in memory.
type Fetcher struct {
	client *http.Client

	mu    sync.Mutex
	cache map[string]cached
	order []string
}

type cached struct {
	p   Preview
	err error
}

// New returns a fetcher that reaches only public addresses.
func New() *Fetcher {
	return newFetcher(publicOnly)
}

func newFetcher(allow func(netip.Addr) bool) *Fetcher {
	d := &net.Dialer{
		Timeout: timeout,
		Control: func(_, address string, _ syscall.RawConn) error {
			ap, err := netip.ParseAddrPort(address)
			if err != nil {
				return err
			}
			if !allow(ap.Addr().Unmap()) {
				return fmt.Errorf("linkpreview: %s is not a public address", ap.Addr())
			}
			return nil
		},
	}
	tr := &http.Transport{
		Proxy:                 nil, // a proxy would dial for us, past the check
		DialContext:           d.DialContext,
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
	}
	return &Fetcher{
		client: &http.Client{
			Transport: tr,
			Timeout:   timeout,
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return errors.New("linkpreview: too many redirects")
				}
				if r.URL.Scheme != "http" && r.URL.Scheme != "https" {
					return errors.New("linkpreview: redirect off the web")
				}
				return nil
			},
		},
		cache: map[string]cached{},
	}
}

// publicOnly refuses loopback, private, link-local, multicast, CGNAT,
// unspecified and documentation networks.
func publicOnly(a netip.Addr) bool {
	if !a.IsValid() || a.IsLoopback() || a.IsPrivate() || a.IsLinkLocalUnicast() || a.IsLinkLocalMulticast() ||
		a.IsInterfaceLocalMulticast() || a.IsMulticast() || a.IsUnspecified() {
		return false
	}
	for _, p := range special {
		if p.Contains(a) {
			return false
		}
	}
	return true
}

var special = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"), // CGNAT
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b::/96"), // NAT64 can reach private IPv4
	netip.MustParsePrefix("2001:db8::/32"),
}

// Fetch returns the preview of a http(s) link (cached, errors too).
func (f *Fetcher) Fetch(ctx context.Context, link string) (Preview, error) {
	u, err := url.Parse(strings.TrimSpace(link))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Preview{}, fmt.Errorf("linkpreview: not a web link: %q", link)
	}
	u.Fragment = ""
	key := u.String()
	f.mu.Lock()
	c, ok := f.cache[key]
	f.mu.Unlock()
	if ok {
		return c.p, c.err
	}
	p, err := f.fetch(ctx, u)
	if ctx.Err() != nil {
		return p, err // cancelled: try again next time
	}
	f.mu.Lock()
	if _, ok := f.cache[key]; !ok {
		f.order = append(f.order, key)
		if len(f.order) > cacheSize {
			delete(f.cache, f.order[0])
			f.order = f.order[1:]
		}
	}
	f.cache[key] = cached{p, err}
	f.mu.Unlock()
	return p, err
}

func (f *Fetcher) get(ctx context.Context, u string, accept string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", accept)
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("linkpreview: %s", resp.Status)
	}
	return resp, nil
}

func (f *Fetcher) fetch(ctx context.Context, u *url.URL) (Preview, error) {
	resp, err := f.get(ctx, u.String(), acceptHTML)
	if err != nil {
		return Preview{}, err
	}
	defer resp.Body.Close()
	ct, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if ct != "text/html" && ct != "application/xhtml+xml" {
		return Preview{}, ErrNoPreview
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxPage))
	if err != nil && len(body) == 0 {
		return Preview{}, err
	}
	base := resp.Request.URL
	p := parse(body, resp.Header.Get("Content-Type"), base)
	p.URL = base.String()
	if p.Title == "" && p.Description == "" {
		return Preview{}, ErrNoPreview
	}
	if p.imageURL != "" {
		p.Image, p.ImageMIME = f.image(ctx, p.imageURL)
	}
	return p.Preview, nil
}

// image fetches og:image; nothing when it is not a picture or too big.
func (f *Fetcher) image(ctx context.Context, u string) ([]byte, string) {
	resp, err := f.get(ctx, u, "image/*")
	if err != nil {
		return nil, ""
	}
	defer resp.Body.Close()
	ct, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if !strings.HasPrefix(ct, "image/") || ct == "image/svg+xml" {
		return nil, ""
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxImage+1))
	if err != nil || len(b) > maxImage || len(b) == 0 {
		return nil, ""
	}
	return b, ct
}

// parsed is a preview with the picture's address, before it is fetched.
type parsed struct {
	Preview
	imageURL string
}

// parse reads the Open Graph tags of a page (Twitter cards and <title> /
// meta description as fallbacks); relative addresses resolve against base.
func parse(page []byte, contentType string, base *url.URL) parsed {
	var r io.Reader = bytes.NewReader(page)
	if cr, err := charset.NewReader(r, contentType); err == nil {
		r = cr
	}
	meta := map[string]string{}
	var title strings.Builder
	inTitle := false
	z := html.NewTokenizer(r)
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}
		name, hasAttr := z.TagName()
		tag := string(name)
		switch {
		case tt == html.StartTagToken && tag == "body":
			goto done // the tags are in the head
		case (tt == html.StartTagToken || tt == html.SelfClosingTagToken) && tag == "meta" && hasAttr:
			var key, content string
			for {
				k, v, more := z.TagAttr()
				switch strings.ToLower(string(k)) {
				case "property", "name":
					if key == "" {
						key = strings.ToLower(strings.TrimSpace(string(v)))
					}
				case "content":
					content = string(v)
				}
				if !more {
					break
				}
			}
			if key != "" && content != "" {
				if _, seen := meta[key]; !seen {
					meta[key] = clean(content)
				}
			}
		case tt == html.StartTagToken && tag == "title":
			inTitle = true
		case tt == html.EndTagToken && tag == "title":
			inTitle = false
		case tt == html.TextToken && inTitle:
			title.Write(z.Text())
		}
	}
done:
	first := func(keys ...string) string {
		for _, k := range keys {
			if v := meta[k]; v != "" {
				return v
			}
		}
		return ""
	}
	var p parsed
	p.SiteName = first("og:site_name")
	if p.SiteName == "" && base != nil {
		p.SiteName = strings.TrimPrefix(base.Hostname(), "www.")
	}
	p.Title = first("og:title", "twitter:title")
	if p.Title == "" {
		p.Title = clean(title.String())
	}
	p.Description = cut(first("og:description", "twitter:description", "description"), maxText)
	if img := first("og:image:secure_url", "og:image", "og:image:url", "twitter:image", "twitter:image:src"); img != "" && base != nil {
		if iu, err := base.Parse(img); err == nil && (iu.Scheme == "http" || iu.Scheme == "https") {
			p.imageURL = iu.String()
		}
	}
	return p
}

func clean(s string) string { return strings.Join(strings.Fields(s), " ") }

func cut(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)
	return strings.TrimSpace(string(r[:n])) + "…"
}
