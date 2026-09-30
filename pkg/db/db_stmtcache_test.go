package db

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
)

func initTempSQLite(t *testing.T, cfg Config) {
	t.Helper()
	cfg.Driver = "sqlite"
	if cfg.DSN == "" {
		cfg.DSN = filepath.Join(t.TempDir(), "test.db") + "?_busy_timeout=5000"
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	t.Cleanup(func() { _ = Close() })
}

// ExecPrepared ilk çağrıda hazırlar, ikinci çağrıda cache'ten kullanır.
func TestStmtCache_ExecPrepared(t *testing.T) {
	initTempSQLite(t, Config{EnableStmtCache: true, StmtCacheSize: 10})
	ctx := context.Background()
	if _, err := ExecString(ctx, `CREATE TABLE kv (k TEXT PRIMARY KEY, v TEXT)`, nil); err != nil {
		t.Fatal(err)
	}
	q := "INSERT INTO kv (k,v) VALUES (${k}, ${v})"
	if _, err := ExecPrepared(ctx, q, map[string]any{"k": "a", "v": "1"}); err != nil {
		t.Fatal(err)
	}
	if n, err := ExecPrepared(ctx, q, map[string]any{"k": "b", "v": "2"}); err != nil || n != 1 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	prepares, hits := StmtMetrics()
	if prepares != 1 || hits != 1 {
		t.Fatalf("prepares=%d hits=%d, beklenen 1/1", prepares, hits)
	}

	type kv struct{ K, V string }
	var out []kv
	if err := QueryPrepared[kv](ctx, "SELECT k, v FROM kv WHERE k IN (${ks}) ORDER BY k", map[string]any{"ks": []string{"a", "b"}}, &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[0].K != "a" || out[1].V != "2" {
		t.Fatalf("beklenmeyen sonuç: %+v", out)
	}
}

// DB-9: prepare hatası yutulmaz.
func TestStmtCache_PrepareErrorReturned(t *testing.T) {
	initTempSQLite(t, Config{EnableStmtCache: true})
	_, err := ExecPrepared(context.Background(), "INSERT INTO yok_boyle_tablo (a) VALUES (${a})", map[string]any{"a": 1})
	if err == nil {
		t.Fatal("prepare hatası bekleniyordu")
	}
}

// DB-6: tahliye edilen statement kullanımdayken kapatılmaz.
func TestStmtCache_EvictionKeepsInUseStatement(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "c.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	c := newStmtCache(1)
	ctx := context.Background()
	a, err := c.acquire(ctx, sqlDB, "SELECT 1")
	if err != nil {
		t.Fatal(err)
	}
	b, err := c.acquire(ctx, sqlDB, "SELECT 2") // a'yı tahliye eder
	if err != nil {
		t.Fatal(err)
	}
	if c.len() != 1 {
		t.Fatalf("cache boyu %d", c.len())
	}
	var v int
	if err := a.stmt.QueryRowContext(ctx).Scan(&v); err != nil || v != 1 {
		t.Fatalf("tahliye edilen ama kullanımdaki stmt çalışmalı: v=%d err=%v", v, err)
	}
	c.release(a)
	if err := a.stmt.QueryRowContext(ctx).Scan(&v); err == nil {
		t.Fatal("release sonrası tahliye edilmiş stmt kapanmalıydı")
	}
	c.release(b)
	if err := b.stmt.QueryRowContext(ctx).Scan(&v); err != nil {
		t.Fatalf("cache'teki stmt açık kalmalı: %v", err)
	}
	c.closeAll()
	if err := b.stmt.QueryRowContext(ctx).Scan(&v); err == nil {
		t.Fatal("closeAll sonrası stmt kapanmalıydı")
	}
}

// DB-6: eşzamanlı kullanım + küçük cache (tahliye baskısı) -race altında temiz olmalı.
func TestStmtCache_Concurrent(t *testing.T) {
	initTempSQLite(t, Config{EnableStmtCache: true, StmtCacheSize: 2, MaxOpenConns: 4})
	ctx := context.Background()
	if _, err := ExecString(ctx, `CREATE TABLE n (id INTEGER PRIMARY KEY, g INTEGER)`, nil); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		if _, err := ExecString(ctx, "INSERT INTO n (id, g) VALUES (${i}, ${g})", map[string]any{"i": i, "g": i % 5}); err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	errs := make(chan error, 200)
	for w := 0; w < 16; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < 25; i++ {
				// 5 farklı SQL metni: cache boyu 2 olduğu için sürekli tahliye
				q := fmt.Sprintf("SELECT id FROM n WHERE g = ${g} /* q%d */", (w+i)%5)
				var ids []struct{ ID int }
				if err := QueryPrepared(ctx, q, map[string]any{"g": (w + i) % 5}, &ids); err != nil {
					errs <- err
					return
				}
				if len(ids) != 4 {
					errs <- fmt.Errorf("beklenen 4 satır, gelen %d", len(ids))
					return
				}
				_, _ = StmtMetrics()
			}
		}(w)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	if p, h := StmtMetrics(); p+h != 16*25 {
		t.Fatalf("prepares+hits=%d, beklenen %d", p+h, 16*25)
	}
}

// DB-6: Init yeniden çağrıldığında cache temizlenir.
func TestInitAgainResetsCache(t *testing.T) {
	initTempSQLite(t, Config{EnableStmtCache: true})
	ctx := context.Background()
	var one []struct{ X int }
	if err := QueryPrepared(ctx, "SELECT 1 AS x", nil, &one); err != nil {
		t.Fatal(err)
	}
	mu.RLock()
	oldCache := cur.cache
	mu.RUnlock()
	if p, _ := StmtMetrics(); p != 1 {
		t.Fatalf("prepares=%d", p)
	}
	initTempSQLite(t, Config{EnableStmtCache: true})
	if p, h := StmtMetrics(); p != 0 || h != 0 {
		t.Fatalf("yeni Init sonrası metrikler sıfır olmalı: %d/%d", p, h)
	}
	if oldCache.len() != 0 || !oldCache.closed {
		t.Fatal("eski cache kapatılmalıydı")
	}
}
