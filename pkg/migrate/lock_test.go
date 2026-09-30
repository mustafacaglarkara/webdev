package migrate

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/mustafacaglarkara/webdev/pkg/sqlutil"
)

// ---- Sorguları kaydeden sahte sürücü (gerçek sunucu olmadan kilit SQL'ini doğrular) ----

type fakeDriver struct {
	mu     sync.Mutex
	log    []string
	fail   string // bu metni içeren Exec hata döner
	nolock bool   // GET_LOCK 0, sp_getapplock -1 döndür (zaman aşımı)
}

func (d *fakeDriver) record(q string) {
	d.mu.Lock()
	d.log = append(d.log, strings.TrimSpace(q))
	d.mu.Unlock()
}

func (d *fakeDriver) statements() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]string(nil), d.log...)
}

func (d *fakeDriver) reset() { d.mu.Lock(); d.log = nil; d.mu.Unlock() }

func (d *fakeDriver) Open(string) (driver.Conn, error) { return &fakeConn{d: d}, nil }

type fakeConn struct{ d *fakeDriver }

func (c *fakeConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare unsupported")
}
func (c *fakeConn) Close() error              { return nil }
func (c *fakeConn) Begin() (driver.Tx, error) { return fakeTx{}, nil }
func (c *fakeConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return fakeTx{}, nil
}

func (c *fakeConn) ExecContext(_ context.Context, q string, _ []driver.NamedValue) (driver.Result, error) {
	c.d.record(q)
	if c.d.fail != "" && strings.Contains(q, c.d.fail) {
		return nil, errors.New("boom")
	}
	return driver.RowsAffected(1), nil
}

func (c *fakeConn) QueryContext(_ context.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
	c.d.record(q)
	switch {
	case strings.Contains(q, "GET_LOCK"):
		v := int64(1)
		if c.d.nolock {
			v = 0
		}
		return &fakeRows{cols: []string{"GET_LOCK"}, vals: [][]driver.Value{{v}}}, nil
	case strings.Contains(q, "sp_getapplock"):
		v := int64(0)
		if c.d.nolock {
			v = -1
		}
		return &fakeRows{cols: []string{"r"}, vals: [][]driver.Value{{v}}}, nil
	case strings.Contains(q, "pg_try_advisory_lock"):
		return &fakeRows{cols: []string{"ok"}, vals: [][]driver.Value{{!c.d.nolock}}}, nil
	}
	return &fakeRows{cols: []string{"version", "name", "applied_at"}}, nil
}

type fakeTx struct{}

func (fakeTx) Commit() error   { return nil }
func (fakeTx) Rollback() error { return nil }

type fakeRows struct {
	cols []string
	vals [][]driver.Value
	i    int
}

func (r *fakeRows) Columns() []string { return r.cols }
func (r *fakeRows) Close() error      { return nil }
func (r *fakeRows) Next(dest []driver.Value) error {
	if r.i >= len(r.vals) {
		return io.EOF
	}
	copy(dest, r.vals[r.i])
	r.i++
	return nil
}

var fake = &fakeDriver{}

func init() { sql.Register("fakemig", fake) }

