// Command crm, github.com/mustafacaglarkara/webdev kütüphanesini gerçek bir Fiber v2 +
// Jet uygulamasında gösteren örnek CRM'dir: oturum, CSRF korumalı giriş/çıkış, flash
// mesajları, casbin yetkilendirmesi, veriden yüklenen menü, Türkçe/İngilizce arayüz,
// form doğrulama ve JSON API.
//
//	go run ./cmd/crm
//	CRM_ENV=production CRM_SESSION_KEY=... CRM_ADMIN_PASSWORD=... CRM_USER_PASSWORD=... go run ./cmd/crm
//
// Şablonlar, statik dosyalar, dil dosyaları ve casbin politikası ikiliye gömülüdür.
// Ayrıntılar: cmd/crm/README.md.
package main

import (
	"context"
	"crypto/hkdf"
	"crypto/sha256"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/template/jet/v2"

	"github.com/mustafacaglarkara/webdev/cmd/crm/config"
	"github.com/mustafacaglarkara/webdev/cmd/crm/controllers"
	"github.com/mustafacaglarkara/webdev/cmd/crm/menusvc"
	"github.com/mustafacaglarkara/webdev/cmd/crm/middleware"
	"github.com/mustafacaglarkara/webdev/cmd/crm/routers"
	"github.com/mustafacaglarkara/webdev/pkg/localization"
	"github.com/mustafacaglarkara/webdev/pkg/logx"
	"github.com/mustafacaglarkara/webdev/pkg/policy"
	"github.com/mustafacaglarkara/webdev/pkg/web"
	"github.com/mustafacaglarkara/webdev/pkg/web/fiberweb"
)

//go:embed templates/*.jet
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

//go:embed policy/model.conf policy/policy.csv
var policyFS embed.FS

