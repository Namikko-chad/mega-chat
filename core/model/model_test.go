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

func TestMentionedText(t *testing.T) {
	text := "hi @Ann and @Bob <3"
	ms := []Mention{{Offset: 3, Length: 4, User: User{ID: "A"}}, {Offset: 12, Length: 4, User: User{ID: "B"}}, {Offset: 40, Length: 2}}
	got := MentionedText(text, ms, func(m Mention, span string) string { return "<" + m.User.ID + ":" + span + ">" }, func(s string) string { return "[" + s + "]" })
	if want := "[hi ]<A:@Ann>[ and ]<B:@Bob>[ <3]"; got != want {
		t.Fatalf("%q", got)
	}
}

func TestMatchUser(t *testing.T) {
	u := User{Name: "Анна Петрова", Username: "anna_p"}
	for q, want := range map[string]bool{"": true, "ann": true, "АН": true, "пет": true, "x": false, "анна п": true} {
		if MatchUser(u, q) != want {
			t.Errorf("%q: %v", q, !want)
		}
	}
}
