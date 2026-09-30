// Package forms, Django benzeri hafif bir form katmanıdır: istekten veri
// toplar, pkg/validation ile doğrular, struct'a bağlar (Bind) ve Clean hook'larını çalıştırır.
package forms

import (
	"errors"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"

	"github.com/mustafacaglarkara/webdev/pkg/validation"
)

// DefaultMaxMemory, multipart formlar ayrıştırılırken belleğe alınacak azami
// bayt sayısıdır (32 MiB, net/http varsayılanı). Fazlası geçici dosyalara yazılır.
const DefaultMaxMemory int64 = 32 << 20

// ErrNilRequest, NewFromRequest'e nil istek verildiğinde ParseError() ile döner.
var ErrNilRequest = errors.New("forms: istek nil")

// Form is a lightweight wrapper to validate and access cleaned data similar to Django forms.
type Form struct {
	Data        map[string]any
	Errors      map[string]string
	cleanedData map[string]any
	Validated   bool
	// Files, multipart isteklerde yüklenen dosyaları içerir (alan adı -> dosyalar).
	Files map[string][]*multipart.FileHeader

	parseErr error
}

// NewFromMap creates a new Form from a generic map (e.g., request data).
func NewFromMap(m map[string]any) *Form {
	if m == nil {
		m = map[string]any{}
	}
	return &Form{Data: m, Errors: map[string]string{}}
}

// NewFromRequest, istek form değerlerini okur (application/x-www-form-urlencoded
// veya multipart/form-data). Multipart istekler DefaultMaxMemory ile ayrıştırılır.
// Ayrıştırma hatası yutulmaz; ParseError() ile okunabilir.
func NewFromRequest(r *http.Request) *Form {
	return NewFromRequestWithMaxMemory(r, DefaultMaxMemory)
}

// NewFromRequestWithMaxMemory, NewFromRequest ile aynıdır; multipart ayrıştırma
// için bellek limiti verilebilir (<= 0 ise DefaultMaxMemory).
func NewFromRequestWithMaxMemory(r *http.Request, maxMemory int64) *Form {
	f := &Form{Data: map[string]any{}, Errors: map[string]string{}}
	if r == nil {
		f.parseErr = ErrNilRequest
		return f
	}
	if maxMemory <= 0 {
		maxMemory = DefaultMaxMemory
	}
	if isMultipart(r) {
		f.parseErr = r.ParseMultipartForm(maxMemory)
		if r.MultipartForm != nil && len(r.MultipartForm.File) > 0 {
			f.Files = r.MultipartForm.File
		}
	} else {
		f.parseErr = r.ParseForm()
	}
	vals := r.Form
	if len(r.PostForm) > 0 { // prefer body form fields when available
		vals = r.PostForm
	}
	f.Data = valuesToMap(vals)
	return f
}

func isMultipart(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		return false
	}
	mt, _, err := mime.ParseMediaType(ct)
	return err == nil && mt == "multipart/form-data"
}

// ParseError, NewFromRequest sırasında oluşan ayrıştırma hatasını döner (yoksa nil).
func (f *Form) ParseError() error { return f.parseErr }

// File, alan için yüklenen ilk dosyayı döner (yoksa nil).
func (f *Form) File(field string) *multipart.FileHeader {
	if fs := f.Files[field]; len(fs) > 0 {
		return fs[0]
	}
	return nil
}

func valuesToMap(v url.Values) map[string]any {
	m := make(map[string]any, len(v))
	for k, vals := range v {
		if len(vals) == 1 {
			m[k] = vals[0]
		} else {
			arr := make([]string, len(vals))
			copy(arr, vals)
			m[k] = arr
		}
	}
	return m
}

// ValidateMap runs validation rules (Laravel-like) and populates Errors/CleanedData.
// Kural sözdizimi için bkz. validation.ValidateMap.
func (f *Form) ValidateMap(rules map[string]string) bool {
	return f.ValidateMapWithMessages(rules, nil)
}