func main() {
	if err := run(); err != nil {
		logx.Error("crm: fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	setupLogger(cfg)
	warnInsecureDefaults(cfg)

	app, err := newApp(cfg)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logx.Info("crm: listening", "addr", cfg.Addr(), "env", cfg.Env)
		errCh <- app.Listen(cfg.Addr())
	}()

	select {
	case err := <-errCh:
		return err // dinleme başlatılamadı (ör. port kullanımda)
	case <-ctx.Done():
	}
	logx.Info("crm: shutting down")
	stop() // ikinci sinyal süreci hemen sonlandırsın
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logx.Info("crm: stopped")
	return nil
}

// setupLogger pkg/logx varsayılan logger'ını kurar: production'da JSON, geliştirmede text.
func setupLogger(cfg config.Config) {
	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	format := "text"
	if cfg.IsProduction() {
		format = "json"
	}
	logx.SetDefault(logx.New(logx.Config{Level: level, Format: format}))
}

// warnInsecureDefaults geliştirme varsayılanları kullanılıyorsa açıkça uyarır.
func warnInsecureDefaults(cfg config.Config) {
	if cfg.SessionKeyGenerated {
		logx.Warn("crm: " + config.EnvSessionKey + " is not set; using a random per-process key. " +
			"Sessions will not survive restarts. Set a random value of at least 32 bytes (e.g. `openssl rand -base64 48`).")
	}
	if cfg.DefaultPasswords {
		logx.Warn("crm: using DEVELOPMENT DEFAULT demo passwords; set "+config.EnvAdminPassword+
			" and "+config.EnvUserPassword+" before exposing this server",
			"admin_default", os.Getenv(config.EnvAdminPassword) == "",
			"user_default", os.Getenv(config.EnvUserPassword) == "")
	}
}

// newApp tüm bağımlılıkları kurar ve Fiber uygulamasını döner. Testler de bunu kullanır.
// Paket düzeyindeki kütüphane globalleri (oturum deposu, varsayılan localization, can
// denetleyicisi, isimli route'lar) burada ayarlanır.
func newApp(cfg config.Config) (*fiber.App, error) {
	// --- Oturum: HKDF ile ayrı imza (64 bayt) ve AES-256 şifreleme (32 bayt) anahtarları.
	hashKey, err := hkdf.Key(sha256.New, cfg.SessionKey, nil, "webdev-crm session hmac", 64)
	if err != nil {
		return nil, fmt.Errorf("session key: %w", err)
	}
	blockKey, err := hkdf.Key(sha256.New, cfg.SessionKey, nil, "webdev-crm session aes", 32)
	if err != nil {
		return nil, fmt.Errorf("session key: %w", err)
	}
	web.InitSessionStoreKeys(hashKey, blockKey)
	opts := web.DefaultSessionOptions()
	opts.Secure = cfg.SecureCookies() // production'da yalnızca HTTPS
	opts.MaxAge = int((8 * time.Hour).Seconds())
	web.SetSessionOptions(&opts)
	web.SetRedirectWhitelist(nil) // yalnızca göreli yönlendirmeler

	// --- i18n
	if err := initLocalization(); err != nil {
		return nil, err
	}
	if !cfg.IsProduction() {
		localization.Default().OnMissing(func(langs []string, id string, _ error) {
			logx.Warn("i18n: missing translation", "id", id, "langs", langs)
		})
	}

	// --- Yetkilendirme (casbin)
	pm, err := loadPolicy()
	if err != nil {
		return nil, err
	}
	policy.SetDefault(pm)
	web.SetCanChecker(policy.Check) // şablondaki can(ctx, ...) ve menü süzmesi

	// --- Kullanıcılar ve menü
	users, err := controllers.NewUserStore(cfg.AdminPassword, cfg.UserPassword)
	if err != nil {
		return nil, err
	}
	menu := menusvc.NewLoader(config.NewMemoryMenuStore(config.DefaultMenuRecords()), time.Minute)
	if err := menu.Reload(context.Background()); err != nil {
		return nil, err
	}

	h := &controllers.Handlers{
		Cfg:       cfg,
		Users:     users,
		Menu:      menu,
		Policy:    pm,
		Langs:     supportedLangs,
		StartedAt: time.Now(),
	}

	// --- Şablonlar (Jet, gömülü)
	tplFS, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		return nil, err
	}
	engine := jet.NewFileSystem(http.FS(tplFS), ".jet")
	engine.AddFuncMap(web.JetTemplateFilters())    // upper, default, date, sanitize ...
	engine.AddFuncMap(web.JetFormHelpers())        // form_error, form_has_error ...
	engine.AddFuncMap(fiberweb.JetGlobalHelpers()) // t(ctx,..), old(ctx,..), can(ctx,..), route, static, dict
	if err := engine.Load(); err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}

	app := fiber.New(fiber.Config{
		AppName:               "webdev-crm",
		Views:                 engine,
		ErrorHandler:          h.ErrorHandler,
		DisableStartupMessage: true,
		ReadTimeout:           15 * time.Second,
		WriteTimeout:          15 * time.Second,
		IdleTimeout:           60 * time.Second,
		BodyLimit:             1 << 20,
	})

	// Sıra önemlidir: Fiber middleware ve route'ları kayıt sırasıyla eşler.
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(middleware.RequestLogger())
	app.Use(middleware.SecurityHeaders(cfg.IsProduction()))

	// 1) Statik dosyalar: oturum/CSRF middleware'lerinden ÖNCE; çerez yazılmaz.
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, err
	}
	app.Use("/static", filesystem.New(filesystem.Config{Root: http.FS(sub), MaxAge: 3600}))

	// 2) Oturumdaki kullanıcı Locals'a (hem API hem web için).
	app.Use(fiberweb.AttachUser)

	// 3) JSON API: kendi CSRF kuralıyla (yalnızca yazma isteklerinde).
	routers.API(app, h)

	// 4) HTML sayfaları: tüm istekler CSRF middleware'inden geçer (token üretimi +
	//    POST/PUT/PATCH/DELETE doğrulaması).
	app.Use(fiberweb.CSRFWithConfig(&fiberweb.CSRFConfig{ErrorHandler: h.CSRFError}))
	routers.Web(app, h, routers.Options{LoginRateLimit: cfg.LoginRateLimit})
	routers.Admin(app, h)

	return app, nil
}

// loadPolicy gömülü casbin model/politika dosyalarını doğrudan embed.FS'ten yükler
// (policy.NewFS, A5-3). Manager.Reload gömülü dosyaları yeniden okur.
func loadPolicy() (*policy.Manager, error) {
	m, err := policy.NewFS(policyFS, "policy/model.conf", "policy/policy.csv")
	if err != nil {
		return nil, fmt.Errorf("policy: %w", err)
	}
	return m, nil
}
