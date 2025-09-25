package web

import (
	"fmt"
	"html/template"
	"reflect"
	"sync"
)

// Basit template tag registry (Django {% my_tag %} benzeri)
// Kullanım:
//   web.RegisterTag("upper", strings.ToUpper)
//   tmpl := template.New("x").Funcs(web.TemplateFuncs(w,r))
//   {{ upper .Title }}
//
// Fonksiyon imzaları esnek; template motoru argümanları yansıma ile çağırır.
// Dönen tek değer string veya template.HTML ise direkt yazılır; (string,error) dönerse hata ignore edilmez.

var (
	tagMu       sync.RWMutex
	tagRegistry = map[string]any{}
)

// RegisterTag adı verilen fonksiyonu registry'e ekler. Üzerine yazmak isterseniz önce UnregisterTag çağırın.
func RegisterTag(name string, fn any) {
	if name == "" || fn == nil {
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
	for k, v := range tagRegistry {
		val := v
		fm[k] = func(args ...any) any { return callTag(k, val, args...) }
	}
	return fm
}

func callTag(name string, fn any, args ...any) any {
	rv := reflect.ValueOf(fn)
	if rv.Kind() != reflect.Func {
		return fmt.Sprintf("tag %s not func", name)
	}
	if rv.Type().NumIn() != len(args) && !rv.Type().IsVariadic() {
		return fmt.Sprintf("tag %s arg mismatch", name)
	}
	in := make([]reflect.Value, 0, len(args))
	for i, a := range args {
		if !rv.Type().IsVariadic() && i < rv.Type().NumIn() {
			at := rv.Type().In(i)
			in = append(in, reflect.ValueOf(convertArg(a, at)))
		} else {
			in = append(in, reflect.ValueOf(a))
		}
	}
	out := rv.Call(in)
	switch len(out) {
	case 1:
		return out[0].Interface()
	case 2:
		if errv, ok := out[1].Interface().(error); ok && errv != nil {
			return fmt.Sprintf("tag %s error: %v", name, errv)
		}
		return out[0].Interface()
	default:
		return nil
	}
}

func convertArg(a any, t reflect.Type) any {
	rv := reflect.ValueOf(a)
	if !rv.IsValid() {
		return reflect.Zero(t).Interface()
	}
	if rv.Type().AssignableTo(t) {
		return a
	}
	if rv.Type().ConvertibleTo(t) {
		return rv.Convert(t).Interface()
	}
	return a
}
