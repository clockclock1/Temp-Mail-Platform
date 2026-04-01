package smtp

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"tempmail/backend/internal/service"

	smtplib "github.com/emersion/go-smtp"
	"gorm.io/gorm"
)

type Server struct {
	mailService *service.MailService
	address     string
	server      *smtplib.Server
}

func New(address string, mailService *service.MailService) *Server {
	backend := &backend{mailService: mailService}
	s := smtplib.NewServer(backend)
	s.Addr = address
	s.Domain = "localhost"
	s.AllowInsecureAuth = true
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxRecipients = 50
	s.MaxMessageBytes = 15 * 1024 * 1024

	return &Server{mailService: mailService, address: address, server: s}
}

func (s *Server) Start() error {
	log.Printf("smtp server listening on %s", s.address)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	ch := make(chan error, 1)
	go func() {
		ch <- s.server.Close()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-ch:
		return err
	}
}

type backend struct {
	mailService *service.MailService
}

func (b *backend) NewSession(c *smtplib.Conn) (smtplib.Session, error) {
	return &session{mailService: b.mailService, remoteAddr: c.Hostname()}, nil
}

type rcptTarget struct {
	envelopeTo string
	local      string
	domain     string
}

type session struct {
	mailService *service.MailService
	remoteAddr  string
	from        string
	targets     []rcptTarget
}

func (s *session) AuthPlain(username, password string) error {
	return nil
}

func (s *session) Mail(from string, opts *smtplib.MailOptions) error {
	s.from = strings.TrimSpace(from)
	s.targets = nil
	return nil
}

func (s *session) Rcpt(to string, opts *smtplib.RcptOptions) error {
	local, domain, err := splitAddress(to)
	if err != nil {
		return fmt.Errorf("invalid recipient")
	}

	_, err = s.mailService.FindMailboxByAddress(local, domain)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("mailbox not found")
		}
		return fmt.Errorf("lookup mailbox: %w", err)
	}

	s.targets = append(s.targets, rcptTarget{envelopeTo: strings.ToLower(strings.TrimSpace(to)), local: local, domain: domain})
	return nil
}

func (s *session) Data(r io.Reader) error {
	if len(s.targets) == 0 {
		return fmt.Errorf("no valid recipients")
	}
	raw, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read email data: %w", err)
	}

	for _, t := range s.targets {
		mailbox, err := s.mailService.FindMailboxByAddress(t.local, t.domain)
		if err != nil {
			continue
		}
		if _, err := s.mailService.StoreIncomingMessage(mailbox, t.envelopeTo, s.from, raw); err != nil {
			log.Printf("store message failed rcpt=%s err=%v", t.envelopeTo, err)
		}
	}
	return nil
}

func (s *session) Reset() {
	s.from = ""
	s.targets = nil
}

func (s *session) Logout() error {
	s.Reset()
	return nil
}

func splitAddress(addr string) (local, domain string, err error) {
	addr = strings.ToLower(strings.TrimSpace(addr))
	if strings.HasPrefix(addr, "<") && strings.HasSuffix(addr, ">") {
		addr = strings.TrimPrefix(strings.TrimSuffix(addr, ">"), "<")
	}
	if host, _, splitErr := net.SplitHostPort(addr); splitErr == nil {
		addr = host
	}
	parts := strings.Split(addr, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid address")
	}
	return parts[0], parts[1], nil
}
