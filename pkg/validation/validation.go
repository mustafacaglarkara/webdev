// Package validation, kural tabanlı doğrulama motorudur.
//
// İki giriş noktası vardır:
//   - ValidateStruct: go-playground/validator `validate:"..."` struct tag'leri.
//   - ValidateMap / ValidateMapWithMessages / ValidateMapRules: Laravel tarzı
//     "required|email|min:3" kural dizgileri (map tabanlı, dinamik veri).
//
// E-posta ve URL kontrolleri her iki yolda da pkg/validate üzerinden yapılır;
// böylece validate.IsEmail, ValidateMap("email") ve `validate:"email"` her zaman
// aynı sonucu verir.
package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"

	"github.com/mustafacaglarkara/webdev/pkg/validate"
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
		// email/url: tek uygulama pkg/validate (ARCH-2). Yerleşik tag'ler ezilir.
		_ = val.RegisterValidation("email", func(fl validator.FieldLevel) bool {
			f := fl.Field()
			return f.Kind() == reflect.String && validate.IsEmail(f.String())
		})
		_ = val.RegisterValidation("url", func(fl validator.FieldLevel) bool {
			f := fl.Field()
			return f.Kind() == reflect.String && validate.IsURL(f.String())
		})
	})
	return val
}

// undefinedTagRe, go-playground/validator'ın tanımsız tag paniğinden alan ve kural adını
// çıkarır ("Undefined validation function 'x' on field 'Name'").
var undefinedTagRe = regexp.MustCompile(`^Undefined validation function '([^']*)' on field '([^']*)'`)

// ValidateStruct: struct tag'lerindeki `validate:"..."` kurallarını uygular.
// Laravel formatına benzer basit mesajlar döndürmek için ValidationErrorList kullanın.
// Tanımsız bir tag panik üretmez; Tag "unknown_rule", Param tanımsız kural adı ve Field
// ilgili alan adı (belirlenemezse "_") olan bir hata döner (A5-7).
func ValidateStruct(s any) (out ValidationErrorList) {
	defer func() {
		if r := recover(); r != nil {
			ve := ValidationError{
				Field: "_",
				Tag:   "unknown_rule",
				Msg:   fmt.Sprintf("doğrulama kuralı hatası: %v", r),
			}
			if msg, ok := r.(string); ok {
				if m := undefinedTagRe.FindStringSubmatch(msg); m != nil {
					ve.Param = m[1]
					if m[2] != "" {
						ve.Field = m[2]
					}
				}
			}
			out = ValidationErrorList{ve}
		}
	}()
	err := v().Struct(s)
	if err == nil {
		return nil
	}
	var ves validator.ValidationErrors
	if errors.As(err, &ves) {
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

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	case bool:
		return strconv.FormatBool(t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(t), 'f', -1, 32)
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprintf("%v", v)
	}
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

// structTagRule, go-playground tag'lerini mesaj anahtarlarına eşler.
var structTagRule = map[string]string{
	"required": "required", "email": "email", "url": "url",
	"min": "min", "gte": "min", "max": "max", "lte": "max", "len": "size",
	"oneof": "in", "eqfield": "same", "nefield": "different",
	"numeric": "numeric", "number": "numeric", "alpha": "alpha", "alphaunicode": "alpha",
	"alphanum": "alpha_num", "alphanumunicode": "alpha_num", "boolean": "boolean",
	"datetime": "date",
}

func defaultMessage(fe validator.FieldError) string {
	rule, ok := structTagRule[fe.Tag()]
	vars := map[string]string{"param": fe.Param(), "rule": fe.Tag()}
	if !ok {
		return message(nil, fe.Field(), "struct_fallback", "", vars)
	}
	kind := ""
	switch fe.Kind() {
	case reflect.String:
		kind = "string"
	case reflect.Slice, reflect.Array, reflect.Map:
		kind = "array"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		kind = "numeric"
	}
	switch rule {
	case "min":
		vars["min"] = fe.Param()
	case "max":
		vars["max"] = fe.Param()
	case "in":
		vars["values"] = strings.Join(strings.Fields(fe.Param()), ", ")
	case "same", "different":
		vars["other"] = fe.Param()
	default:
		if rule != "size" {
			kind = ""
		}
	}
	return message(nil, fe.Field(), rule, kind, vars)
}

// DefaultMaxJSONBytes, BindAndValidateJSON'un kabul ettiği azami gövde boyutudur (1 MiB).
const DefaultMaxJSONBytes int64 = 1 << 20

var (
	// ErrEmptyBody, istek gövdesi yoksa veya boşsa döner.
	ErrEmptyBody = errors.New("validation: istek gövdesi boş")
	// ErrBodyTooLarge, gövde azami boyutu aşarsa döner.
	ErrBodyTooLarge = errors.New("validation: istek gövdesi çok büyük")
	// ErrTrailingData, JSON değerinden sonra fazladan veri varsa döner.
	ErrTrailingData = errors.New("validation: JSON gövdesinde fazladan veri var")
)

// BindAndValidateJSON: HTTP JSON body -> struct bind + validate.
// Gövde DefaultMaxJSONBytes ile sınırlıdır; tek bir JSON değerinden sonra gelen
// veri ErrTrailingData ile reddedilir.
func BindAndValidateJSON(r *http.Request, dest any) (ValidationErrorList, error) {
	return BindAndValidateJSONLimit(r, dest, DefaultMaxJSONBytes)
}

// BindAndValidateJSONLimit, BindAndValidateJSON'un ayarlanabilir boyut limitli
// sürümüdür. maxBytes <= 0 ise DefaultMaxJSONBytes kullanılır.
// Dönen hata errors.Is ile ErrEmptyBody / ErrBodyTooLarge / ErrTrailingData
// karşılaştırılabilir; diğer durumlarda JSON çözme hatasıdır.
func BindAndValidateJSONLimit(r *http.Request, dest any, maxBytes int64) (ValidationErrorList, error) {
	if r == nil || r.Body == nil || r.Body == http.NoBody {
		return nil, ErrEmptyBody
	}
	defer r.Body.Close()
	if maxBytes <= 0 {
		maxBytes = DefaultMaxJSONBytes
	}
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, maxBytes))
	if err := dec.Decode(dest); err != nil {
		return nil, wrapJSONErr(err)
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		if err != nil {
			if e := wrapJSONErr(err); errors.Is(e, ErrBodyTooLarge) {
				return nil, e
			}
		}
		return nil, ErrTrailingData
	}
	return ValidateStruct(dest), nil
}

func wrapJSONErr(err error) error {
	var mbe *http.MaxBytesError
	if errors.As(err, &mbe) {
		return fmt.Errorf("%w (azami %d bayt)", ErrBodyTooLarge, mbe.Limit)
	}
	if errors.Is(err, io.EOF) {
		return ErrEmptyBody
	}
	return err
}
