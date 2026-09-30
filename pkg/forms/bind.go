package forms

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TimeLayouts, time.Time alanlarına bağlarken sırayla denenen biçimlerdir.
// Alan bazında `time_format:"02.01.2006"` etiketiyle tek bir biçim zorlanabilir.
var TimeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02T15:04", // <input type="datetime-local">
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02", // <input type="date">
	"02.01.2006 15:04",
	"02.01.2006",
}

// ErrBindTarget, Bind'e struct işaretçisi dışında bir şey verildiğinde döner.
var ErrBindTarget = errors.New("forms: Bind hedefi nil olmayan bir struct işaretçisi olmalı")

// BindError, bir veya daha fazla alanın tip dönüşümü başarısız olduğunda döner.
// Fields anahtarları Go alan adlarıdır (ValidateStruct hatalarıyla aynı anahtarlar).
type BindError struct {
	Fields map[string]string
}

func (e *BindError) Error() string {
	keys := make([]string, 0, len(e.Fields))
	for k := range e.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+": "+e.Fields[k])
	}
	return "forms: bağlama hatası: " + strings.Join(parts, "; ")
}

// Bind, form verisini (f.Data) dest struct'ına doldurur.
//
// Alan adı eşleme sırası: `form:"ad"` etiketi, `json:"ad"` etiketi, Go alan adı
// (önce birebir, sonra büyük/küçük harf duyarsız). `form:"-"` alanı atlar.
// Gömülü (anonymous) struct alanları düzleştirilerek doldurulur.
//
// Desteklenen tipler: string, int*, uint*, float*, bool, time.Time, bunların
// dilimleri ve işaretçileri. Boş dizge sayısal/bool/zaman alanlarında sıfır
// değer (işaretçide nil) bırakır. Dönüştürülemeyen değerler *BindError içinde
// toplanır; diğer alanlar yine doldurulur.
func (f *Form) Bind(dest any) error {
	rv := reflect.ValueOf(dest)
	if !rv.IsValid() || rv.Kind() != reflect.Pointer || rv.IsNil() || rv.Elem().Kind() != reflect.Struct {
		return ErrBindTarget
	}
	errs := map[string]string{}
	bindStruct(rv.Elem(), f.Data, errs)
	if len(errs) > 0 {
		return &BindError{Fields: errs}
	}
	return nil
}

func bindStruct(sv reflect.Value, data map[string]any, errs map[string]string) {
	st := sv.Type()
	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		fv := sv.Field(i)
		if sf.Anonymous {
			t := sf.Type
			if t.Kind() == reflect.Struct && t != timeType {
				bindStruct(fv, data, errs)
				continue
			}
			if t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Struct && t.Elem() != timeType && sf.IsExported() {
				if fv.IsNil() {
					fv.Set(reflect.New(t.Elem()))
				}
				bindStruct(fv.Elem(), data, errs)
				continue
			}
		}
		if !sf.IsExported() || !fv.CanSet() {
			continue
		}
		key, skip := fieldKey(sf)
		if skip {
			continue
		}
		raw, ok := lookup(data, key, sf.Name)
		if !ok {
			continue
		}
		if err := setValue(fv, raw, sf.Tag.Get("time_format")); err != nil {
			errs[sf.Name] = fmt.Sprintf("%s alanı için geçersiz değer: %v", sf.Name, err)
		}
	}
}

// fieldKey, alan için form anahtarını döner: form etiketi > json etiketi > alan adı.
func fieldKey(sf reflect.StructField) (string, bool) {
	for _, tagName := range []string{"form", "json"} {
		tag, ok := sf.Tag.Lookup(tagName)
		if !ok {
			continue
		}
		name := strings.TrimSpace(strings.Split(tag, ",")[0])
		if name == "-" {
			return "", true
		}
		if name != "" {
			return name, false
		}
	}
	return sf.Name, false
}

func lookup(data map[string]any, key, goName string) (any, bool) {
	if v, ok := data[key]; ok {
		return v, true
	}
	if key != goName {
		return nil, false // etiket verilmişse yalnızca etiket adı kullanılır
	}
	for k, v := range data {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}
	return nil, false
}

var timeType = reflect.TypeOf(time.Time{})

