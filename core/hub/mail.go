package hub

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/SelfHold/parley/core/eventlog"
	"github.com/SelfHold/parley/core/model"
	"github.com/SelfHold/parley/core/provider"
)

// Mail and calendars are not cached: every read goes to the providers,
// several accounts at once, and the hub merges the answers into one inbox
// or one calendar. Only the folder lists (for their unread counts) and
// the calendar lists are kept in memory.

const (
	// EvMailNew: Mail is a new message in an account's inbox.
	EvMailNew EventKind = 100 + iota
	// EvMail: the mail of AccountID changed: reload what is shown.
	EvMail
	// EvCalendar: the events of AccountID changed.
	EvCalendar
)

// folderTTL: how long a folder list is used before it is fetched again
// (changes the session reports drop it earlier).
const folderTTL = 5 * time.Minute

// perAccount bounds one account's part of a merged read.
const perAccount = 60 * time.Second

type mailCache struct {
	mu        sync.Mutex
	folders   map[string]folderEntry // by account
	calendars map[string]calEntry
}

type folderEntry struct {
	list []model.Folder
	at   time.Time
}

type calEntry struct {
	list []model.Calendar
	at   time.Time
}

// AccountError is the part of a merged read that failed.
type AccountError struct {
	Account string
	Err     error
}

func (e AccountError) Error() string { return e.Account + ": " + e.Err.Error() }

// session returns the session of an account that can be used now.
func (h *Hub) session(acct string) (provider.Session, error) {
	h.mu.Lock()
	r := h.runs[acct]
	h.mu.Unlock()
	if r == nil {
		return nil, ErrNoAccount
	}
	if r.sess == nil || r.acct.Disabled {
		return nil, provider.ErrOffline
	}
	return r.sess, nil
}

func (h *Hub) mailer(acct string) (provider.Mailer, error) {
	s, err := h.session(acct)
	if err != nil {
		return nil, err
	}
	m, ok := s.(provider.Mailer)
	if !ok {
		return nil, provider.ErrUnsupported
	}
	return m, nil
}

func (h *Hub) calendarsOf(acct string) (provider.Calendars, error) {
	s, err := h.session(acct)
	if err != nil {
		return nil, err
	}
	c, ok := s.(provider.Calendars)
	if !ok {
		return nil, provider.ErrUnsupported
	}
	return c, nil
}

// withCap returns the given accounts, or every enabled account whose
// session passes has when none are given.
func (h *Hub) withCap(accounts []string, has func(provider.Session) bool) []string {
	if len(accounts) > 0 {
		return accounts
	}
	var out []string
	for _, a := range h.v.Config().Accounts {
		if a.Disabled {
			continue
		}
		if s, err := h.session(a.ID); err == nil && has(s) {
			out = append(out, a.ID)
		}
	}
	return out
}

func isMailer(s provider.Session) bool { _, ok := s.(provider.Mailer); return ok }

func isCalendars(s provider.Session) bool { _, ok := s.(provider.Calendars); return ok }

