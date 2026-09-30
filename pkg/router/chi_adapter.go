package router

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// HandleNamed registers a route on chi.Router and also registers the name->pattern mapping
// so ReverseURL / ReverseURLWithQuery can resolve it. method should be UPPERCASE (GET, POST...).
//
// Not: chi alt yönlendiricilerinde (r.Route / r.Mount) yalnızca göreli desen kaydedilir;
// tam yol için Chi(r, "/prefix") ile ChiGroup kullanın (A5-8).
func HandleNamed(r chi.Router, method, pattern, name string, h http.HandlerFunc) {
	RegisterRoute(name, pattern)
	r.Method(method, pattern, h)
}

// Helper shortcuts
func GetNamed(r chi.Router, pattern, name string, h http.HandlerFunc) {
	HandleNamed(r, "GET", pattern, name, h)
}
func PostNamed(r chi.Router, pattern, name string, h http.HandlerFunc) {
	HandleNamed(r, "POST", pattern, name, h)
}
func PutNamed(r chi.Router, pattern, name string, h http.HandlerFunc) {
	HandleNamed(r, "PUT", pattern, name, h)
}
func DeleteNamed(r chi.Router, pattern, name string, h http.HandlerFunc) {
	HandleNamed(r, "DELETE", pattern, name, h)
}

// ChiGroup, bir chi.Router'ı bağlı olduğu yol önekiyle birlikte taşır. Rotalar chi'ye
// göreli desenle, ters URL kayıt defterine ise TAM yolla (önek + desen) yazılır; böylece
// alt yönlendiricilerde ReverseURL doğru sonuç verir (A5-8).
type ChiGroup struct {
	r      chi.Router
	prefix string
}

// Chi, r için ChiGroup oluşturur. prefix, r'nin bağlı olduğu yol önekidir (kök için boş
// bırakın). Alt yönlendiriciler için Route kullanın; önek otomatik birleştirilir.
func Chi(r chi.Router, prefix ...string) *ChiGroup {
	g := &ChiGroup{r: r}
	if len(prefix) > 0 {
		g.prefix = cleanPrefix(prefix[0])
	}
	return g
}

// Router alttaki chi.Router'ı döner.
func (g *ChiGroup) Router() chi.Router { return g.r }

// Prefix grubun tam yol önekini döner (kök için "").
func (g *ChiGroup) Prefix() string { return g.prefix }

// Handle rotayı chi'ye göreli desenle, kayıt defterine tam yolla kaydeder.
func (g *ChiGroup) Handle(method, pattern, name string, h http.HandlerFunc) {
	RegisterRoute(name, g.Full(pattern))
	g.r.Method(method, pattern, h)
}

func (g *ChiGroup) Get(pattern, name string, h http.HandlerFunc) {
	g.Handle(http.MethodGet, pattern, name, h)
}
func (g *ChiGroup) Post(pattern, name string, h http.HandlerFunc) {
	g.Handle(http.MethodPost, pattern, name, h)
}
func (g *ChiGroup) Put(pattern, name string, h http.HandlerFunc) {
	g.Handle(http.MethodPut, pattern, name, h)
}
func (g *ChiGroup) Patch(pattern, name string, h http.HandlerFunc) {
	g.Handle(http.MethodPatch, pattern, name, h)
}
func (g *ChiGroup) Delete(pattern, name string, h http.HandlerFunc) {
	g.Handle(http.MethodDelete, pattern, name, h)
}

// Route, chi.Router.Route gibi alt yönlendirici oluşturur; fn'e öneki birleştirilmiş
// yeni bir ChiGroup verir.
func (g *ChiGroup) Route(pattern string, fn func(*ChiGroup)) chi.Router {
	return g.r.Route(pattern, func(sub chi.Router) {
		fn(&ChiGroup{r: sub, prefix: g.Full(pattern)})
	})
}

// Group, chi.Router.Group gibi aynı önekte inline grup açar (middleware kapsamı için).
func (g *ChiGroup) Group(fn func(*ChiGroup)) chi.Router {
	return g.r.Group(func(sub chi.Router) {
		fn(&ChiGroup{r: sub, prefix: g.prefix})
	})
}

// Full, göreli deseni grubun önekiyle birleştirip tam yolu döner.
func (g *ChiGroup) Full(pattern string) string {
	if g.prefix == "" {
		return pattern
	}
	if pattern == "" || pattern == "/" {
		return g.prefix + pattern
	}
	return g.prefix + "/" + strings.TrimPrefix(pattern, "/")
}

// cleanPrefix: "/" ile başlar, sonda "/" olmaz; "/" ve "" → "".
func cleanPrefix(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimSuffix(p, "/")
	if p == "" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}
