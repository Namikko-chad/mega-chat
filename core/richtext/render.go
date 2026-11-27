package richtext

import (
	"html"
	"sort"
	"strings"
	"unicode"

	"github.com/SelfHold/parley/core/model"
)

// style says how a target markup writes each kind of span.
type style struct {
	marks   map[model.SpanKind][2]string
	link    func(url string) (string, string)
	esc     func(string) string // plain text
	code    func(string) string // the inside of code
	pre     func(lang, body string) string
	quote   func(body string) string
	heading func(body string) string
	keep    func(i int, span string) string // a Doc.Keep range; nil: as text
	// lines: blocks must stand on lines of their own (not HTML).
	lines bool
}

type byteSpan struct {
	kind   model.SpanKind
	lo, hi int
	url    string
	lang   string
}

func isBlock(k model.SpanKind) bool {
	return k == model.SpanPre || k == model.SpanQuote || k == model.SpanHeading
}

// byteSpans turns UTF-16 spans into byte spans of text; broken ones go, and
// marks do not hug spaces (Markdown would not read them back).
func byteSpans(text string, spans []model.TextSpan) []byteSpan {
	at := make([]int, 0, len(text)+1)
	for i, r := range text {
		at = append(at, i)
		if r >= 0x10000 {
			at = append(at, i)
		}
	}
	at = append(at, len(text))
	var out []byteSpan
	for _, s := range spans {
		if s.Length <= 0 || s.Offset < 0 || s.Offset+s.Length >= len(at) {
			continue
		}
		b := byteSpan{kind: s.Kind, lo: at[s.Offset], hi: at[s.Offset+s.Length], url: s.URL, lang: s.Lang}
		if s.Kind != model.SpanPre && s.Kind != model.SpanCode {
			t := text[b.lo:b.hi]
			b.lo += len(t) - len(strings.TrimLeftFunc(t, unicode.IsSpace))
			b.hi -= len(t) - len(strings.TrimRightFunc(t, unicode.IsSpace))
		}
		if b.hi > b.lo {
			out = append(out, b)
		}
	}
	return out
}

func render(d Doc, st style) string {
	spans := byteSpans(d.Text, d.Spans)
	var blocks, inline []byteSpan
	for _, s := range spans {
		if isBlock(s.kind) {
			blocks = append(blocks, s)
		} else {
			inline = append(inline, s)
		}
	}
	sort.SliceStable(blocks, func(i, j int) bool { return blocks[i].lo < blocks[j].lo })
	var b strings.Builder
	at := 0
	for _, s := range blocks {
		if s.lo < at {
			continue // overlaps the block before it
		}
		b.WriteString(renderInline(d, st, inline, at, s.lo))
		if st.lines && b.Len() > 0 && !strings.HasSuffix(b.String(), "\n") {
			b.WriteString("\n")
		}
		switch s.kind {
		case model.SpanPre:
			b.WriteString(st.pre(s.lang, st.code(d.Text[s.lo:s.hi])))
		case model.SpanQuote:
			b.WriteString(st.quote(renderInline(d, st, inline, s.lo, s.hi)))
		case model.SpanHeading:
			b.WriteString(st.heading(renderInline(d, st, inline, s.lo, s.hi)))
		}
		at = s.hi
		switch {
		case at >= len(d.Text):
		case d.Text[at] == '\n':
			if !st.lines {
				at++ // the block ends the line by itself
			}
		case st.lines:
			b.WriteString("\n")
		}
	}
	b.WriteString(renderInline(d, st, inline, at, len(d.Text)))
	return b.String()
}

// renderInline writes text[lo:hi]. Overlapping spans nest: a mark closes
// and reopens where needed, and none crosses a line.
func renderInline(d Doc, st style, spans []byteSpan, lo, hi int) string {
	if lo >= hi {
		return ""
	}
	cuts := []int{lo, hi}
	for _, s := range spans {
		if s.hi > lo && s.lo < hi {
			cuts = append(cuts, max(s.lo, lo), min(s.hi, hi))
		}
	}
	keepAt := map[int]int{}
	for i, k := range d.Keep {
		if k.Offset >= lo && k.Offset+k.Length <= hi && k.Length > 0 {
			keepAt[k.Offset] = i
			cuts = append(cuts, k.Offset, k.Offset+k.Length)
		}
	}
	for i := lo; i < hi; i++ {
		if d.Text[i] == '\n' {
			cuts = append(cuts, i, i+1)
		}
	}
	sort.Ints(cuts)

	var b strings.Builder
	var stack []*byteSpan
	closeTo := func(n int) {
		for len(stack) > n {
			b.WriteString(closeMark(st, stack[len(stack)-1]))
			stack = stack[:len(stack)-1]
		}
	}
	for c := 0; c+1 < len(cuts); c++ {
		s, e := cuts[c], cuts[c+1]
		if s == e {
			continue
		}
		seg := d.Text[s:e]
		if seg == "\n" {
			closeTo(0)
			b.WriteString(st.esc(seg))
			continue
		}
		var want []*byteSpan
		code := false
		for i := range spans {
			if spans[i].lo <= s && spans[i].hi >= e {
				want = append(want, &spans[i])
				code = code || spans[i].kind == model.SpanCode
			}
		}
		// Keep what is open and still wanted, from the outside in.
		n := 0
		for n < len(stack) && hasSpan(want, stack[n]) {
			n++
		}
		closeTo(n)
		sort.SliceStable(want, func(i, j int) bool { return want[i].hi-want[i].lo > want[j].hi-want[j].lo })
		for _, w := range want {
			if !hasSpan(stack, w) {
				b.WriteString(openMark(st, w))
				stack = append(stack, w)
			}
		}
		i, kept := keepAt[s]
		switch {
		case kept && st.keep != nil && d.Keep[i].Length == e-s:
			b.WriteString(st.keep(i, seg))
		case code:
			b.WriteString(st.code(seg))
		default:
			b.WriteString(st.esc(seg))
		}
	}
	closeTo(0)
	return b.String()
}

