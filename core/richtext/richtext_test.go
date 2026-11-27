package richtext

import (
	"strings"
	"testing"

	"github.com/SelfHold/parley/core/model"
)

func kinds(d Doc) map[model.SpanKind]string {
	u := []rune(d.Text)
	out := map[model.SpanKind]string{}
	for _, s := range d.Spans {
		out[s.Kind] = string(u[s.Offset : s.Offset+s.Length])
	}
	return out
}

func TestMarkdownInline(t *testing.T) {
	d := Parse("a **bold** *it* ~~no~~ ||hid|| <u>un</u> `c*d` [site](https://x.io/a_b)", Markdown)
	if d.Text != "a bold it no hid un c*d site" {
		t.Fatalf("text %q", d.Text)
	}
	k := kinds(d)
	for kind, want := range map[model.SpanKind]string{
		model.SpanBold: "bold", model.SpanItalic: "it", model.SpanStrike: "no", model.SpanSpoiler: "hid",
		model.SpanUnderline: "un", model.SpanCode: "c*d", model.SpanLink: "site",
	} {
		if k[kind] != want {
			t.Errorf("kind %d = %q, want %q", kind, k[kind], want)
		}
	}
}

func TestPlainTextStaysText(t *testing.T) {
	for _, s := range []string{
		"2*3*4 = 24", "snake_case_name", "https://x.io/a_b_c*d", "C:\\dir\\_file", "a * b * c", "user_1@mail.ru",
		"[x](javascript:alert(1))", "@[user_name@host]", "#hashtag", "-- ", "price: 5 * 3",
	} {
		if d := Parse(s, Markdown); d.Text != s || len(d.Spans) != 0 {
			t.Errorf("%q became %q %v", s, d.Text, d.Spans)
		}
	}
}

func TestBlocks(t *testing.T) {
	d := Parse("# Title\n> one\n> two\n- item\n```go\nx := *p\n```\nend", Markdown)
	if d.Text != "Title\none\ntwo\n• item\nx := *p\nend" {
		t.Fatalf("text %q", d.Text)
	}
	k := kinds(d)
	if k[model.SpanHeading] != "Title" || k[model.SpanQuote] != "one\ntwo" || k[model.SpanPre] != "x := *p" {
		t.Fatalf("spans %v", k)
	}
}

func TestSlackDialect(t *testing.T) {
	d := Parse("*bold* _it_ ~gone~\n```a\nb```", Slack)
	if d.Text != "bold it gone\na\nb" {
		t.Fatalf("text %q", d.Text)
	}
	k := kinds(d)
	if k[model.SpanBold] != "bold" || k[model.SpanItalic] != "it" || k[model.SpanStrike] != "gone" || k[model.SpanPre] != "a\nb" {
		t.Fatalf("spans %v", k)
	}
}

func TestUTF16Offsets(t *testing.T) {
	d := Parse("😀 **жир**", Markdown)
	if len(d.Spans) != 1 || d.Spans[0].Offset != 3 || d.Spans[0].Length != 3 {
		t.Fatalf("spans %+v", d.Spans)
	}
}

func TestKeepRanges(t *testing.T) {
	src := "**hi** @some_user_x ok"
	at := strings.Index(src, "@")
	d := Parse(src, Markdown, Range{Offset: at, Length: len("@some_user_x")})
	k := d.Keep[0]
	if d.Text[k.Offset:k.Offset+k.Length] != "@some_user_x" {
		t.Fatalf("keep %+v in %q", k, d.Text)
	}
}

func TestToMarkdownRoundTrip(t *testing.T) {
	for _, src := range []string{
		"plain", "a **bold** b", "*it* and `code`", "[l](https://x.io)", "> q1\n> q2\nafter", "```go\ncode\n```\nafter", "||s|| <u>u</u> ~~x~~",
	} {
		d := Parse(src, Markdown)
		back := Parse(ToMarkdown(d.Text, d.Spans), Markdown)
		if back.Text != d.Text || !sameSpans(back.Spans, d.Spans) {
			t.Errorf("%q: got %q %v, want %q %v", src, back.Text, back.Spans, d.Text, d.Spans)
		}
	}
}

func TestToMarkdownEscapesLiteralMarks(t *testing.T) {
	text := "**not bold**"
	back := Parse(ToMarkdown(text, nil), Markdown)
	if back.Text != text || len(back.Spans) != 0 {
		t.Fatalf("got %q %v", back.Text, back.Spans)
	}
}

func TestToSlack(t *testing.T) {
	src := "**b** *i* ~~s~~ [l](https://x.io) a<b @ann"
	at := strings.Index(src, "@ann")
	d := Parse(src, Markdown, Range{Offset: at, Length: 4})
	got := ToSlack(d, func(int) string { return "<@U1>" })
	if want := "*b* _i_ ~s~ <https://x.io|l> a&lt;b <@U1>"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFromHTML(t *testing.T) {
	d := FromHTML(`<html><head><title>x</title><style>p{}</style></head><body>
	<div style="display:none">preheader</div><h1>Hello</h1>
	<p>Some <b>bold</b> and <a href="https://x.io">link</a> <a href="javascript:alert(1)">bad</a></p>
	<script>alert(1)</script><img src="https://track.example/p.gif"><img src="x" alt="Logo">
	<ul><li>one</li><li>two</li></ul><blockquote>quoted</blockquote><pre>a  b
c</pre></body></html>`)
	for _, bad := range []string{"preheader", "alert", "track", "p{}"} {
		if strings.Contains(d.Text, bad) {
			t.Errorf("%q leaked into %q", bad, d.Text)
		}
	}
	for _, want := range []string{"Hello", "Some bold and link bad", "[Logo]", "• one\n• two", "a  b\nc"} {
		if !strings.Contains(d.Text, want) {
			t.Errorf("no %q in %q", want, d.Text)
		}
	}
	k := kinds(d)
	if k[model.SpanHeading] != "Hello" || k[model.SpanBold] != "bold" || k[model.SpanLink] != "link" || k[model.SpanQuote] != "quoted" {
		t.Fatalf("spans %v", k)
	}
	for _, s := range d.Spans {
		if s.Kind == model.SpanLink && s.URL != "https://x.io" {
			t.Fatalf("unsafe link kept: %q", s.URL)
		}
	}
}

func TestToHTMLEscapes(t *testing.T) {
	got := ToHTML(Parse("**<script>** [x](https://a.io/?q=\"1)", Markdown))
	if strings.Contains(got, "<script>") || !strings.Contains(got, "<b>&lt;script&gt;</b>") {
		t.Fatalf("got %q", got)
	}
}

func TestHostileInputIsBounded(t *testing.T) {
	for _, s := range []string{
		strings.Repeat("*a ", 5000), strings.Repeat("[", 9000), strings.Repeat("**", 4000), strings.Repeat("> ", 4000), strings.Repeat("`", 9000),
	} {
		Parse(s, Markdown)
		Parse(s, Slack)
	}
	FromHTML(strings.Repeat("<div>", 20000) + "x")
}
