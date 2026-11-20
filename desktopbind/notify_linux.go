package main

import (
	"slices"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
)

// Linux: org.freedesktop.Notifications on the session bus (GNOME, KDE,
// dunst, mako, ...). conn and markup belong to the notifier goroutine;
// ids is shared with the signal reader.
var fdo struct {
	conn   *dbus.Conn
	markup bool // the server parses markup in the body

	mu  sync.Mutex
	ids map[string]uint32 // tag → id of the shown notification
}

const (
	fdoName = "org.freedesktop.Notifications"
	fdoPath = "/org/freedesktop/Notifications"
)

func fdoConn() *dbus.Conn {
	if fdo.conn != nil && fdo.conn.Connected() {
		return fdo.conn
	}
	// A private connection: the shared one is the keychain's.
	conn, err := dbus.SessionBusPrivate()
	if err != nil {
		return nil
	}
	if err := conn.Auth(nil); err != nil {
		conn.Close()
		return nil
	}
	if err := conn.Hello(); err != nil {
		conn.Close()
		return nil
	}
	fdo.conn = conn
	fdo.mu.Lock()
	fdo.ids = map[string]uint32{}
	fdo.mu.Unlock()
	var caps []string
	if conn.Object(fdoName, fdoPath).Call(fdoName+".GetCapabilities", 0).Store(&caps) == nil {
		fdo.markup = slices.Contains(caps, "body-markup")
	}
	_ = conn.AddMatchSignal(dbus.WithMatchInterface(fdoName), dbus.WithMatchMember("ActionInvoked"))
	_ = conn.AddMatchSignal(dbus.WithMatchInterface(fdoName), dbus.WithMatchMember("NotificationClosed"))
	sig := make(chan *dbus.Signal, 16)
	conn.Signal(sig)
	go func() {
		for s := range sig {
			if len(s.Body) < 1 {
				continue
			}
			id, _ := s.Body[0].(uint32)
			tag := fdoTag(id, s.Name == fdoName+".NotificationClosed")
			if s.Name == fdoName+".ActionInvoked" && tag != "" {
				notes.clicked(tag)
			}
		}
	}()
	return conn
}

// fdoTag finds the tag of a notification id; forget: it was closed.
func fdoTag(id uint32, forget bool) string {
	fdo.mu.Lock()
	defer fdo.mu.Unlock()
	for t, x := range fdo.ids {
		if x == id {
			if forget {
				delete(fdo.ids, t)
			}
			return t
		}
	}
	return ""
}

func showNotice(n notice) {
	conn := fdoConn()
	if conn == nil {
		return
	}
	body := n.Body
	if fdo.markup {
		body = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(body)
	}
	hints := map[string]dbus.Variant{"category": dbus.MakeVariant("im.received")}
	if n.Entry != "" {
		hints["desktop-entry"] = dbus.MakeVariant(n.Entry)
	}
	fdo.mu.Lock()
	replaces := fdo.ids[n.Tag]
	fdo.mu.Unlock()
	var id uint32
	err := conn.Object(fdoName, fdoPath).Call(fdoName+".Notify", 0,
		"MegaChat", replaces, n.Icon, n.Title, body,
		[]string{"default", "Open"}, hints, int32(-1),
	).Store(&id)
	if err == nil && n.Tag != "" {
		fdo.mu.Lock()
		fdo.ids[n.Tag] = id
		fdo.mu.Unlock()
	}
}

func closeNotice(tag string) {
	fdo.mu.Lock()
	id, ok := fdo.ids[tag]
	delete(fdo.ids, tag)
	fdo.mu.Unlock()
	if ok && fdo.conn != nil {
		fdo.conn.Object(fdoName, fdoPath).Call(fdoName+".CloseNotification", 0, id)
	}
}
