// Package router, isimli route kaydı ve ters URL (reverse URL) üretimi sağlar.
//
// Desteklenen desen sözdizimleri:
//
//	chi / gorilla mux : /user/{id}, /user/{id:[0-9]+}, /d/{code:[a-z]{2,3}}
//	fiber             : /user/:id, /user/:id?, /files/*, /files/+
//	eski printf       : /dl/%s (yalnızca başka parametre yoksa; kullanımdan kaldırıldı)
//
// Kayıt defteri (registry) eşzamanlı kullanım için güvenlidir.
package router

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

var (
	tmplMu    sync.RWMutex
	templates = make(map[string]string)
)

// Convenience error values
var (
	// ErrRouteNotFound, isimli route kayıtlı değilse döner (errors.Is ile kontrol edin).
	ErrRouteNotFound = errors.New("route not found")
	// ErrInvalidParams, parametre sayısı/adları desenle uyuşmadığında döner.
	ErrInvalidParams = errors.New("invalid route params")
	// ErrInvalidPattern, desen ayrıştırılamadığında döner (örn. kapanmamış '{').
	ErrInvalidPattern = errors.New("invalid route pattern")
)

// registerTemplate stores the route pattern template for reverse lookups.
func registerTemplate(name, pattern string) {
	tmplMu.Lock()
	defer tmplMu.Unlock()
	templates[name] = pattern
}

// RegisterTemplate exposes registerTemplate to other packages.
func RegisterTemplate(name, pattern string) {
	registerTemplate(name, pattern)
}

// RegisterRoute stores a route template for reverse URL lookups. Same as RegisterTemplate.
func RegisterRoute(name, pattern string) {
	registerTemplate(name, pattern)
}

// UnregisterRoute, isimli route kaydını siler (yoksa bir şey yapmaz).
func UnregisterRoute(name string) {
	tmplMu.Lock()
	defer tmplMu.Unlock()
	delete(templates, name)
}

// GetTemplate returns the registered template for a route name.
func GetTemplate(name string) (string, bool) {
	tmplMu.RLock()
	defer tmplMu.RUnlock()
	p, ok := templates[name]
	return p, ok
}

// Routes, kayıtlı tüm route'ların (ad -> desen) bir kopyasını döner.
func Routes() map[string]string {
	tmplMu.RLock()
	defer tmplMu.RUnlock()
	out := make(map[string]string, len(templates))
	for k, v := range templates {
		out[k] = v
	}
	return out
}

// ---- desen ayrıştırma ----

type paramKind int

const (
	pNamed    paramKind = iota // {id}, {id:re}, :id
	pOptional                  // :id?
	pWildcard                  // chi /*, fiber * ve :ad* (isteğe bağlı)
	pPrintf                    // %s, %d, %v
)

type segment struct {
	literal string
	isParam bool
	name    string
	kind    paramKind
	greedy  bool // değerdeki '/' korunur (joker parametreler)
}

func (s segment) required() bool { return s.kind == pNamed || s.kind == pPrintf }

