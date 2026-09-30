package policy

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"testing/fstest"

	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
)

const (
	modelPath  = "testdata/model.conf"
	policyPath = "testdata/policy.csv"
)

func resetDefault(t *testing.T) {
	t.Helper()
	defaultMu.Lock()
	defaultManager = nil
	pendingAdapter = nil
	defaultMu.Unlock()
	t.Cleanup(func() {
		defaultMu.Lock()
		defaultManager = nil
		pendingAdapter = nil
		defaultMu.Unlock()
	})
}

func TestEnforceNotInitialised(t *testing.T) {
	resetDefault(t)
	ok, err := Enforce("admin", "/admin", "GET")
	if ok || !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("got %v %v", ok, err)
	}
	if err := Reload(); !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("Reload: %v", err)
	}
	if _, err := AddPolicyRule("a", "b", "c"); !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("AddPolicyRule: %v", err)
	}
}

// POL-1: middleware Init'ten önce oluşturulsa bile enforcer yokken reddeder, sonra çalışır.
func TestDefaultMiddlewareFailClosedThenWorks(t *testing.T) {
	resetDefault(t)
	mw := DefaultMiddleware(func(r *http.Request) any { return r.Header.Get("X-Role") }, nil, nil)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))

	req := func(role, path string) int {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.Header.Set("X-Role", role)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		return rec.Code
	}
	if c := req("admin", "/admin"); c != 403 {
		t.Fatalf("uninitialised: code=%d want 403", c)
	}
	if err := Init(modelPath, policyPath); err != nil {
		t.Fatal(err)
	}
	if c := req("admin", "/admin/users"); c != 200 {
		t.Fatalf("admin: code=%d", c)
	}
	if c := req("guest", "/admin"); c != 403 {
		t.Fatalf("guest: code=%d", c)
	}
}

// POL-4: nil fonksiyonlar varsayılanları kullanır; Enforce hatası 500 döner.
func TestMiddlewareNilFuncsAndEnforceError(t *testing.T) {
	m, err := New(modelPath, policyPath)
	if err != nil {
		t.Fatal(err)
	}
	h := m.Middleware(nil, nil, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/public", nil)) // subject → guest
	if rec.Code != 200 {
		t.Fatalf("guest /public: code=%d", rec.Code)
	}

	rm, err := New("testdata/regex_model.conf", policyPath)
	if err != nil {
		t.Fatal(err)
	}
	h = rm.Middleware(nil, func(*http.Request) any { return 42 }, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler must not run")
	}))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rec.Code != 500 {
		t.Fatalf("enforce error: code=%d want 500", rec.Code)
	}
}

// POL-2: eşzamanlı Enforce / AddPolicyRule / Reload / SetDefault yarışsız (-race).
func TestConcurrentUse(t *testing.T) {
	resetDefault(t)
	if err := Init(modelPath, policyPath); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_, _ = Enforce("admin", "/admin", "GET")
				_ = PolicyStats()
			}
		}()
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_, _ = AddPolicyRule(fmt.Sprintf("u%d", i), fmt.Sprintf("/r%d", j), "GET")
				_, _ = RemovePolicyRule(fmt.Sprintf("u%d", i), fmt.Sprintf("/r%d", j), "GET")
			}
		}(i)
		go func() {
			defer wg.Done()
			_ = LastReload()
			_ = DefaultManager()
		}()
	}
	wg.Wait()
	// Reload dosya adaptöründen yeniden yükler (eklenen kurallar dosyaya yazılmadıysa kaybolur)
	if err := Reload(); err != nil {
		t.Fatal(err)
	}
	if ok, _ := Enforce("admin", "/admin", "GET"); !ok {
		t.Fatal("policy lost after reload")
	}
}

// POL-3: SetAdapter enforcer'dan önce çağrılırsa Init onu uygular.
func TestPendingAdapterAppliedOnInit(t *testing.T) {
	resetDefault(t)
	if err := SetAdapter(nil); err == nil {
		t.Fatal("nil adapter must error")
	}
	if err := SetAdapter(fileadapter.NewAdapter(policyPath)); err != nil {
		t.Fatal(err)
	}
	if err := Init(modelPath, ""); err != nil {
		t.Fatal(err)
	}
	if ok, err := Enforce("guest", "/public", "GET"); !ok || err != nil {
		t.Fatalf("pending adapter not applied: %v %v", ok, err)
	}
	if len(GetPolicyRules()) != 3 {
		t.Fatalf("rules=%v", GetPolicyRules())
	}
}

