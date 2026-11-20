package mailutil

import (
	"net/mail"
	"strings"
	"testing"
	"time"

	"megachat/core/model"
)

func TestHTMLToText(t *testing.T) {
	got := HTMLToText(`<html><head><style>p{}</style></head><body><p>Hello <b>world</b>,</p>
<p>see <a href="https://example.com/x">the page</a><br>bye</p><script>evil()</script><ul><li>one</li><li>two</li></ul></body></html>`)
	want := "Hello world,\n\nsee the page <https://example.com/x>\nbye\n\n• one\n• two"
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestComposeParses(t *testing.T) {
	raw := Compose(model.Address{Name: "Иван", Email: "ivan@mail.ru"}, model.OutMail{
		To: []model.Address{{Name: "Anna", Email: "anna@example.com"}}, Bcc: []model.Address{{Email: "hidden@example.com"}},
		Subject: "Привет", Text: "line 1\nстрока 2",
	}, &Reply{MessageID: "<a@b>", References: "<z@b>"}, time.Unix(1700000000, 0))
	m, err := mail.ReadMessage(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatal(err)
	}
	if s := DecodeHeader(m.Header.Get("Subject")); s != "Привет" {
		t.Fatalf("subject %q", s)
	}
	if strings.Contains(string(raw), "hidden@") {
		t.Fatal("bcc leaked into the message")
	}
	if m.Header.Get("In-Reply-To") != "<a@b>" || m.Header.Get("References") != "<z@b> <a@b>" {
		t.Fatalf("threading headers %v", m.Header)
	}
	if a := Address(m.Header.Get("From")); a.Name != "Иван" || a.Email != "ivan@mail.ru" {
		t.Fatalf("from %+v", a)
	}
	if got := Recipients(model.OutMail{To: []model.Address{{Email: "a"}}, Bcc: []model.Address{{Email: "b"}}}); len(got) != 2 {
		t.Fatal(got)
	}
}

func TestAddressesFallback(t *testing.T) {
	as := Addresses(`"Doe, John" <john@x.com>, jane@y.org`)
	if len(as) != 2 || as[0].Name != "Doe, John" || as[1].Email != "jane@y.org" {
		t.Fatalf("%+v", as)
	}
	if s := Snippet("a  b\n\nc", 3); s != "a b…" {
		t.Fatalf("%q", s)
	}
}
