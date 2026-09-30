package web

import (
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/mustafacaglarkara/webdev/pkg/security"
)

// ---- WEB-1 / WEB-2 / WEB-15: güvenli yönlendirme ----

func TestIsSafeRedirect(t *testing.T) {
	SetRedirectWhitelist([]string{"Example.com"})
	t.Cleanup(func() { SetRedirectWhitelist(nil) })
	safe := []string{"/", "/dashboard?x=1", "https://example.com/a", "http://app.example.com", ""}
	unsafe := []string{
		"//evil.com", "/\\evil.com", "\\\\evil.com", "/\t/evil.com", "/\n/evil.com",
		"https://evil.com", "javascript:alert(1)", "ftp://example.com/x", "data:text/html,x",
		"https://user@example.com", "foo/bar", "https://example.com.evil.com",
	}
	for _, s := range safe {
		if !IsSafeRedirect(s) {
			t.Errorf("expected safe: %q", s)
		}
	}
	for _, s := range unsafe {
		if IsSafeRedirect(s) {
			t.Errorf("expected unsafe: %q", s)
		}
	}
	if got := NormalizeSafeRedirect("//evil.com", "/home"); got != "/home" {
		t.Fatalf("normalize: %q", got)
	}
	if got := NormalizeSafeRedirect("  /ok  "); got != "/ok" {
		t.Fatalf("normalize trim: %q", got)
	}
	// A5-4: açıkça verilen boş fallback aynen döner; "next yok" ile "next=/" ayrılır.
	if got := NormalizeSafeRedirect("", ""); got != "" {
		t.Fatalf("empty fallback: %q", got)
	}
	if got := NormalizeSafeRedirect("//evil.com", ""); got != "" {
		t.Fatalf("empty fallback unsafe: %q", got)
	}
	if got := NormalizeSafeRedirect("/", ""); got != "/" {
		t.Fatalf("root next: %q", got)
	}
	if got := NormalizeSafeRedirect(""); got != "/" {
		t.Fatalf("default fallback: %q", got)
	}
}

// A5-6: yalnızca text/html (veya xhtml) içeren Accept HTML sayılır; boş ve */* sayılmaz.
func TestWantsHTML(t *testing.T) {
	yes := []string{"text/html", "text/html,application/xhtml+xml,*/*;q=0.8", "application/xhtml+xml"}
	no := []string{"", "*/*", "application/json", "application/json, */*"}
	for _, a := range yes {
		if !WantsHTML(a) {
			t.Errorf("expected html: %q", a)
		}
	}
	for _, a := range no {
		if WantsHTML(a) {
			t.Errorf("expected non-html: %q", a)
		}
	}
	// LoginRequired: Accept'siz istek 401 JSON alır, tarayıcı 302.
	SetAuthChecker(func(*http.Request) bool { return false })
	t.Cleanup(func() { SetAuthChecker(nil) })
	h := LoginRequired(func(w http.ResponseWriter, r *http.Request) {}, nil)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no accept: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.Header.Set("Accept", "text/html")
	h(rec, r)
	if rec.Code != http.StatusFound {
		t.Fatalf("html accept: %d", rec.Code)
	}
}

func TestRedirectWhitelistConcurrent(t *testing.T) {
	t.Cleanup(func() { SetRedirectWhitelist(nil) })
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); SetRedirectWhitelist([]string{"a.com"}) }()
		go func() { defer wg.Done(); _ = IsSafeRedirect("https://a.com") }()
	}
	wg.Wait()
}

// ---- WEB-4: fail-closed LoginRequired ----

func TestLoginRequiredFailClosed(t *testing.T) {
	SetAuthChecker(nil)
	called := false
	h := LoginRequired(func(w http.ResponseWriter, r *http.Request) { called = true }, nil)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/secret", nil)
	r.Header.Set("Accept", "application/json")
	h(rec, r)
	if called || rec.Code != http.StatusUnauthorized {
		t.Fatalf("no checker must deny: called=%v code=%d", called, rec.Code)
	}
	mw := LoginRequiredMiddleware()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	rec = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/secret", nil)
	r.Header.Set("Accept", "text/html") // tarayıcı isteği → 302 (A5-6: Accept'siz istek 401 alır)
	mw.ServeHTTP(rec, r)
	if called || rec.Code != http.StatusFound {
		t.Fatalf("middleware no checker: called=%v code=%d", called, rec.Code)
	}
	SetAuthChecker(func(*http.Request) bool { return true })
	t.Cleanup(func() { SetAuthChecker(nil) })
	rec = httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/secret", nil))
	if !called {
		t.Fatal("checker true must allow")
	}
}