// ValidateMapWithMessages, ValidateMap ile aynıdır; çağrıya özel mesaj
// şablonları verilebilir (bkz. validation.ValidateMapWithMessages).
func (f *Form) ValidateMapWithMessages(rules map[string]string, messages map[string]string) bool {
	errs, ok := validation.ValidateMapWithMessages(f.Data, rules, messages)
	f.Errors = errs
	f.Validated = true
	if ok {
		// naive cleaned data = input
		f.cleanedData = make(map[string]any, len(f.Data))
		for k, v := range f.Data {
			f.cleanedData[k] = v
		}
	}
	return ok
}

// ValidateStruct validates dest using struct tags; Errors are filled from ValidationErrorList.
// Ardından Clean<Alan>() ve Clean() hook'ları çalışır. Doğrulama hatası olan
// alanların Clean<Alan> hook'u çalışmaz ve hook hataları mevcut hataları ezmez.
func (f *Form) ValidateStruct(dest any) validation.ValidationErrorList {
	list := validation.ValidateStruct(dest)
	f.Validated = true
	f.Errors = list.ToMap()
	// expose posted values to templates regardless of validity
	f.cleanedData = make(map[string]any, len(f.Data))
	for k, v := range f.Data {
		f.cleanedData[k] = v
	}
	// Apply Clean<Field>() and Clean() hooks on dest if available, then mirror updated values
	f.applyCleanHooks(dest)
	return list
}

// BindAndValidate, Bind + ValidateStruct'ı birlikte çalıştırır. Bağlama
// (tip dönüşümü) hataları, aynı alanın doğrulama hatasından önceliklidir.
// dest struct işaretçisi değilse hata "_" anahtarına yazılır.
func (f *Form) BindAndValidate(dest any) bool {
	bindErr := f.Bind(dest)
	var be *BindError
	if bindErr != nil && !errors.As(bindErr, &be) {
		f.Validated = true
		f.Errors = map[string]string{"_": bindErr.Error()}
		return false
	}
	f.ValidateStruct(dest)
	if be != nil {
		for k, msg := range be.Fields {
			f.AddError(k, msg)
		}
	}
	return f.IsValid()
}

// IsValid returns true if the form was validated and has no errors.
func (f *Form) IsValid() bool { return f.Validated && len(f.Errors) == 0 }

// CleanedData returns the map of cleaned values (after validation).
func (f *Form) CleanedData() map[string]any { return f.cleanedData }

// Cleaned returns a single cleaned field value (if present).
func (f *Form) Cleaned(field string) any {
	if f.cleanedData == nil {
		return nil
	}
	return f.cleanedData[field]
}

// Error returns first error message for a given field (or empty string).
func (f *Form) Error(field string) string {
	if f.Errors == nil {
		return ""
	}
	return f.Errors[field]
}

// AddError adds or overrides an error for given field.
func (f *Form) AddError(field, msg string) {
	if f.Errors == nil {
		f.Errors = map[string]string{}
	}
	f.Errors[field] = msg
}

// addErrorIfAbsent, alanda hata yoksa ekler; mevcut (doğrulama) hatasını ezmez.
func (f *Form) addErrorIfAbsent(field, msg string) {
	if f.Errors == nil {
		f.Errors = map[string]string{}
	}
	if _, ok := f.Errors[field]; !ok {
		f.Errors[field] = msg
	}
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// safeConvertible, reflect dönüşümüne izin verilip verilmeyeceğini söyler.
// Tam sayı <-> string dönüşümü Go'da rune dönüşümüdür (65 -> "A"); engellenir.
func safeConvertible(from, to reflect.Type) bool {
	if !from.ConvertibleTo(to) {
		return false
	}
	if isIntKind(from.Kind()) && to.Kind() == reflect.String {
		return false
	}
	if from.Kind() == reflect.String && isIntKind(to.Kind()) {
		return false
	}
	return true
}

func isIntKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	}
	return false
}