// setValue, raw değerini fv'ye tip dönüşümüyle atar.
func setValue(fv reflect.Value, raw any, layout string) error {
	ft := fv.Type()
	if raw == nil {
		fv.Set(reflect.Zero(ft))
		return nil
	}
	rt := reflect.TypeOf(raw)
	// Doğrudan atanabilen değer (NewFromMap ile gelen tipli veri).
	if rt.AssignableTo(ft) {
		fv.Set(reflect.ValueOf(raw))
		return nil
	}

	switch {
	case ft.Kind() == reflect.Pointer:
		s, isStr := singleString(raw)
		if isStr && strings.TrimSpace(s) == "" && ft.Elem().Kind() != reflect.String {
			fv.Set(reflect.Zero(ft))
			return nil
		}
		nv := reflect.New(ft.Elem())
		if err := setValue(nv.Elem(), raw, layout); err != nil {
			return err
		}
		fv.Set(nv)
		return nil
	case ft.Kind() == reflect.Slice && ft.Elem().Kind() != reflect.Uint8:
		items := toItems(raw)
		out := reflect.MakeSlice(ft, len(items), len(items))
		for i, it := range items {
			if err := setValue(out.Index(i), it, layout); err != nil {
				return err
			}
		}
		fv.Set(out)
		return nil
	}

	// Tekil hedef: çok değerli girdide ilk değer kullanılır.
	if items := toItems(raw); len(items) != 1 || !isScalar(raw) {
		if len(items) == 0 {
			fv.Set(reflect.Zero(ft))
			return nil
		}
		raw = items[0]
	}
	return setScalar(fv, raw, layout)
}

func isScalar(raw any) bool {
	k := reflect.TypeOf(raw).Kind()
	return k != reflect.Slice && k != reflect.Array
}

func toItems(raw any) []any {
	switch t := raw.(type) {
	case []string:
		out := make([]any, len(t))
		for i, s := range t {
			out[i] = s
		}
		return out
	case []any:
		return t
	}
	rv := reflect.ValueOf(raw)
	if (rv.Kind() == reflect.Slice && rv.Type().Elem().Kind() != reflect.Uint8) || rv.Kind() == reflect.Array {
		out := make([]any, rv.Len())
		for i := range out {
			out[i] = rv.Index(i).Interface()
		}
		return out
	}
	return []any{raw}
}

func singleString(raw any) (string, bool) {
	switch t := raw.(type) {
	case string:
		return t, true
	case []string:
		if len(t) > 0 {
			return t[0], true
		}
		return "", true
	}
	return "", false
}

func scalarString(raw any) string {
	switch t := raw.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(t), 'f', -1, 32)
	case fmt.Stringer:
		return t.String()
	}
	return fmt.Sprint(raw)
}

func setScalar(fv reflect.Value, raw any, layout string) error {
	ft := fv.Type()
	if rt := reflect.TypeOf(raw); rt.AssignableTo(ft) {
		fv.Set(reflect.ValueOf(raw))
		return nil
	}
	if ft == timeType {
		s := strings.TrimSpace(scalarString(raw))
		if s == "" {
			fv.Set(reflect.Zero(ft))
			return nil
		}
		layouts := TimeLayouts
		if layout != "" {
			layouts = []string{layout}
		}
		for _, l := range layouts {
			if t, err := time.Parse(l, s); err == nil {
				fv.Set(reflect.ValueOf(t))
				return nil
			}
		}
		return fmt.Errorf("tarih çözümlenemedi: %q", s)
	}

	s := scalarString(raw)
	trimmed := strings.TrimSpace(s)
	switch ft.Kind() {
	case reflect.String:
		fv.SetString(s)
	case reflect.Bool:
		if trimmed == "" {
			fv.SetBool(false)
			return nil
		}
		b, err := parseBool(trimmed)
		if err != nil {
			return err
		}
		fv.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if trimmed == "" {
			fv.SetInt(0)
			return nil
		}
		n, err := strconv.ParseInt(trimmed, 10, ft.Bits())
		if err != nil {
			return fmt.Errorf("tam sayı değil: %q", trimmed)
		}
		fv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if trimmed == "" {
			fv.SetUint(0)
			return nil
		}
		n, err := strconv.ParseUint(trimmed, 10, ft.Bits())
		if err != nil {
			return fmt.Errorf("pozitif tam sayı değil: %q", trimmed)
		}
		fv.SetUint(n)
	case reflect.Float32, reflect.Float64:
		if trimmed == "" {
			fv.SetFloat(0)
			return nil
		}
		// Türkçe ondalık virgülü ("3,5") desteklenir; binlik ayraçlar desteklenmez.
		if !strings.Contains(trimmed, ".") && strings.Count(trimmed, ",") == 1 {
			trimmed = strings.Replace(trimmed, ",", ".", 1)
		}
		n, err := strconv.ParseFloat(trimmed, ft.Bits())
		if err != nil {
			return fmt.Errorf("sayı değil: %q", trimmed)
		}
		fv.SetFloat(n)
	default:
		return fmt.Errorf("desteklenmeyen alan tipi %s", ft)
	}
	return nil
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "1", "true", "on", "yes", "evet":
		return true, nil
	case "0", "false", "off", "no", "hayır", "hayir":
		return false, nil
	}
	return false, fmt.Errorf("mantıksal değer değil: %q", s)
}