// ---- WEB-16: tag registry ----

func TestCallTagErrors(t *testing.T) {
	RegisterTag("upper", strings.ToUpper)
	RegisterTag("add", func(a, b int) int { return a + b })
	RegisterTag("join", func(sep string, parts ...string) string { return strings.Join(parts, sep) })
	RegisterTag("fail", func() (string, error) { return "", errors.New("secret internal detail") })
	RegisterTag("boom", func() string { panic("kaboom") })
	t.Cleanup(func() {
		for _, n := range []string{"upper", "add", "join", "fail", "boom"} {
			UnregisterTag(n)
		}
	})

	if v := CallTag("upper", "abc"); v != "ABC" {
		t.Fatalf("upper=%v", v)
	}
	if v := CallTag("add", 1, int64(2)); v != 3 {
		t.Fatalf("add=%v", v)
	}
	if v := CallTag("join", "-", "a", "b"); v != "a-b" {
		t.Fatalf("join=%v", v)
	}
	bad := [][]any{
		{"upper"},             // eksik argüman
		{"upper", "a", "b"},   // fazla argüman
		{"add", "x", 1},       // tip uyuşmazlığı (panik yerine hata)
		{"upper", 65},         // int → string rune çevrimi yasak
		{"join", "-", 1, "b"}, // variadic tip uyuşmazlığı
		{"join"},              // variadic eksik
		{"fail"},              // tag hatası
		{"boom"},              // tag paniği
		{"yok"},               // kayıtsız
	}
	for _, args := range bad {
		name := args[0].(string)
		if _, err := CallTagE(name, args[1:]...); err == nil {
			t.Errorf("CallTagE(%v) expected error", args)
		}
		if v := CallTag(name, args[1:]...); v != "" {
			t.Errorf("CallTag(%v) = %v, want \"\" (no internal error text)", args, v)
		}
	}
	// html/template üzerinden de sayfaya hata metni basılmaz
	tpl := template.Must(template.New("x").Funcs(TemplateFuncs(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))).Parse(`[{{ fail }}]`))
	var sb strings.Builder
	if err := tpl.Execute(&sb, nil); err != nil {
		t.Fatal(err)
	}
	if sb.String() != "[]" {
		t.Fatalf("template leaked: %q", sb.String())
	}
}

// ---- rol çıkarımı / Can / menü ----

type roleUser struct{ r string }

func (u roleUser) GetRole() string { return u.r }

func TestExtractUserRoleAndCan(t *testing.T) {
	if ExtractUserRole(map[string]any{"role": "admin"}, "guest") != "admin" ||
		ExtractUserRole(map[string]string{"role": "editor"}, "guest") != "editor" ||
		ExtractUserRole(roleUser{"ops"}, "guest") != "ops" ||
		ExtractUserRole(nil, "guest") != "guest" ||
		ExtractUserRole(map[string]any{"role": 5}, "guest") != "guest" {
		t.Fatal("ExtractUserRole")
	}
	if !HasRole(map[string]any{"role": "admin"}, "admin") || HasRole(map[string]any{}, "") {
		t.Fatal("HasRole")
	}

	SetCanChecker(nil)
	if Can(map[string]any{"role": "admin"}, "/admin", "GET") {
		t.Fatal("no checker must deny")
	}
	var gotSub string
	SetCanChecker(func(sub, obj, act string) (bool, error) {
		gotSub = sub
		if sub == "err" {
			return true, errors.New("x")
		}
		return sub == "admin", nil
	})
	t.Cleanup(func() { SetCanChecker(nil) })
	if !Can(map[string]any{"role": "admin"}, "/admin", "GET") || Can(nil, "/admin", "GET") || gotSub != GuestRole {
		t.Fatal("Can")
	}
	if Can(map[string]any{"role": "err"}, "/x", "GET") {
		t.Fatal("checker error must deny")
	}

	items := []MenuItem{
		{LabelKey: "home", URL: "/"},
		{LabelKey: "admin", URL: "/admin", Object: "/admin", Action: "GET"},
		{LabelKey: "grp", HideIfEmptyChildren: true, Children: []MenuItem{{LabelKey: "c", URL: "/admin/c", Object: "/admin", Action: "GET"}}},
	}
	if got := BuildMenu("/", nil, items); len(got) != 1 {
		t.Fatalf("guest menu: %+v", got)
	}
	got := BuildMenu("/admin/c", map[string]any{"role": "admin"}, items)
	if len(got) != 3 || !got[1].Active || !got[2].Active {
		t.Fatalf("admin menu: %+v", got)
	}
}

