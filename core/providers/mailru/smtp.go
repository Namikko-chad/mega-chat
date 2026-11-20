package mailru

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"

	"megachat/core/model"
	"megachat/core/provider"
	"megachat/core/providers/mailutil"
)

// sendSMTP delivers raw to the recipients. Port 465 is TLS from the start,
// any other port upgrades with STARTTLS (required unless Plain, in tests).
func (p *Provider) sendSMTP(ctx context.Context, cfg config, from string, rcpt []string, raw []byte) error {
	host, port, err := net.SplitHostPort(cfg.SMTP)
	if err != nil {
		return fmt.Errorf("mailru: bad SMTP server %q", cfg.SMTP)
	}
	tcfg := &tls.Config{ServerName: host}
	if p.TLS != nil {
		tcfg = p.TLS.Clone()
	}
	d := &net.Dialer{Timeout: 30 * time.Second}
	var conn net.Conn
	if port == "465" && !p.Plain {
		conn, err = tls.DialWithDialer(d, "tcp", cfg.SMTP, tcfg)
	} else {
		conn, err = d.DialContext(ctx, "tcp", cfg.SMTP)
	}
	if err != nil {
		return fmt.Errorf("SMTP %s: %v", cfg.SMTP, err)
	}
	deadline := time.Now().Add(2 * time.Minute)
	if dl, ok := ctx.Deadline(); ok && dl.Before(deadline) {
		deadline = dl
	}
	conn.SetDeadline(deadline)
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("SMTP: %v", err)
	}
	defer c.Close()
	if _, isTLS := conn.(*tls.Conn); !isTLS && !p.Plain {
		if ok, _ := c.Extension("STARTTLS"); !ok {
			return fmt.Errorf("SMTP %s offers no TLS", cfg.SMTP)
		}
		if err := c.StartTLS(tcfg); err != nil {
			return fmt.Errorf("SMTP STARTTLS: %v", err)
		}
	}
	if ok, _ := c.Extension("AUTH"); ok {
		auth := smtp.PlainAuth("", cfg.Email, cfg.Password, host)
		if p.Plain {
			auth = plainAnyway{cfg.Email, cfg.Password}
		}
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("SMTP login: %w: %v", provider.ErrAuth, err)
		}
	}
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("SMTP: %v", err)
	}
	for _, r := range rcpt {
		if err := c.Rcpt(r); err != nil {
			return fmt.Errorf("SMTP: %s: %v", r, err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("SMTP: %v", err)
	}
	if _, err := w.Write(raw); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("SMTP: %v", err)
	}
	return c.Quit()
}

// plainAnyway is PLAIN auth over a connection without TLS (tests only).
type plainAnyway struct{ user, pass string }

func (a plainAnyway) Start(*smtp.ServerInfo) (string, []byte, error) {
	return "PLAIN", []byte("\x00" + a.user + "\x00" + a.pass), nil
}

func (a plainAnyway) Next([]byte, bool) ([]byte, error) { return nil, nil }

// replyInfo reads the threading headers of the message a reply answers.
func (s *session) replyInfo(id string) (*mailutil.Reply, error) {
	var out *mailutil.Reply
	err := s.imapDo(func(c *client.Client) error {
		uid, _, err := selectID(c, id, true)
		if err != nil {
			return err
		}
		refs := &imap.BodySectionName{BodyPartName: imap.BodyPartName{Specifier: imap.HeaderSpecifier, Fields: []string{"References"}}, Peek: true}
		msgs, err := fetch(c, true, uidSet(uid), []imap.FetchItem{imap.FetchEnvelope, refs.FetchItem()})
		if err != nil || len(msgs) == 0 || msgs[0].Envelope == nil {
			return fmt.Errorf("mailru: the message replied to is gone")
		}
		out = &mailutil.Reply{MessageID: msgs[0].Envelope.MessageId}
		if lit := msgs[0].GetBody(refs); lit != nil {
			raw, _ := io.ReadAll(lit)
			if _, v, ok := strings.Cut(string(raw), ":"); ok {
				out.References = strings.Join(strings.Fields(v), " ")
			}
		}
		return nil
	})
	return out, err
}

func (s *session) SendMail(ctx context.Context, out model.OutMail) error {
	var reply *mailutil.Reply
	if out.ReplyTo != "" {
		var err error
		if reply, err = s.replyInfo(out.ReplyTo); err != nil {
			return err
		}
	}
	now := time.Now()
	raw := mailutil.Compose(model.Address{Email: s.cfg.Email}, out, reply, now)
	if err := s.p.sendSMTP(ctx, s.cfg, s.cfg.Email, mailutil.Recipients(out), raw); err != nil {
		return err
	}
	if !s.cfg.SentCopy {
		return nil
	}
	return s.imapDo(func(c *client.Client) error {
		sent, err := s.folderWith(c, model.FolderSent)
		if err != nil || sent == "" {
			return err
		}
		return c.Append(sent, []string{imap.SeenFlag}, now, literal{strings.NewReader(string(raw)), len(raw)})
	})
}

// literal is an IMAP literal of a known size.
type literal struct {
	*strings.Reader
	n int
}

func (l literal) Len() int { return l.n }