// fanOut runs fn for every account at once and collects what failed.
func fanOut(ctx context.Context, accounts []string, fn func(ctx context.Context, acct string) error) []AccountError {
	var (
		mu   sync.Mutex
		errs []AccountError
		wg   sync.WaitGroup
	)
	for _, a := range accounts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			actx, cancel := context.WithTimeout(ctx, perAccount)
			defer cancel()
			if err := fn(actx, a); err != nil {
				mu.Lock()
				errs = append(errs, AccountError{Account: a, Err: err})
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	slices.SortFunc(errs, func(a, b AccountError) int { return cmp.Compare(a.Account, b.Account) })
	return errs
}

// --- folders -------------------------------------------------------------

// folders returns the folder list of an account, from memory when fresh.
func (h *Hub) folders(ctx context.Context, acct string) ([]model.Folder, error) {
	h.mail.mu.Lock()
	e, ok := h.mail.folders[acct]
	h.mail.mu.Unlock()
	if ok && time.Since(e.at) < folderTTL {
		return e.list, nil
	}
	m, err := h.mailer(acct)
	if err != nil {
		return nil, err
	}
	list, err := m.Folders(ctx)
	if err != nil {
		return nil, err
	}
	h.mail.mu.Lock()
	if h.mail.folders == nil {
		h.mail.folders = map[string]folderEntry{}
	}
	old := h.mail.folders[acct]
	h.mail.folders[acct] = folderEntry{list: list, at: time.Now()}
	h.mail.mu.Unlock()
	if inboxUnread(old.list) != inboxUnread(list) || old.list == nil {
		h.emitAccount(acct)
	}
	return list, nil
}

func inboxUnread(fs []model.Folder) int {
	n := 0
	for _, f := range fs {
		if f.Role == model.FolderInbox {
			n += f.Unread
		}
	}
	return n
}

// mailUnread is the unread count of an account's inbox as last fetched.
func (h *Hub) mailUnread(acct string) int {
	h.mail.mu.Lock()
	defer h.mail.mu.Unlock()
	return inboxUnread(h.mail.folders[acct].list)
}

// dropFolders forgets an account's folder list: the next read fetches it.
func (h *Hub) dropFolders(acct string) {
	h.mail.mu.Lock()
	if e, ok := h.mail.folders[acct]; ok {
		e.at = time.Time{}
		h.mail.folders[acct] = e
	}
	h.mail.mu.Unlock()
}

// forgetMail drops what is kept of a removed or stopped account.
func (h *Hub) forgetMail(acct string) {
	h.mail.mu.Lock()
	delete(h.mail.folders, acct)
	delete(h.mail.calendars, acct)
	h.mail.mu.Unlock()
}

// AccountFolders are the folders of one account.
type AccountFolders struct {
	Account string
	Folders []model.Folder
}

// Folders returns the folders of the given accounts (all mail accounts
// when empty).
func (h *Hub) Folders(ctx context.Context, accounts []string) ([]AccountFolders, []AccountError) {
	accounts = h.withCap(accounts, isMailer)
	var mu sync.Mutex
	got := map[string][]model.Folder{}
	errs := fanOut(ctx, accounts, func(ctx context.Context, a string) error {
		fs, err := h.folders(ctx, a)
		if err == nil {
			mu.Lock()
			got[a] = fs
			mu.Unlock()
		}
		return err
	})
	var out []AccountFolders
	for _, a := range accounts {
		if fs, ok := got[a]; ok {
			out = append(out, AccountFolders{Account: a, Folders: fs})
		}
	}
	return out, errs
}

// refreshMail loads the folders of a connected mail account, so its
// unread count is known before the user opens the mail.
func (h *Hub) refreshMail(acct string) {
	if _, err := h.mailer(acct); err != nil {
		return
	}
	h.dropFolders(acct)
	ctx, cancel := context.WithTimeout(h.ctx, perAccount)
	defer cancel()
	if _, err := h.folders(ctx, acct); err != nil && h.ctx.Err() == nil {
		h.logf(eventlog.Warn, "mail", acct, "folders: %v", err)
	}
}

// --- messages ------------------------------------------------------------

// MailCursor is where a merged mail list stopped: its last message.
type MailCursor struct {
	Time    time.Time
	Account string
	ID      string
}

// MailQuery asks for one page of a merged mail list.
type MailQuery struct {
	// Accounts to read; empty: every mail account.
	Accounts []string
	// Role picks the folder of each account (all inboxes, ...). Folder,
	// with exactly one account, names one folder instead.
	Role   model.FolderRole
	Folder string
	// After is the cursor of the previous page; nil: the latest.
	After *MailCursor
	Limit int
}

// MailPage is one page of a merged mail list, newest first.
type MailPage struct {
	Mails  []model.Mail
	More   bool
	Next   *MailCursor // the cursor for the next page (nil: none)
	Failed []AccountError
}

// mailLess orders a merged list: newest first, then by account and id, so
// that a cursor names a place in it exactly.
func mailLess(a, b model.Mail) int {
	if c := cmp.Compare(b.Time.UnixMilli(), a.Time.UnixMilli()); c != 0 {
		return c
	}
	if c := cmp.Compare(a.AccountID, b.AccountID); c != 0 {
		return c
	}
	return cmp.Compare(a.ID, b.ID)
}

// pastCursor reports whether m comes after the cursor in the merged order.
func pastCursor(c *MailCursor, m model.Mail) bool {
	if c == nil {
		return true
	}
	return mailLess(model.Mail{Time: c.Time, AccountID: c.Account, ID: c.ID}, m) < 0
}

// Mail returns one page of the merged mail of several accounts. Accounts
// that fail are reported in MailPage.Failed; the others still count.
func (h *Hub) Mail(ctx context.Context, q MailQuery) (MailPage, error) {
	if q.Limit <= 0 || q.Limit > 200 {
		q.Limit = 50
	}
	if q.Folder != "" && len(q.Accounts) != 1 {
		return MailPage{}, errors.New("a folder id needs exactly one account")
	}
	accounts := h.withCap(q.Accounts, isMailer)
	var before time.Time
	if q.After != nil {
		before = q.After.Time
	}
	type part struct {
		mails []model.Mail
		more  bool
	}
	var mu sync.Mutex
	parts := map[string]part{}
	failed := fanOut(ctx, accounts, func(ctx context.Context, a string) error {
		m, err := h.mailer(a)
		if err != nil {
			return err
		}
		folder := q.Folder
		if folder == "" {
			fs, err := h.folders(ctx, a)
			if err != nil {
				return err
			}
			i := slices.IndexFunc(fs, func(f model.Folder) bool { return f.Role == q.Role })
			if i < 0 {
				return nil // this account has no such folder
			}
			folder = fs[i].ID
		}
		mails, more, err := m.Mail(ctx, folder, before, q.Limit)
		if err != nil {
			return err
		}
		kept := mails[:0]
		for _, x := range mails {
			x.AccountID = a
			if x.FolderID == "" {
				x.FolderID = folder
			}
			if pastCursor(q.After, x) {
				kept = append(kept, x)
			}
		}
		mu.Lock()
		parts[a] = part{kept, more}
		mu.Unlock()
		return nil
	})
	if len(failed) > 0 && len(failed) == len(accounts) && len(accounts) == 1 {
		return MailPage{}, failed[0].Err
	}
	// An account with more to come is complete only down to its oldest
	// message: anything older from the others waits for the next page.
	var floor time.Time
	var all []model.Mail
	more := false
	for _, p := range parts {
		all = append(all, p.mails...)
		if p.more {
			more = true
			if n := len(p.mails); n > 0 && p.mails[n-1].Time.After(floor) {
				floor = p.mails[n-1].Time
			}
		}
	}
	slices.SortFunc(all, mailLess)
	if !floor.IsZero() {
		cut := len(all)
		for cut > 0 && all[cut-1].Time.UnixMilli() < floor.UnixMilli() {
			cut--
		}
		all = all[:cut]
	}
	if len(all) > q.Limit {
		all, more = all[:q.Limit], true
	}
	page := MailPage{Mails: all, More: more, Failed: failed}
	if more && len(all) > 0 {
		l := all[len(all)-1]
		page.Next = &MailCursor{Time: l.Time, Account: l.AccountID, ID: l.ID}
	} else if more && q.After != nil {
		page.Next = q.After
	}
	return page, nil
}

// MailBody returns one message with its body.
func (h *Hub) MailBody(ctx context.Context, acct, id string) (model.Mail, error) {
	m, err := h.mailer(acct)
	if err != nil {
		return model.Mail{}, err
	}
	out, err := m.MailBody(ctx, id)
	out.AccountID = acct
	for i := range out.Attachments {
		if out.Attachments[i].Kind == model.AttachUnknown {
			out.Attachments[i].Kind = model.AttachFile
		}
	}
	return out, err
}

// SendMail sends a message from an account.
func (h *Hub) SendMail(ctx context.Context, acct string, out model.OutMail) error {
	if len(out.To)+len(out.Cc)+len(out.Bcc) == 0 {
		return errors.New("no recipients")
	}
	m, err := h.mailer(acct)
	if err != nil {
		return err
	}
	if err := m.SendMail(ctx, out); err != nil {
		return err
	}
	h.logf(eventlog.Info, "mail", acct, "sent %q", out.Subject)
	h.mailChanged(acct)
	return nil
}

// UpdateMail marks, stars or trashes messages of an account.
func (h *Hub) UpdateMail(ctx context.Context, acct string, ids []string, action model.MailAction) error {
	if len(ids) == 0 {
		return nil
	}
	if action == model.MailUnknownAction {
		return fmt.Errorf("unknown mail action")
	}
	m, err := h.mailer(acct)
	if err != nil {
		return err
	}
	if err := m.UpdateMail(ctx, ids, action); err != nil {
		return err
	}
	h.mailChanged(acct)
	return nil
}

// MoveMail moves messages of an account from one folder to another.
func (h *Hub) MoveMail(ctx context.Context, acct string, ids []string, from, to string) error {
	if len(ids) == 0 || from == to {
		return nil
	}
	if to == "" {
		return fmt.Errorf("no folder to move to")
	}
	m, err := h.mailer(acct)
	if err != nil {
		return err
	}
	mv, ok := m.(provider.MailMover)
	if !ok {
		return provider.ErrUnsupported
	}
	if err := mv.MoveMail(ctx, ids, from, to); err != nil {
		return err
	}
	h.mailChanged(acct)
	return nil
}

// mailChanged drops the folder list and tells subscribers to reload. The
// new unread count follows in the background.
func (h *Hub) mailChanged(acct string) {
	h.dropFolders(acct)
	h.emit(Event{Kind: EvMail, AccountID: acct})
	go func() {
		ctx, cancel := context.WithTimeout(h.ctx, perAccount)
		defer cancel()
		_, _ = h.folders(ctx, acct)
	}()
}

// applyMail handles the mail and calendar events of a session.
func (h *Hub) applyMail(acct string, ev model.Event) {
	switch ev.Kind {
	case model.EvMailNew:
		if ev.Mail == nil {
			return
		}
		m := *ev.Mail
		m.AccountID = acct
		h.emit(Event{Kind: EvMailNew, AccountID: acct, Mail: &m})
		h.mailChanged(acct)
	case model.EvMailChanged:
		h.mailChanged(acct)
	case model.EvCalendarChanged:
		h.mail.mu.Lock()
		delete(h.mail.calendars, acct)
		h.mail.mu.Unlock()
		h.emit(Event{Kind: EvCalendar, AccountID: acct})
	}
}