// ---- WEB-19: yardımcı haritası ----

func TestJetGlobalHelpersBuiltOnceAndCloned(t *testing.T) {
	a := JetGlobalHelpers()
	a["static"] = "mutated"
	b := JetGlobalHelpers()
	if _, ok := b["static"].(func(string, ...any) string); !ok {
		t.Fatal("helpers map shared between callers")
	}
	if got := b["assets"].(func(string, ...any) string)("css/app.css", 3); got != "/static/css/app.css?v=3" {
		t.Fatalf("assets=%q", got)
	}
	if got := b["old"].(func(any, string) string)(url.Values{"x": {"1"}}, "x"); got != "1" {
		t.Fatalf("old=%q", got)
	}
}

func TestRequestLangsUsesSessionPreference(t *testing.T) {
	InitSessionStore([]byte(testKey))
	r, w := newRequest(http.MethodGet, "/", "")
	if err := SetPreferredLang(w, r, "de"); err != nil {
		t.Fatal(err)
	}
	if err := SetPreferredLang(w, r, "x\"><script>"); !errors.Is(err, ErrInvalidLang) {
		t.Fatalf("invalid lang accepted: %v", err)
	}
	r2 := next(t, w)
	r2.Header.Set("Accept-Language", "en-US,en;q=0.9")
	langs := RequestLangs(r2, "tr")
	if len(langs) == 0 || langs[0] != "de" {
		t.Fatalf("langs=%v", langs)
	}
}

// ---- WEB-17 / ARCH-3: CSRF (net/http, oturum tabanlı, security çekirdeği) ----

func TestCSRFMiddlewareSessionToken(t *testing.T) {
	InitSessionStore([]byte(testKey))
	var tok string
	h := CSRFMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok = security.CSRFTokenFromContext(r.Context())
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/form", nil))
	if tok == "" {
		t.Fatal("no token")
	}
	post := func(provided string) int {
		form := url.Values{"csrf_token": {provided}}
		r, w := newRequest(http.MethodPost, "/form", form.Encode())
		applyCookies(rec, r)
		h.ServeHTTP(w, r)
		return w.Code
	}
	if c := post(tok); c != 200 {
		t.Fatalf("valid token: %d", c)
	}
	wrong := []byte(tok)
	if wrong[0] == 'a' {
		wrong[0] = 'b'
	} else {
		wrong[0] = 'a'
	}
	if c := post(string(wrong)); c != 403 {
		t.Fatalf("wrong token: %d", c)
	}
	if c := post(""); c != 403 {
		t.Fatalf("missing token: %d", c)
	}
	r, _ := newRequest(http.MethodPost, "/form", "")
	applyCookies(rec, r)
	if err := VerifyCSRFToken(r, tok); err != nil {
		t.Fatalf("VerifyCSRFToken: %v", err)
	}
}

func TestSanitizeFilterUsesSecurity(t *testing.T) {
	f := JetTemplateFilters()["sanitize"].(func(string, ...string) template.HTML)
	if got := f(`<a href="javascript:x">l</a><script>x</script>`); strings.Contains(string(got), "script") || strings.Contains(string(got), "javascript") {
		t.Fatalf("relaxed: %q", got)
	}
	if got := f("<b>x</b>", "strict"); got != "x" {
		t.Fatalf("strict: %q", got)
	}
}