func openFake(t *testing.T) *sql.DB {
	t.Helper()
	fake.reset()
	fake.fail = ""
	fake.nolock = false
	db, err := sql.Open("fakemig", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func has(list []string, sub string) int {
	for i, s := range list {
		if strings.Contains(s, sub) {
			return i
		}
	}
	return -1
}

// I5-1: her diyalekt için kilit alınır, migration'lar arada çalışır, sonunda bırakılır.
func TestLock_PerDialectSQL(t *testing.T) {
	cases := []struct {
		d            sqlutil.Dialect
		lock, unlock string
	}{
		{sqlutil.Postgres, "pg_advisory_lock($1)", "pg_advisory_unlock($1)"},
		{sqlutil.MySQL, "GET_LOCK(?, ?)", "RELEASE_LOCK(?)"},
		{sqlutil.SQLServer, "sp_getapplock", "sp_releaseapplock"},
	}
	fsys := fstest.MapFS{"1_a.up.sql": {Data: []byte("CREATE TABLE a (id INT);")}}
	for _, c := range cases {
		t.Run(string(c.d), func(t *testing.T) {
			db := openFake(t)
			m := New(db, fsys, ".", WithDialect(c.d))
			ran, err := m.Up(context.Background())
			if err != nil || len(ran) != 1 {
				t.Fatalf("up: %v %v", ran, err)
			}
			st := fake.statements()
			li, ci, mi, ui := has(st, c.lock), has(st, "CREATE TABLE"), has(st, "CREATE TABLE a"), has(st, c.unlock)
			if li < 0 || ci < 0 || mi < 0 || ui < 0 {
				t.Fatalf("eksik ifade: %v", st)
			}
			if !(li < ci && ci < mi && mi < ui) || ui != len(st)-1 {
				t.Fatalf("sıra yanlış (lock < ensure < migration < unlock): %v", st)
			}
			// Down da kilitler.
			fake.reset()
			if _, err := m.DownAll(context.Background()); err != nil {
				t.Fatal(err)
			}
			st = fake.statements()
			if has(st, c.lock) != 0 || has(st, c.unlock) != len(st)-1 {
				t.Fatalf("down kilitlemedi: %v", st)
			}
		})
	}
}

// Migration hata verse de kilit bırakılır.
func TestLock_ReleasedOnError(t *testing.T) {
	db := openFake(t)
	fake.fail = "BOOM"
	fsys := fstest.MapFS{"1_a.up.sql": {Data: []byte("BOOM;")}}
	_, err := New(db, fsys, ".", WithDialect(sqlutil.Postgres)).Up(context.Background())
	if err == nil {
		t.Fatal("hata bekleniyordu")
	}
	st := fake.statements()
	if len(st) == 0 || !strings.Contains(st[len(st)-1], "pg_advisory_unlock") {
		t.Fatalf("kilit bırakılmadı: %v", st)
	}
}

// Kilit alınamazsa (zaman aşımı) ErrLockTimeout döner ve migration çalışmaz.
func TestLock_Timeout(t *testing.T) {
	fsys := fstest.MapFS{"1_a.up.sql": {Data: []byte("CREATE TABLE a (id INT);")}}
	for _, d := range []sqlutil.Dialect{sqlutil.MySQL, sqlutil.SQLServer, sqlutil.Postgres} {
		t.Run(string(d), func(t *testing.T) {
			db := openFake(t)
			fake.nolock = true
			m := New(db, fsys, ".", WithDialect(d), WithLockTimeout(300*time.Millisecond))
			_, err := m.Up(context.Background())
			if !errors.Is(err, ErrLockTimeout) {
				t.Fatalf("ErrLockTimeout bekleniyordu: %v", err)
			}
			if has(fake.statements(), "CREATE TABLE") >= 0 {
				t.Fatalf("kilitsiz migration çalıştı: %v", fake.statements())
			}
		})
	}
	// Postgres süreli bekleme try_lock kullanır; süresiz bekleme blocking lock.
	db := openFake(t)
	if _, err := New(db, fsys, ".", WithDialect(sqlutil.Postgres), WithLockTimeout(time.Second)).Up(context.Background()); err != nil {
		t.Fatal(err)
	}
	if has(fake.statements(), "pg_try_advisory_lock") < 0 {
		t.Fatalf("try_lock bekleniyordu: %v", fake.statements())
	}
}

// WithLock(false) ve SQLite: kilit ifadesi yok.
func TestLock_DisabledAndSQLiteNoop(t *testing.T) {
	fsys := fstest.MapFS{"1_a.up.sql": {Data: []byte("CREATE TABLE a (id INT);")}}
	db := openFake(t)
	if _, err := New(db, fsys, ".", WithDialect(sqlutil.Postgres), WithLock(false)).Up(context.Background()); err != nil {
		t.Fatal(err)
	}
	if has(fake.statements(), "advisory") >= 0 {
		t.Fatalf("kilit kapalıyken ifade var: %v", fake.statements())
	}
	// Gerçek SQLite: no-op, Up çalışır.
	sq := openDB(t)
	if _, err := New(sq, fsys, ".").Up(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !tableExists(t, sq, "a") {
		t.Fatal("sqlite migration çalışmadı")
	}
	// Kilit adı MySQL sınırına sığar.
	m := New(sq, fsys, ".", WithTable(strings.Repeat("t", 60)))
	if n := m.lockName(); len(n) > 64 {
		t.Fatalf("kilit adı çok uzun: %d", len(n))
	}
}
