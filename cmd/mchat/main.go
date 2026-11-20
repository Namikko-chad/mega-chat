// Command mchat is the debugging CLI of MegaChat. It talks to the core of
// the running app (the cache is locked to one process), found through
// endpoint.json, so it sees exactly what the app sees.
//
//	mchat state | create | unlock [remembered] | lock
//	mchat providers | accounts | login PROVIDER [METHOD] | relogin ACCOUNT
//	mchat enable|disable|remove ACCOUNT
//	mchat chats [-n N] [ACCOUNT...]
//	mchat history ACCOUNT CHAT [-n N] [-before ID]
//	mchat send ACCOUNT CHAT TEXT...
//	mchat read ACCOUNT CHAT
//	mchat folders [ACCOUNT...]
//	mchat mail [-n N] [-before CURSOR] [ACCOUNT...] | mail ACCOUNT FOLDER
//	mchat mailread ACCOUNT ID
//	mchat sendmail ACCOUNT TO[,TO] SUBJECT TEXT...
//	mchat events [-n DAYS] [ACCOUNT...]
//	mchat watch | logs
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/term"
	"rsc.io/qr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"megachat/core/rpc/pb"
)

func main() {
	cfgDir, _ := os.UserConfigDir()
	endpoint := flag.String("endpoint", filepath.Join(cfgDir, "megachat", "endpoint.json"), "endpoint file written by the running app")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: mchat [-endpoint FILE] COMMAND ...  (see the package doc: go doc megachat/cmd/mchat)")
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	c, ctx, err := connect(*endpoint)
	if err != nil {
		fail(err)
	}
	if err := run(ctx, c, flag.Arg(0), flag.Args()[1:]); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "mchat:", err)
	os.Exit(1)
}

func connect(path string) (pb.MegaChatClient, context.Context, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("is the MegaChat app running? %w", err)
	}
	var ep struct{ Address, Token string }
	if err := json.Unmarshal(data, &ep); err != nil {
		return nil, nil, err
	}
	target := ep.Address
	if strings.HasPrefix(target, "unix://") {
		target = "unix:" + strings.TrimPrefix(target, "unix://")
	}
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+ep.Token)
	return pb.NewMegaChatClient(conn), ctx, nil
}

var stdin = bufio.NewReader(os.Stdin)

func ask(label string, secret bool) string {
	fmt.Fprint(os.Stderr, label)
	if secret && term.IsTerminal(int(os.Stdin.Fd())) {
		b, _ := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		return string(b)
	}
	line, _ := stdin.ReadString('\n')
	return strings.TrimRight(line, "\r\n")
}

