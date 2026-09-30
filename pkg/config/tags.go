package config

import "reflect"

// GetTag struct'ın field alanındaki tag anahtarının değerini döner.
// Gömülü (embedded) struct'lardan yükseltilen (promoted) alanlar da bulunur.
// v struct veya struct işaretçisi olmalıdır (nil işaretçi de kabul edilir).
//
//	GetTag(MyConfig{}, "FieldName", "validate")
func GetTag(v any, field, tag string) (string, bool) {
	rt := structType(v)
	if rt == nil {
		return "", false
	}
	if f, ok := rt.FieldByName(field); ok {
		return f.Tag.Lookup(tag)
	}
	return "", false
}

// GetTags dışa açık alanlar için alanAdı -> etiketDeğeri haritası döner.
// Etiketi olmayan alanlar atlanır.
//
// Gömülü struct'lar: etiketsiz gömülü bir struct'ın (dışa açık olmayan tip
// dahil) dışa açık alanları, encoding/json'daki gibi dış struct'ın alanıymış
// gibi eklenir; dış düzeydeki aynı adlı alan önceliklidir. Gömülü alanın
// kendisi tag anahtarına sahipse içine inilmez, alan kendi adıyla eklenir.
func GetTags(v any, tag string) map[string]string {
	out := map[string]string{}
	rt := structType(v)
	if rt == nil {
		return out
	}
	collectTags(rt, tag, out, map[reflect.Type]bool{})
	return out
}

func collectTags(rt reflect.Type, tag string, out map[string]string, visiting map[reflect.Type]bool) {
	if visiting[rt] {
		return
	}
	visiting[rt] = true
	defer delete(visiting, rt)

	// Önce bu düzeyin alanları (dış düzey önceliklidir), sonra gömülüler.
	names := map[string]bool{}
	var embedded []reflect.Type
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		names[f.Name] = true
		val, hasTag := f.Tag.Lookup(tag)
		if f.Anonymous && !hasTag {
			et := f.Type
			if et.Kind() == reflect.Pointer {
				et = et.Elem()
			}
			if et.Kind() == reflect.Struct {
				embedded = append(embedded, et)
				continue
			}
		}
		if !f.IsExported() {
			continue
		}
		if hasTag {
			out[f.Name] = val
		}
	}
	for _, et := range embedded {
		inner := map[string]string{}
		collectTags(et, tag, inner, visiting)
		for k, v := range inner {
			if names[k] {
				continue
			}
			if _, exists := out[k]; !exists {
				out[k] = v
			}
		}
	}
}

func structType(v any) reflect.Type {
	if v == nil {
		return nil
	}
	rt := reflect.TypeOf(v)
	for rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return nil
	}
	return rt
}
