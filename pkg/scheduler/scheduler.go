package scheduler

import (
	"time"

	"github.com/robfig/cron/v3"
)

// Manager, cron tabanlı zamanlayıcı yöneticisi.
type Manager struct {
	c *cron.Cron
}

type options struct {
	withSeconds bool
	loc         *time.Location
}

type Option func(*options)

// WithSeconds cron ifadelerinde saniye alanını etkinleştirir.
func WithSeconds() Option { return func(o *options) { o.withSeconds = true } }

// WithLocation zamanlayıcıyı verilen lokasyonla başlatır.
func WithLocation(loc *time.Location) Option { return func(o *options) { o.loc = loc } }

// New yeni bir Manager döner.
func New(opts ...Option) *Manager {
	o := &options{loc: time.Local}
	for _, fn := range opts {
		fn(o)
	}

	copts := []cron.Option{cron.WithLocation(o.loc)}
	if o.withSeconds {
		copts = append(copts, cron.WithSeconds())
	}
	return &Manager{c: cron.New(copts...)}
}

// Start zamanlayıcıyı başlatır.
func (m *Manager) Start() { m.c.Start() }

// Stop zamanlayıcıyı durdurur ve kalan işleri beklemeden sonlandırır.
func (m *Manager) Stop() { m.c.Stop() }

// AddFunc cron ifadesiyle bir iş ekler.
func (m *Manager) AddFunc(spec string, cmd func()) (cron.EntryID, error) {
	return m.c.AddFunc(spec, cmd)
}

// AddJob cron ifadesiyle bir cron.Job ekler.
func (m *Manager) AddJob(spec string, job cron.Job) (cron.EntryID, error) {
	return m.c.AddJob(spec, job)
}

// Entries planlanmış işlerin listesini döner.
func (m *Manager) Entries() []cron.Entry { return m.c.Entries() }

// --- Varsayılan zamanlayıcı ---
var Default = New()

func Start()                                               { Default.Start() }
func Stop()                                                { Default.Stop() }
func AddFunc(spec string, f func()) (cron.EntryID, error)  { return Default.AddFunc(spec, f) }
func AddJob(spec string, j cron.Job) (cron.EntryID, error) { return Default.AddJob(spec, j) }
func Entries() []cron.Entry                                { return Default.Entries() }