func run(ctx context.Context, c pb.MegaChatClient, cmd string, args []string) error {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	n := fs.Int("n", 30, "how many")
	before := fs.String("before", "", "older than this message id")
	// Flags may come after the positional arguments ("history A C -n 5").
	var pos []string
	for {
		fs.Parse(args)
		if fs.NArg() == 0 {
			break
		}
		pos = append(pos, fs.Arg(0))
		args = fs.Args()[1:]
	}
	args = pos
	need := func(k int) error {
		if len(args) < k {
			return fmt.Errorf("%s needs %d arguments", cmd, k)
		}
		return nil
	}
	switch cmd {
	case "state":
		st, err := c.GetState(ctx, &pb.Empty{})
		if err != nil {
			return err
		}
		fmt.Printf("vault %s exists=%v unlocked=%v remembered=%v accounts=%d online=%d core=%s (%s)\n",
			st.VaultPath, st.VaultExists, st.Unlocked, st.Remembered, st.Accounts, st.AccountsOnline, st.CoreVersion, st.CoreBuild)
	case "create", "unlock":
		if cmd == "unlock" && len(args) > 0 && args[0] == "remembered" {
			_, err := c.UnlockRemembered(ctx, &pb.Empty{})
			return err
		}
		pw := ask("Password: ", true)
		var err error
		if cmd == "create" {
			if pw != ask("Again: ", true) {
				return fmt.Errorf("passwords differ")
			}
			_, err = c.CreateVault(ctx, &pb.PasswordRequest{Password: pw})
		} else {
			_, err = c.Unlock(ctx, &pb.PasswordRequest{Password: pw})
		}
		return err
	case "lock":
		_, err := c.LockVault(ctx, &pb.Empty{})
		return err
	case "providers":
		r, err := c.Providers(ctx, &pb.Empty{})
		if err != nil {
			return err
		}
		for _, p := range r.Providers {
			fmt.Printf("%-10s %s\n", p.Type, p.Name)
			for _, m := range p.AuthMethods {
				fmt.Printf("    %-10s %s — %s\n", m.Id, m.Name, m.Description)
			}
		}
	case "accounts":
		r, err := c.Accounts(ctx, &pb.Empty{})
		if err != nil {
			return err
		}
		for _, a := range r.Accounts {
			st := strings.ToLower(strings.TrimPrefix(a.Status.String(), "ACCOUNT_STATUS_"))
			fmt.Printf("%s  %-9s %-24s %-30s %-12s unread=%d %s\n", a.Id, a.Provider, a.Name, a.Login, st, a.Unread, a.StatusMessage)
		}
	case "login", "relogin":
		if err := need(1); err != nil {
			return err
		}
		var step *pb.LoginStepReply
		var err error
		if cmd == "login" {
			method := ""
			if len(args) > 1 {
				method = args[1]
			}
			step, err = c.StartLogin(ctx, &pb.StartLoginRequest{Provider: args[0], Method: method})
		} else {
			step, err = c.Relogin(ctx, &pb.IdRequest{Id: args[0]})
		}
		if err != nil {
			return err
		}
		return loginDialog(ctx, c, step)
	case "enable", "disable":
		if err := need(1); err != nil {
			return err
		}
		_, err := c.SetAccountEnabled(ctx, &pb.SetEnabledRequest{Id: args[0], Enabled: cmd == "enable"})
		return err
	case "remove":
		if err := need(1); err != nil {
			return err
		}
		_, err := c.RemoveAccount(ctx, &pb.IdRequest{Id: args[0]})
		return err
	case "chats":
		r, err := c.Chats(ctx, &pb.ChatsRequest{AccountIds: args, Limit: int32(*n)})
		if err != nil {
			return err
		}
		for _, ch := range r.Chats {
			last := ""
			if m := ch.LastMessage; m != nil {
				last = m.Sender.GetName() + ": " + oneLine(m.Text, 50)
			}
			fmt.Printf("%s %-24s %-30s %3d  %s\n", ch.AccountId, ch.Id, oneLine(ch.Title, 30), ch.Unread, last)
		}
	case "history":
		if err := need(2); err != nil {
			return err
		}
		r, err := c.History(ctx, &pb.HistoryRequest{AccountId: args[0], ChatId: args[1], BeforeId: *before, Limit: int32(*n)})
		if err != nil {
			return err
		}
		for i := len(r.Messages) - 1; i >= 0; i-- {
			printMessage(r.Messages[i])
		}
		if r.More {
			fmt.Printf("… older: -before %s\n", r.Messages[len(r.Messages)-1].GetId())
		}
	case "send":
		if err := need(3); err != nil {
			return err
		}
		m, err := c.Send(ctx, &pb.SendRequest{AccountId: args[0], ChatId: args[1], Text: strings.Join(args[2:], " ")})
		if err != nil {
			return err
		}
		printMessage(m)
	case "read":
		if err := need(2); err != nil {
			return err
		}
		_, err := c.MarkRead(ctx, &pb.MarkReadRequest{AccountId: args[0], ChatId: args[1]})
		return err
	case "watch":
		s, err := c.Watch(ctx, &pb.Empty{})
		if err != nil {
			return err
		}
		for {
			ev, err := s.Recv()
			if err != nil {
				return err
			}
			kind := strings.ToLower(strings.TrimPrefix(ev.Kind.String(), "KIND_"))
			switch {
			case ev.Message != nil:
				fmt.Printf("%-14s %s/%s ", kind, ev.AccountId, ev.Message.ChatId)
				printMessage(ev.Message)
			case ev.Chat != nil:
				fmt.Printf("%-14s %s/%s %q unread=%d\n", kind, ev.AccountId, ev.Chat.Id, ev.Chat.Title, ev.Chat.Unread)
			case ev.Account != nil:
				fmt.Printf("%-14s %s %s %s\n", kind, ev.Account.Id, ev.Account.Status, ev.Account.StatusMessage)
			default:
				fmt.Println(kind)
			}
		}
	case "logs":
		r, err := c.Logs(ctx, &pb.LogsRequest{Limit: int32(*n)})
		if err != nil {
			return err
		}
		for _, e := range r.Entries {
			fmt.Printf("%s %-5s %-8s %s %s (x%d)\n", time.UnixMilli(e.TimeMs).Format(time.DateTime), e.Level, e.Source, e.AccountId, e.Message, e.Count)
		}
	default:
		if ok, err := runMail(ctx, c, cmd, args, *n, *before); ok {
			return err
		}
		return fmt.Errorf("unknown command %q", cmd)
	}
	return nil
}

