// Package scheduler robfig/cron üzerinde panik korumalı, loglu bir zamanlayıcı sarmalayıcısıdır.
package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

// Manager, cron tabanlı zamanlayıcı yöneticisi.
//
// Tüm işler cron.Recover ile sarılır: bir işteki panik loglanır, zamanlayıcıyı
// ve diğer işleri durdurmaz.
type Manager struct {
	c *cron.Cron
}

type options struct {
	withSeconds  bool
	loc          *time.Location
	logger       cron.Logger
	skipRunning  bool
	delayRunning bool
}

type Option func(*options)

// WithSeconds cron ifadelerinde saniye alanını etkinleştirir.
func WithSeconds() Option { return func(o *options) { o.withSeconds = true } }

// WithLocation zamanlayıcıyı verilen lokasyonla başlatır.
func WithLocation(loc *time.Location) Option { return func(o *options) { o.loc = loc } }

// WithLogger cron.Logger kullanır (panikler ve atlanan çalışmalar buraya loglanır).
func WithLogger(l cron.Logger) Option { return func(o *options) { o.logger = l } }

// WithSlog slog.Logger kullanır. cron'un bilgi mesajları Debug, hatalar Error seviyesinde yazılır.
func WithSlog(l *slog.Logger) Option {
	return func(o *options) {
		if l != nil {
			o.logger = SlogLogger(l)
		}
	}
}

// WithSkipIfStillRunning bir işin önceki çalışması bitmemişse yeni tetiklemeyi atlar.
func WithSkipIfStillRunning() Option { return func(o *options) { o.skipRunning = true } }

// WithDelayIfStillRunning bir işin önceki çalışması bitene kadar yeni tetiklemeyi geciktirir.
func WithDelayIfStillRunning() Option { return func(o *options) { o.delayRunning = true } }

// slogAdapter cron.Logger arayüzünü slog'a uyarlar.
type slogAdapter struct{ l *slog.Logger }

// SlogLogger bir slog.Logger'ı cron.Logger'a uyarlar.
func SlogLogger(l *slog.Logger) cron.Logger { return slogAdapter{l: l} }

func (a slogAdapter) Info(msg string, keysAndValues ...interface{}) {
	a.l.Debug("scheduler: "+msg, keysAndValues...)
}
func (a slogAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	a.l.Error("scheduler: "+msg, append([]any{"err", err}, keysAndValues...)...)
}

// New yeni bir Manager döner. Varsayılan logger slog.Default()'tur.
func New(opts ...Option) *Manager {
	o := &options{loc: time.Local}
	for _, fn := range opts {
		if fn != nil {
			fn(o)
		}
	}
	if o.loc == nil {
		o.loc = time.Local
	}
	if o.logger == nil {
		o.logger = SlogLogger(slog.Default())
	}
	wrappers := []cron.JobWrapper{cron.Recover(o.logger)}
	switch {
	case o.skipRunning:
		wrappers = append(wrappers, cron.SkipIfStillRunning(o.logger))
	case o.delayRunning:
		wrappers = append(wrappers, cron.DelayIfStillRunning(o.logger))
	}
	copts := []cron.Option{
		cron.WithLocation(o.loc),
		cron.WithLogger(o.logger),
		cron.WithChain(wrappers...),
	}
	if o.withSeconds {
		copts = append(copts, cron.WithSeconds())
	}
	return &Manager{c: cron.New(copts...)}
}

// Start zamanlayıcıyı arka planda başlatır (zaten çalışıyorsa bir şey yapmaz).
func (m *Manager) Start() { m.c.Start() }

// Stop yeni tetiklemeleri durdurur ve çalışmakta olan işler bittiğinde
// kapanan (Done) bir context döner. Çalışan işler kesilmez; beklemek için:
//
//	ctx := m.Stop()
//	select { case <-ctx.Done(): case <-time.After(10 * time.Second): }
func (m *Manager) Stop() context.Context { return m.c.Stop() }

// Shutdown Stop'u çağırır ve çalışan işlerin bitmesini ctx süresince bekler.
func (m *Manager) Shutdown(ctx context.Context) error {
	done := m.c.Stop()
	select {
	case <-done.Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// AddFunc cron ifadesiyle bir iş ekler.
func (m *Manager) AddFunc(spec string, cmd func()) (cron.EntryID, error) {
	return m.c.AddFunc(spec, cmd)
}

// AddJob cron ifadesiyle bir cron.Job ekler.
func (m *Manager) AddJob(spec string, job cron.Job) (cron.EntryID, error) {
	return m.c.AddJob(spec, job)
}

// Remove bir işi kaldırır.
func (m *Manager) Remove(id cron.EntryID) { m.c.Remove(id) }

// Entries planlanmış işlerin listesini döner.
func (m *Manager) Entries() []cron.Entry { return m.c.Entries() }

// --- Varsayılan zamanlayıcı ---
var Default = New()

func Start()                                               { Default.Start() }
func Stop() context.Context                                { return Default.Stop() }
func AddFunc(spec string, f func()) (cron.EntryID, error)  { return Default.AddFunc(spec, f) }
func AddJob(spec string, j cron.Job) (cron.EntryID, error) { return Default.AddJob(spec, j) }
func Entries() []cron.Entry                                { return Default.Entries() }
