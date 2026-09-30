package fs

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeSMTP yalnızca test için minimal bir SMTP sunucusudur (loopback).
type fakeSMTP struct {
	ln net.Listener
	// mailReply n'inci bağlantıdaki MAIL komutuna verilecek yanıt (boşsa 250).
	mailReply func(conn int) string
	// hangAfterData true ise mesaj gövdesinden sonra yanıt verilmez.
	hangAfterData bool

	conns     atomic.Int32
	delivered atomic.Int32
	wg        sync.WaitGroup
}

func newFakeSMTP(t *testing.T) *fakeSMTP {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &fakeSMTP{ln: ln}
	t.Cleanup(func() { ln.Close(); s.wg.Wait() })
	return s
}

func (s *fakeSMTP) start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			c, err := s.ln.Accept()
			if err != nil {
				return
			}
			n := int(s.conns.Add(1))
			s.wg.Add(1)
			go func() { defer s.wg.Done(); s.handle(c, n) }()
		}
	}()
}

func (s *fakeSMTP) handle(c net.Conn, n int) {
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(5 * time.Second))
	r := bufio.NewReader(c)
	w := func(line string) { fmt.Fprintf(c, "%s\r\n", line) }
	w("220 fake ESMTP")
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		cmd := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(cmd, "EHLO"), strings.HasPrefix(cmd, "HELO"):
			w("250-fake")
			w("250 8BITMIME")
		case strings.HasPrefix(cmd, "MAIL"):
			if s.mailReply != nil {
				if rep := s.mailReply(n); rep != "" {
					w(rep)
					continue
				}
			}
			w("250 ok")
		case strings.HasPrefix(cmd, "RCPT"):
			w("250 ok")
		case cmd == "DATA":
			w("354 go ahead")
			tr := textproto.NewReader(r)
			if _, err := tr.ReadDotBytes(); err != nil {
				return
			}
			s.delivered.Add(1)
			if s.hangAfterData {
				_, _ = io.Copy(io.Discard, r) // yanıt vermeden bekle
				return
			}
			w("250 queued")
		case cmd == "QUIT":
			w("221 bye")
			return
		case cmd == "RSET", cmd == "NOOP":
			w("250 ok")
		default:
			w("502 not implemented")
		}
	}
}

func (s *fakeSMTP) cfg() SMTPCfg {
	host, portStr, _ := net.SplitHostPort(s.ln.Addr().String())
	port, _ := strconv.Atoi(portStr)
	return SMTPCfg{Host: host, Port: port, Timeout: 2 * time.Second, RetryDelay: 10 * time.Millisecond}
}

func TestIsTransientSMTPError(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{&textproto.Error{Code: 421, Msg: "busy"}, true},
		{&textproto.Error{Code: 451, Msg: "later"}, true},
		{&textproto.Error{Code: 550, Msg: "mailbox 421 unavailable"}, false},
		{errors.New("550 user 452 unknown"), false},
		{errors.New("452 too many recipients"), true},
		{errors.New("something 421 happened"), false},
		{nil, false},
	}
	for _, c := range cases {
		if got := isTransientSMTPError(c.err); got != c.want {
			t.Errorf("isTransientSMTPError(%v) = %v, want %v", c.err, got, c.want)
		}
	}
}

