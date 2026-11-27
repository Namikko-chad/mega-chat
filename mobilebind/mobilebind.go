// Package mobilebind runs the core inside the Android app. It is built
// with "gomobile bind" into an AAR. The Kotlin side starts it from a
// foreground service (sessions stay connected in the background); the
// Flutter UI talks to it over the same gRPC API as on desktops, through a
// unix socket in the app's private files dir (core/embed).
package mobilebind

import (
	"encoding/json"
	"os"
	"sync"

	"megachat/core/cloudsync"
	"megachat/core/embed"
	"megachat/core/rpc/pb"
)

// Listener gets core events (implemented in Kotlin). kind is Event.Kind
// of megachat.proto. For a message worth a notification (kind 4, or 9 for
// a reply in a followed thread or a mention there; Event.notify) json is {"account_id","chat_id",
// "chat_title","sender","text","mention"}; such messages that are not
// worth one are not passed. For a call (kind 10) see callNotice; other
// kinds carry no json.
type Listener interface {
	OnEvent(kind int, json string)
}

var (
	mu        sync.Mutex
	core      *embed.Core
	unsub     func()
	cloudSync CloudSyncBinder
)

// CloudSyncBinder is CloudSync's bound service (ICloudSyncConnector.aidl,
// CONNECTOR.md § 3.2), bound by the Kotlin side. Call returns the
// serialized CallReply of the method; an error means CloudSync is not
// installed or the binding broke. Calls come from Go's goroutines, never
// the main thread, and may block (AwaitPairing up to 60 s).
type CloudSyncBinder interface {
	Call(method string, request []byte) ([]byte, error)
	// Watch starts a watch that sends serialized Change messages to w,
	// then one OnEnd; it returns the handle for Unwatch.
	Watch(w CloudSyncWatcher) (int64, error)
	Unwatch(handle int64)
}

// CloudSyncWatcher gets one watch's changes (implemented in Go).
type CloudSyncWatcher interface {
	OnChange(change []byte)
	OnEnd(reply []byte)
}

// SetCloudSync gives the core CloudSync's binder; call it before Start.
func SetCloudSync(b CloudSyncBinder) {
	mu.Lock()
	defer mu.Unlock()
	cloudSync = b
}

type binder struct{ b CloudSyncBinder }

func (b binder) Call(method string, req []byte) ([]byte, error) { return b.b.Call(method, req) }
func (b binder) Watch(w cloudsync.BinderWatcher) (int64, error) { return b.b.Watch(w) }
func (b binder) Unwatch(h int64)                                { b.b.Unwatch(h) }

// Start runs the core: the vault and the cache live in dataDir, temporary
// files go to cacheDir, gRPC listens on the unix socket sock. It returns
// the bearer token clients must send; calling it again returns the same.
func Start(dataDir, cacheDir, sock string) (string, error) {
	mu.Lock()
	defer mu.Unlock()
	if core != nil {
		return core.Token, nil
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return "", err
	}
	// Android has no HOME and its default temp dir is not writable by apps.
	os.Setenv("HOME", dataDir)
	if cacheDir != "" {
		os.MkdirAll(cacheDir, 0o700)
		os.Setenv("TMPDIR", cacheDir)
	}
	// Only this app can reach its files dir: a socket left by a killed
	// process is stale.
	os.Remove(sock)
	o := embed.Options{Dir: dataDir, Platform: "android", Listen: "unix:" + sock}
	if cloudSync != nil {
		o.CloudSync = cloudsync.Android{Binder: binder{cloudSync}}
	}
	c, err := embed.Start(o)
	if err != nil {
		return "", err
	}
	core = c
	return c.Token, nil
}

type notice struct {
	AccountID string `json:"account_id"`
	ChatID    string `json:"chat_id"`
	ChatTitle string `json:"chat_title"`
	Sender    string `json:"sender"`
	Text      string `json:"text"`
	Mention   bool   `json:"mention,omitempty"`
}

// noticeText is the text of a message, or what it carries without text.
func noticeText(m *pb.Message) string {
	if m.Text != "" || len(m.Attachments) == 0 {
		return m.Text
	}
	if a := m.Attachments[0]; a.Name != "" {
		return "📎 " + a.Name
	}
	return "📎"
}

// callNotice is the json of a KIND_CALL event: the incoming call
// notification and the foreground service type follow it.
type callNotice struct {
	AccountID string `json:"account_id"`
	CallID    string `json:"call_id"`
	Peer      string `json:"peer"`
	Video     bool   `json:"video"`
	Outgoing  bool   `json:"outgoing"`
	State     int    `json:"state"` // CallState of megachat.proto
}

// SetListener registers l for core events (nil to stop).
func SetListener(l Listener) {
	mu.Lock()
	defer mu.Unlock()
	if unsub != nil {
		unsub()
		unsub = nil
	}
	if l == nil || core == nil {
		return
	}
	unsub = core.Service.Subscribe(func(ev *pb.Event) {
		js := ""
		if ev.Kind == pb.Event_KIND_MESSAGE_NEW || ev.Kind == pb.Event_KIND_THREAD_MESSAGE {
			m := ev.Message
			if m == nil || !ev.Notify {
				return
			}
			b, _ := json.Marshal(notice{
				AccountID: ev.AccountId, ChatID: m.ChatId, ChatTitle: ev.Chat.GetTitle(),
				Sender: m.Sender.GetName(), Text: noticeText(m), Mention: m.Mentioned,
			})
			js = string(b)
		}
		if ev.Kind == pb.Event_KIND_MAIL_NEW {
			// Shown like a message: the sender as the title, the subject.
			m := ev.Mail
			if m == nil {
				return
			}
			from := m.From.GetName()
			if from == "" {
				from = m.From.GetEmail()
			}
			b, _ := json.Marshal(notice{AccountID: ev.AccountId, ChatID: "mail:" + m.Id, ChatTitle: from, Text: m.Subject})
			js = string(b)
		}
		if ev.Kind == pb.Event_KIND_CALL {
			c := ev.Call
			if c == nil {
				return
			}
			b, _ := json.Marshal(callNotice{
				AccountID: ev.AccountId, CallID: c.Id, Peer: c.Peer.GetName(),
				Video: c.Video, Outgoing: c.Outgoing, State: int(c.State),
			})
			js = string(b)
		}
		l.OnEvent(int(ev.Kind), js)
	})
}

// Stop locks the vault and stops the server.
func Stop() {
	mu.Lock()
	defer mu.Unlock()
	if core == nil {
		return
	}
	if unsub != nil {
		unsub()
		unsub = nil
	}
	core.Stop()
	core = nil
}
