package mailru

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime/quotedprintable"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/charset"

	"megachat/core/model"
	"megachat/core/provider"
	"megachat/core/providers/mailutil"
)

// Folder names that tell a role on servers without SPECIAL-USE (Mail.ru's
// Russian names among them).
var roleNames = map[string]model.FolderRole{
	"sent": model.FolderSent, "sent items": model.FolderSent, "sent messages": model.FolderSent, "отправленные": model.FolderSent,
	"drafts": model.FolderDrafts, "черновики": model.FolderDrafts,
	"trash": model.FolderTrash, "deleted": model.FolderTrash, "deleted items": model.FolderTrash, "корзина": model.FolderTrash,
	"spam": model.FolderSpam, "junk": model.FolderSpam, "спам": model.FolderSpam,
	"archive": model.FolderArchive, "архив": model.FolderArchive,
}

func roleOf(info *imap.MailboxInfo) model.FolderRole {
	if strings.EqualFold(info.Name, imap.InboxName) {
		return model.FolderInbox
	}
	for _, a := range info.Attributes {
		switch a {
		case imap.SentAttr:
			return model.FolderSent
		case imap.DraftsAttr:
			return model.FolderDrafts
		case imap.TrashAttr:
			return model.FolderTrash
		case imap.JunkAttr:
			return model.FolderSpam
		case imap.ArchiveAttr, imap.AllAttr:
			return model.FolderArchive
		}
	}
	leaf := info.Name
	if info.Delimiter != "" {
		if i := strings.LastIndex(leaf, info.Delimiter); i >= 0 {
			leaf = leaf[i+len(info.Delimiter):]
		}
	}
	return roleNames[strings.ToLower(leaf)]
}

var roleOrder = map[model.FolderRole]int{
	model.FolderInbox: 0, model.FolderSent: 1, model.FolderDrafts: 2, model.FolderArchive: 3,
	model.FolderSpam: 4, model.FolderTrash: 5, model.FolderOther: 6,
}

func (s *session) listFolders(c *client.Client) ([]*imap.MailboxInfo, error) {
	ch := make(chan *imap.MailboxInfo, 32)
	done := make(chan error, 1)
	go func() { done <- c.List("", "*", ch) }()
	var out []*imap.MailboxInfo
	for m := range ch {
		if !slices.Contains(m.Attributes, imap.NoSelectAttr) && !slices.Contains(m.Attributes, "\\NonExistent") {
			out = append(out, m)
		}
	}
	return out, <-done
}

func (s *session) Folders(ctx context.Context) ([]model.Folder, error) {
	var out []model.Folder
	err := s.imapDo(func(c *client.Client) error {
		infos, err := s.listFolders(c)
		if err != nil {
			return err
		}
		out = out[:0]
		seen := map[model.FolderRole]bool{}
		for _, info := range infos {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			st, err := c.Status(info.Name, []imap.StatusItem{imap.StatusMessages, imap.StatusUnseen})
			if err != nil {
				continue // listed but not readable
			}
			role := roleOf(info)
			if role != model.FolderOther && seen[role] {
				role = model.FolderOther // one folder per role
			}
			seen[role] = true
			name := info.Name
			if role == model.FolderInbox {
				name = "Inbox"
			} else if info.Delimiter != "" {
				name = strings.ReplaceAll(name, info.Delimiter, " / ")
			}
			out = append(out, model.Folder{ID: info.Name, Name: name, Role: role, Unread: int(st.Unseen), Total: int(st.Messages)})
		}
		return nil
	})
	sort.SliceStable(out, func(i, j int) bool { return roleOrder[out[i].Role] < roleOrder[out[j].Role] })
	return out, err
}

// mailID names a message: uid validity, uid and folder.
func mailID(validity, uid uint32, folder string) string {
	return fmt.Sprintf("%d:%d:%s", validity, uid, folder)
}

func parseID(id string) (validity, uid uint32, folder string, err error) {
	parts := strings.SplitN(id, ":", 3)
	if len(parts) == 3 {
		v, e1 := strconv.ParseUint(parts[0], 10, 32)
		u, e2 := strconv.ParseUint(parts[1], 10, 32)
		if e1 == nil && e2 == nil && parts[2] != "" {
			return uint32(v), uint32(u), parts[2], nil
		}
	}
	return 0, 0, "", fmt.Errorf("mailru: bad mail id")
}

