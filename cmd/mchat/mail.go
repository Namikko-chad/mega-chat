package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"megachat/core/rpc/pb"
)

// runMail serves the mail and calendar commands; ok is false for others.
func runMail(ctx context.Context, c pb.MegaChatClient, cmd string, args []string, n int, before string) (ok bool, err error) {
	failed := func(fs []*pb.AccountError) {
		for _, f := range fs {
			fmt.Printf("! %s: %s\n", f.AccountId, f.Error)
		}
	}
	switch cmd {
	case "folders":
		r, err := c.Folders(ctx, &pb.AccountsRequest{AccountIds: args})
		if err != nil {
			return true, err
		}
		for _, f := range r.Folders {
			fmt.Printf("%s  %-9s %-30q unread %d of %d\n", f.AccountId, strings.TrimPrefix(f.Role.String(), "FOLDER_ROLE_"), f.Name, f.Unread, f.Total)
		}
		failed(r.Failed)
	case "mail":
		req := &pb.MailListRequest{Role: pb.FolderRole_FOLDER_ROLE_INBOX, Cursor: before, Limit: int32(n)}
		if len(args) == 2 {
			req.AccountIds, req.FolderId = args[:1], args[1]
		} else {
			req.AccountIds = args
		}
		r, err := c.MailList(ctx, req)
		if err != nil {
			return true, err
		}
		for _, m := range r.Mails {
			mark := " "
			if m.Unread {
				mark = "*"
			}
			from := m.From.GetName()
			if from == "" {
				from = m.From.GetEmail()
			}
			fmt.Printf("%s %s %s  %-24s %s\n    %s %s\n", mark, time.UnixMilli(m.TimeMs).Format("01-02 15:04"), m.AccountId, oneLine(from, 24), oneLine(m.Subject, 60), m.AccountId, m.Id)
		}
		failed(r.Failed)
		if r.More {
			fmt.Printf("… older: -before %s\n", r.Next)
		}
	case "mailread":
		if len(args) < 2 {
			return true, fmt.Errorf("mailread needs ACCOUNT ID")
		}
		m, err := c.GetMail(ctx, &pb.GetMailRequest{AccountId: args[0], Id: args[1]})
		if err != nil {
			return true, err
		}
		fmt.Printf("From: %s <%s>\nSubject: %s\nDate: %s\n\n%s\n", m.From.GetName(), m.From.GetEmail(), m.Subject, time.UnixMilli(m.TimeMs).Format(time.DateTime), m.Text)
		for _, a := range m.Attachments {
			fmt.Printf("[%s, %d bytes] %s\n", a.Name, a.Size, a.Ref)
		}
	case "sendmail":
		if len(args) < 4 {
			return true, fmt.Errorf("sendmail needs ACCOUNT TO SUBJECT TEXT...")
		}
		var to []*pb.Address
		for _, a := range strings.Split(args[1], ",") {
			to = append(to, &pb.Address{Email: strings.TrimSpace(a)})
		}
		_, err := c.SendMail(ctx, &pb.SendMailRequest{AccountId: args[0], To: to, Subject: args[2], Text: strings.Join(args[3:], " ")})
		return true, err
	case "events":
		from := time.Now().Truncate(24 * time.Hour)
		r, err := c.Events(ctx, &pb.EventsRequest{AccountIds: args, FromMs: from.UnixMilli(), ToMs: from.AddDate(0, 0, n).UnixMilli()})
		if err != nil {
			return true, err
		}
		for _, e := range r.Events {
			when := time.UnixMilli(e.StartMs).Format("Mon 01-02 15:04")
			if e.AllDay {
				when = time.UnixMilli(e.StartMs).UTC().Format("Mon 01-02") + " all day"
			}
			fmt.Printf("%-22s %s  %s\n", when, e.AccountId, oneLine(e.Title, 60))
		}
		failed(r.Failed)
	default:
		return false, nil
	}
	return true, nil
}
