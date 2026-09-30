package menusvc

import (
	"context"
	"testing"
	"time"

	"github.com/mustafacaglarkara/webdev/cmd/crm/config"
)

func TestBuildTreeSortsAndNests(t *testing.T) {
	items, err := BuildTree(config.DefaultMenuRecords())
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, f := range Flatten(items) {
		keys = append(keys, f.Item.LabelKey)
	}
	want := []string{"menu.home", "menu.about", "menu.demos", "menu.demo_menu", "menu.formdemo", "menu.profile", "menu.admin"}
	if len(keys) != len(want) {
		t.Fatalf("keys = %v", keys)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("keys = %v, want %v", keys, want)
		}
	}
	if len(items[2].Children) != 2 || !items[2].HideIfEmptyChildren {
		t.Fatalf("demos group = %+v", items[2])
	}
}

func TestBuildTreeRejectsBadData(t *testing.T) {
	cases := map[string][]config.MenuRecord{
		"unknown parent": {{ID: 1, ParentID: 9}},
		"duplicate id":   {{ID: 1}, {ID: 1}},
		"invalid id":     {{ID: 0}},
		"cycle":          {{ID: 1}, {ID: 2, ParentID: 3}, {ID: 3, ParentID: 2}},
	}
	for name, recs := range cases {
		if _, err := BuildTree(recs); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestLoaderCachesAndReloads(t *testing.T) {
	store := config.NewMemoryMenuStore([]config.MenuRecord{{ID: 1, LabelKey: "a"}})
	l := NewLoader(store, time.Minute)
	now := time.Unix(1000, 0)
	l.now = func() time.Time { return now }
	ctx := context.Background()

	if got := l.Items(ctx); len(got) != 1 || got[0].LabelKey != "a" {
		t.Fatalf("items = %+v", got)
	}
	store.Replace([]config.MenuRecord{{ID: 1, LabelKey: "b"}})
	if got := l.Items(ctx); got[0].LabelKey != "a" {
		t.Fatal("cache not used within TTL")
	}
	now = now.Add(2 * time.Minute)
	if got := l.Items(ctx); got[0].LabelKey != "b" {
		t.Fatal("cache not refreshed after TTL")
	}
	// Bozuk veri: eski ağaç korunur.
	store.Replace([]config.MenuRecord{{ID: 1, ParentID: 5}})
	now = now.Add(2 * time.Minute)
	if got := l.Items(ctx); len(got) != 1 || got[0].LabelKey != "b" {
		t.Fatal("stale menu should be served when reload fails")
	}
}