func addr(a *imap.Address) model.Address {
	if a == nil {
		return model.Address{}
	}
	email := a.MailboxName
	if a.HostName != "" {
		email += "@" + a.HostName
	}
	return model.Address{Name: mailutil.DecodeHeader(a.PersonalName), Email: email}
}

func addrs(as []*imap.Address) []model.Address {
	var out []model.Address
	for _, a := range as {
		if a != nil && a.MailboxName != "" {
			out = append(out, addr(a))
		}
	}
	return out
}

var summaryItems = []imap.FetchItem{imap.FetchUid, imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, imap.FetchBodyStructure}

func summary(m *imap.Message, validity uint32, folder string) model.Mail {
	out := model.Mail{
		ID: mailID(validity, m.Uid, folder), FolderID: folder, Time: m.InternalDate,
		Unread: !slices.Contains(m.Flags, imap.SeenFlag), Starred: slices.Contains(m.Flags, imap.FlaggedFlag),
	}
	if e := m.Envelope; e != nil {
		out.Subject = mailutil.DecodeHeader(e.Subject)
		if len(e.From) > 0 {
			out.From = addr(e.From[0])
		}
		out.To, out.Cc, out.ReplyTo = addrs(e.To), addrs(e.Cc), addrs(e.ReplyTo)
		out.MessageID = e.MessageId
		if out.Time.IsZero() {
			out.Time = e.Date
		}
	}
	if bs := m.BodyStructure; bs != nil {
		bs.Walk(func(_ []int, p *imap.BodyStructure) bool {
			if isAttachment(p) {
				out.HasAttachments = true
			}
			return true
		})
	}
	return out
}

func isAttachment(p *imap.BodyStructure) bool {
	if strings.EqualFold(p.MIMEType, "multipart") {
		return false
	}
	if strings.EqualFold(p.Disposition, "attachment") {
		return true
	}
	name, _ := p.Filename()
	return name != "" && !strings.EqualFold(p.MIMEType, "text")
}

func fetch(c *client.Client, uid bool, set *imap.SeqSet, items []imap.FetchItem) ([]*imap.Message, error) {
	ch := make(chan *imap.Message, 64)
	done := make(chan error, 1)
	go func() {
		if uid {
			done <- c.UidFetch(set, items, ch)
		} else {
			done <- c.Fetch(set, items, ch)
		}
	}()
	var out []*imap.Message
	for m := range ch {
		out = append(out, m)
	}
	return out, <-done
}

func (s *session) Mail(ctx context.Context, folder string, before time.Time, limit int) ([]model.Mail, bool, error) {
	var out []model.Mail
	more := false
	err := s.imapDo(func(c *client.Client) error {
		out, more = nil, false
		st, err := c.Select(folder, true)
		if err != nil {
			return err
		}
		if st.Messages == 0 {
			return nil
		}
		// Sequence numbers follow arrival, so the newest are at the end.
		var seqs []uint32
		if before.IsZero() {
			for n := st.Messages; n >= 1 && len(seqs) < limit; n-- {
				seqs = append(seqs, n)
			}
			slices.Reverse(seqs)
		} else {
			// SEARCH takes whole days (server time); the exact cut follows.
			crit := imap.NewSearchCriteria()
			crit.Before = before.AddDate(0, 0, 2)
			if seqs, err = c.Search(crit); err != nil {
				return err
			}
			sort.Slice(seqs, func(i, j int) bool { return seqs[i] < seqs[j] })
		}
		end := len(seqs)
		for end > 0 && len(out) < limit {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			start := max(0, end-limit-10)
			set := new(imap.SeqSet)
			set.AddNum(seqs[start:end]...)
			msgs, err := fetch(c, false, set, summaryItems)
			if err != nil {
				return err
			}
			for _, m := range msgs {
				x := summary(m, st.UidValidity, folder)
				if before.IsZero() || !x.Time.After(before) {
					out = append(out, x)
				}
			}
			end = start
		}
		sort.SliceStable(out, func(i, j int) bool { return out[i].Time.After(out[j].Time) })
		if len(out) > limit {
			out, more = out[:limit], true
		}
		if end > 0 || (before.IsZero() && len(seqs) > 0 && seqs[0] > 1) {
			more = true
		}
		return nil
	})
	return out, more, err
}

