package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

type userDTO struct {
	Email string `validate:"required,email"`
	Pass  string `validate:"required,min=6"`
}

func TestNewFromMap_ValidateMap(t *testing.T) {
	f := NewFromMap(map[string]any{"email": "john@example.com", "pass": "secret123"})
	ok := f.ValidateMap(map[string]string{
		"email": "required|email",
		"pass":  "required|min=6",
	})
	if !ok || !f.IsValid() {
		t.Fatalf("expected valid form, got errors=%v", f.Errors)
	}
	if f.CleanedData()["email"] != "john@example.com" {
		t.Fatalf("cleaned data missing or wrong")
	}
}

func TestNewFromRequest_ValidateStruct(t *testing.T) {
	// prepare a form request
	form := url.Values{}
	form.Set("Email", "bad-email")
	form.Set("Pass", "123")
	r := httptest.NewRequest("POST", "/", nil)
	r.PostForm = form
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	f := NewFromRequest(r)
	// bind into dto and validate
	dto := userDTO{Email: f.Data["Email"].(string), Pass: f.Data["Pass"].(string)}
	errs := f.ValidateStruct(&dto)
	if len(errs) == 0 || f.IsValid() {
		t.Fatalf("expected validation errors, got none")
	}
	if f.Error("Email") == "" {
		t.Fatalf("expected Email error")
	}
}
