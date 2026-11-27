package hub

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

func TestChatFoldersAndActions(t *testing.T) {
	e := newEnv(t)
	e.srv.Special = true
	ctx := context.Background()
	a := e.login(t, "alice")
	waitFor(t, "chat list", func() bool { cs, _ := e.h.Chats(nil, 0); return len(cs) == 11 })
	info, _ := e.h.Account(a.ID)
	if !slices.Contains(info.Caps, provider.CapChatFolders) || !slices.Contains(info.Caps, provider.CapManageChats) || len(info.ChatActions) != 7 {
		t.Fatalf("caps %v actions %v", info.Caps, info.ChatActions)
	}
	chat := func(id string) model.Chat {
		c, ok, err := e.st.Chat(a.ID, id)
		if !ok || err != nil {
			t.Fatalf("chat %s: %v %v", id, ok, err)
		}
		return c
	}
	if !chat("saved").Saved || !chat("c10").Archived || !chat("c3").Pinned || chat("c1").Archived {
		t.Fatalf("special chats: %+v %+v", chat("saved"), chat("c10"))
	}
	if f := chat("c1").Folders; !slices.Equal(f, []string{"people"}) {
		t.Fatalf("folders of c1: %v", f)
	}

	afs, failed := e.h.ChatFolders(ctx, nil)
	if len(failed) != 0 || len(afs) != 1 || afs[0].Account != a.ID || len(afs[0].Folders) != 2 {
		t.Fatalf("folders: %+v %v", afs, failed)
	}
	teams := afs[0].Folders[1]
	if !teams.Contains(chat("c3")) || teams.Contains(chat("c1")) {
		t.Fatal("teams folder membership")
	}

	// Archive, mute, pin: stored and told.
	c, err := e.h.ManageChat(ctx, a.ID, "c3", model.ChatArchive)
	if err != nil || !c.Archived || !chat("c3").Archived {
		t.Fatalf("archive: %+v %v", c, err)
	}
	if teams.Contains(chat("c3")) {
		t.Fatal("an archived chat stays in a folder that excludes the archive")
	}
	if c, err = e.h.ManageChat(ctx, a.ID, "c1", model.ChatMute); err != nil || !c.Muted {
		t.Fatalf("mute: %+v %v", c, err)
	}
	e.mu.Lock()
	told := slices.ContainsFunc(e.events, func(ev Event) bool { return ev.Kind == EvChat && ev.Chat.ID == "c1" && ev.Chat.Muted })
	e.mu.Unlock()
	if !told {
		t.Fatal("no chat event after muting")
	}
	// A refresh keeps what the provider now says.
	if err := e.h.RefreshChats(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if !chat("c1").Muted || !chat("c3").Archived {
		t.Fatal("flags lost after a refresh")
	}

	// A change made elsewhere comes as EvChatFlags.
	off := false
	e.h.apply(a.ID, model.Event{Kind: model.EvChatFlags, Chat: &model.Chat{ID: "c3"}, Flags: &model.ChatFlags{Archived: &off}})
	if chat("c3").Archived {
		t.Fatal("EvChatFlags not applied")
	}

	// Leaving: groups only; the chat goes away.
	if _, err := e.h.ManageChat(ctx, a.ID, "c1", model.ChatLeave); !errors.Is(err, ErrCannotLeave) {
		t.Fatalf("leave a direct chat: %v", err)
	}
	if _, err := e.h.ManageChat(ctx, a.ID, "c3", model.ChatLeave); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := e.st.Chat(a.ID, "c3"); ok {
		t.Fatal("left chat still cached")
	}
	if _, err := e.h.ManageChat(ctx, a.ID, "c1", model.ChatActionUnknown); !errors.Is(err, provider.ErrUnsupported) {
		t.Fatalf("unknown action: %v", err)
	}
}
