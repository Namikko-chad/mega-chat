// Package mailutil holds what the mail providers share: composing an RFC
// 5322 message, reading addresses and headers, and turning HTML into text
// (the core shows mail as text and never loads remote content).
package mailutil

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/emersion/go-message/charset"
	xhtml "golang.org/x/net/html"

	"megachat/core/model"
)

var dec = &mime.WordDecoder{CharsetReader: charset.Reader}

var spaces = regexp.MustCompile(`[ \t\r\n\f\x{a0}]+`)

// DecodeHeader decodes RFC 2047 encoded words; undecodable input is kept.
func DecodeHeader(s string) string {
	out, err := dec.DecodeHeader(s)
	if err != nil {
		return s
	}
	return out
}

// Addresses parses an address list header; parts that do not parse are
// kept as bare addresses.
func Addresses(s string) []model.Address {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	p := mail.AddressParser{WordDecoder: dec}
	list, err := p.ParseList(s)
	if err != nil {
		var out []model.Address
		for _, part := range strings.Split(s, ",") {
			if a := strings.TrimSpace(part); a != "" {
				out = append(out, model.Address{Email: strings.Trim(a, "<>")})
			}
		}
		return out
	}
	out := make([]model.Address, len(list))
	for i, a := range list {
		out[i] = model.Address{Name: a.Name, Email: a.Address}
	}
	return out
}

// Address parses one address.
func Address(s string) model.Address {
	if as := Addresses(s); len(as) > 0 {
		return as[0]
	}
	return model.Address{}
}

func formatAddress(a model.Address) string {
	return (&mail.Address{Name: a.Name, Address: a.Email}).String()
}

func formatList(as []model.Address) string {
	parts := make([]string, len(as))
	for i, a := range as {
		parts[i] = formatAddress(a)
	}
	return strings.Join(parts, ", ")
}

// Reply is what a reply takes from the message it answers.
type Reply struct {
	MessageID  string
	References string
	Subject    string
}

// ReplySubject puts "Re: " before a subject that has none.
func ReplySubject(s string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(s)), "re:") {
		return s
	}
	return "Re: " + s
}

// Compose builds a text/plain UTF-8 message. Bcc is not written into it:
// the sender passes those recipients to the server separately.
func Compose(from model.Address, out model.OutMail, reply *Reply, now time.Time) []byte {
	var b bytes.Buffer
	h := func(k, v string) {
		if v != "" {
			fmt.Fprintf(&b, "%s: %s\r\n", k, v)
		}
	}
	h("From", formatAddress(from))
	h("To", formatList(out.To))
	h("Cc", formatList(out.Cc))
	h("Subject", mime.QEncoding.Encode("utf-8", out.Subject))
	h("Date", now.Format(time.RFC1123Z))
	h("Message-ID", newMessageID(from.Email))
	if reply != nil && reply.MessageID != "" {
		h("In-Reply-To", reply.MessageID)
		h("References", strings.TrimSpace(reply.References+" "+reply.MessageID))
	}
	h("MIME-Version", "1.0")
	h("Content-Type", "text/plain; charset=utf-8")
	h("Content-Transfer-Encoding", "quoted-printable")
	b.WriteString("\r\n")
	w := quotedprintable.NewWriter(&b)
	text := strings.ReplaceAll(strings.ReplaceAll(out.Text, "\r\n", "\n"), "\n", "\r\n")
	w.Write([]byte(text))
	w.Close()
	return b.Bytes()
}

func newMessageID(from string) string {
	var r [12]byte
	rand.Read(r[:])
	host := "megachat.local"
	if i := strings.LastIndexByte(from, '@'); i >= 0 && i < len(from)-1 {
		host = from[i+1:]
	}
	return "<" + hex.EncodeToString(r[:]) + "@" + host + ">"
}

// Recipients lists every address a message goes to.
func Recipients(out model.OutMail) []string {
	var rs []string
	for _, l := range [][]model.Address{out.To, out.Cc, out.Bcc} {
		for _, a := range l {
			rs = append(rs, a.Email)
		}
	}
	return rs
}

// Snippet is the start of a text on one line.
func Snippet(text string, n int) string {
	text = strings.Join(strings.Fields(text), " ")
	if utf8.RuneCountInString(text) <= n {
		return text
	}
	r := []rune(text)
	return strings.TrimRightFunc(string(r[:n]), unicode.IsSpace) + "…"
}

// HTMLToText renders an HTML part as readable text: blocks become lines,
// links keep their address, scripts and styles go.
func HTMLToText(src string) string {
	doc, err := xhtml.Parse(strings.NewReader(src))
	if err != nil {
		return html.UnescapeString(src)
	}
	var b strings.Builder
	nl := func(n int) {
		s := b.String()
		have := len(s) - len(strings.TrimRight(s, "\n"))
		if len(s) == 0 {
			return
		}
		for ; have < n; have++ {
			b.WriteByte('\n')
		}
	}
	var walk func(n *xhtml.Node, pre bool)
	walk = func(n *xhtml.Node, pre bool) {
		switch n.Type {
		case xhtml.TextNode:
			t := n.Data
			if !pre {
				t = spaces.ReplaceAllString(t, " ")
			}
			b.WriteString(t)
			return
		case xhtml.ElementNode:
			switch n.Data {
			case "script", "style", "head", "title":
				return
			case "br":
				b.WriteByte('\n')
				return
			case "p", "div", "table", "ul", "ol", "h1", "h2", "h3", "h4", "h5", "h6", "blockquote", "hr":
				nl(2)
			case "tr", "li":
				nl(1)
				if n.Data == "li" {
					b.WriteString("• ")
				}
			case "td", "th":
				if s := b.String(); s != "" && !strings.HasSuffix(s, "\n") {
					b.WriteString("  ")
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, pre || (n.Type == xhtml.ElementNode && n.Data == "pre"))
		}
		if n.Type == xhtml.ElementNode {
			switch n.Data {
			case "a":
				for _, a := range n.Attr {
					if a.Key == "href" && (strings.HasPrefix(a.Val, "http") || strings.HasPrefix(a.Val, "mailto:")) {
						label := strings.TrimSpace(textOf(n))
						if label != a.Val && strings.TrimPrefix(a.Val, "mailto:") != label {
							b.WriteString(" <" + a.Val + ">")
						}
					}
				}
			case "p", "div", "table", "ul", "ol", "h1", "h2", "h3", "h4", "h5", "h6", "blockquote", "pre":
				nl(2)
			case "tr", "li":
				nl(1)
			}
		}
	}
	walk(doc, false)
	lines := strings.Split(b.String(), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSpace(spaces.ReplaceAllString(l, " "))
	}
	out := strings.Join(lines, "\n")
	for strings.Contains(out, "\n\n\n") {
		out = strings.ReplaceAll(out, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(out)
}

func textOf(n *xhtml.Node) string {
	if n.Type == xhtml.TextNode {
		return n.Data
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(textOf(c))
	}
	return b.String()
}

// AttachmentKind guesses how an attachment is shown from its MIME type.
func AttachmentKind(mimeType string) model.AttachmentKind {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return model.AttachImage
	case strings.HasPrefix(mimeType, "video/"):
		return model.AttachVideo
	case strings.HasPrefix(mimeType, "audio/"):
		return model.AttachAudio
	}
	return model.AttachFile
}
