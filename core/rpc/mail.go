package rpc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"megachat/core/hub"
	"megachat/core/model"
	"megachat/core/richtext"
	"megachat/core/rpc/pb"
)

func init() {
	eventKindToPB[hub.EvMailNew] = pb.Event_KIND_MAIL_NEW
	eventKindToPB[hub.EvMail] = pb.Event_KIND_MAIL
	eventKindToPB[hub.EvCalendar] = pb.Event_KIND_CALENDAR
}

// --- mail ------------------------------------------------------------------

func failedToPB(errs []hub.AccountError) []*pb.AccountError {
	out := make([]*pb.AccountError, len(errs))
	for i, e := range errs {
		out[i] = &pb.AccountError{AccountId: e.Account, Error: status.Convert(toStatus(e.Err)).Message()}
	}
	return out
}

func (s *Service) Folders(ctx context.Context, r *pb.AccountsRequest) (*pb.FoldersReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	afs, errs := h.Folders(ctx, r.AccountIds)
	out := &pb.FoldersReply{Failed: failedToPB(errs)}
	for _, af := range afs {
		for _, f := range af.Folders {
			out.Folders = append(out.Folders, &pb.Folder{
				AccountId: af.Account, Id: f.ID, Name: f.Name, Role: pb.FolderRole(f.Role),
				Unread: int32(f.Unread), Total: int32(f.Total),
			})
		}
	}
	return out, nil
}

// cursor is MailListReply.next: opaque to clients.
type cursor struct {
	T int64  `json:"t"`
	A string `json:"a"`
	I string `json:"i"`
}

func encodeCursor(c *hub.MailCursor) string {
	if c == nil {
		return ""
	}
	js, _ := json.Marshal(cursor{T: c.Time.UnixMilli(), A: c.Account, I: c.ID})
	return base64.RawURLEncoding.EncodeToString(js)
}

func decodeCursor(s string) (*hub.MailCursor, error) {
	if s == "" {
		return nil, nil
	}
	js, err := base64.RawURLEncoding.DecodeString(s)
	var c cursor
	if err == nil {
		err = json.Unmarshal(js, &c)
	}
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad cursor")
	}
	return &hub.MailCursor{Time: time.UnixMilli(c.T), Account: c.A, ID: c.I}, nil
}

func (s *Service) MailList(ctx context.Context, r *pb.MailListRequest) (*pb.MailListReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	after, err := decodeCursor(r.Cursor)
	if err != nil {
		return nil, err
	}
	page, err := h.Mail(ctx, hub.MailQuery{
		Accounts: r.AccountIds, Role: model.FolderRole(r.Role), Folder: r.FolderId, After: after, Limit: int(r.Limit),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	out := &pb.MailListReply{More: page.More, Next: encodeCursor(page.Next), Failed: failedToPB(page.Failed)}
	for i := range page.Mails {
		out.Mails = append(out.Mails, mailToPB(&page.Mails[i]))
	}
	return out, nil
}

func (s *Service) GetMail(ctx context.Context, r *pb.GetMailRequest) (*pb.Mail, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	m, err := h.MailBody(ctx, r.AccountId, r.Id)
	if err != nil {
		return nil, toStatus(err)
	}
	return mailToPB(&m), nil
}

func (s *Service) SendMail(ctx context.Context, r *pb.SendMailRequest) (*pb.Empty, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	out := model.OutMail{
		To: addresses(r.To), Cc: addresses(r.Cc), Bcc: addresses(r.Bcc),
		Subject: r.Subject, Text: r.Text, ReplyTo: r.ReplyToId,
	}
	if len(out.To)+len(out.Cc)+len(out.Bcc) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no recipients")
	}
	return &pb.Empty{}, toStatus(h.SendMail(ctx, r.AccountId, out))
}

func (s *Service) UpdateMail(ctx context.Context, r *pb.UpdateMailRequest) (*pb.Empty, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	if r.Action == pb.MailAction_MAIL_ACTION_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "no action")
	}
	return &pb.Empty{}, toStatus(h.UpdateMail(ctx, r.AccountId, r.Ids, model.MailAction(r.Action)))
}

func addressToPB(a model.Address) *pb.Address { return &pb.Address{Name: a.Name, Email: a.Email} }

func addressesToPB(as []model.Address) []*pb.Address {
	out := make([]*pb.Address, len(as))
	for i, a := range as {
		out[i] = addressToPB(a)
	}
	return out
}

func addresses(as []*pb.Address) []model.Address {
	var out []model.Address
	for _, a := range as {
		if a.GetEmail() != "" {
			out = append(out, model.Address{Name: a.GetName(), Email: a.GetEmail()})
		}
	}
	return out
}

