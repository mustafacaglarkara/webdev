package fs

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jordan-wright/email"
)

type SMTPCfg struct {
	Host          string
	Port          int
	User          string
	Pass          string
	UseTLS        bool
	Timeout       time.Duration
	RateInterval  time.Duration
	RateMax       int
	RetryAttempts int
	RetryDelay    time.Duration
}

var (
	emailRateMu          sync.Mutex
	emailRateWindowStart time.Time
	emailRateCount       int
)

func allowEmailSend(cfg SMTPCfg) bool {
	if cfg.RateMax <= 0 || cfg.RateInterval <= 0 {
		return true
	}
	now := time.Now()
	emailRateMu.Lock()
	defer emailRateMu.Unlock()
	if emailRateWindowStart.IsZero() || now.Sub(emailRateWindowStart) > cfg.RateInterval {
		emailRateWindowStart = now
		emailRateCount = 0
	}
	if emailRateCount >= cfg.RateMax {
		return false
	}
	emailRateCount++
	return true
}

func isTransientSMTPError(err error) bool {
	if err == nil {
		return false
	}
	var nerr net.Error
	if errors.As(err, &nerr) {
		if nerr.Timeout() || nerr.Temporary() {
			return true
		}
	}
	es := err.Error()
	if strings.Contains(es, " 421") || strings.Contains(es, " 450") || strings.Contains(es, " 451") || strings.Contains(es, " 452") {
		return true
	}
	if strings.Contains(strings.ToLower(es), "timeout") {
		return true
	}
	return false
}

func SendEmail(ctx context.Context, cfg SMTPCfg, from string, to, cc, bcc []string, subject string, textBody, htmlBody string, attachments []string) error {
	if !allowEmailSend(cfg) {
		return errors.New("email rate limit exceeded")
	}
	e := email.NewEmail()
	e.From = from
	e.To = to
	if len(cc) > 0 {
		e.Cc = cc
	}
	if len(bcc) > 0 {
		e.Bcc = bcc
	}
	e.Subject = subject
	if htmlBody != "" {
		e.HTML = []byte(htmlBody)
	}
	if textBody != "" {
		e.Text = []byte(textBody)
	}
	for _, a := range attachments {
		if strings.HasPrefix(a, "http://") || strings.HasPrefix(a, "https://") {
			continue
		}
		if fi, err := os.Stat(a); err == nil && !fi.IsDir() {
			if _, err := e.AttachFile(a); err != nil {
				return err
			}
		}
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	attempts := cfg.RetryAttempts
	if attempts <= 1 {
		attempts = 1
	}
	delay := cfg.RetryDelay
	if delay <= 0 {
		delay = 2 * time.Second
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		attemptCtx := ctx
		cancel := func() {}
		if cfg.Timeout > 0 {
			attemptCtx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		}
		done := make(chan error, 1)
		go func() {
			if cfg.UseTLS {
				done <- e.SendWithTLS(addr, auth, &tls.Config{ServerName: cfg.Host})
			} else {
				done <- e.Send(addr, auth)
			}
		}()
		select {
		case <-attemptCtx.Done():
			lastErr = attemptCtx.Err()
			cancel()
		case err := <-done:
			if err == nil {
				cancel()
				return nil
			}
			lastErr = err
			cancel()
			if !isTransientSMTPError(err) {
				return err
			}
		}
		if i < attempts-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return lastErr
}

// Helper for joining attachment paths (ensures dir exists)
func EnsureAttachmentDir(path string) error { return os.MkdirAll(filepath.Dir(path), 0o755) }
