package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestLoginRateLimit(t *testing.T) {
	app := fiber.New()
	app.All("/login", LoginRateLimit(3, time.Minute), func(c *fiber.Ctx) error { return c.SendString("ok") })
	codes := []int{}
	for range 5 {
		resp, err := app.Test(httptest.NewRequest(http.MethodPost, "/login", nil))
		if err != nil {
			t.Fatal(err)
		}
		codes = append(codes, resp.StatusCode)
	}
	want := []int{200, 200, 200, 429, 429}
	for i := range want {
		if codes[i] != want[i] {
			t.Fatalf("codes = %v, want %v", codes, want)
		}
	}
	// GET istekleri sınırlanmaz.
	resp, _ := app.Test(httptest.NewRequest(http.MethodGet, "/login", nil))
	if resp.StatusCode != 200 {
		t.Fatalf("GET code = %d", resp.StatusCode)
	}
}

func TestOnlyUnsafeMethods(t *testing.T) {
	deny := func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusForbidden) }
	app := fiber.New()
	app.Use(OnlyUnsafeMethods(deny))
	app.All("/", func(c *fiber.Ctx) error { return c.SendString("ok") })
	for method, want := range map[string]int{http.MethodGet: 200, http.MethodHead: 200, http.MethodPost: 403, http.MethodDelete: 403} {
		resp, err := app.Test(httptest.NewRequest(method, "/", nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != want {
			t.Errorf("%s: code = %d, want %d", method, resp.StatusCode, want)
		}
	}
}