func loginDialog(ctx context.Context, c pb.MegaChatClient, step *pb.LoginStepReply) error {
	shownQR := ""
	for !step.Done {
		if step.Qr != "" && step.Qr == shownQR && len(step.Fields) == 0 {
			// Same code, still waiting for the scan.
			next, err := c.LoginStep(ctx, &pb.LoginStepRequest{LoginId: step.LoginId})
			if err != nil {
				return err
			}
			step = next
			continue
		}
		shownQR = step.Qr
		if step.Title != "" {
			fmt.Fprintf(os.Stderr, "\n== %s\n", step.Title)
		}
		if step.Message != "" {
			fmt.Fprintln(os.Stderr, step.Message)
		}
		if step.Error != "" {
			fmt.Fprintln(os.Stderr, "!", step.Error)
		}
		if step.OpenUrl != "" {
			fmt.Fprintln(os.Stderr, "Open in a browser:", step.OpenUrl)
		}
		if step.Qr != "" {
			printQR(step.Qr)
		}
		values := map[string]string{}
		for _, f := range step.Fields {
			label := f.Label
			if f.Help != "" {
				label += " (" + f.Help + ")"
			}
			if f.DefaultValue != "" && !f.Secret {
				label += " [" + f.DefaultValue + "]"
			}
			v := ask(label+": ", f.Secret)
			if v == "" {
				v = f.DefaultValue
			}
			values[f.Key] = v
		}
		if (step.OpenUrl != "" && values["callback"] == "") || (step.Qr != "" && len(step.Fields) == 0) {
			values = nil // wait for the redirect or the scan
		}
		next, err := c.LoginStep(ctx, &pb.LoginStepRequest{LoginId: step.LoginId, Values: values})
		if err != nil {
			return err
		}
		step = next
	}
	a := step.Account
	fmt.Printf("added %s: %s (%s)\n", a.Id, a.Name, a.Login)
	return nil
}

// printQR draws a QR code with half blocks (two rows per line), with the
// quiet zone scanners need.
func printQR(text string) {
	c, err := qr.Encode(text, qr.L)
	if err != nil {
		fmt.Fprintln(os.Stderr, "QR:", text)
		return
	}
	const quiet = 2
	dark := func(x, y int) bool {
		x, y = x-quiet, y-quiet
		return x >= 0 && y >= 0 && x < c.Size && y < c.Size && c.Black(x, y)
	}
	n := c.Size + 2*quiet
	var b strings.Builder
	for y := 0; y < n; y += 2 {
		for x := 0; x < n; x++ {
			// Light on dark terminals: a dark module is a space.
			switch top, bottom := !dark(x, y), !dark(x, y+1); {
			case top && bottom:
				b.WriteRune('█')
			case top:
				b.WriteRune('▀')
			case bottom:
				b.WriteRune('▄')
			default:
				b.WriteRune(' ')
			}
		}
		b.WriteByte('\n')
	}
	fmt.Fprint(os.Stderr, b.String())
}

func oneLine(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > max {
		return string(r[:max-1]) + "…"
	}
	return s
}

func printMessage(m *pb.Message) {
	t := time.UnixMilli(m.TimeMs).Format("01-02 15:04")
	mark := ""
	if m.EditedMs != 0 {
		mark = " (edited)"
	}
	if m.Deleted {
		mark = " (deleted)"
	}
	text := m.Text
	for _, a := range m.Attachments {
		text += fmt.Sprintf(" [%s %s]", strings.ToLower(strings.TrimPrefix(a.Kind.String(), "ATTACHMENT_KIND_")), a.Name)
	}
	fmt.Printf("%s %s  %s: %s%s\n", t, m.Id, m.Sender.GetName(), text, mark)
}
