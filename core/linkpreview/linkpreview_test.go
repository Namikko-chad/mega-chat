package linkpreview

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"testing"
)

const page = `<!doctype html><html><head>
<meta charset="utf-8"><title> Fallback   title </title>
<meta property="og:site_name" content="Example">
<meta property="og:title" content="A  page">
<meta name="description" content="plain description">
<meta property="og:description" content="What it is about">
<meta property="og:image" content="/img.png">
</head><body><meta property="og:title" content="not this"></body></html>`

func TestParse(t *testing.T) {
	base, _ := url.Parse("https://example.com/a/b")
	p := parse([]byte(page), "text/html; charset=utf-8", base)
	if p.SiteName != "Example" || p.Title != "A page" || p.Description != "What it is about" || p.imageURL != "https://example.com/img.png" {
		t.Fatalf("%+v", p)
	}
	p = parse([]byte(`<html><head><title>Only title</title><meta name="description" content="d"></head></html>`), "text/html", base)
	if p.Title != "Only title" || p.Description != "d" || p.SiteName != "example.com" || p.imageURL != "" {
		t.Fatalf("fallbacks: %+v", p)
	}
	p = parse([]byte(`<head><meta property="og:title" content="x"><meta property="og:image" content="javascript:alert(1)"></head>`), "text/html", base)
	if p.imageURL != "" {
		t.Fatalf("image off the web: %q", p.imageURL)
	}
	long := strings.Repeat("я", 400)
	if p = parse([]byte(`<head><meta property="og:description" content="`+long+`"></head>`), "text/html", base); len([]rune(p.Description)) != maxText+1 {
		t.Fatalf("description not cut: %d", len([]rune(p.Description)))
	}
}

func TestFetch(t *testing.T) {
	png := "\x89PNG\r\n\x1a\nfake"
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/page":
			hits++
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(page))
		case "/img.png":
			w.Header().Set("Content-Type", "image/png")
			w.Write([]byte(png))
		case "/moved":
			http.Redirect(w, r, "/page", http.StatusFound)
		case "/file":
			w.Header().Set("Content-Type", "application/pdf")
			w.Write([]byte("%PDF"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	f := newFetcher(func(netip.Addr) bool { return true })
	ctx := context.Background()
	p, err := f.Fetch(ctx, srv.URL+"/moved#frag")
	if err != nil || p.Title != "A page" || string(p.Image) != png || p.ImageMIME != "image/png" || p.URL != srv.URL+"/page" {
		t.Fatalf("%+v %v", p, err)
	}
	if _, err := f.Fetch(ctx, srv.URL+"/moved"); err != nil || hits != 1 {
		t.Fatalf("not cached: %d %v", hits, err)
	}
	if _, err := f.Fetch(ctx, srv.URL+"/file"); err != ErrNoPreview {
		t.Fatalf("pdf: %v", err)
	}
	if _, err := f.Fetch(ctx, "ftp://example.com/x"); err == nil {
		t.Fatal("ftp accepted")
	}
	// The real fetcher does not reach this machine or the local network.
	if _, err := New().Fetch(ctx, srv.URL+"/page"); err == nil || !strings.Contains(err.Error(), "not a public address") {
		t.Fatalf("loopback reached: %v", err)
	}
}

func TestPublicOnly(t *testing.T) {
	for _, s := range []string{"127.0.0.1", "10.1.2.3", "192.168.0.1", "172.16.0.1", "169.254.169.254", "100.64.0.1", "0.0.0.0", "::1", "fe80::1", "fc00::1", "64:ff9b::a00:1"} {
		if publicOnly(netip.MustParseAddr(s)) {
			t.Errorf("%s allowed", s)
		}
	}
	for _, s := range []string{"93.184.216.34", "2606:2800:220:1::1"} {
		if !publicOnly(netip.MustParseAddr(s)) {
			t.Errorf("%s refused", s)
		}
	}
}
