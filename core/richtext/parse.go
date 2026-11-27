// Package richtext is the one model of formatted text: visible text plus
// style spans (model.TextSpan, UTF-16 offsets). It reads Markdown, Slack's
// mrkdwn and HTML, and writes Markdown, mrkdwn and HTML. Nothing here fetches
// or executes anything: HTML is reduced to text and spans.
package richtext

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/SelfHold/parley/core/model"
)

// Dialect is the markup a source text is written in.
type Dialect int

const (
	// Markdown: **bold**, *italic*, _italic_, ~~strike~~, ||spoiler||,
	// <u>underline</u>, `code`, fences, [label](url), > quotes, # headings,
	// "- " lists.
	Markdown Dialect = iota
	// Slack: *bold*, _italic_, ~strike~, `code`, ``` blocks, > quotes.
	Slack
)

const (
	maxSource = 1 << 20 // a longer text is shown as it is
	maxLine   = 1 << 13 // so is a longer line: matching marks is quadratic
	maxDepth  = 6
)

// escapable are the characters a backslash makes literal. Not the backslash
// itself: "\\server" and "C:\dir" stay what they are.
const escapable = "*_~`[]|<"

// Range is a span of bytes of a text.
type Range struct{ Offset, Length int }

// Doc is parsed text. Keep holds where each range given to Parse ended up
// in Text, in the same order; Offset is -1 for one that did not survive.
type Doc struct {
	Text  string
	Spans []model.TextSpan
	Keep  []Range
}

// Parse reads src. The keep ranges (mentions) are copied verbatim and never
// read as markup.
func Parse(src string, d Dialect, keep ...Range) Doc {
	if len(src) > maxSource {
		return Doc{Text: src, Keep: keep}
	}
	if len(keep) == 0 && strings.Contains(src, "\r") {
		src = strings.ReplaceAll(src, "\r\n", "\n")
	}
	p := &parser{src: src, d: d, keep: make(map[int]int, len(keep)), index: make(map[int]int, len(keep))}
	p.kept = make([]Range, len(keep))
	for i, k := range keep {
		p.kept[i] = Range{Offset: -1}
		if k.Length > 0 && k.Offset >= 0 && k.Offset+k.Length <= len(src) {
			p.keep[k.Offset], p.index[k.Offset] = k.Offset+k.Length, i
			p.starts = append(p.starts, k.Offset)
		}
	}
	sort.Ints(p.starts)
	p.blocks()
	sort.SliceStable(p.spans, func(i, j int) bool { return p.spans[i].Offset < p.spans[j].Offset })
	return Doc{Text: p.out.String(), Spans: p.spans, Keep: p.kept}
}

type parser struct {
	src    string
	d      Dialect
	keep   map[int]int // start -> end of a verbatim range
	index  map[int]int // start -> its index in the Parse call
	starts []int
	kept   []Range
	out    strings.Builder
	u16    int
	spans  []model.TextSpan
}

func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		n += max(utf16.RuneLen(r), 1)
	}
	return n
}

func (p *parser) write(s string) {
	p.out.WriteString(s)
	p.u16 += utf16Len(s)
}

// raw copies src[lo:hi] as it is, noting where verbatim ranges land.
func (p *parser) raw(lo, hi int) {
	at := lo
	for _, s := range p.starts[sort.SearchInts(p.starts, lo):] {
		if s >= hi {
			break
		}
		e := p.keep[s]
		if s < at || e > hi {
			continue
		}
		p.write(p.src[at:s])
		p.kept[p.index[s]] = Range{Offset: p.out.Len(), Length: e - s}
		p.write(p.src[s:e])
		at = e
	}
	p.write(p.src[at:hi])
}

func (p *parser) span(k model.SpanKind, start int, url, lang string) {
	if p.u16 > start {
		p.spans = append(p.spans, model.TextSpan{Kind: k, Offset: start, Length: p.u16 - start, URL: url, Lang: lang})
	}
}

func (p *parser) lineEnd(pos int) int {
	if i := strings.IndexByte(p.src[pos:], '\n'); i >= 0 {
		return pos + i
	}
	return len(p.src)
}

