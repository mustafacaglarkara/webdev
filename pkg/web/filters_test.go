package web

import (
	"html/template"
	"testing"
	"time"
)

func TestJetTemplateFilters_Basics(t *testing.T) {
	fm := JetTemplateFilters()

	// upper
	if got := fm["upper"].(func(string) string)("deneme"); got != "DENEME" {
		t.Fatalf("upper failed: got=%q", got)
	}
	// default
	if got := fm["default"].(func(any, any) any)("", "x"); got != "x" {
		t.Fatalf("default failed for empty string: got=%v", got)
	}
	if got := fm["default"].(func(any, any) any)("val", "x"); got != "val" {
		t.Fatalf("default failed for non-empty: got=%v", got)
	}
	// length
	if ln := fm["length"].(func(any) int)("ışık"); ln != 4 { // rune count
		t.Fatalf("length failed for runes, got=%d", ln)
	}
	// truncatewords
	if s := fm["truncatewords"].(func(string, int) string)("a b c d", 2); s != "a b..." {
		t.Fatalf("truncatewords failed: got=%q", s)
	}
	// date
	tm := time.Date(2024, 2, 3, 10, 11, 0, 0, time.UTC)
	if ds := fm["date"].(func(any, string) string)(tm, "2006-01-02"); ds != "2024-02-03" {
		t.Fatalf("date failed: got=%q", ds)
	}
	// safe
	if v := fm["safe"].(func(string) template.HTML)("<b>x</b>"); v != template.HTML("<b>x</b>") {
		t.Fatalf("safe failed: got=%v", v)
	}
}
