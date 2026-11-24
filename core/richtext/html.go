package richtext

import (
	"strings"

	xhtml "golang.org/x/net/html"

	"megachat/core/model"
)

// FromHTML reduces an HTML part to text and spans. Only structure and
// emphasis survive: no scripts, styles, forms or remote content (a picture is
// its alt text), and links only to web and mail addresses.
func FromHTML(src string) Doc {
	if len(src) > 4*maxSource {
		src = src[:4*maxSource]
	}
	root, err := xhtml.Parse(strings.NewReader(src))
	if err != nil {
		return Doc{}
	}
	w := &htmlWriter{}
	w.walk(root, 0)
	return w.doc()
}

type htmlWriter struct {
	b     strings.Builder
	u16   int
	space bool // a collapsed space is due before the next word
	spans []model.TextSpan
}

// The HTML tree is a page from a stranger: its depth is limited.
const maxHTMLDepth = 256

func (w *htmlWriter) put(s string) {
	w.b.WriteString(s)
	w.u16 += utf16Len(s)
}

func (w *htmlWriter) text(t string, pre bool) {
	if pre {
		w.flushSpace()
		w.put(strings.ReplaceAll(t, "\r\n", "\n"))
		return
	}
	if t == "" {
		return
	}
	if strings.TrimLeftFunc(t, isHTMLSpace) != t {
		w.space = true
	}
	for i, f := range strings.FieldsFunc(t, isHTMLSpace) {
		if i > 0 {
			w.space = true
		}
		w.flushSpace()
		w.put(f)
	}
	if strings.TrimRightFunc(t, isHTMLSpace) != t {
		w.space = true
	}
}

func isHTMLSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f' || r == 0xa0
}

func (w *htmlWriter) flushSpace() {
	if w.space {
		if s := w.b.String(); s != "" && !strings.HasSuffix(s, "\n") && !strings.HasSuffix(s, " ") {
			w.put(" ")
		}
		w.space = false
	}
}

// line makes sure n line breaks end the text so far.
func (w *htmlWriter) line(n int) {
	w.space = false
	s := w.b.String()
	if s == "" {
		return
	}
	have := len(s) - len(strings.TrimRight(s, "\n"))
	for ; have < n; have++ {
		w.put("\n")
	}
}

func attr(n *xhtml.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// hidden: preheaders and tracking text are hidden by style.
func hidden(n *xhtml.Node) bool {
	if _, ok := attrOK(n, "hidden"); ok {
		return true
	}
	st := strings.ReplaceAll(strings.ToLower(attr(n, "style")), " ", "")
	return strings.Contains(st, "display:none") || strings.Contains(st, "visibility:hidden")
}

func attrOK(n *xhtml.Node, key string) (string, bool) {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val, true
		}
	}
	return "", false
}

// styleKinds reads the emphasis a style attribute gives.
func styleKinds(n *xhtml.Node) []model.SpanKind {
	st := strings.ReplaceAll(strings.ToLower(attr(n, "style")), " ", "")
	if st == "" {
		return nil
	}
	var out []model.SpanKind
	for _, d := range strings.Split(st, ";") {
		k, v, _ := strings.Cut(d, ":")
		switch k {
		case "font-weight":
			if v == "bold" || v == "bolder" || v == "600" || v == "700" || v == "800" || v == "900" {
				out = append(out, model.SpanBold)
			}
		case "font-style":
			if v == "italic" || v == "oblique" {
				out = append(out, model.SpanItalic)
			}
		case "text-decoration", "text-decoration-line":
			if strings.Contains(v, "underline") {
				out = append(out, model.SpanUnderline)
			}
			if strings.Contains(v, "line-through") {
				out = append(out, model.SpanStrike)
			}
		}
	}
	return out
}

var tagKinds = map[string]model.SpanKind{
	"b": model.SpanBold, "strong": model.SpanBold, "i": model.SpanItalic, "em": model.SpanItalic,
	"cite": model.SpanItalic, "u": model.SpanUnderline, "ins": model.SpanUnderline, "s": model.SpanStrike,
	"strike": model.SpanStrike, "del": model.SpanStrike, "code": model.SpanCode, "kbd": model.SpanCode,
	"samp": model.SpanCode, "tt": model.SpanCode,
	"h1": model.SpanHeading, "h2": model.SpanHeading, "h3": model.SpanHeading, "h4": model.SpanHeading,
	"h5": model.SpanHeading, "h6": model.SpanHeading, "blockquote": model.SpanQuote, "pre": model.SpanPre,
}

func isBlockTag(t string) bool {
	switch t {
	case "p", "div", "table", "ul", "ol", "dl", "section", "article", "header", "footer", "aside", "nav", "main",
		"h1", "h2", "h3", "h4", "h5", "h6", "blockquote", "pre", "hr", "figure", "address":
		return true
	}
	return false
}

func (w *htmlWriter) walk(n *xhtml.Node, depth int) { w.node(n, depth, false) }

