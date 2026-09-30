package fs

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jordan-wright/email"
)

// DefaultSMTPTimeout SMTPCfg.Timeout verilmediğinde her deneme için kullanılan süre sınırıdır.
const DefaultSMTPTimeout = 30 * time.Second

type SMTPCfg struct {
	Host string
	Port int
	User string
	Pass string
	// UseTLS true ise doğrudan TLS (implicit, genelde 465) kullanılır; false ise
	// sunucu destekliyorsa STARTTLS ile yükseltilir.
	UseTLS bool
	// Timeout her denemenin (bağlantı + gönderim) süre sınırıdır; 0 ise DefaultSMTPTimeout.
	Timeout time.Duration
	// RateInterval / RateMax: aynı yapılandırma (host, port, user) için
	// RateInterval içinde en fazla RateMax başarılı gönderim.
	RateInterval  time.Duration
	RateMax       int
	RetryAttempts int
	RetryDelay    time.Duration
}

// ErrEmailRateLimited gönderim hız sınırı aşıldığında döner.
var ErrEmailRateLimited = errors.New("email rate limit exceeded")

// --- yapılandırma başına hız sınırı ---

type rateWindow struct {
	mu    sync.Mutex
	start time.Time
	count int
}

var emailRates sync.Map // map[string]*rateWindow

func rateKey(cfg SMTPCfg) string {
	return cfg.Host + ":" + strconv.Itoa(cfg.Port) + "|" + cfg.User + "|" +
		cfg.RateInterval.String() + "|" + strconv.Itoa(cfg.RateMax)
}

// reserveEmailSlot bir gönderim hakkı ayırır; dönen release fonksiyonu
// gönderim başarısız olursa çağrılır ve hakkı geri verir.
func reserveEmailSlot(cfg SMTPCfg) (release func(), ok bool) {
	if cfg.RateMax <= 0 || cfg.RateInterval <= 0 {
		return func() {}, true
	}
	v, _ := emailRates.LoadOrStore(rateKey(cfg), &rateWindow{})
	w := v.(*rateWindow)
	now := time.Now()
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.start.IsZero() || now.Sub(w.start) > cfg.RateInterval {
		w.start = now
		w.count = 0
	}
	if w.count >= cfg.RateMax {
		return nil, false
	}
	w.count++
	start := w.start
	return func() {
		w.mu.Lock()
		defer w.mu.Unlock()
		if w.start.Equal(start) && w.count > 0 {
			w.count--
		}
	}, true
}

// isTransientSMTPError geçici SMTP hatalarını (4xx: 421, 450, 451, 452) ve ağ
// zaman aşımlarını tanır. Kod, yanıtın başında eşleştirilir; mesajın içinde
// geçen "421" gibi sayılar dikkate alınmaz.
func isTransientSMTPError(err error) bool {
	if err == nil {
		return false
	}
	var tpErr *textproto.Error
	if errors.As(err, &tpErr) {
		return isTransientSMTPCode(tpErr.Code)
	}
	var nerr net.Error
	if errors.As(err, &nerr) && nerr.Timeout() {
		return true
	}
	es := strings.TrimSpace(err.Error())
	if len(es) >= 3 {
		if code, cerr := strconv.Atoi(es[:3]); cerr == nil && (len(es) == 3 || es[3] == ' ' || es[3] == '-') {
			return isTransientSMTPCode(code)
		}
	}
	return false
}

func isTransientSMTPCode(code int) bool {
	switch code {
	case 421, 450, 451, 452:
		return true
	}
	return false
}

// smtpSendError gönderimin hangi aşamada başarısız olduğunu taşır.
// ambiguous=true ise mesaj sunucuya iletilmiş olabilir; yeniden denenmez.
type smtpSendError struct {
	err       error
	ambiguous bool
}

func (e *smtpSendError) Error() string { return e.err.Error() }
func (e *smtpSendError) Unwrap() error { return e.err }