// blocks reads the text line by line.
func (p *parser) blocks() {
	for pos := 0; pos < len(p.src); {
		if strings.HasPrefix(p.src[pos:], "```") {
			if next, ok := p.fence(pos); ok {
				pos = next
				continue
			}
		}
		e := p.lineEnd(pos)
		line := p.src[pos:e]
		switch {
		case quoteBody(line) >= 0:
			e = p.quote(pos)
		case p.d == Markdown && headingBody(line) > 0:
			start := p.u16
			p.inline(pos+headingBody(line), e, 0)
			p.span(model.SpanHeading, start, "", "")
		case p.d == Markdown && listBody(line) > 0:
			b := listBody(line)
			p.write(line[:b-2] + "• ")
			p.inline(pos+b, e, 0)
		default:
			p.inline(pos, e, 0)
		}
		if e < len(p.src) {
			p.write("\n")
		}
		pos = e + 1
	}
}

// fence reads a code block at pos and returns where to go on. Markdown's
// closes on a line of its own and may name a language; Slack's closes at the
// next ``` anywhere and has none.
func (p *parser) fence(pos int) (int, bool) {
	var lo, hi, next int
	lang := ""
	if p.d == Slack {
		lo = pos + 3
		i := strings.Index(p.src[lo:], "```")
		if i <= 0 {
			return 0, false
		}
		hi, next = lo+i, lo+i+3
		if p.src[lo] == '\n' {
			lo++
		}
		if hi > lo && p.src[hi-1] == '\n' {
			hi--
		}
	} else {
		e := p.lineEnd(pos)
		info := strings.TrimSpace(p.src[pos+3 : e])
		if e == len(p.src) || strings.ContainsAny(info, " \t`") {
			return 0, false
		}
		lang, lo = info, e+1
		found := false
		for at := lo; at <= len(p.src); {
			le := p.lineEnd(at)
			if strings.TrimRight(p.src[at:le], " \t\r") == "```" {
				hi, next, found = max(lo, at-1), le, true
				break
			}
			if le == len(p.src) {
				break
			}
			at = le + 1
		}
		if !found {
			return 0, false
		}
	}
	start := p.u16
	p.raw(lo, hi)
	p.span(model.SpanPre, start, "", lang)
	switch {
	case next >= len(p.src):
	case p.src[next] == '\n':
		p.write("\n")
		next++
	default:
		p.write("\n")
	}
	return next, true
}

// quoteBody is where the text of a "> " line starts; -1 for other lines.
// Nested marks fold into one level.
func quoteBody(line string) int {
	i := 0
	for i < len(line) && line[i] == '>' {
		i++
		if i < len(line) && line[i] == ' ' {
			i++
		}
	}
	if i == 0 {
		return -1
	}
	return i
}

// quote reads a run of quoted lines and returns the end of the last one.
func (p *parser) quote(pos int) int {
	start := p.u16
	for {
		e := p.lineEnd(pos)
		p.inline(pos+quoteBody(p.src[pos:e]), e, 0)
		if e >= len(p.src)-1 || quoteBody(p.src[e+1:p.lineEnd(e+1)]) < 0 {
			p.span(model.SpanQuote, start, "", "")
			return e
		}
		p.write("\n")
		pos = e + 1
	}
}

func headingBody(line string) int {
	i := 0
	for i < len(line) && i < 6 && line[i] == '#' {
		i++
	}
	if i == 0 || i >= len(line) || line[i] != ' ' || strings.TrimSpace(line[i:]) == "" {
		return 0
	}
	return i + 1
}

// listBody is where an item's text starts after its indent and "- ", 0 for
// other lines.
func listBody(line string) int {
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	if i+2 >= len(line) || !strings.ContainsRune("-*+", rune(line[i])) || line[i+1] != ' ' {
		return 0
	}
	if strings.TrimSpace(line[i+2:]) == "" {
		return 0
	}
	return i + 2
}

