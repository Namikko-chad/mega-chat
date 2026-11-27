package mailru

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
)

// Calendar objects are read and written here rather than through
// go-webdav's caldav client, which rejects a whole answer for one thing it
// does not like. Mail.ru sends ETags without quotes and answers a
// calendar-query with responses lacking calendar-data (404 in their
// propstat); both are taken as they come.

// davRaw sends requests to a CalDAV server.
type davRaw struct {
	hc   webdav.HTTPClient // basic auth and keepSlash already in
	base *url.URL
}

// calObject is a calendar object: its path, ETag (unquoted) and data, nil
// when not asked for or not given.
type calObject struct {
	Path, ETag string
	Data       *ical.Calendar
}

func (d davRaw) url(path string) string {
	return d.base.ResolveReference(&url.URL{Path: path}).String()
}

func (d davRaw) do(ctx context.Context, method, path string, hdr map[string]string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, d.url(path), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	resp, err := d.hc.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		resp.Body.Close()
		return nil, fmt.Errorf("caldav: %s %s: %s", method, path, resp.Status)
	}
	return resp, nil
}

// unquoteETag takes an ETag quoted, weak or bare.
func unquoteETag(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "W/")
	return strings.Trim(s, `"`)
}

type msPropstat struct {
	Status string `xml:"DAV: status"`
	Prop   struct {
		ETag string `xml:"DAV: getetag"`
		Data string `xml:"urn:ietf:params:xml:ns:caldav calendar-data"`
	} `xml:"DAV: prop"`
}

type multistatus struct {
	Responses []struct {
		Href      string       `xml:"DAV: href"`
		Status    string       `xml:"DAV: status"`
		Propstats []msPropstat `xml:"DAV: propstat"`
	} `xml:"DAV: response"`
}

// rawObject is one response: data still text, "" when missing.
type rawObject struct {
	path, etag, data string
}

func (d davRaw) report(ctx context.Context, path, body string) ([]rawObject, error) {
	resp, err := d.do(ctx, "REPORT", path, map[string]string{"Depth": "1", "Content-Type": "application/xml; charset=utf-8"}, []byte(xml.Header+body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var ms multistatus
	if err := xml.NewDecoder(io.LimitReader(resp.Body, 64<<20)).Decode(&ms); err != nil {
		return nil, fmt.Errorf("caldav: REPORT %s: %w", path, err)
	}
	var out []rawObject
	for _, r := range ms.Responses {
		u, err := url.Parse(strings.TrimSpace(r.Href))
		if err != nil || u.Path == "" || strings.HasSuffix(u.Path, "/") {
			continue // a collection, not an object
		}
		o := rawObject{path: u.Path}
		for _, ps := range r.Propstats {
			if ps.Status != "" && !strings.Contains(ps.Status, " 200") {
				continue
			}
			if ps.Prop.ETag != "" {
				o.etag = unquoteETag(ps.Prop.ETag)
			}
			if strings.TrimSpace(ps.Prop.Data) != "" {
				o.data = ps.Prop.Data
			}
		}
		out = append(out, o)
	}
	return out, nil
}

const davTime = "20060102T150405Z"

const davNS = `xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav"`

// query lists the objects of a calendar with an event in [from, to); with
// data it reads them too, asking again (multiget, then GET) for those the
// query came back without. An object that cannot be read is left out and
// reported to logf.
func (d davRaw) query(ctx context.Context, cal string, from, to time.Time, data bool, logf func(string, ...any)) ([]calObject, error) {
	prop := "<D:getetag/>"
	if data {
		prop += "<C:calendar-data/>"
	}
	body := fmt.Sprintf(`<C:calendar-query %s><D:prop>%s</D:prop><C:filter><C:comp-filter name="VCALENDAR"><C:comp-filter name="VEVENT"><C:time-range start="%s" end="%s"/></C:comp-filter></C:comp-filter></C:filter></C:calendar-query>`,
		davNS, prop, from.UTC().Format(davTime), to.UTC().Format(davTime))
	raws, err := d.report(ctx, cal, body)
	if err != nil {
		return nil, err
	}
	if !data {
		out := make([]calObject, len(raws))
		for i, r := range raws {
			out[i] = calObject{Path: r.path, ETag: r.etag}
		}
		return out, nil
	}
	var missing []string
	for _, r := range raws {
		if r.data == "" {
			missing = append(missing, r.path)
		}
	}
	if len(missing) > 0 {
		got := d.multiget(ctx, cal, missing)
		for i, r := range raws {
			if r.data == "" {
				if g, ok := got[r.path]; ok {
					raws[i].data = g.data
					if g.etag != "" {
						raws[i].etag = g.etag
					}
				}
			}
		}
	}
	var out []calObject
	for _, r := range raws {
		if r.data == "" {
			o, err := d.get(ctx, r.path)
			if err != nil {
				logf("caldav: %s: no calendar data: %v", r.path, err)
				continue
			}
			out = append(out, o)
			continue
		}
		c, err := ical.NewDecoder(strings.NewReader(r.data)).Decode()
		if err != nil {
			logf("caldav: %s: %v", r.path, err)
			continue
		}
		out = append(out, calObject{Path: r.path, ETag: r.etag, Data: c})
	}
	return out, nil
}

// multiget reads objects by path; what fails is simply not in the answer.
func (d davRaw) multiget(ctx context.Context, cal string, paths []string) map[string]rawObject {
	var b strings.Builder
	fmt.Fprintf(&b, `<C:calendar-multiget %s><D:prop><D:getetag/><C:calendar-data/></D:prop>`, davNS)
	for _, p := range paths {
		b.WriteString("<D:href>")
		xml.EscapeText(&b, []byte((&url.URL{Path: p}).EscapedPath()))
		b.WriteString("</D:href>")
	}
	b.WriteString("</C:calendar-multiget>")
	raws, err := d.report(ctx, cal, b.String())
	out := map[string]rawObject{}
	if err != nil {
		return out
	}
	for _, r := range raws {
		if r.data != "" {
			out[r.path] = r
		}
	}
	return out
}

// get reads one object.
func (d davRaw) get(ctx context.Context, path string) (calObject, error) {
	resp, err := d.do(ctx, http.MethodGet, path, map[string]string{"Accept": "text/calendar"}, nil)
	if err != nil {
		return calObject{}, err
	}
	defer resp.Body.Close()
	c, err := ical.NewDecoder(io.LimitReader(resp.Body, 16<<20)).Decode()
	if err != nil {
		return calObject{}, fmt.Errorf("caldav: %s: %w", path, err)
	}
	return calObject{Path: path, ETag: unquoteETag(resp.Header.Get("ETag")), Data: c}, nil
}

// put writes one object.
func (d davRaw) put(ctx context.Context, path string, cal *ical.Calendar) error {
	var buf bytes.Buffer
	if err := ical.NewEncoder(&buf).Encode(cal); err != nil {
		return err
	}
	resp, err := d.do(ctx, http.MethodPut, path, map[string]string{"Content-Type": ical.MIMEType}, buf.Bytes())
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
