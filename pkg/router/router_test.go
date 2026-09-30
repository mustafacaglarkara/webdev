package router

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/mux"
)

func TestReverseURL_NormalizeAndPositional(t *testing.T) {
	RegisterTemplate("u1", "/user/{id:[0-9]+}")
	got, err := ReverseURL("u1", "123")
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}
	if got != "/user/123" {
		t.Fatalf("unexpected URL: %s", got)
	}
}

func TestReverseURL_NamedMapAndEscape(t *testing.T) {
	RegisterTemplate("u2", "/file/{name}")
	got, err := ReverseURL("u2", map[string]string{"name": "a/b c"})
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}
	// a/b c -> a%2Fb%20c
	if got != "/file/a%2Fb%20c" {
		t.Fatalf("unexpected escaped URL: %s", got)
	}
}

func TestReverseURL_PrintfAndEscape(t *testing.T) {
	RegisterTemplate("p1", "/dl/%s")
	got, err := ReverseURL("p1", "a/b c")
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}
	if got != "/dl/a%2Fb%20c" {
		t.Fatalf("unexpected printf escaped url: %s", got)
	}
}

func check(t *testing.T, name string, want string, params ...any) {
	t.Helper()
	got, err := ReverseURL(name, params...)
	if err != nil {
		t.Fatalf("%s: beklenmeyen hata: %v", name, err)
	}
	if got != want {
		t.Fatalf("%s: got %q want %q", name, got, want)
	}
}

func checkErr(t *testing.T, name string, target error, params ...any) {
	t.Helper()
	_, err := ReverseURL(name, params...)
	if !errors.Is(err, target) {
		t.Fatalf("%s: %v hatası bekleniyordu, alınan: %v", name, target, err)
	}
}

// RT-1: '%' içeren desenler Sprintf'ten geçmez, bozulmaz.
func TestReverseURL_PercentInPattern(t *testing.T) {
	RegisterRoute("rt1.a", "/indirim/%50/{id}")
	check(t, "rt1.a", "/indirim/%50/7", "7")
	RegisterRoute("rt1.b", "/a%20b/%s/%d")
	check(t, "rt1.b", "/a%20b/x/5", "x", 5)
	RegisterRoute("rt1.c", "/yuzde/%%/%s")
	check(t, "rt1.c", "/yuzde/%/%C5%9F", "ş") // "%%" tek literal %
}

// RT-2: iç içe süslü parantezli regex.
func TestReverseURL_NestedBraces(t *testing.T) {
	RegisterRoute("rt2.a", "/d/{code:[a-z]{2,3}}/{n:\\d{1,4}}")
	check(t, "rt2.a", "/d/tr/42", "tr", 42)
	check(t, "rt2.a", "/d/en/1", map[string]any{"code": "en", "n": 1})
	RegisterRoute("rt2.b", "/x/{id:(?:[0-9]{2}){1,2}}/son")
	check(t, "rt2.b", "/x/1234/son", "1234")
	RegisterRoute("rt2.bad", "/x/{id:[0-9]{2}")
	checkErr(t, "rt2.bad", ErrInvalidPattern, "1")
}

// RT-3: fiber tarzı parametreler.
func TestReverseURL_FiberStyle(t *testing.T) {
	RegisterRoute("f.user", "/user/:id")
	check(t, "f.user", "/user/42", "42")
	check(t, "f.user", "/user/42", map[string]string{"id": "42"})

	RegisterRoute("f.opt", "/user/:id?")
	check(t, "f.opt", "/user/5", "5")
	check(t, "f.opt", "/user")
	check(t, "f.opt", "/user", map[string]string{})

	RegisterRoute("f.optmid", "/a/:x?/b/:y")
	check(t, "f.optmid", "/a/b/2", "2")
	check(t, "f.optmid", "/a/1/b/2", "1", "2")

	RegisterRoute("f.multi", "/flights/:from-:to")
	check(t, "f.multi", "/flights/IST-ESB", "IST", "ESB")

	RegisterRoute("f.wild", "/static/*")
	check(t, "f.wild", "/static/css/site%20ana.css", "css/site ana.css")
	check(t, "f.wild", "/static/", map[string]string{})
	check(t, "f.wild", "/static/js/a.js", map[string]string{"*": "js/a.js"})

	RegisterRoute("f.plus", "/files/+")
	check(t, "f.plus", "/files/a/b", "a/b")
	checkErr(t, "f.plus", ErrInvalidParams)

	RegisterRoute("f.greedy", "/docs/:path+")
	check(t, "f.greedy", "/docs/tr/giri%C5%9F", "tr/giriş")
}

// RT-4: eksik/fazla parametre hatadır; bulunamayan route ErrRouteNotFound döner.
func TestReverseURL_ParamCountAndNotFound(t *testing.T) {
	RegisterRoute("rt4", "/p/{a}/{b}")
	checkErr(t, "rt4", ErrInvalidParams, "1")
	checkErr(t, "rt4", ErrInvalidParams, "1", "2", "3")
	checkErr(t, "rt4", ErrInvalidParams, map[string]string{"a": "1"})
	checkErr(t, "rt4", ErrInvalidParams, map[string]string{"a": "1", "b": "2", "c": "3"})
	checkErr(t, "rt4", ErrInvalidParams, "", "2")
	checkErr(t, "rt4.yok", ErrRouteNotFound)

	RegisterRoute("rt4.static", "/hakkimizda")
	check(t, "rt4.static", "/hakkimizda")
	checkErr(t, "rt4.static", ErrInvalidParams, "fazla")

	if _, err := ReverseURLWithQuery("rt4.yok", nil, nil); !errors.Is(err, ErrRouteNotFound) {
		t.Fatalf("ReverseURLWithQuery ErrRouteNotFound döndürmeli: %v", err)
	}
}