// A5-3: model/politika gömülü dosyadan veya metinden yüklenir; Reload dosyayı yeniden okur.
func TestNewFSAndTextReload(t *testing.T) {
	modelText, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatal(err)
	}
	policyText, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatal(err)
	}
	fsys := fstest.MapFS{
		"conf/model.conf": {Data: modelText},
		"conf/policy.csv": {Data: policyText},
	}
	m, err := NewFS(fsys, "conf/model.conf", "conf/policy.csv")
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := m.Enforce("admin", "/admin/x", "GET"); !ok || err != nil {
		t.Fatalf("fs policy: %v %v", ok, err)
	}
	if rules, _ := m.Policies(); len(rules) != 3 {
		t.Fatalf("rules=%v", rules)
	}
	// AddPolicy bellekte çalışır (salt okunur adaptör hatası yutulur).
	if added, err := m.AddPolicy("editor", "/posts", "POST"); err != nil || !added {
		t.Fatalf("add: %v %v", added, err)
	}
	// Dosya değişti → Reload yeni politikayı okur, bellekteki ekleme gider.
	fsys["conf/policy.csv"] = &fstest.MapFile{Data: []byte("p, guest, /only, GET\n# yorum\n\n")}
	before := m.LastReload()
	if err := m.Reload(); err != nil {
		t.Fatal(err)
	}
	if !m.LastReload().After(before) && m.LastReload() != before {
		t.Fatal("LastReload not updated")
	}
	if ok, _ := m.Enforce("admin", "/admin/x", "GET"); ok {
		t.Fatal("old rule survived reload")
	}
	if ok, _ := m.Enforce("guest", "/only", "GET"); !ok {
		t.Fatal("new rule not loaded")
	}
	if ok, _ := m.Enforce("editor", "/posts", "POST"); ok {
		t.Fatal("in-memory rule must be dropped on reload")
	}
	// Model bozulursa Reload hata döner.
	fsys["conf/model.conf"] = &fstest.MapFile{Data: []byte("[request_definition]\nr = sub")}
	if err := m.Reload(); err == nil {
		t.Fatal("broken model must error")
	}

	// Metin tabanlı.
	tm, err := NewFromText(string(modelText), string(policyText))
	if err != nil {
		t.Fatal(err)
	}
	if ok, _ := tm.Enforce("guest", "/public", "GET"); !ok {
		t.Fatal("text policy not enforced")
	}
	if err := tm.Reload(); err != nil {
		t.Fatalf("text reload: %v", err)
	}
	if ok, _ := tm.Enforce("guest", "/public", "GET"); !ok {
		t.Fatal("text policy lost after reload")
	}
	// Politika metni boş → bellek içi, AddPolicy çalışır.
	em, err := NewFromText(string(modelText), "")
	if err != nil {
		t.Fatal(err)
	}
	if ok, _ := em.Enforce("guest", "/public", "GET"); ok {
		t.Fatal("empty policy must deny")
	}
	if _, err := em.AddPolicy("guest", "/public", "GET"); err != nil {
		t.Fatal(err)
	}
	if ok, _ := em.Enforce("guest", "/public", "GET"); !ok {
		t.Fatal("in-memory add failed")
	}
	// Hatalı girdiler.
	if _, err := NewFromText("", ""); err == nil {
		t.Fatal("empty model must error")
	}
	if _, err := NewFS(fsys, "../model.conf", ""); err == nil {
		t.Fatal("invalid path must error")
	}
	if _, err := NewFS(fsys, "conf/missing.conf", ""); err == nil {
		t.Fatal("missing model must error")
	}
	if _, err := NewWithModel(nil, nil); err == nil {
		t.Fatal("nil loader must error")
	}
	if _, err := NewFSWithAdapter(fsys, "conf/model.conf", nil); err == nil {
		t.Fatal("nil adapter must error")
	}
	// Varsayılan olarak da kullanılabilir (SetDefault) ve Reload gömülü dosyayı okur.
	resetDefault(t)
	fsys["conf/model.conf"] = &fstest.MapFile{Data: modelText}
	dm, err := NewFS(fsys, "conf/model.conf", "conf/policy.csv")
	if err != nil {
		t.Fatal(err)
	}
	SetDefault(dm)
	if err := Reload(); err != nil {
		t.Fatal(err)
	}
	if ok, _ := Check("guest", "/only", "GET"); !ok {
		t.Fatal("default manager from fs not working")
	}
}

func TestCheckMatchesCanCheckerSignature(t *testing.T) {
	resetDefault(t)
	var fn func(string, string, string) (bool, error) = Check
	if ok, err := fn("admin", "/admin", "GET"); ok || !errors.Is(err, ErrNotInitialized) {
		t.Fatal("Check must fail closed")
	}
}

func TestAddPolicyValidation(t *testing.T) {
	m, err := New(modelPath, policyPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.AddPolicy("", "/x", "GET"); err == nil {
		t.Fatal("empty field must error")
	}
	added, err := m.AddPolicy("editor", "/posts", "POST")
	if err != nil || !added {
		t.Fatalf("add: %v %v", added, err)
	}
	if ok, _ := m.Enforce("editor", "/posts", "POST"); !ok {
		t.Fatal("added rule not enforced")
	}
}
