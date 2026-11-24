package rpc

import (
	"testing"

	"megachat/core/rpc/pb"
)

func TestMentionSpans(t *testing.T) {
	// "😀" is two UTF-16 units and four bytes; "Аня" six bytes.
	text := "😀 @Аня, @Bob"
	u := func(id string) *pb.User { return &pb.User{Id: id, Name: id} }
	got := mentions(text, []*pb.Mention{
		{Offset: 9, Length: 4, User: u("B")},
		{Offset: 3, Length: 4, User: u("A")},
		{Offset: 4, Length: 2, User: u("X")}, // overlaps A
		{Offset: 1, Length: 2, User: u("Y")}, // starts inside the surrogate pair
		{Offset: 9, Length: 9, User: u("Z")}, // off the end
		{Offset: 0, Length: 1, User: nil},    // nobody
	})
	if len(got) != 2 {
		t.Fatalf("%+v", got)
	}
	if s := text[got[0].Offset : got[0].Offset+got[0].Length]; s != "@Аня" || got[0].User.ID != "A" {
		t.Fatalf("first: %q %+v", s, got[0])
	}
	if s := text[got[1].Offset : got[1].Offset+got[1].Length]; s != "@Bob" || got[1].User.ID != "B" {
		t.Fatalf("second: %q", s)
	}
}