func TestReverseURL_TurkishEscaping(t *testing.T) {
	RegisterRoute("tr.slug", "/blog/{slug}")
	check(t, "tr.slug", "/blog/%C3%A7a%C4%9Flar%3F", "çağlar?")
}

func TestReverseURLWithQuery(t *testing.T) {
	RegisterRoute("q.user", "/user/{id}")
	got, err := ReverseURLWithQuery("q.user", []any{"9"}, map[string]any{"tab": "profil", "etiket": []string{"a", "b"}, "bos": nil})
	if err != nil {
		t.Fatal(err)
	}
	if got != "/user/9?etiket=a&etiket=b&tab=profil" {
		t.Fatalf("query: %s", got)
	}
}

func TestReverseURLMust(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("panik bekleniyordu")
		}
	}()
	ReverseURLMust("olmayan.route")
}

func TestRegistryHelpers(t *testing.T) {
	RegisterRoute("reg.x", "/x")
	if p, ok := GetTemplate("reg.x"); !ok || p != "/x" {
		t.Fatal("GetTemplate")
	}
	if Routes()["reg.x"] != "/x" {
		t.Fatal("Routes")
	}
	UnregisterRoute("reg.x")
	if _, ok := GetTemplate("reg.x"); ok {
		t.Fatal("UnregisterRoute")
	}
}

func TestChiAdapter(t *testing.T) {
	r := chi.NewRouter()
	GetNamed(r, "/chi/user/{id:[0-9]+}", "chi.user", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(chi.URLParam(req, "id")))
	})
	PostNamed(r, "/chi/user", "chi.create", func(w http.ResponseWriter, req *http.Request) {})
	PutNamed(r, "/chi/user/{id}", "chi.update", func(w http.ResponseWriter, req *http.Request) {})
	DeleteNamed(r, "/chi/user/{id}", "chi.delete", func(w http.ResponseWriter, req *http.Request) {})
	u, err := ReverseURL("chi.user", 77)
	if err != nil || u != "/chi/user/77" {
		t.Fatalf("chi reverse: %q %v", u, err)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, u, nil))
	if rec.Body.String() != "77" {
		t.Fatalf("üretilen URL route ile eşleşmedi: %q", rec.Body.String())
	}
}

// A5-8: ChiGroup alt yönlendiricilerde tam yolu kaydeder; chi eşleşmesi göreli desenle sürer.
func TestChiGroupFullPath(t *testing.T) {
	r := chi.NewRouter()
	g := Chi(r)
	g.Get("/", "grp.home", func(w http.ResponseWriter, req *http.Request) { _, _ = w.Write([]byte("home")) })
	g.Route("/admin", func(a *ChiGroup) {
		a.Get("/", "grp.admin", func(w http.ResponseWriter, req *http.Request) { _, _ = w.Write([]byte("admin")) })
		a.Route("/users", func(u *ChiGroup) {
			u.Get("/{id:[0-9]+}", "grp.admin.user", func(w http.ResponseWriter, req *http.Request) {
				_, _ = w.Write([]byte("u" + chi.URLParam(req, "id")))
			})
			u.Group(func(x *ChiGroup) {
				x.Post("/", "grp.admin.user.create", func(w http.ResponseWriter, req *http.Request) {})
			})
		})
	})
	// Bağlı alt router (Mount) için önek elle verilir.
	sub := chi.NewRouter()
	Chi(sub, "/api/v1/").Delete("/items/{id}", "grp.api.item.delete", func(w http.ResponseWriter, req *http.Request) {})
	r.Mount("/api/v1", sub)

	cases := map[string]string{
		"grp.home":              "/",
		"grp.admin":             "/admin/",
		"grp.admin.user":        "/admin/users/5",
		"grp.admin.user.create": "/admin/users/",
		"grp.api.item.delete":   "/api/v1/items/5",
	}
	for name, want := range cases {
		var got string
		var err error
		if strings.Contains(want, "5") {
			got, err = ReverseURL(name, 5)
		} else {
			got, err = ReverseURL(name)
		}
		if err != nil || got != want {
			t.Fatalf("%s: got %q want %q (%v)", name, got, want, err)
		}
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/users/5", nil))
	if rec.Body.String() != "u5" {
		t.Fatalf("chi did not match generated url: %q", rec.Body.String())
	}
	if p := Chi(chi.NewRouter(), "admin").Prefix(); p != "/admin" {
		t.Fatalf("prefix normalize: %q", p)
	}
}

func TestMuxAdapter(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/mux/product/{sku:[A-Z]{3}[0-9]+}", func(w http.ResponseWriter, req *http.Request) {}).Name("mux.product")
	r.HandleFunc("/mux/isimsiz", func(w http.ResponseWriter, req *http.Request) {})
	if err := RegisterMuxRoutes(r); err != nil {
		t.Fatal(err)
	}
	u, err := ReverseURL("mux.product", map[string]string{"sku": "ABC123"})
	if err != nil || u != "/mux/product/ABC123" {
		t.Fatalf("mux reverse: %q %v", u, err)
	}
	var m mux.RouteMatch
	if !r.Match(httptest.NewRequest(http.MethodGet, u, nil), &m) {
		t.Fatal("üretilen URL mux route ile eşleşmedi")
	}
}

func TestConcurrentRegistry(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			RegisterRoute(fmt.Sprintf("c.%d", i%5), "/c/{id}")
		}(i)
		go func(i int) {
			defer wg.Done()
			_, _ = ReverseURL(fmt.Sprintf("c.%d", i%5), i)
			_ = Routes()
		}(i)
	}
	wg.Wait()
}