// SendEmail e-posta gönderir.
//   - Her denemenin süre sınırı cfg.Timeout'tur (0 ise DefaultSMTPTimeout); süre
//     bağlantıya uygulanır, arka planda çalışmaya devam eden goroutine kalmaz.
//   - Yalnızca geçici hatalar (421/450/451/452, ağ zaman aşımı) yeniden denenir.
//     Mesaj gövdesi gönderilmeye başladıktan sonra oluşan hatalar yeniden
//     denenmez (çift gönderimi önlemek için).
//   - Okunamayan veya uzak (http/https) ekler hata döndürür.
//   - Başarısız gönderim hız sınırı hakkı tüketmez.
func SendEmail(ctx context.Context, cfg SMTPCfg, from string, to, cc, bcc []string, subject string, textBody, htmlBody string, attachments []string) error {
	if ctx == nil {
		ctx = context.Background()
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
			return fmt.Errorf("fs: remote attachments are not supported: %s", a)
		}
		fi, err := os.Stat(a)
		if err != nil {
			return fmt.Errorf("fs: attachment %s: %w", a, err)
		}
		if fi.IsDir() {
			return fmt.Errorf("fs: attachment %s is a directory", a)
		}
		if _, err := e.AttachFile(a); err != nil {
			return fmt.Errorf("fs: attachment %s: %w", a, err)
		}
	}
	sender, rcpts, err := envelope(e)
	if err != nil {
		return err
	}
	raw, err := e.Bytes()
	if err != nil {
		return err
	}

	release, ok := reserveEmailSlot(cfg)
	if !ok {
		return ErrEmailRateLimited
	}
	sent := false
	defer func() {
		if !sent {
			release()
		}
	}()

	attempts := cfg.RetryAttempts
	if attempts <= 1 {
		attempts = 1
	}
	delay := cfg.RetryDelay
	if delay <= 0 {
		delay = 2 * time.Second
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = DefaultSMTPTimeout
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			if lastErr != nil {
				return fmt.Errorf("%w (last error: %v)", err, lastErr)
			}
			return err
		}
		err := sendOnce(ctx, cfg, timeout, sender, rcpts, raw)
		if err == nil {
			sent = true
			return nil
		}
		lastErr = err
		var se *smtpSendError
		if errors.As(err, &se) && se.ambiguous {
			return err
		}
		if ctx.Err() != nil || !isTransientSMTPError(err) {
			return err
		}
		if i < attempts-1 {
			t := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				t.Stop()
				return fmt.Errorf("%w (last error: %v)", ctx.Err(), lastErr)
			case <-t.C:
			}
		}
	}
	return lastErr
}

func envelope(e *email.Email) (string, []string, error) {
	all := make([]string, 0, len(e.To)+len(e.Cc)+len(e.Bcc))
	all = append(append(append(all, e.To...), e.Cc...), e.Bcc...)
	rcpts := make([]string, 0, len(all))
	for _, a := range all {
		addr, err := mail.ParseAddress(a)
		if err != nil {
			return "", nil, err
		}
		rcpts = append(rcpts, addr.Address)
	}
	if e.From == "" || len(rcpts) == 0 {
		return "", nil, errors.New("must specify at least one From address and one To address")
	}
	from, err := mail.ParseAddress(e.From)
	if err != nil {
		return "", nil, err
	}
	return from.Address, rcpts, nil
}

// sendOnce tek bir SMTP oturumu açar. Bağlantıya mutlak bir son tarih
// (deadline) konur ve ctx iptali bağlantıyı hemen keser; böylece hiçbir
// goroutine arka planda asılı kalmaz.
func sendOnce(ctx context.Context, cfg SMTPCfg, timeout time.Duration, from string, to []string, msg []byte) (err error) {
	attemptCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	tlsCfg := &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12}

	var conn net.Conn
	d := &net.Dialer{}
	if cfg.UseTLS {
		conn, err = (&tls.Dialer{NetDialer: d, Config: tlsCfg}).DialContext(attemptCtx, "tcp", addr)
	} else {
		conn, err = d.DialContext(attemptCtx, "tcp", addr)
	}
	if err != nil {
		return err
	}
	deadline, _ := attemptCtx.Deadline()
	_ = conn.SetDeadline(deadline)
	stop := context.AfterFunc(attemptCtx, func() { _ = conn.SetDeadline(time.Unix(1, 0)) })
	defer stop()

	c, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()
	if err := c.Hello("localhost"); err != nil {
		return err
	}
	if !cfg.UseTLS {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(tlsCfg); err != nil {
				return err
			}
		}
	}
	if cfg.User != "" {
		if ok, _ := c.Extension("AUTH"); !ok {
			return errors.New("smtp: server doesn't support AUTH")
		}
		if err := c.Auth(smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)); err != nil {
			return err
		}
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	for _, r := range to {
		if err := c.Rcpt(r); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	// Bu noktadan sonra sunucu mesajı almış olabilir: hata belirsizdir.
	if _, err := w.Write(msg); err != nil {
		return &smtpSendError{err: err, ambiguous: true}
	}
	if err := w.Close(); err != nil {
		var tpErr *textproto.Error
		if errors.As(err, &tpErr) {
			// Sunucu açık bir ret kodu döndürdü; mesaj kabul edilmedi.
			return err
		}
		return &smtpSendError{err: err, ambiguous: true}
	}
	_ = c.Quit()
	return nil
}

// EnsureAttachmentDir ek dosyası yolu için üst dizini oluşturur.
func EnsureAttachmentDir(path string) error { return os.MkdirAll(filepath.Dir(path), 0o755) }
