package forms

import (
	"net/http"
	"net/url"
	"reflect"

	"github.com/mustafacaglarkara/webdev/pkg/validation"
)

// Form is a lightweight wrapper to validate and access cleaned data similar to Django forms.
type Form struct {
	Data        map[string]any
	Errors      map[string]string
	cleanedData map[string]any
	Validated   bool
}

// NewFromMap creates a new Form from a generic map (e.g., request data).
func NewFromMap(m map[string]any) *Form {
	if m == nil {
		m = map[string]any{}
	}
	return &Form{Data: m, Errors: map[string]string{}}
}

// NewFromRequest parses form values from an *http.Request (application/x-www-form-urlencoded or multipart).
func NewFromRequest(r *http.Request) *Form {
	_ = r.ParseForm()
	vals := r.Form
	if r.PostForm != nil && len(r.PostForm) > 0 { // prefer body form fields when available
		vals = r.PostForm
	}
	return &Form{Data: valuesToMap(vals), Errors: map[string]string{}}
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
func (f *Form) ValidateMap(rules map[string]string) bool {
	errs, ok := validation.ValidateMap(f.Data, rules)
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
	if rv.Kind() != reflect.Ptr {
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
		mname := "Clean" + fld.Name
		meth := rv.Addr().MethodByName(mname)
		if !meth.IsValid() {
			continue
		}
		switch meth.Type().NumIn() {
		case 0:
			// expects: func() error
			if meth.Type().NumOut() == 1 && meth.Type().Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				out := meth.Call(nil)
				if !out[0].IsNil() {
					err := out[0].Interface().(error)
					f.AddError(fld.Name, err.Error())
				}
			}
		case 1:
			// expects: func(v T) (T, error) — T alan tipiyle aynı veya pointer karşılığı olabilir
			argT := meth.Type().In(0)
			retCount := meth.Type().NumOut()
			if retCount == 2 && meth.Type().Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				fv := rv.Field(i)
				if !fv.IsValid() {
					break
				}
				var callArg reflect.Value
				ft := fv.Type()
				// Doğrudan atanan tür
				if ft.AssignableTo(argT) {
					callArg = fv
				} else if ft.Kind() == reflect.Ptr && ft.Elem().AssignableTo(argT) {
					// Alan *X, arg X ise: nil ise zero X, değilse *alan'ı değer olarak gönder
					if fv.IsNil() {
						callArg = reflect.Zero(argT)
					} else {
						callArg = fv.Elem()
					}
				} else if argT.Kind() == reflect.Ptr && ft.AssignableTo(argT.Elem()) {
					// Alan X, arg *X ise: alanın adresini gönder
					if fv.CanAddr() {
						callArg = fv.Addr()
					}
				} else if ft.ConvertibleTo(argT) {
					callArg = fv.Convert(argT)
				} else {
					// desteklenmeyen tür eşleşmesi
					break
				}

				out := meth.Call([]reflect.Value{callArg})
				newV := out[0]
				errV := out[1]
				if !errV.IsNil() {
					err := errV.Interface().(error)
					f.AddError(fld.Name, err.Error())
					break
				}
				// Dönen değeri alana uygula: pointer/non-pointer durumlarını ele al
				if !fv.CanSet() {
					break
				}
				// 1) newV alan tipine atanabiliyorsa doğrudan set et
				if newV.Type().AssignableTo(ft) {
					fv.Set(newV)
					break
				}
				// 2) Alan *X, dönüş X ise: yeni değerden *X oluşturup ata
				if ft.Kind() == reflect.Ptr && newV.Type().AssignableTo(ft.Elem()) {
					ptr := reflect.New(ft.Elem())
					ptr.Elem().Set(newV)
					fv.Set(ptr)
					break
				}
				// 3) Alan X, dönüş *X ise: nil değilse deref ederek ata; nil ise zero değer ata
				if newV.Kind() == reflect.Ptr && newV.Type().Elem().AssignableTo(ft) {
					if newV.IsNil() {
						fv.Set(reflect.Zero(ft))
					} else {
						fv.Set(newV.Elem())
					}
					break
				}
				// 4) Dönüş değeri dönüştürülebiliyorsa çevirerek ata
				if newV.Type().ConvertibleTo(ft) {
					fv.Set(newV.Convert(ft))
				}
			}
		}
	}
	// form-level cleaner: Clean() error
	if m := rv.Addr().MethodByName("Clean"); m.IsValid() && m.Type().NumIn() == 0 && m.Type().NumOut() == 1 && m.Type().Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		out := m.Call(nil)
		if !out[0].IsNil() {
			err := out[0].Interface().(error)
			// non-field errors under "_"
			f.AddError("_", err.Error())
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
		fv := rv.Field(i)
		f.cleanedData[fld.Name] = fv.Interface()
	}
}
