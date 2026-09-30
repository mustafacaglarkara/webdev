package web

import (
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"reflect"
	"sync"
)

// Basit template tag registry (Django {% my_tag %} benzeri)
// Kullanım:
//   web.RegisterTag("upper", strings.ToUpper)
//   tmpl := template.New("x").Funcs(web.TemplateFuncs(w,r))
//   {{ upper .Title }}
//
// Fonksiyon imzaları esnek; argümanlar yansıma ile doğrulanıp çevrilir. Argüman sayısı veya
// tipi uyuşmazsa ya da tag panik atarsa CallTagE hata döner, CallTag ise "" döner ve hatayı
// loglar; iç hata metni sayfaya basılmaz (WEB-16).

var (
	tagMu       sync.RWMutex
	tagRegistry = map[string]any{}
)

// ErrTagNotFound kayıtlı olmayan tag çağrıldığında döner.
var ErrTagNotFound = errors.New("web: tag not registered")

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// RegisterTag adı verilen fonksiyonu registry'e ekler (varsa üzerine yazar).
// fn fonksiyon değilse kayıt yapılmaz.
func RegisterTag(name string, fn any) {
	if name == "" || fn == nil || reflect.TypeOf(fn).Kind() != reflect.Func {
		return
	}
	tagMu.Lock()
	defer tagMu.Unlock()
	tagRegistry[name] = fn
}

// UnregisterTag bir tag'i siler.
func UnregisterTag(name string) {
	tagMu.Lock()
	defer tagMu.Unlock()
	delete(tagRegistry, name)
}

// listRegisteredTags internal kopya döner.
func listRegisteredTags() template.FuncMap {
	tagMu.RLock()
	defer tagMu.RUnlock()
	fm := make(template.FuncMap, len(tagRegistry))
	for k := range tagRegistry {
		name := k
		fm[name] = func(args ...any) any { return CallTag(name, args...) }
	}
	return fm
}

func callTag(name string, fn any, args ...any) (res any, err error) {
	defer func() {
		if p := recover(); p != nil {
			res, err = nil, fmt.Errorf("web: tag %q panicked: %v", name, p)
		}
	}()
	rv := reflect.ValueOf(fn)
	if rv.Kind() != reflect.Func {
		return nil, fmt.Errorf("web: tag %q is not a function", name)
	}
	ft := rv.Type()
	numIn := ft.NumIn()
	if ft.IsVariadic() {
		if len(args) < numIn-1 {
			return nil, fmt.Errorf("web: tag %q expects at least %d args, got %d", name, numIn-1, len(args))
		}
	} else if len(args) != numIn {
		return nil, fmt.Errorf("web: tag %q expects %d args, got %d", name, numIn, len(args))
	}
	in := make([]reflect.Value, len(args))
	for i, a := range args {
		var pt reflect.Type
		if ft.IsVariadic() && i >= numIn-1 {
			pt = ft.In(numIn - 1).Elem()
		} else {
			pt = ft.In(i)
		}
		v, ok := convertArg(a, pt)
		if !ok {
			return nil, fmt.Errorf("web: tag %q arg %d: cannot use %T as %s", name, i, a, pt)
		}
		in[i] = v
	}
	out := rv.Call(in)
	switch len(out) {
	case 0:
		return nil, nil
	case 1:
		if ft.Out(0) == errorType {
			if e, _ := out[0].Interface().(error); e != nil {
				return nil, e
			}
			return nil, nil
		}
		return out[0].Interface(), nil
	default:
		if ft.Out(len(out)-1) == errorType {
			if e, _ := out[len(out)-1].Interface().(error); e != nil {
				return nil, e
			}
		}
		return out[0].Interface(), nil
	}
}

func isIntKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	}
	return false
}

// convertArg a'yı t tipine çevirir. nil → sıfır değer. Sayıdan string'e (rune) dönüşüm
// gibi şaşırtıcı çevrimlere izin verilmez.
func convertArg(a any, t reflect.Type) (reflect.Value, bool) {
	rv := reflect.ValueOf(a)
	if !rv.IsValid() {
		return reflect.Zero(t), true
	}
	if rv.Type().AssignableTo(t) {
		return rv, true
	}
	if t.Kind() == reflect.String && isIntKind(rv.Kind()) {
		return reflect.Value{}, false
	}
	if rv.Type().ConvertibleTo(t) {
		return rv.Convert(t), true
	}
	return reflect.Value{}, false
}

// CallTagE kayıtlı tag'i ismiyle çağırır; hataları döner.
func CallTagE(name string, args ...any) (any, error) {
	tagMu.RLock()
	fn, ok := tagRegistry[name]
	tagMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrTagNotFound, name)
	}
	return callTag(name, fn, args...)
}

// CallTag kayıtlı tag'i ismiyle çağırır. Hata durumunda hatayı loglar ve "" döner
// (şablona iç hata metni basılmaz).
func CallTag(name string, args ...any) any {
	v, err := CallTagE(name, args...)
	if err != nil {
		slog.Warn("web: tag call failed", "tag", name, "err", err)
		return ""
	}
	return v
}