func isNameByte(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// parsePattern, deseni sabit metin ve parametre parçalarına ayırır.
// {ad:regex} içindeki regex iç içe süslü parantez ve `\{` kaçışları içerebilir.
func parsePattern(p string) ([]segment, error) {
	var segs []segment
	var lit strings.Builder
	flush := func() {
		if lit.Len() > 0 {
			segs = append(segs, segment{literal: lit.String()})
			lit.Reset()
		}
	}
	for i := 0; i < len(p); i++ {
		c := p[i]
		switch {
		case c == '{':
			depth := 1
			j := i + 1
			colon := -1
			for ; j < len(p) && depth > 0; j++ {
				switch p[j] {
				case '\\':
					j++ // kaçışlı karakteri atla
				case '{':
					depth++
				case '}':
					depth--
				case ':':
					if depth == 1 && colon < 0 {
						colon = j
					}
				}
			}
			if depth != 0 {
				return nil, fmt.Errorf("%w: kapanmamış '{' (%q)", ErrInvalidPattern, p)
			}
			end := j - 1 // kapanış '}' indeksi
			nameEnd := end
			if colon >= 0 {
				nameEnd = colon
			}
			name := strings.TrimSpace(p[i+1 : nameEnd])
			if name == "" {
				return nil, fmt.Errorf("%w: boş parametre adı (%q)", ErrInvalidPattern, p)
			}
			flush()
			segs = append(segs, segment{isParam: true, name: name, kind: pNamed})
			i = end
		case c == ':' && i+1 < len(p) && isNameByte(p[i+1]) && (i == 0 || strings.IndexByte("/-.", p[i-1]) >= 0):
			// fiber: :ad, :ad? (isteğe bağlı), :ad+ / :ad* (açgözlü)
			j := i + 1
			for j < len(p) && isNameByte(p[j]) {
				j++
			}
			seg := segment{isParam: true, name: p[i+1 : j], kind: pNamed}
			if j < len(p) {
				switch p[j] {
				case '?':
					seg.kind = pOptional
					j++
				case '+':
					seg.greedy = true
					j++
				case '*':
					seg.kind, seg.greedy = pWildcard, true
					j++
				}
			}
			flush()
			segs = append(segs, seg)
			i = j - 1
		case (c == '*' || c == '+') && (i == 0 || p[i-1] == '/'):
			// chi "/*", fiber "*" (isteğe bağlı) ve "+" (zorunlu); anahtar "*" veya "+"
			seg := segment{isParam: true, name: string(c), kind: pWildcard, greedy: true}
			if c == '+' {
				seg.kind = pNamed
			}
			flush()
			segs = append(segs, seg)
		default:
			lit.WriteByte(c)
		}
	}
	flush()
	return segs, nil
}

// parsePrintf, yalnızca %s/%d/%v fiilleri içeren eski desenleri ayrıştırır.
// "%%" tek '%' olur; diğer '%' dizileri (örn. "%20") olduğu gibi kalır.
func parsePrintf(p string) ([]segment, int) {
	var segs []segment
	var lit strings.Builder
	n := 0
	for i := 0; i < len(p); i++ {
		if p[i] == '%' && i+1 < len(p) {
			switch p[i+1] {
			case 's', 'd', 'v':
				if lit.Len() > 0 {
					segs = append(segs, segment{literal: lit.String()})
					lit.Reset()
				}
				segs = append(segs, segment{isParam: true, name: fmt.Sprintf("%d", n), kind: pPrintf})
				n++
				i++
				continue
			case '%':
				lit.WriteByte('%')
				i++
				continue
			}
		}
		lit.WriteByte(p[i])
	}
	if lit.Len() > 0 {
		segs = append(segs, segment{literal: lit.String()})
	}
	return segs, n
}

// ReverseURL builds a URL from a named route template.
//
// params şunlardan biri olabilir:
//   - tek bir map[string]string / map[string]any: ada göre doldurma
//     (joker '*' parametresi için anahtar "*" kullanılır);
//   - sıralı parametreler: desendeki parametreler sırayla doldurulur.
//
// Kurallar:
//   - Route yoksa hata errors.Is(err, ErrRouteNotFound) sağlar.
//   - Eksik zorunlu veya fazla parametre ErrInvalidParams ile hata döner.
//   - İsteğe bağlı fiber parametresi (:id?) verilmezse önündeki '/' ile birlikte düşer.
//   - Değerler url.PathEscape ile kaçışlanır; joker (*) değerlerde '/' korunur.
//   - Desen asla fmt.Sprintf'ten geçirilmez; '%' içeren desenler bozulmaz.
//
// Örnekler:
//
//	ReverseURL("user.show", map[string]string{"id": "123"}) // /user/123
//	ReverseURL("user.show", "123")                          // /user/123
func ReverseURL(name string, params ...any) (string, error) {
	pattern, ok := GetTemplate(name)
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrRouteNotFound, name)
	}
	segs, err := parsePattern(pattern)
	if err != nil {
		return "", fmt.Errorf("route %s: %w", name, err)
	}
	nParams := 0
	for _, s := range segs {
		if s.isParam {
			nParams++
		}
	}
	// Eski printf desenleri: yalnızca başka parametre yoksa ve sıralı argüman verildiyse.
	if nParams == 0 && len(params) > 0 && !isMap(params) {
		if ps, n := parsePrintf(pattern); n > 0 {
			segs, nParams = ps, n
		}
	}
	if nParams == 0 {
		if len(params) > 0 && !(len(params) == 1 && isEmptyMap(params[0])) {
			return "", fmt.Errorf("%w: route %s parametre almaz, %d verildi", ErrInvalidParams, name, len(params))
		}
		return pattern, nil
	}

	values := make([]*string, len(segs))
	if len(params) == 1 && isMap(params) {
		m := toStringMap(params[0])
		used := 0
		for i, s := range segs {
			if !s.isParam {
				continue
			}
			if v, ok := m[s.name]; ok {
				vv := v
				values[i] = &vv
				used++
			} else if s.required() {
				return "", fmt.Errorf("%w: route %s için '%s' parametresi eksik", ErrInvalidParams, name, s.name)
			}
		}
		if used != len(m) {
			return "", fmt.Errorf("%w: route %s için bilinmeyen parametre(ler) verildi", ErrInvalidParams, name)
		}
	} else {
		required := 0
		for _, s := range segs {
			if s.isParam && s.required() {
				required++
			}
		}
		if len(params) < required || len(params) > nParams {
			return "", fmt.Errorf("%w: route %s için %d-%d parametre gerekli, %d verildi", ErrInvalidParams, name, required, nParams, len(params))
		}
		// Zorunlu parametreler her zaman doldurulur; isteğe bağlılar kalan
		// argüman sayısı kadar soldan sağa doldurulur.
		spare := len(params) - required
		k := 0
		for i, s := range segs {
			if !s.isParam {
				continue
			}
			if !s.required() {
				if spare == 0 {
					continue
				}
				spare--
			}
			v := fmt.Sprint(params[k])
			values[i] = &v
			k++
		}
	}

	var b strings.Builder
	for i, s := range segs {
		if !s.isParam {
			b.WriteString(s.literal)
			continue
		}
		v := values[i]
		if v == nil || (s.kind == pOptional && *v == "") {
			if s.kind == pOptional {
				// "/user/:id?" -> "/user"
				out := b.String()
				if strings.HasSuffix(out, "/") && (i+1 >= len(segs) || strings.HasPrefix(segs[i+1].literal, "/")) {
					b.Reset()
					b.WriteString(strings.TrimSuffix(out, "/"))
				}
			}
			continue
		}
		if s.kind == pNamed && *v == "" {
			return "", fmt.Errorf("%w: route %s için '%s' parametresi boş", ErrInvalidParams, name, s.name)
		}
		if s.greedy {
			b.WriteString(escapePathKeepSlash(*v))
		} else {
			b.WriteString(url.PathEscape(*v))
		}
	}
	out := b.String()
	if out == "" {
		out = "/"
	}
	return out, nil
}