// selectID selects the folder of a message and checks it is still the
// mailbox the id was made for.
func selectID(c *client.Client, id string, readOnly bool) (uint32, string, error) {
	validity, uid, folder, err := parseID(id)
	if err != nil {
		return 0, "", err
	}
	st, err := c.Select(folder, readOnly)
	if err != nil {
		return 0, "", err
	}
	if st.UidValidity != validity {
		return 0, "", fmt.Errorf("mailru: the folder changed; reload it")
	}
	return uid, folder, nil
}

func uidSet(uid uint32) *imap.SeqSet {
	set := new(imap.SeqSet)
	set.AddNum(uid)
	return set
}

func pathString(path []int) string {
	parts := make([]string, len(path))
	for i, p := range path {
		parts[i] = strconv.Itoa(p)
	}
	return strings.Join(parts, ".")
}

// section fetches one body part and undoes its transfer encoding.
func section(c *client.Client, uid uint32, path string, enc string, limit int64) ([]byte, error) {
	name, err := imap.ParseBodySectionName(imap.FetchItem("BODY.PEEK[" + path + "]"))
	if err != nil {
		return nil, err
	}
	msgs, err := fetch(c, true, uidSet(uid), []imap.FetchItem{name.FetchItem()})
	if err != nil {
		return nil, err
	}
	if len(msgs) == 0 {
		return nil, fmt.Errorf("mailru: message is gone")
	}
	lit := msgs[0].GetBody(name)
	if lit == nil {
		return nil, fmt.Errorf("mailru: no such part")
	}
	var r io.Reader = lit
	switch strings.ToLower(enc) {
	case "base64":
		r = base64.NewDecoder(base64.StdEncoding, newlineStripper{lit})
	case "quoted-printable":
		r = quotedprintable.NewReader(lit)
	}
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if int64(len(data)) > limit {
		return nil, provider.ErrTooLarge
	}
	return data, err
}

// newlineStripper drops line breaks from base64 text.
type newlineStripper struct{ r io.Reader }

func (n newlineStripper) Read(p []byte) (int, error) {
	k, err := n.r.Read(p)
	j := 0
	for _, b := range p[:k] {
		if b != '\r' && b != '\n' && b != ' ' && b != '\t' {
			p[j] = b
			j++
		}
	}
	return j, err
}

func decodeText(data []byte, cs string) string {
	if cs == "" || strings.EqualFold(cs, "utf-8") || strings.EqualFold(cs, "us-ascii") {
		return string(data)
	}
	r, err := charset.Reader(cs, bytes.NewReader(data))
	if err != nil {
		return string(data)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		return string(data)
	}
	return string(out)
}

const maxText = 4 << 20

