package migrate

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/mustafacaglarkara/webdev/pkg/sqlutil"
)

// ErrLockTimeout: advisory kilit verilen süre içinde alınamadı (başka bir koşucu çalışıyor).
var ErrLockTimeout = errors.New("migrate: kilit zaman aşımı (başka bir migration koşucusu çalışıyor olabilir)")

// WithLock: eşzamanlı koşuculara karşı veritabanı advisory kilidini açar/kapatır
// (varsayılan: açık). PostgreSQL pg_advisory_lock, MySQL GET_LOCK, SQL Server
// sp_getapplock kullanır; SQLite ve tanınmayan diyalektlerde kilit yoktur (no-op).
func WithLock(enabled bool) Option { return func(m *Migrator) { m.lock = enabled } }

// WithLockTimeout: kilidi beklemek için azami süre. 0 (varsayılan) ise context iptal
// edilene kadar beklenir. Süre dolarsa ErrLockTimeout döner.
func WithLockTimeout(d time.Duration) Option { return func(m *Migrator) { m.lockTimeout = d } }

// lockName: takip tablosuna bağlı kilit adı (MySQL 64, SQL Server 255 karakter sınırı).
func (m *Migrator) lockName() string {
	name := "webdev_migrate:" + m.table
	if len(name) > 64 {
		h := fnv.New64a()
		_, _ = h.Write([]byte(name))
		name = "webdev_migrate:" + hex.EncodeToString(h.Sum(nil))
	}
	return name
}

// lockKey: PostgreSQL için 64 bit sayısal anahtar.
func (m *Migrator) lockKey() int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(m.lockName()))
	return int64(h.Sum64()) //nolint:gosec // bilinçli taşma: yalnızca anahtar olarak kullanılır
}

// acquireLock: kilidi alır ve bırakma fonksiyonunu döner. Kilit ayrı bir bağlantı
// (sql.Conn) üzerinde oturum düzeyinde tutulur; migration'lar havuzdaki diğer
// bağlantılarla çalışabilir. Kilit desteklenmiyorsa no-op döner.
func (m *Migrator) acquireLock(ctx context.Context) (release func() error, err error) {
	noop := func() error { return nil }
	if !m.lock {
		return noop, nil
	}
	switch m.dialect {
	case sqlutil.Postgres, sqlutil.MySQL, sqlutil.SQLServer:
	default:
		return noop, nil
	}
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("migrate: kilit bağlantısı açılamadı: %w", err)
	}
	waitCtx := ctx
	if m.lockTimeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, m.lockTimeout)
		defer cancel()
	}
	if err := m.lockOn(waitCtx, conn); err != nil {
		_ = conn.Close()
		if errors.Is(err, context.DeadlineExceeded) && m.lockTimeout > 0 && ctx.Err() == nil {
			return nil, ErrLockTimeout
		}
		return nil, err
	}
	return func() error {
		// Bırakma, çağıranın context'i iptal edilmiş olsa da denenir.
		rctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		uerr := m.unlockOn(rctx, conn)
		if cerr := conn.Close(); uerr == nil {
			uerr = cerr
		}
		return uerr
	}, nil
}

func (m *Migrator) lockOn(ctx context.Context, conn *sql.Conn) error {
	switch m.dialect {
	case sqlutil.Postgres:
		if m.lockTimeout <= 0 {
			_, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", m.lockKey())
			if err != nil {
				return fmt.Errorf("migrate: pg_advisory_lock: %w", err)
			}
			return nil
		}
		// Süreli bekleme: try_lock ile kısa aralıklarla dene.
		for {
			var got bool
			if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", m.lockKey()).Scan(&got); err != nil {
				return fmt.Errorf("migrate: pg_try_advisory_lock: %w", err)
			}
			if got {
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(200 * time.Millisecond):
			}
		}
	case sqlutil.MySQL:
		timeout := int64(-1) // sonsuz
		if m.lockTimeout > 0 {
			timeout = int64(m.lockTimeout / time.Second)
			if timeout < 1 {
				timeout = 1
			}
		}
		var got sql.NullInt64
		if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", m.lockName(), timeout).Scan(&got); err != nil {
			return fmt.Errorf("migrate: GET_LOCK: %w", err)
		}
		if !got.Valid {
			return errors.New("migrate: GET_LOCK hata döndü (NULL)")
		}
		if got.Int64 != 1 {
			return ErrLockTimeout
		}
		return nil
	case sqlutil.SQLServer:
		timeout := int64(-1)
		if m.lockTimeout > 0 {
			timeout = int64(m.lockTimeout / time.Millisecond)
		}
		var rc int64
		q := "DECLARE @r INT; EXEC @r = sp_getapplock @Resource = @p1, @LockMode = 'Exclusive', @LockOwner = 'Session', @LockTimeout = @p2; SELECT @r"
		if err := conn.QueryRowContext(ctx, q, m.lockName(), timeout).Scan(&rc); err != nil {
			return fmt.Errorf("migrate: sp_getapplock: %w", err)
		}
		if rc == -1 {
			return ErrLockTimeout
		}
		if rc < 0 {
			return fmt.Errorf("migrate: sp_getapplock başarısız (kod %d)", rc)
		}
		return nil
	}
	return nil
}

func (m *Migrator) unlockOn(ctx context.Context, conn *sql.Conn) error {
	var err error
	switch m.dialect {
	case sqlutil.Postgres:
		_, err = conn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", m.lockKey())
	case sqlutil.MySQL:
		_, err = conn.ExecContext(ctx, "SELECT RELEASE_LOCK(?)", m.lockName())
	case sqlutil.SQLServer:
		_, err = conn.ExecContext(ctx, "EXEC sp_releaseapplock @Resource = @p1, @LockOwner = 'Session'", m.lockName())
	}
	if err != nil {
		return fmt.Errorf("migrate: kilit bırakılamadı: %w", err)
	}
	return nil
}
