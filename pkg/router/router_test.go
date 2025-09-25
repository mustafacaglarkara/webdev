package router

import (
	"testing"
)

func TestReverseURL_NormalizeAndPositional(t *testing.T) {
	RegisterTemplate("u1", "/user/{id:[0-9]+}")
	got, err := ReverseURL("u1", "123")
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}
	if got != "/user/123" {
		t.Fatalf("unexpected URL: %s", got)
	}
}

func TestReverseURL_NamedMapAndEscape(t *testing.T) {
	RegisterTemplate("u2", "/file/{name}")
	got, err := ReverseURL("u2", map[string]string{"name": "a/b c"})
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}
	// a/b c -> a%2Fb%20c
	if got != "/file/a%2Fb%20c" {
		t.Fatalf("unexpected escaped URL: %s", got)
	}
}

func TestReverseURL_PrintfAndEscape(t *testing.T) {
	RegisterTemplate("p1", "/dl/%s")
	got, err := ReverseURL("p1", "a/b c")
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}
	if got != "/dl/a%2Fb%20c" {
		t.Fatalf("unexpected printf escaped url: %s", got)
	}
}