func (s *session) MailBody(ctx context.Context, id string) (model.Mail, error) {
	var out model.Mail
	err := s.imapDo(func(c *client.Client) error {
		uid, folder, err := selectID(c, id, true)
		if err != nil {
			return err
		}
		refs := &imap.BodySectionName{BodyPartName: imap.BodyPartName{Specifier: imap.HeaderSpecifier, Fields: []string{"References"}}, Peek: true}
		msgs, err := fetch(c, true, uidSet(uid), append(slices.Clone(summaryItems), refs.FetchItem()))
		if err != nil {
			return err
		}
		if len(msgs) == 0 {
			return fmt.Errorf("mailru: message is gone")
		}
		m := msgs[0]
		validity, _, _, _ := parseID(id)
		out = summary(m, validity, folder)
		if lit := m.GetBody(refs); lit != nil {
			raw, _ := io.ReadAll(lit)
			if _, v, ok := strings.Cut(string(raw), ":"); ok {
				out.References = strings.Join(strings.Fields(v), " ")
			}
		}
		if m.BodyStructure == nil {
			return nil
		}
		type textPart struct {
			path string
			p    *imap.BodyStructure
		}
		var plain, html *textPart
		m.BodyStructure.Walk(func(path []int, p *imap.BodyStructure) bool {
			ps := pathString(path)
			switch {
			case isAttachment(p):
				name, _ := p.Filename()
				if name == "" {
					name = "attachment"
				}
				mt := strings.ToLower(p.MIMEType + "/" + p.MIMESubType)
				out.Attachments = append(out.Attachments, model.Attachment{
					Kind: mailutil.AttachmentKind(mt), Name: name, MIME: mt, Size: int64(p.Size), Ref: "part:" + id + "#" + ps,
				})
			case strings.EqualFold(p.MIMEType, "text") && strings.EqualFold(p.MIMESubType, "plain") && plain == nil:
				plain = &textPart{ps, p}
			case strings.EqualFold(p.MIMEType, "text") && strings.EqualFold(p.MIMESubType, "html") && html == nil:
				html = &textPart{ps, p}
			}
			return true
		})
		if plain != nil {
			if data, err := section(c, uid, plain.path, plain.p.Encoding, maxText); err == nil {
				out.Text = decodeText(data, plain.p.Params["charset"])
			}
		}
		if html != nil {
			if data, err := section(c, uid, html.path, html.p.Encoding, maxText); err == nil {
				out.HTML = decodeText(data, html.p.Params["charset"])
			}
		}
		if strings.TrimSpace(out.Text) == "" && out.HTML != "" {
			out.Text = mailutil.HTMLToText(out.HTML)
		}
		out.Snippet = mailutil.Snippet(out.Text, 140)
		out.HasAttachments = len(out.Attachments) > 0
		return nil
	})
	return out, err
}

// Media fetches an attachment: "part:<mail id>#<part path>".
func (s *session) Media(ctx context.Context, ref string, limit int64) ([]byte, string, error) {
	rest, ok := strings.CutPrefix(ref, "part:")
	i := strings.LastIndexByte(rest, '#')
	if !ok || i < 0 {
		return nil, "", fmt.Errorf("mailru: unknown media ref")
	}
	id, path := rest[:i], rest[i+1:]
	var data []byte
	var mt string
	err := s.imapDo(func(c *client.Client) error {
		uid, _, err := selectID(c, id, true)
		if err != nil {
			return err
		}
		msgs, err := fetch(c, true, uidSet(uid), []imap.FetchItem{imap.FetchBodyStructure})
		if err != nil || len(msgs) == 0 || msgs[0].BodyStructure == nil {
			return fmt.Errorf("mailru: message is gone")
		}
		var part *imap.BodyStructure
		msgs[0].BodyStructure.Walk(func(p []int, bs *imap.BodyStructure) bool {
			if pathString(p) == path {
				part = bs
			}
			return true
		})
		if part == nil {
			return fmt.Errorf("mailru: no such part")
		}
		mt = strings.ToLower(part.MIMEType + "/" + part.MIMESubType)
		data, err = section(c, uid, path, part.Encoding, limit)
		return err
	})
	return data, mt, err
}

// folderWith returns the name of the account's folder with a role.
func (s *session) folderWith(c *client.Client, role model.FolderRole) (string, error) {
	infos, err := s.listFolders(c)
	if err != nil {
		return "", err
	}
	for _, info := range infos {
		if roleOf(info) == role {
			return info.Name, nil
		}
	}
	return "", nil
}