func mailToPB(m *model.Mail) *pb.Mail {
	if m == nil {
		return nil
	}
	out := &pb.Mail{
		AccountId: m.AccountID, Id: m.ID, FolderId: m.FolderID, ThreadId: m.ThreadID,
		From: addressToPB(m.From), To: addressesToPB(m.To), Cc: addressesToPB(m.Cc), ReplyTo: addressesToPB(m.ReplyTo),
		Subject: m.Subject, Snippet: m.Snippet, TimeMs: ms(m.Time),
		Unread: m.Unread, Starred: m.Starred, HasAttachments: m.HasAttachments || len(m.Attachments) > 0,
		Text: m.Text, Html: m.HTML,
	}
	// The body as text and styles: the HTML part when it says anything,
	// else the plain part read as Markdown.
	var body richtext.Doc
	if m.HTML != "" {
		body = richtext.FromHTML(m.HTML)
	}
	if strings.TrimSpace(body.Text) == "" && m.Text != "" {
		body = richtext.Parse(m.Text, richtext.Markdown)
	}
	if body.Text != "" {
		out.Text, out.Format = body.Text, spansToPB(body.Spans)
	}
	for _, a := range m.Attachments {
		out.Attachments = append(out.Attachments, &pb.Attachment{
			Kind: pb.AttachmentKind(a.Kind), Name: a.Name, Mime: a.MIME, Size: a.Size,
			Ref: a.Ref, Thumb: a.Thumb, Width: int32(a.Width), Height: int32(a.Height),
		})
	}
	return out
}

// --- calendars -------------------------------------------------------------

func (s *Service) Calendars(ctx context.Context, r *pb.AccountsRequest) (*pb.CalendarsReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	cs, errs := h.Calendars(ctx, r.AccountIds)
	out := &pb.CalendarsReply{Failed: failedToPB(errs)}
	for _, c := range cs {
		out.Calendars = append(out.Calendars, &pb.Calendar{
			AccountId: c.AccountID, Id: c.ID, Name: c.Name, Color: int32(c.Color), ReadOnly: c.ReadOnly, Primary: c.Primary,
		})
	}
	return out, nil
}

func (s *Service) Events(ctx context.Context, r *pb.EventsRequest) (*pb.EventsReply, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	q := hub.EventsQuery{Accounts: r.AccountIds, From: time.UnixMilli(r.FromMs), To: time.UnixMilli(r.ToMs)}
	for _, c := range r.Calendars {
		q.Calendars = append(q.Calendars, hub.CalendarKey{Account: c.AccountId, Calendar: c.CalendarId})
	}
	evs, errs, err := h.Events(ctx, q)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	out := &pb.EventsReply{Failed: failedToPB(errs)}
	for i := range evs {
		out.Events = append(out.Events, eventToCalPB(&evs[i]))
	}
	return out, nil
}

func (s *Service) SaveEvent(ctx context.Context, r *pb.CalendarEvent) (*pb.CalendarEvent, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	ev, err := h.SaveEvent(ctx, calEvent(r))
	if err != nil {
		return nil, toStatus(err)
	}
	return eventToCalPB(&ev), nil
}

func (s *Service) DeleteEvent(ctx context.Context, r *pb.DeleteEventRequest) (*pb.Empty, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, toStatus(h.DeleteEvent(ctx, r.AccountId, r.CalendarId, r.Id))
}

func eventToCalPB(e *model.CalEvent) *pb.CalendarEvent {
	out := &pb.CalendarEvent{
		AccountId: e.AccountID, CalendarId: e.CalendarID, Id: e.ID, Title: e.Title,
		Location: e.Location, Description: e.Description,
		StartMs: e.Start.UnixMilli(), EndMs: e.End.UnixMilli(), AllDay: e.AllDay, Recurring: e.Recurring,
		Link: e.Link,
	}
	if e.Organizer.Email != "" || e.Organizer.Name != "" {
		out.Organizer = addressToPB(e.Organizer)
	}
	for _, a := range e.Attendees {
		out.Attendees = append(out.Attendees, &pb.Attendee{Address: addressToPB(a.Address), Status: pb.Rsvp(a.Status), Self: a.Self})
	}
	return out
}

func calEvent(r *pb.CalendarEvent) model.CalEvent {
	e := model.CalEvent{
		AccountID: r.AccountId, CalendarID: r.CalendarId, ID: r.Id, Title: r.Title,
		Location: r.Location, Description: r.Description,
		Start: time.UnixMilli(r.StartMs), End: time.UnixMilli(r.EndMs), AllDay: r.AllDay,
	}
	if e.AllDay {
		e.Start, e.End = e.Start.UTC(), e.End.UTC()
	}
	for _, a := range r.Attendees {
		if a.GetAddress().GetEmail() != "" {
			e.Attendees = append(e.Attendees, model.Attendee{
				Address: model.Address{Name: a.Address.Name, Email: a.Address.Email}, Status: model.RSVP(a.Status),
			})
		}
	}
	return e
}

func (s *Service) MoveMail(ctx context.Context, r *pb.MoveMailRequest) (*pb.Empty, error) {
	h, err := s.hub()
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, toStatus(h.MoveMail(ctx, r.AccountId, r.Ids, r.FromFolderId, r.ToFolderId))
}
