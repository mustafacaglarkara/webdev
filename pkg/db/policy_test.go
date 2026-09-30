package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"

	"github.com/mustafacaglarkara/webdev/pkg/resilience"
)

type fakePgErr struct{ code string }

func (e fakePgErr) Error() string    { return "pg " + e.code }
func (e fakePgErr) SQLState() string { return e.code }

type fakeMssqlErr struct{ n int32 }

func (e fakeMssqlErr) Error() string         { return fmt.Sprintf("mssql %d", e.n) }
func (e fakeMssqlErr) SQLErrorNumber() int32 { return e.n }

func TestErrorClassification(t *testing.T) {
	transient := []error{
		driver.ErrBadConn,
		fmt.Errorf("wrap: %w", syscall.ECONNRESET),
		fakePgErr{"40001"}, fakePgErr{"40P01"}, fakePgErr{"08006"},
		fakeMssqlErr{1205},
		&mysql.MySQLError{Number: 1213},
		sqlite3.Error{Code: sqlite3.ErrBusy},
	}
	for _, e := range transient {
		if !IsTransient(e) {
			t.Errorf("geçici olmalı: %v", e)
		}
	}
	notTransient := []error{
		context.Canceled, context.DeadlineExceeded, sql.ErrNoRows, gorm.ErrRecordNotFound,
		errors.New("syntax error"), fakePgErr{"23505"}, fakePgErr{"42601"},
		fakeMssqlErr{2627}, &mysql.MySQLError{Number: 1062}, sqlite3.Error{Code: sqlite3.ErrConstraint},
	}
	for _, e := range notTransient {
		if IsTransient(e) {
			t.Errorf("geçici OLMAMALI: %v", e)
		}
	}
	constraint := []error{fakePgErr{"23505"}, fakeMssqlErr{547}, &mysql.MySQLError{Number: 1452}, sqlite3.Error{Code: sqlite3.ErrConstraint}, gorm.ErrDuplicatedKey}
	for _, e := range constraint {
		if !IsConstraintViolation(e) {
			t.Errorf("kısıt ihlali olmalı: %v", e)
		}
		if isBreakerFailure(e) {
			t.Errorf("kısıt ihlali breaker hatası sayılmamalı: %v", e)
		}
	}
	for _, e := range []error{context.Canceled, sql.ErrNoRows, fmt.Errorf("x: %w", gorm.ErrRecordNotFound)} {
		if isBreakerFailure(e) {
			t.Errorf("breaker hatası sayılmamalı: %v", e)
		}
	}
	if !IsTimeout(fakePgErr{"57014"}) || IsTimeout(context.DeadlineExceeded) {
		t.Error("IsTimeout sınıflandırması hatalı")
	}
}

func newTestRuntime(cfg Config) *runtime {
	rt := &runtime{cfg: cfg}
	if cfg.EnableBreaker {
		rt.cb = resilience.NewCircuitBreaker(cfg.BreakerFailThreshold, cfg.BreakerOpenTimeout, resilience.WithFailurePredicate(isBreakerFailure))
	}
	return rt
}