func hasSpan(list []*byteSpan, s *byteSpan) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func openMark(st style, s *byteSpan) string {
	if s.kind == model.SpanLink {
		o, _ := st.link(s.url)
		return o
	}
	return st.marks[s.kind][0]
}

func closeMark(st style, s *byteSpan) string {
	if s.kind == model.SpanLink {
		_, c := st.link(s.url)
		return c
	}
	return st.marks[s.kind][1]
}

func quoteLines(body string) string {
	lines := strings.Split(body, "\n")
	for i, l := range lines {
		lines[i] = "> " + l
	}
	return strings.Join(lines, "\n")
}

func identity(s string) string { return s }

// mdEscape backslashes the marks of s. A "_" inside a word is left alone: it
// opens nothing, and the parser keeps a backslash there (Windows paths).
func mdEscape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if strings.IndexByte(escapable, c) >= 0 && (c != '_' || !wordBefore(s, i)) {
			b.WriteByte('\\')
		}
		b.WriteByte(c)
	}
	return b.String()
}

func markdownStyle(esc func(string) string) style {
	return style{
		marks: map[model.SpanKind][2]string{
			model.SpanBold: {"**", "**"}, model.SpanItalic: {"*", "*"}, model.SpanUnderline: {"<u>", "</u>"},
			model.SpanStrike: {"~~", "~~"}, model.SpanSpoiler: {"||", "||"}, model.SpanCode: {"`", "`"},
		},
		link:    func(u string) (string, string) { return "[", "](" + u + ")" },
		esc:     esc,
		code:    identity,
		pre:     func(lang, body string) string { return "```" + lang + "\n" + body + "\n```" },
		quote:   quoteLines,
		heading: func(body string) string { return "## " + body },
		lines:   true,
	}
}

// ToMarkdown writes text and spans as Markdown that parses back to them: what
// the composer shows when a message is edited. Literal marks are escaped only
// when they would otherwise be read as formatting.
func ToMarkdown(text string, spans []model.TextSpan) string {
	d := Doc{Text: text, Spans: spans}
	md := render(d, markdownStyle(identity))
	if back := Parse(md, Markdown); back.Text == text && sameSpans(back.Spans, spans) {
		return md
	}
	return render(d, markdownStyle(mdEscape))
}

func sameSpans(a, b []model.TextSpan) bool {
	if len(a) != len(b) {
		return false
	}
	key := func(s model.TextSpan) model.TextSpan { s.Lang = ""; return s }
	seen := map[model.TextSpan]int{}
	for _, s := range a {
		seen[key(s)]++
	}
	for _, s := range b {
		if seen[key(s)] == 0 {
			return false
		}
		seen[key(s)]--
	}
	return true
}

// ToSlack writes a parsed Markdown text as mrkdwn, escaped; mention gives
// the text of the i-th Doc.Keep range. Slack has no underline or spoiler:
// those stay plain.
func ToSlack(d Doc, mention func(i int) string) string {
	esc := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace
	st := style{
		marks: map[model.SpanKind][2]string{
			model.SpanBold: {"*", "*"}, model.SpanItalic: {"_", "_"}, model.SpanStrike: {"~", "~"}, model.SpanCode: {"`", "`"},
		},
		link: func(u string) (string, string) {
			return "<" + strings.NewReplacer("|", "%7C", ">", "%3E", "<", "%3C").Replace(u) + "|", ">"
		},
		esc:     esc,
		code:    esc,
		pre:     func(_, body string) string { return "```\n" + body + "\n```" },
		quote:   quoteLines,
		heading: func(body string) string { return "*" + body + "*" },
		lines:   true,
	}
	if mention != nil {
		st.keep = func(i int, _ string) string { return mention(i) }
	}
	return render(d, st)
}

// ToHTML writes a parsed text as a safe HTML fragment for an e-mail.
func ToHTML(d Doc) string {
	esc := func(s string) string { return strings.ReplaceAll(html.EscapeString(s), "\n", "<br>\n") }
	st := style{
		marks: map[model.SpanKind][2]string{
			model.SpanBold: {"<b>", "</b>"}, model.SpanItalic: {"<i>", "</i>"}, model.SpanUnderline: {"<u>", "</u>"},
			model.SpanStrike: {"<s>", "</s>"}, model.SpanCode: {"<code>", "</code>"}, model.SpanSpoiler: {"<span>", "</span>"},
		},
		link: func(u string) (string, string) { return `<a href="` + html.EscapeString(u) + `">`, "</a>" },
		esc:  esc,
		code: html.EscapeString,
		pre:  func(_, body string) string { return "<pre>" + body + "</pre>\n" },
		quote: func(body string) string {
			return `<blockquote style="margin:0 0 0 8px;padding-left:8px;border-left:2px solid #ccc">` + body + "</blockquote>\n"
		},
		heading: func(body string) string { return "<h3>" + body + "</h3>\n" },
	}
	return render(d, st)
}
