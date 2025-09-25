package validation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

// global validator instance (thread-safe)
var (
	valOnce sync.Once
	val     *validator.Validate
)

func v() *validator.Validate {
	valOnce.Do(func() {
		val = validator.New(validator.WithRequiredStructEnabled())
		// Tag alias examples (laravel benzeri kısayol)
		val.RegisterAlias("int", "numeric")
	})
	return val
}

// ValidateStruct: struct tag'lerindeki `validate:"..."` kurallarını uygular.
// Laravel formatına benzer basit mesajlar döndürmek için ValidationErrorList kullanın.
func ValidateStruct(s any) ValidationErrorList {
	err := v().Struct(s)
	if err == nil {
		return nil
	}
	if ves, ok := err.(validator.ValidationErrors); ok {
		list := make(ValidationErrorList, 0, len(ves))
		for _, fe := range ves {
			list = append(list, ValidationError{
				Field: fe.Field(),
				Tag:   fe.Tag(),
				Param: fe.Param(),
				Msg:   defaultMessage(fe),
			})
		}
		return list
	}
	return ValidationErrorList{{Field: "_", Tag: "error", Msg: err.Error()}}
}

// ValidateMap: harita tabanlı (dinamik) giriş doğrulama (Laravel $request->validate([...]) benzeri)
// rules: field => rule string (örn: "email: required,email", "password: required,min=8")
// Dönüş: errors map (alan -> ilk hata mesajı), ok bool (true => geçerli)
func ValidateMap(data map[string]any, rules map[string]string) (map[string]string, bool) {
	errs := make(map[string]string)
	for field, ruleStr := range rules {
		valStr := toString(data[field])
		for _, rule := range strings.Split(ruleStr, "|") {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}
			// simple switch; daha karmaşıklar için validator.Var kullanılabilir
			switch {
			case rule == "required":
				if valStr == "" {
					errs[field] = fmt.Sprintf("%s alanı zorunlu", field)
				}
			case rule == "email":
				if errs[field] == "" && v().Var(valStr, "email") != nil {
					errs[field] = fmt.Sprintf("%s geçerli bir email olmalı", field)
				}
			case strings.HasPrefix(rule, "min=") || strings.HasPrefix(rule, "min:"):
				if errs[field] == "" {
					if m := parseNum(rule); m > 0 && len(valStr) < m {
						errs[field] = fmt.Sprintf("%s en az %d karakter olmalı", field, m)
					}
				}
			case strings.HasPrefix(rule, "max=") || strings.HasPrefix(rule, "max:"):
				if errs[field] == "" {
					if m := parseNum(rule); m > 0 && len(valStr) > m {
						errs[field] = fmt.Sprintf("%s en fazla %d karakter olmalı", field, m)
					}
				}
			default:
				// fallback go-playground (örn: numeric, uuid4 ...)
				if errs[field] == "" && v().Var(valStr, rule) != nil {
					errs[field] = fmt.Sprintf("%s kuralı sağlanmıyor (%s)", field, rule)
				}
			}
			if errs[field] != "" {
				break
			} // ilk hatada dur
		}
	}
	return errs, len(errs) == 0
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func parseNum(rule string) int {
	var p string
	if i := strings.IndexAny(rule, ":="); i >= 0 {
		p = rule[i+1:]
	} else {
		return 0
	}
	var n int
	fmt.Sscanf(p, "%d", &n)
	return n
}

// ValidationError tek alan hatası
type ValidationError struct {
	Field string
	Tag   string
	Param string
	Msg   string
}

// ValidationErrorList toplu hatalar
// Error() => birleştirilmiş mesaj
// ToMap() => field -> mesaj ilk hatayı döner

type ValidationErrorList []ValidationError

func (l ValidationErrorList) Error() string {
	if len(l) == 0 {
		return ""
	}
	parts := make([]string, 0, len(l))
	for _, e := range l {
		parts = append(parts, fmt.Sprintf("%s: %s", e.Field, e.Msg))
	}
	return strings.Join(parts, "; ")
}
func (l ValidationErrorList) ToMap() map[string]string {
	m := make(map[string]string, len(l))
	for _, e := range l {
		if _, ok := m[e.Field]; !ok {
			m[e.Field] = e.Msg
		}
	}
	return m
}

func defaultMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s alanı zorunlu", fe.Field())
	case "email":
		return fmt.Sprintf("%s geçerli bir email olmalı", fe.Field())
	case "min":
		return fmt.Sprintf("%s en az %s karakter", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s en fazla %s karakter", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("%s geçersiz (%s)", fe.Field(), fe.Tag())
	}
}

// BindAndValidateJSON: HTTP JSON body -> struct bind + validate
func BindAndValidateJSON(r *http.Request, dest any) (ValidationErrorList, error) {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dest); err != nil {
		return nil, err
	}
	return ValidateStruct(dest), nil
}
