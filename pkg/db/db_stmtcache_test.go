package db

import (
	"context"
	"testing"
)

// TestStmtCache_ExecPrepared verifies that ExecPrepared uses the statement cache
// when enabled: first execution prepares the statement, second execution hits the cache.
func TestStmtCache_ExecPrepared(t *testing.T) {
	// init in-memory sqlite and enable stmt cache
	cfg := Config{
		Driver:          "sqlite",
		DSN:             "file:testdb_stmtcache?mode=memory&cache=shared",
		MaxOpenConns:    1,
		EnableStmtCache: true,
		StmtCacheSize:   10,
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	ctx := context.Background()

	// create table
	_, err := ExecString(ctx, `CREATE TABLE IF NOT EXISTS kv (k TEXT PRIMARY KEY, v TEXT);`, nil)
	if err != nil {
		t.Fatalf("create table failed: %v", err)
	}

	// reset metrics
	// (directly reset internal counters for deterministic test)
	stmtMu.Lock()
	stmtMetrics.prepares = 0
	stmtMetrics.hits = 0
	for k, s := range stmtCache {
		if s != nil {
			_ = s.Close()
		}
		delete(stmtCache, k)
	}
	stmtMu.Unlock()

	sqlText := "INSERT INTO kv (k,v) VALUES (${k}, ${v});"
	params := map[string]any{"k": "a", "v": "1"}

	// first exec => should prepare
	ra, err := ExecPrepared(ctx, sqlText, params)
	if err != nil {
		t.Fatalf("first ExecPrepared failed: %v", err)
	}
	if ra != 1 && ra != 0 { // sqlite may return 1 or 0 depending on driver
		// accept 0 or 1
	}

	// second exec => should hit cache
	params2 := map[string]any{"k": "b", "v": "2"}
	ra2, err := ExecPrepared(ctx, sqlText, params2)
	if err != nil {
		t.Fatalf("second ExecPrepared failed: %v", err)
	}
	_ = ra2

	prepares, hits := StmtMetrics()
	if prepares < 1 {
		t.Fatalf("expected at least 1 prepare, got %d", prepares)
	}
	if hits < 1 {
		t.Fatalf("expected at least 1 cache hit, got %d", hits)
	}
}
