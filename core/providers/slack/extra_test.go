package slack

import (
	"net/url"
	"testing"

	"megachat/core/provider"
)

func TestEmoji(t *testing.T) {
	for name, e := range map[string]string{"+1": "👍", "eyes": "👀", "thumbsup::skin-tone-3": "👍", "party-parrot": ":party-parrot:"} {
		if got := emojiOf(name); got != e {
			t.Fatalf("%s: %s", name, got)
		}
	}
	if emojiName("👍") != "+1" || emojiName("👀") != "eyes" || emojiName(":party-parrot:") != "party-parrot" {
		t.Fatal("emoji names")
	}
}

func TestMediaHosts(t *testing.T) {
	p := New()
	sess, _ := p.Open(provider.Account{Params: map[string]string{"method": "session", "workspace_url": "https://acme.slack.com/", "token": "xoxc-1"}}, provider.Env{State: provider.NewMemKV()})
	s := sess.(*session)
	for raw, want := range map[string][2]bool{
		"https://files.slack.com/files-pri/T1-F1/a.png": {true, true},
		"https://acme.slack.com/files/x":                {true, true},
		"https://avatars.slack-edge.com/a.png":          {true, false},
		"https://secure.gravatar.com/avatar/1":          {true, false},
		"https://evil.example.com/a.png":                {false, false},
		"http://files.slack.com/a.png":                  {false, false},
		"https://files.slack.com.evil.com/a.png":        {false, false},
	} {
		u, _ := url.Parse(raw)
		ok, creds := s.mediaHost(u)
		if [2]bool{ok, creds} != want {
			t.Fatalf("%s: %v %v", raw, ok, creds)
		}
	}
}
