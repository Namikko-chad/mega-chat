// Command desktopbind is the core of the desktop app. It is built as a C
// archive (Linux; linked into the app's executable) or a DLL (Windows) and
// called from Dart through FFI (app/lib/data/rpc/native_core.dart). The
// app then talks to it over the gRPC API, like the Android app does with
// mobilebind.
//
// It writes endpoint.json to the config dir, so the mchat CLI reaches the
// core of the running app. MEGACHAT_DEMO=1 adds the Demo messenger.
package main

/*
#include <stdlib.h>

typedef void (*megachat_click_fn)(char*);

static void megachat_click(megachat_click_fn f, char* tag) { f(tag); }
*/
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"time"
	"unsafe"

	"megachat/core/embed"
)

var (
	mu   sync.Mutex
	core *embed.Core
)

type reply struct {
	Address string `json:"address,omitempty"`
	Token   string `json:"token,omitempty"`
	// The core runs in another instance of the app; address and token are its.
	External bool   `json:"external,omitempty"`
	Error    string `json:"error,omitempty"`
}

// MegachatStart starts the core (again: returns the running one) and
// returns JSON {"address","token"} or {"error"}; free it with MegachatFree.
//
//export MegachatStart
func MegachatStart() *C.char {
	return cjson(start())
}

func start() reply {
	mu.Lock()
	defer mu.Unlock()
	if core != nil {
		return reply{Address: core.Address, Token: core.Token}
	}
	dir := embed.DefaultDir()
	endpoint := filepath.Join(dir, "endpoint.json")
	c, err := embed.Start(embed.Options{
		Dir:      dir,
		Listen:   embed.DefaultListen(dir),
		Endpoint: endpoint,
		LogFile:  filepath.Join(dir, "megachat.log"),
		Log:      os.Stderr,
		Demo:     os.Getenv("MEGACHAT_DEMO") == "1",
	})
	if errors.Is(err, embed.ErrRunning) {
		if addr, token, ok := embed.ReadEndpoint(endpoint); ok {
			return reply{Address: addr, Token: token, External: true}
		}
	}
	if err != nil {
		return reply{Error: err.Error()}
	}
	core = c
	return reply{Address: c.Address, Token: c.Token}
}

// stopWait is how long MegachatStop waits for the core before it gives up.
const stopWait = 5 * time.Second

// MegachatStop locks the vault, closes the sessions and stops the server.
// It is called on the app's UI thread right before the process exits, so
// it never waits longer than stopWait: a core that does not stop in time
// leaves its goroutine stacks in stop-hang.log (config dir) and is left to
// die with the process.
//
//export MegachatStop
func MegachatStop() {
	mu.Lock()
	defer mu.Unlock()
	if core == nil {
		return
	}
	c := core
	core = nil
	done := make(chan struct{})
	go func() { c.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(stopWait):
		dumpStacks(filepath.Join(embed.DefaultDir(), "stop-hang.log"))
	}
}

// dumpStacks writes every goroutine's stack (no argument values) to path.
func dumpStacks(path string) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "core did not stop in %s, %s\n\n", stopWait, time.Now().Format(time.RFC3339))
	pprof.Lookup("goroutine").WriteTo(f, 1)
}

// MegachatFree frees a string returned by MegachatStart.
//
//export MegachatFree
func MegachatFree(p *C.char) {
	C.free(unsafe.Pointer(p))
}

func cjson(r reply) *C.char {
	b, _ := json.Marshal(r)
	return C.CString(string(b))
}

// MegachatNotify shows a desktop notification, JSON
// {"tag","title","body","icon","entry"} (see notice); it does not wait.
//
//export MegachatNotify
func MegachatNotify(js *C.char) {
	var n notice
	if json.Unmarshal([]byte(C.GoString(js)), &n) == nil {
		notes.show(n)
	}
}

// MegachatNotifyClose takes away the notification with this tag.
//
//export MegachatNotifyClose
func MegachatNotifyClose(tag *C.char) {
	notes.close(C.GoString(tag))
}

// MegachatNotifyClicks registers f (NULL: none) to be called with the tag
// of a clicked notification, from any thread; f frees the string with
// MegachatFree.
//
//export MegachatNotifyClicks
func MegachatNotifyClicks(f C.megachat_click_fn) {
	if f == nil {
		notes.setClick(nil)
		return
	}
	notes.setClick(func(tag string) { C.megachat_click(f, C.CString(tag)) })
}

func main() {}