func (s *session) UpdateMail(ctx context.Context, ids []string, action model.MailAction) error {
	byFolder := map[string][]string{}
	for _, id := range ids {
		_, _, folder, err := parseID(id)
		if err != nil {
			return err
		}
		byFolder[folder] = append(byFolder[folder], id)
	}
	return s.imapDo(func(c *client.Client) error {
		trash := ""
		if action == model.MailTrash {
			var err error
			if trash, err = s.folderWith(c, model.FolderTrash); err != nil {
				return err
			}
		}
		for folder, fids := range byFolder {
			set := new(imap.SeqSet)
			for _, id := range fids {
				uid, _, err := selectID(c, id, false)
				if err != nil {
					return err
				}
				set.AddNum(uid)
			}
			var err error
			op := func(add bool, flag string) error {
				item := imap.FormatFlagsOp(imap.RemoveFlags, true)
				if add {
					item = imap.FormatFlagsOp(imap.AddFlags, true)
				}
				return c.UidStore(set, item, []any{flag}, nil)
			}
			switch action {
			case model.MailRead:
				err = op(true, imap.SeenFlag)
			case model.MailUnread:
				err = op(false, imap.SeenFlag)
			case model.MailStar:
				err = op(true, imap.FlaggedFlag)
			case model.MailUnstar:
				err = op(false, imap.FlaggedFlag)
			case model.MailTrash:
				if trash == "" || trash == folder {
					// Already in the trash (or there is none): delete.
					if err = op(true, imap.DeletedFlag); err == nil {
						err = c.Expunge(nil)
					}
				} else {
					err = move(c, set, trash)
				}
			default:
				err = provider.ErrUnsupported
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// MoveMail moves messages to folder to (the folder each is in comes from
// its id).
func (s *session) MoveMail(ctx context.Context, ids []string, _, to string) error {
	byFolder := map[string][]string{}
	for _, id := range ids {
		_, _, folder, err := parseID(id)
		if err != nil {
			return err
		}
		if folder != to {
			byFolder[folder] = append(byFolder[folder], id)
		}
	}
	if len(byFolder) == 0 {
		return nil
	}
	return s.imapDo(func(c *client.Client) error {
		for _, fids := range byFolder {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			set := new(imap.SeqSet)
			for _, id := range fids {
				uid, _, err := selectID(c, id, false)
				if err != nil {
					return err
				}
				set.AddNum(uid)
			}
			if err := move(c, set, to); err != nil {
				return err
			}
		}
		return nil
	})
}

// pollInbox looks at the inbox's counters: new UIDs are new mail, other
// changes of the counters a change.
func (s *session) pollInbox(sink provider.Sink) error {
	var news []model.Mail
	changed := false
	err := s.imapDo(func(c *client.Client) error {
		st, err := c.Status(imap.InboxName, []imap.StatusItem{imap.StatusUidNext, imap.StatusUnseen, imap.StatusMessages, imap.StatusUidValidity})
		if err != nil {
			return err
		}
		s.mu.Lock()
		prev := s.inbox
		s.inbox = st
		s.mu.Unlock()
		if prev == nil || prev.UidValidity != st.UidValidity && prev.UidValidity != 0 {
			changed = prev != nil
			return nil
		}
		if st.UidNext > prev.UidNext && prev.UidNext > 0 {
			sel, err := c.Select(imap.InboxName, true)
			if err != nil {
				return err
			}
			set := new(imap.SeqSet)
			set.AddRange(prev.UidNext, st.UidNext-1)
			msgs, err := fetch(c, true, set, summaryItems)
			if err != nil {
				return err
			}
			for _, m := range msgs {
				if m.Uid < prev.UidNext {
					continue
				}
				if x := summary(m, sel.UidValidity, imap.InboxName); x.Unread {
					news = append(news, x)
				}
			}
		}
		changed = st.Unseen != prev.Unseen || st.Messages != prev.Messages || st.UidNext != prev.UidNext
		return nil
	})
	if err != nil {
		return err
	}
	for i := range news {
		sink.Emit(model.Event{Kind: model.EvMailNew, Mail: &news[i]})
	}
	if changed && len(news) == 0 {
		sink.Emit(model.Event{Kind: model.EvMailChanged})
	}
	return nil
}

// move moves messages, by COPY and expunging where MOVE does not work.
func move(c *client.Client, set *imap.SeqSet, dest string) error {
	if ok, _ := c.Support("MOVE"); ok {
		// Some servers announce MOVE and refuse it for some folders.
		if err := c.UidMove(set, dest); err == nil || c.State() == imap.LogoutState {
			return err
		}
	}
	if err := c.UidCopy(set, dest); err != nil {
		return err
	}
	if err := c.UidStore(set, imap.FormatFlagsOp(imap.AddFlags, true), []any{imap.DeletedFlag}, nil); err != nil {
		return err
	}
	return c.Expunge(nil)
}