func (w *htmlWriter) node(n *xhtml.Node, depth int, pre bool) {
	if depth > maxHTMLDepth {
		return
	}
	switch n.Type {
	case xhtml.TextNode:
		w.text(n.Data, pre)
		return
	case xhtml.ElementNode:
	default:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			w.node(c, depth+1, pre)
		}
		return
	}
	tag := n.Data
	switch tag {
	case "script", "style", "head", "title", "template", "noscript", "iframe", "object", "embed", "svg", "math",
		"form", "select", "textarea", "button", "audio", "video", "canvas", "meta", "link":
		return
	case "br":
		w.put("\n")
		w.space = false
		return
	case "img":
		if alt := strings.TrimSpace(attr(n, "alt")); alt != "" {
			w.flushSpace()
			w.put("[" + strings.Join(strings.Fields(alt), " ") + "]")
		}
		return
	}
	if hidden(n) {
		return
	}
	block := isBlockTag(tag)
	switch {
	case tag == "hr":
		w.line(2)
		w.put("———")
		w.line(2)
		return
	case block:
		if tag == "p" || tag == "table" || tag == "ul" || tag == "ol" || tag == "blockquote" || tag == "pre" || tag[0] == 'h' && len(tag) == 2 {
			w.line(2)
		} else {
			w.line(1)
		}
	case tag == "tr" || tag == "dt" || tag == "dd":
		w.line(1)
	case tag == "li":
		w.line(1)
		w.put(listMark(n))
	case tag == "td" || tag == "th":
		if s := w.b.String(); s != "" && !strings.HasSuffix(s, "\n") {
			w.space = false
			w.put("  ")
		}
	default:
		// A space due before a styled element stays outside its spans.
		if _, styled := tagKinds[tag]; styled || tag == "a" || tag == "span" || tag == "font" {
			w.flushSpace()
		}
	}

	type open struct {
		kind model.SpanKind
		url  string
		from int
	}
	var opens []open
	if k, ok := tagKinds[tag]; ok {
		opens = append(opens, open{kind: k, from: w.u16})
	}
	for _, k := range styleKinds(n) {
		opens = append(opens, open{kind: k, from: w.u16})
	}
	if tag == "a" {
		if h := strings.TrimSpace(attr(n, "href")); SafeURL(h) {
			opens = append(opens, open{kind: model.SpanLink, url: h, from: w.u16})
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		w.node(c, depth+1, pre || tag == "pre")
	}
	for _, o := range opens {
		if w.u16 > o.from {
			w.spans = append(w.spans, model.TextSpan{Kind: o.kind, Offset: o.from, Length: w.u16 - o.from, URL: o.url})
		}
	}
	switch {
	case block:
		if tag == "div" || tag == "section" || tag == "article" {
			w.line(1)
		} else {
			w.line(2)
		}
	case tag == "tr" || tag == "li" || tag == "dt" || tag == "dd":
		w.line(1)
	}
}

// listMark is "• ", or the item's number in an ordered list.
func listMark(n *xhtml.Node) string {
	if n.Parent == nil || n.Parent.Data != "ol" {
		return "• "
	}
	i := 1
	for c := n.Parent.FirstChild; c != nil && c != n; c = c.NextSibling {
		if c.Type == xhtml.ElementNode && c.Data == "li" {
			i++
		}
	}
	return itoa(i) + ". "
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}

// doc trims the outer blank lines, moving the spans with them, and caps runs
// of blank lines.
func (w *htmlWriter) doc() Doc {
	s := w.b.String()
	trimmed := strings.TrimLeft(s, " \n")
	shift := utf16Len(s[:len(s)-len(trimmed)])
	text := strings.TrimRight(trimmed, " \n")
	limit := utf16Len(text)
	var spans []model.TextSpan
	for _, sp := range w.spans {
		lo, hi := max(sp.Offset-shift, 0), min(sp.Offset+sp.Length-shift, limit)
		if hi > lo {
			sp.Offset, sp.Length = lo, hi-lo
			spans = append(spans, sp)
		}
	}
	return Doc{Text: text, Spans: normalize(text, spans)}
}

// normalize drops the line breaks that block spans picked up at their ends.
func normalize(text string, spans []model.TextSpan) []model.TextSpan {
	u := []rune(text)
	// UTF-16 index -> rune index is only needed when astral runes exist.
	idx := make([]int, 0, len(u)+1)
	for i, r := range u {
		idx = append(idx, i)
		if r >= 0x10000 {
			idx = append(idx, i)
		}
	}
	idx = append(idx, len(u))
	out := spans[:0]
	for _, s := range spans {
		if s.Offset+s.Length >= len(idx) {
			continue
		}
		for s.Length > 0 && u[idx[s.Offset+s.Length-1]] == '\n' {
			s.Length--
		}
		for s.Length > 0 && u[idx[s.Offset]] == '\n' {
			s.Offset++
			s.Length--
		}
		if s.Length > 0 {
			out = append(out, s)
		}
	}
	return out
}
