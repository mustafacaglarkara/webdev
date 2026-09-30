package db

import (
	"container/list"
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
)

// stmtCache: LRU prepared statement cache'i. Girdiler referans sayımlıdır:
// tahliye edilen (evicted) bir statement, onu kullanan son çağrı bitene kadar
// kapatılmaz.
type stmtCache struct {
	mu     sync.Mutex
	size   int // <= 0 sınırsız
	items  map[string]*stmtEntry
	lru    *list.List // ön: en son kullanılan
	closed bool

	prepares atomic.Int64
	hits     atomic.Int64
}

type stmtEntry struct {
	key     string
	stmt    *sql.Stmt
	refs    int
	evicted bool
	elem    *list.Element
}

func newStmtCache(size int) *stmtCache {
	return &stmtCache{size: size, items: make(map[string]*stmtEntry), lru: list.New()}
}

// acquire: key için hazırlanmış statement döner (gerekirse hazırlar).
// Kullanım bitince release çağrılmalıdır.
func (c *stmtCache) acquire(ctx context.Context, sqlDB *sql.DB, key string) (*stmtEntry, error) {
	c.mu.Lock()
	if e, ok := c.items[key]; ok && !c.closed {
		e.refs++
		c.lru.MoveToFront(e.elem)
		c.mu.Unlock()
		c.hits.Add(1)
		return e, nil
	}
	c.mu.Unlock()

	s, err := sqlDB.PrepareContext(ctx, key)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		// cache kapatıldı: tek kullanımlık girdi; release'de kapanır
		c.prepares.Add(1)
		return &stmtEntry{key: key, stmt: s, refs: 1, evicted: true}, nil
	}
	if e, ok := c.items[key]; ok {
		// başka goroutine aynı anda hazırladı
		_ = s.Close()
		e.refs++
		c.lru.MoveToFront(e.elem)
		c.hits.Add(1)
		return e, nil
	}
	e := &stmtEntry{key: key, stmt: s, refs: 1}
	e.elem = c.lru.PushFront(e)
	c.items[key] = e
	c.prepares.Add(1)
	for c.size > 0 && len(c.items) > c.size {
		back := c.lru.Back()
		if back == nil {
			break
		}
		c.evictLocked(back.Value.(*stmtEntry))
	}
	return e, nil
}

func (c *stmtCache) evictLocked(e *stmtEntry) {
	if e.evicted {
		return
	}
	e.evicted = true
	if e.elem != nil {
		c.lru.Remove(e.elem)
		e.elem = nil
	}
	delete(c.items, e.key)
	if e.refs == 0 {
		_ = e.stmt.Close()
	}
}

// release: acquire ile alınan girdiyi bırakır.
func (c *stmtCache) release(e *stmtEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e.refs--
	if e.evicted && e.refs == 0 {
		_ = e.stmt.Close()
	}
}

// closeAll: tüm girdileri tahliye eder; kullanımda olanlar release'de kapanır.
func (c *stmtCache) closeAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	for _, e := range c.items {
		c.evictLocked(e)
	}
}

func (c *stmtCache) len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}