func isMap(params []any) bool {
	if len(params) != 1 {
		return false
	}
	switch params[0].(type) {
	case map[string]string, map[string]any:
		return true
	}
	return false
}

func isEmptyMap(p any) bool {
	switch m := p.(type) {
	case map[string]string:
		return len(m) == 0
	case map[string]any:
		return len(m) == 0
	}
	return false
}

func toStringMap(p any) map[string]string {
	switch m := p.(type) {
	case map[string]string:
		return m
	case map[string]any:
		sm := make(map[string]string, len(m))
		for k, v := range m {
			sm[k] = fmt.Sprint(v)
		}
		return sm
	}
	return nil
}

func escapePathKeepSlash(s string) string {
	parts := strings.Split(s, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}

// ReverseURLMust is like ReverseURL but panics on error (convenience)
func ReverseURLMust(name string, params ...any) string {
	s, err := ReverseURL(name, params...)
	if err != nil {
		panic(err)
	}
	return s
}

// ReverseURLWithQuery: ReverseURL + query ekleme (query map[string]any; slice değerler çoklu param olur)
func ReverseURLWithQuery(name string, params []any, query map[string]any) (string, error) {
	base, err := ReverseURL(name, params...)
	if err != nil {
		return "", err
	}
	if len(query) == 0 {
		return base, nil
	}
	vals := url.Values{}
	for k, v := range query {
		if v == nil {
			continue
		}
		switch t := v.(type) {
		case []string:
			for _, s := range t {
				vals.Add(k, s)
			}
		case []any:
			for _, iv := range t {
				vals.Add(k, fmt.Sprint(iv))
			}
		default:
			vals.Add(k, fmt.Sprint(t))
		}
	}
	enc := vals.Encode()
	if enc == "" {
		return base, nil
	}
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return base + sep + enc, nil
}