// DB-2: yalnızca geçici hatalar; yazmalar varsayılan olarak yeniden denenmez.
func TestDoWithPolicies_Retry(t *testing.T) {
	ctx := context.Background()
	rt := newTestRuntime(Config{RetryAttempts: 3, RetryDelay: time.Millisecond})

	calls := 0
	flaky := func() error {
		calls++
		if calls < 3 {
			return driver.ErrBadConn
		}
		return nil
	}
	if err := rt.doWithPolicies(ctx, opRead, flaky); err != nil || calls != 3 {
		t.Fatalf("okuma 3 kez denenmeli: calls=%d err=%v", calls, err)
	}

	calls = 0
	if err := rt.doWithPolicies(ctx, opWrite, flaky); !errors.Is(err, driver.ErrBadConn) || calls != 1 {
		t.Fatalf("yazma varsayılan olarak denenmemeli: calls=%d err=%v", calls, err)
	}
	calls = 0
	if err := rt.doWithPolicies(ctx, opTx, flaky); calls != 1 || err == nil {
		t.Fatalf("tx varsayılan olarak denenmemeli: calls=%d", calls)
	}

	calls = 0
	if err := rt.doWithPolicies(WithWriteRetry(ctx, true), opWrite, flaky); err != nil || calls != 3 {
		t.Fatalf("opt-in yazma denenmeli: calls=%d err=%v", calls, err)
	}
	rtW := newTestRuntime(Config{RetryAttempts: 3, RetryDelay: time.Millisecond, RetryWrites: true})
	calls = 0
	if err := rtW.doWithPolicies(ctx, opTx, flaky); err != nil || calls != 3 {
		t.Fatalf("RetryWrites ile tx denenmeli: calls=%d err=%v", calls, err)
	}
	calls = 0
	if err := rtW.doWithPolicies(WithWriteRetry(ctx, false), opWrite, flaky); calls != 1 || err == nil {
		t.Fatalf("ctx ile kapatılan yazma denenmemeli: calls=%d", calls)
	}

	for _, perm := range []error{errors.New("syntax"), sql.ErrNoRows, sqlite3.Error{Code: sqlite3.ErrConstraint}, context.Canceled} {
		calls = 0
		err := rt.doWithPolicies(ctx, opRead, func() error { calls++; return perm })
		if calls != 1 || !errors.Is(err, perm) {
			t.Errorf("%v yeniden denenmemeli: calls=%d err=%v", perm, calls, err)
		}
	}

	// zaman aşımı yalnızca okumada denenir
	calls = 0
	_ = rt.doWithPolicies(ctx, opRead, func() error { calls++; return fakePgErr{"57014"} })
	if calls != 3 {
		t.Errorf("okuma zaman aşımı denenmeli: %d", calls)
	}
	calls = 0
	_ = rtW.doWithPolicies(ctx, opWrite, func() error { calls++; return fakePgErr{"57014"} })
	if calls != 1 {
		t.Errorf("yazma zaman aşımı denenmemeli: %d", calls)
	}
}

// DB-3: breaker; ErrNoRows, kısıt ihlali, context iptali hata sayılmaz.
func TestDoWithPolicies_BreakerIgnoresNonFailures(t *testing.T) {
	ctx := context.Background()
	rt := newTestRuntime(Config{EnableBreaker: true, BreakerFailThreshold: 2, BreakerOpenTimeout: time.Hour})
	for i := 0; i < 5; i++ {
		_ = rt.doWithPolicies(ctx, opRead, func() error { return sql.ErrNoRows })
		_ = rt.doWithPolicies(ctx, opWrite, func() error { return sqlite3.Error{Code: sqlite3.ErrConstraint} })
		_ = rt.doWithPolicies(ctx, opRead, func() error { return context.Canceled })
		cctx, cancel := context.WithCancel(ctx)
		cancel()
		_ = rt.doWithPolicies(cctx, opRead, func() error { return errors.New("çağrılmamalı") })
	}
	if st := rt.cb.State(); st != "closed" {
		t.Fatalf("breaker kapalı kalmalı, durum=%s", st)
	}
	for i := 0; i < 2; i++ {
		_ = rt.doWithPolicies(ctx, opRead, func() error { return errors.New("boom") })
	}
	if err := rt.doWithPolicies(ctx, opRead, func() error { return nil }); !errors.Is(err, resilience.ErrBreakerOpen) {
		t.Fatalf("gerçek hatalardan sonra breaker açılmalı: %v", err)
	}
}

// Breaker + retry eşzamanlı kullanımda yarışsız olmalı.
func TestDoWithPolicies_Concurrent(t *testing.T) {
	rt := newTestRuntime(Config{EnableBreaker: true, BreakerFailThreshold: 1000, RetryAttempts: 2})
	var n atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = rt.doWithPolicies(context.Background(), opRead, func() error {
					n.Add(1)
					if (i+j)%7 == 0 {
						return errors.New("x")
					}
					return nil
				})
			}
		}(i)
	}
	wg.Wait()
	if n.Load() < 32*50 {
		t.Fatalf("çağrı sayısı %d", n.Load())
	}
}
