package model

import "testing"

func TestQuotedText(t *testing.T) {
	for _, c := range []struct {
		o    Outgoing
		want string
	}{
		{Outgoing{Text: "hi"}, "hi"},
		{Outgoing{Text: "hi", Quote: "  \n "}, "hi"},
		{Outgoing{Text: "yes", Quote: "one\ntwo\n"}, "> one\n> two\nyes"},
	} {
		if got := c.o.QuotedText(); got != c.want {
			t.Errorf("%+v: %q, want %q", c.o, got, c.want)
		}
	}
}
