package forms

import (
	"errors"
	"strings"
	"testing"
)

type regForm struct {
	Email string `validate:"required,email"`
	Age   int    `validate:"min=18"`
}

// Field-level cleaner: normalize email to lowercase
func (r *regForm) CleanEmail(v string) (string, error) {
	return strings.ToLower(v), nil
}

// Form-level cleaner: disallow a specific email
func (r *regForm) Clean() error {
	if r.Email == "banned@example.com" {
		return errors.New("this email is banned")
	}
	return nil
}

func TestCleanHooks_FieldAndForm(t *testing.T) {
	// prepare form data
	f := NewFromMap(map[string]any{
		"Email": "TEST@EXAMPLE.COM",
		"Age":   22,
	})
	// bind into DTO and validate + clean
	dto := regForm{Email: f.Data["Email"].(string), Age: 22}
	_ = f.ValidateStruct(&dto)
	// email should be lowercased by CleanEmail
	if got := f.Cleaned("Email"); got != "test@example.com" {
		t.Fatalf("expected cleaned Email to be lowercased, got %v", got)
	}
	// no errors so far
	if f.Error("_") != "" {
		t.Fatalf("unexpected non-field error: %s", f.Error("_"))
	}
	// run with banned email to trigger Clean() error
	f2 := NewFromMap(map[string]any{
		"Email": "banned@example.com",
		"Age":   30,
	})
	dto2 := regForm{Email: f2.Data["Email"].(string), Age: 30}
	_ = f2.ValidateStruct(&dto2)
	if f2.Error("_") == "" {
		t.Fatalf("expected non-field error from Clean()")
	}
}

type ptrForm struct {
	Name *string `validate:"required"`
}

// CleanName with value arg, pointer field: func(v string) (string, error)
func (p *ptrForm) CleanName(v string) (string, error) {
	return strings.TrimSpace(v), nil
}

func TestCleanHooks_PointerField_ValueArg(t *testing.T) {
	name := "  Alice  "
	f := NewFromMap(map[string]any{"Name": name})
	dto := ptrForm{Name: &name}
	_ = f.ValidateStruct(&dto)
	if dto.Name == nil || *dto.Name != "Alice" {
		t.Fatalf("expected Name to be trimmed to 'Alice', got %#v", dto.Name)
	}
	if got := f.Cleaned("Name"); got == nil || got.(*string) == nil || *(got.(*string)) != "Alice" {
		t.Fatalf("cleaned Name mismatch, got %#v", got)
	}
}

type ptrForm2 struct {
	Name *string `validate:"required"`
}

// CleanName with pointer arg, pointer field: func(v *string) (*string, error)
func (p *ptrForm2) CleanName(v *string) (*string, error) {
	if v == nil {
		return v, errors.New("name is nil")
	}
	trimmed := strings.TrimSpace(*v)
	return &trimmed, nil
}

func TestCleanHooks_PointerField_PointerArg(t *testing.T) {
	name := " Bob "
	f := NewFromMap(map[string]any{"Name": name})
	dto := ptrForm2{Name: &name}
	_ = f.ValidateStruct(&dto)
	if dto.Name == nil || *dto.Name != "Bob" {
		t.Fatalf("expected Name to be trimmed to 'Bob', got %#v", dto.Name)
	}
}