func (p *parser) inline(lo, hi, depth int) {
	if hi-lo > maxLine {
		p.raw(lo, hi)
		return
	}
	for i := lo; i < hi; {
		if e, ok := p.keep[i]; ok && e <= hi {
			p.raw(i, e)
			i = e
			continue
		}
		next := 0
		switch c := p.src[i]; c {
		case '\\':
			// "dir\_file" is a path: "_" inside a word is no mark to escape.
			if p.d == Markdown && i+1 < hi && strings.IndexByte(escapable, p.src[i+1]) >= 0 && (p.src[i+1] != '_' || !wordBefore(p.src, i)) {
				p.write(p.src[i+1 : i+2])
				i += 2
				continue
			}
		case '`':
			next = p.code(i, hi)
		case '[':
			if p.d == Markdown {
				next = p.link(i, hi, depth)
			}
		case '*', '_', '~', '|', '<':
			if depth < maxDepth {
				next = p.emphasis(i, hi, depth)
			}
		case 'h', 'H', 'w', 'W':
			if e := p.urlEnd(i, hi); e > i {
				p.raw(i, e)
				next = e
			}
		case '@':
			// A VK Teams mention, @[user_id@host].
			if i+1 < hi && p.src[i+1] == '[' {
				if j := strings.IndexByte(p.src[i:hi], ']'); j > 0 {
					p.raw(i, i+j+1)
					next = i + j + 1
				}
			}
		}
		if next > i {
			i = next
			continue
		}
		_, w := utf8.DecodeRuneInString(p.src[i:hi])
		p.write(p.src[i : i+w])
		i += w
	}
}

// code reads a `code` span at i; 0 when it is not closed.
func (p *parser) code(i, hi int) int {
	n := 0
	for i+n < hi && p.src[i+n] == '`' {
		n++
	}
	close := p.codeClose(i+n, hi, n)
	if close <= i+n {
		return 0
	}
	start := p.u16
	p.raw(i+n, close)
	if n >= 3 {
		p.span(model.SpanPre, start, "", "")
	} else {
		p.span(model.SpanCode, start, "", "")
	}
	return close + n
}

// codeClose finds a run of exactly n backticks from at; -1 without one.
func (p *parser) codeClose(at, hi, n int) int {
	for j := at; j < hi; {
		if p.src[j] != '`' {
			j++
			continue
		}
		r := 0
		for j+r < hi && p.src[j+r] == '`' {
			r++
		}
		if r == n {
			return j
		}
		j += r
	}
	return -1
}

// link reads [label](url); only web and mail addresses make a link.
func (p *parser) link(i, hi, depth int) int {
	nest, close := 0, -1
	for j := i; j < hi && close < 0; j++ {
		switch p.src[j] {
		case '\\':
			j++
		case '[':
			nest++
		case ']':
			if nest--; nest == 0 {
				close = j
			}
		}
	}
	if close <= i+1 || close+1 >= hi || p.src[close+1] != '(' {
		return 0
	}
	nest = 0
	end := -1
	for j := close + 1; j < hi && end < 0; j++ {
		switch p.src[j] {
		case '(':
			nest++
		case ')':
			if nest--; nest == 0 {
				end = j
			}
		case ' ', '\t':
			return 0
		}
	}
	if end < 0 {
		return 0
	}
	url := p.src[close+2 : end]
	if !SafeURL(url) {
		return 0
	}
	start := p.u16
	p.inline(i+1, close, depth+1)
	p.span(model.SpanLink, start, url, "")
	return end + 1
}

// SafeURL tells whether a link may be opened: web and mail addresses only.
func SafeURL(u string) bool {
	l := strings.ToLower(u)
	for _, s := range []string{"http://", "https://", "mailto:"} {
		if strings.HasPrefix(l, s) && len(l) > len(s) {
			return true
		}
	}
	return false
}

