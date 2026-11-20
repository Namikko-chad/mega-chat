package main

import "sync"

// notice is a desktop notification about a message. Tag names its chat:
// a newer notice with the same tag replaces the shown one, a click on it
// reports the tag (onClick), closeNotice takes it away.
type notice struct {
	Tag   string `json:"tag"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon,omitempty"`  // absolute path of a PNG (Linux)
	Entry string `json:"entry,omitempty"` // the app's desktop entry id (Linux)
}

// notifier shows notices one at a time off the caller's thread: the app
// calls in from its UI thread and must not wait for the desktop.
type notifier struct {
	once sync.Once
	ch   chan func()

	mu      sync.Mutex
	onClick func(tag string)
}

var notes notifier

func (n *notifier) do(f func()) {
	n.once.Do(func() {
		n.ch = make(chan func(), 64)
		go func() {
			for f := range n.ch {
				f()
			}
		}()
	})
	select {
	case n.ch <- f:
	default: // the desktop is stuck: drop it rather than pile up
	}
}

func (n *notifier) show(no notice) { n.do(func() { showNotice(no) }) }

func (n *notifier) close(tag string) { n.do(func() { closeNotice(tag) }) }

func (n *notifier) setClick(f func(tag string)) {
	n.mu.Lock()
	n.onClick = f
	n.mu.Unlock()
}

// clicked is called by the platform code when a notice was clicked.
func (n *notifier) clicked(tag string) {
	n.mu.Lock()
	f := n.onClick
	n.mu.Unlock()
	if f != nil {
		f(tag)
	}
}
