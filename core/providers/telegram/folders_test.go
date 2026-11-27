package telegram

import (
	"slices"
	"testing"

	"github.com/gotd/td/tg"

	"github.com/SelfHold/parley/core/model"
)

func TestFolderRules(t *testing.T) {
	s := &session{
		selfID:  1,
		kinds:   map[string]model.ChatKind{"g5": model.ChatGroup, "c6": model.ChatChannel, "c7": model.ChatGroup},
		contact: map[string]bool{"u2": true},
		bots:    map[string]bool{"u4": true},
		filters: []tg.DialogFilterClass{
			&tg.DialogFilter{ID: 2, Title: tg.TextWithEntities{Text: "People"}, Contacts: true, NonContacts: true,
				ExcludePeers: []tg.InputPeerClass{&tg.InputPeerUser{UserID: 3}}},
			&tg.DialogFilter{ID: 3, Title: tg.TextWithEntities{Text: "Work"}, Groups: true, Emoticon: "💼",
				IncludePeers: []tg.InputPeerClass{&tg.InputPeerUser{UserID: 2}}, PinnedPeers: []tg.InputPeerClass{&tg.InputPeerChannel{ChannelID: 6}}},
			&tg.DialogFilter{ID: 4, Title: tg.TextWithEntities{Text: "News"}, Broadcasts: true, Bots: true, ExcludeRead: true},
			&tg.DialogFilterDefault{},
		},
		filtersOK: true,
	}
	for key, want := range map[string][]string{
		"u2": {"2"}, "u3": {}, "u4": {"4"}, "g5": {"3"}, "c6": {"4"}, "c7": {"3"}, "u1": {"2"},
	} {
		if got := s.folderIDs(key); !slices.Equal(got, want) {
			t.Errorf("folders of %s: %v, want %v", key, got, want)
		}
	}
	s.mu.Lock()
	work := model.ChatFolder{ID: "3"}
	work.Always = s.peerKeys(s.filters[1].(*tg.DialogFilter).PinnedPeers, s.filters[1].(*tg.DialogFilter).IncludePeers)
	s.mu.Unlock()
	if !slices.Equal(work.Always, []string{"c6", "u2"}) {
		t.Fatalf("always: %v", work.Always)
	}
	if !work.Contains(model.Chat{ID: "u2"}) || !work.Contains(model.Chat{ID: "g5", Folders: []string{"3"}}) {
		t.Fatal("membership")
	}
	news := model.ChatFolder{ID: "4", ExcludeRead: true}
	if news.Contains(model.Chat{ID: "c6", Folders: []string{"4"}}) || !news.Contains(model.Chat{ID: "c6", Folders: []string{"4"}, Unread: 2}) {
		t.Fatal("exclude read")
	}
}