// urlEnd is the end of a bare address at i (0: none there), so that the
// underscores and stars inside one are not read as marks.
func (p *parser) urlEnd(i, hi int) int {
	rest := p.src[i:hi]
	ok := false
	for _, s := range []string{"http://", "https://", "www."} {
		if len(rest) > len(s) && strings.EqualFold(rest[:len(s)], s) {
			ok = true
		}
	}
	if !ok || wordBefore(p.src, i) {
		return 0
	}
	j := i
	for j < hi && !strings.ContainsRune(" \t\n<>\"", rune(p.src[j])) {
		j++
	}
	for j > i && strings.IndexByte("*_~|`", p.src[j-1]) >= 0 {
		j--
	}
	return j
}

func wordBefore(s string, i int) bool {
	if i <= 0 {
		return false
	}
	r, _ := utf8.DecodeLastRuneInString(s[:i])
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func wordAfter(s string, i, hi int) bool {
	if i >= hi {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s[i:hi])
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

type mark struct {
	open, close string
	kind        model.SpanKind
	// inWord: may touch letters outside it. Single marks may not, so that
	// 2*3*4 and snake_case_name stay text.
	inWord bool
}

var markdownMarks = []mark{
	{"<u>", "</u>", model.SpanUnderline, true},
	{"**", "**", model.SpanBold, true},
	{"__", "__", model.SpanBold, false},
	{"~~", "~~", model.SpanStrike, true},
	{"||", "||", model.SpanSpoiler, true},
	{"*", "*", model.SpanItalic, false},
	{"_", "_", model.SpanItalic, false},
}

var slackMarks = []mark{
	{"**", "**", model.SpanBold, true},
	{"~~", "~~", model.SpanStrike, true},
	{"*", "*", model.SpanBold, false},
	{"_", "_", model.SpanItalic, false},
	{"~", "~", model.SpanStrike, false},
}

// emphasis reads a marked run at i; 0 when the mark opens nothing.
func (p *parser) emphasis(i, hi, depth int) int {
	marks := markdownMarks
	if p.d == Slack {
		marks = slackMarks
	}
	for _, m := range marks {
		if !strings.HasPrefix(p.src[i:hi], m.open) {
			continue
		}
		body := i + len(m.open)
		if body >= hi || isSpace(p.src, body, hi) || (!m.inWord && wordBefore(p.src, i)) {
			return 0
		}
		j := p.closing(m, body, hi)
		if j < 0 {
			continue
		}
		start := p.u16
		p.inline(body, j, depth+1)
		p.span(m.kind, start, "", "")
		return j + len(m.close)
	}
	return 0
}

func isSpace(s string, i, hi int) bool {
	r, _ := utf8.DecodeRuneInString(s[i:hi])
	return unicode.IsSpace(r)
}

// closing finds the mark closing a run whose text starts at body; -1 if
// none. Code, addresses and verbatim ranges are not looked into.
func (p *parser) closing(m mark, body, hi int) int {
	for j := body; j < hi; {
		if e, ok := p.keep[j]; ok && e <= hi {
			j = e
			continue
		}
		switch p.src[j] {
		case '\\':
			j += 2
			continue
		case '`':
			n := 0
			for j+n < hi && p.src[j+n] == '`' {
				n++
			}
			if c := p.codeClose(j+n, hi, n); c > 0 {
				j = c + n
			} else {
				j += n
			}
			continue
		case 'h', 'H', 'w', 'W':
			if e := p.urlEnd(j, hi); e > j {
				j = e
				continue
			}
		}
		if !strings.HasPrefix(p.src[j:hi], m.close) {
			j++
			continue
		}
		if len(m.close) == 1 {
			// A single mark is not a piece of a double one.
			r := 0
			for j+r < hi && p.src[j+r] == m.close[0] {
				r++
			}
			if r > 1 {
				j += r
				continue
			}
		} else if m.close[0] != '<' && j+len(m.close) < hi && p.src[j+len(m.close)] == m.close[0] {
			j++ // "***": the last two close, the first belongs to the text
			continue
		}
		if j > body && !endsSpace(p.src[body:j]) && (m.inWord || !wordAfter(p.src, j+len(m.close), hi)) {
			return j
		}
		j += len(m.close)
	}
	return -1
}

func endsSpace(s string) bool {
	r, _ := utf8.DecodeLastRuneInString(s)
	return unicode.IsSpace(r)
}