// applyCleanHooks runs Clean<Field>() and Clean() on dest if they exist.
// Supported signatures:
//   - Clean<Field>() error                              // reads/modifies struct field directly
//   - Clean<Field>(v T) (T, error)                      // returns new value for the field
//   - Clean() error                                     // form-level checks
func (f *Form) applyCleanHooks(dest any) {
	rv := reflect.ValueOf(dest)
	if !rv.IsValid() {
		return
	}
	// ensure pointer to struct to access methods and set fields
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return
	}
	rt := rv.Type()
	// field-level cleaners
	for i := 0; i < rt.NumField(); i++ {
		fld := rt.Field(i)
		if fld.PkgPath != "" { // unexported
			continue
		}
		// Django davranışı: doğrulaması başarısız alanın cleaner'ı çalışmaz.
		if _, failed := f.Errors[fld.Name]; failed {
			continue
		}
		mname := "Clean" + fld.Name
		meth := rv.Addr().MethodByName(mname)
		if !meth.IsValid() {
			continue
		}
		mt := meth.Type()
		switch mt.NumIn() {
		case 0:
			// expects: func() error
			if mt.NumOut() == 1 && mt.Out(0).Implements(errorType) {
				out := meth.Call(nil)
				if !out[0].IsNil() {
					f.addErrorIfAbsent(fld.Name, out[0].Interface().(error).Error())
				}
			}
		case 1:
			// expects: func(v T) (T, error) — T alan tipiyle aynı veya pointer karşılığı olabilir
			if mt.NumOut() != 2 || !mt.Out(1).Implements(errorType) {
				break
			}
			argT := mt.In(0)
			fv := rv.Field(i)
			var callArg reflect.Value
			ft := fv.Type()
			switch {
			case ft.AssignableTo(argT):
				callArg = fv
			case ft.Kind() == reflect.Ptr && ft.Elem().AssignableTo(argT):
				// Alan *X, arg X ise: nil ise zero X, değilse *alan'ı değer olarak gönder
				if fv.IsNil() {
					callArg = reflect.Zero(argT)
				} else {
					callArg = fv.Elem()
				}
			case argT.Kind() == reflect.Ptr && ft.AssignableTo(argT.Elem()):
				// Alan X, arg *X ise: alanın adresini gönder
				if !fv.CanAddr() {
					continue
				}
				callArg = fv.Addr()
			case safeConvertible(ft, argT):
				callArg = fv.Convert(argT)
			default:
				// desteklenmeyen tür eşleşmesi
				continue
			}

			out := meth.Call([]reflect.Value{callArg})
			newV, errV := out[0], out[1]
			if !errV.IsNil() {
				f.addErrorIfAbsent(fld.Name, errV.Interface().(error).Error())
				break
			}
			if !fv.CanSet() {
				break
			}
			setCleanedValue(fv, newV)
		}
	}
	// form-level cleaner: Clean() error
	if m := rv.Addr().MethodByName("Clean"); m.IsValid() && m.Type().NumIn() == 0 && m.Type().NumOut() == 1 && m.Type().Out(0).Implements(errorType) {
		out := m.Call(nil)
		if !out[0].IsNil() {
			// non-field errors under "_"
			f.addErrorIfAbsent("_", out[0].Interface().(error).Error())
		}
	}
	// mirror updated struct values into cleanedData
	if f.cleanedData == nil {
		f.cleanedData = map[string]any{}
	}
	for i := 0; i < rt.NumField(); i++ {
		fld := rt.Field(i)
		if fld.PkgPath != "" {
			continue
		}
		f.cleanedData[fld.Name] = rv.Field(i).Interface()
	}
}

// setCleanedValue, cleaner'ın döndürdüğü değeri alana uygular
// (pointer/non-pointer durumları dahil).
func setCleanedValue(fv, newV reflect.Value) {
	ft := fv.Type()
	if !newV.IsValid() {
		fv.Set(reflect.Zero(ft))
		return
	}
	nt := newV.Type()
	switch {
	case nt.AssignableTo(ft):
		fv.Set(newV)
	case ft.Kind() == reflect.Ptr && nt.AssignableTo(ft.Elem()):
		// Alan *X, dönüş X ise: yeni değerden *X oluşturup ata
		ptr := reflect.New(ft.Elem())
		ptr.Elem().Set(newV)
		fv.Set(ptr)
	case newV.Kind() == reflect.Ptr && nt.Elem().AssignableTo(ft):
		// Alan X, dönüş *X ise: nil değilse deref ederek ata; nil ise zero değer ata
		if newV.IsNil() {
			fv.Set(reflect.Zero(ft))
		} else {
			fv.Set(newV.Elem())
		}
	case safeConvertible(nt, ft):
		fv.Set(newV.Convert(ft))
	}
}