func TestSendEmail_Success(t *testing.T) {
	s := newFakeSMTP(t)
	s.start()
	err := SendEmail(context.Background(), s.cfg(), "a@example.com", []string{"b@example.com"}, nil, []string{"c@example.com"}, "hi", "body", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if s.delivered.Load() != 1 {
		t.Fatalf("delivered = %d", s.delivered.Load())
	}
}

func TestSendEmail_RetriesTransientOnly(t *testing.T) {
	s := newFakeSMTP(t)
	s.mailReply = func(n int) string {
		if n == 1 {
			return "451 try again later"
		}
		return ""
	}
	s.start()
	cfg := s.cfg()
	cfg.RetryAttempts = 3
	if err := SendEmail(context.Background(), cfg, "a@example.com", []string{"b@example.com"}, nil, nil, "s", "b", "", nil); err != nil {
		t.Fatal(err)
	}
	if s.conns.Load() != 2 || s.delivered.Load() != 1 {
		t.Fatalf("conns=%d delivered=%d", s.conns.Load(), s.delivered.Load())
	}

	// Kalıcı hata (550) yeniden denenmez.
	p := newFakeSMTP(t)
	p.mailReply = func(int) string { return "550 mailbox 421 unavailable" }
	p.start()
	cfg2 := p.cfg()
	cfg2.RetryAttempts = 3
	if err := SendEmail(context.Background(), cfg2, "a@example.com", []string{"b@example.com"}, nil, nil, "s", "b", "", nil); err == nil {
		t.Fatal("expected error")
	}
	if p.conns.Load() != 1 {
		t.Fatalf("permanent error retried: conns=%d", p.conns.Load())
	}
}

func TestSendEmail_TimeoutAfterDataIsNotRetried(t *testing.T) {
	s := newFakeSMTP(t)
	s.hangAfterData = true
	s.start()
	cfg := s.cfg()
	cfg.Timeout = 300 * time.Millisecond
	cfg.RetryAttempts = 3
	start := time.Now()
	err := SendEmail(context.Background(), cfg, "a@example.com", []string{"b@example.com"}, nil, nil, "s", "b", "", nil)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if time.Since(start) > 3*time.Second {
		t.Fatalf("timeout not honoured: %v", time.Since(start))
	}
	if s.delivered.Load() != 1 || s.conns.Load() != 1 {
		t.Fatalf("double send risk: conns=%d delivered=%d", s.conns.Load(), s.delivered.Load())
	}
}

func TestSendEmail_ContextCancelStopsImmediately(t *testing.T) {
	s := newFakeSMTP(t)
	s.hangAfterData = true
	s.start()
	cfg := s.cfg()
	cfg.Timeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	if err := SendEmail(ctx, cfg, "a@example.com", []string{"b@example.com"}, nil, nil, "s", "b", "", nil); err == nil {
		t.Fatal("expected error")
	}
	if time.Since(start) > 3*time.Second {
		t.Fatalf("ctx cancel not honoured: %v", time.Since(start))
	}
}

func TestSendEmail_FailedSendDoesNotConsumeRateSlot(t *testing.T) {
	s := newFakeSMTP(t)
	s.mailReply = func(n int) string {
		if n == 1 {
			return "550 rejected"
		}
		return ""
	}
	s.start()
	cfg := s.cfg()
	cfg.RateMax = 1
	cfg.RateInterval = time.Hour
	if err := SendEmail(context.Background(), cfg, "a@example.com", []string{"b@example.com"}, nil, nil, "s", "b", "", nil); err == nil {
		t.Fatal("expected first send to fail")
	}
	if err := SendEmail(context.Background(), cfg, "a@example.com", []string{"b@example.com"}, nil, nil, "s", "b", "", nil); err != nil {
		t.Fatalf("second send should use the returned slot: %v", err)
	}
	if err := SendEmail(context.Background(), cfg, "a@example.com", []string{"b@example.com"}, nil, nil, "s", "b", "", nil); !errors.Is(err, ErrEmailRateLimited) {
		t.Fatalf("expected rate limit, got %v", err)
	}
	// Farklı yapılandırma (başka kullanıcı) kendi penceresine sahiptir.
	other := cfg
	other.RateMax = 2
	if _, ok := reserveEmailSlot(other); !ok {
		t.Fatal("rate limiter must be per configuration")
	}
}

func TestSendEmail_UnreadableAttachmentFails(t *testing.T) {
	s := newFakeSMTP(t)
	s.start()
	missing := filepath.Join(t.TempDir(), "nope.pdf")
	err := SendEmail(context.Background(), s.cfg(), "a@example.com", []string{"b@example.com"}, nil, nil, "s", "b", "", []string{missing})
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected not-exist error, got %v", err)
	}
	if err := SendEmail(context.Background(), s.cfg(), "a@example.com", []string{"b@example.com"}, nil, nil, "s", "b", "", []string{"https://x/y.pdf"}); err == nil {
		t.Fatal("expected error for remote attachment")
	}
	if s.conns.Load() != 0 {
		t.Fatal("nothing should be sent when attachments fail")
	}
}

// --- FTP ---

func TestNewFTPClient_LoginFailureClosesConnection(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	closed := make(chan string, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		_ = c.SetDeadline(time.Now().Add(5 * time.Second))
		r := bufio.NewReader(c)
		fmt.Fprint(c, "220 fake ftp\r\n")
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				closed <- "eof"
				return
			}
			cmd := strings.ToUpper(strings.TrimSpace(line))
			switch {
			case strings.HasPrefix(cmd, "USER"):
				fmt.Fprint(c, "331 password please\r\n")
			case strings.HasPrefix(cmd, "PASS"):
				fmt.Fprint(c, "530 login incorrect\r\n")
			case cmd == "QUIT":
				fmt.Fprint(c, "221 bye\r\n")
				closed <- "quit"
				return
			default:
				fmt.Fprint(c, "502 no\r\n")
			}
		}
	}()
	_, err = NewFTPClientWithOptions(ln.Addr().String(), "u", "bad", FTPOptions{Timeout: 2 * time.Second})
	if err == nil {
		t.Fatal("expected login error")
	}
	select {
	case how := <-closed:
		if how != "quit" && how != "eof" {
			t.Fatalf("unexpected: %s", how)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("connection was not closed after failed login")
	}
}
