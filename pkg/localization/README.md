# localization

Çoklu dil desteği için yardımcı katman. Laravel/Django'daki gibi JSON/array tabanlı dil dosyaları ile çalışır. Context ile dil yönetimi, parametreli mesajlar ve fallback desteği sunar.

## Tavsiye Edilen Paket: [nicksnyder/go-i18n](https://github.com/nicksnyder/go-i18n)

---

## Temel Kullanım

### 1. Manager Oluşturma ve JSON Dosyası Yükleme

```go
import (
    "embed"
    "github.com/nicksnyder/go-i18n/v2/i18n"
    "golang.org/x/text/language"
    "your/module/path/pkg/localization"
)

//go:embed locales/*.json
var locales embed.FS

mgr := localization.New(language.Turkish)
err := mgr.LoadFS(locales, "locales/*.json")
if err != nil {
    panic(err)
}
```

### 2. Basit Mesaj Çözümleme

```go
msg := mgr.T([]string{"tr"}, "hello", nil) // "Merhaba!"
msg2 := mgr.T([]string{"en"}, "hello", nil) // "Hello!"
```

### 3. Parametreli Mesaj

```go
msg := mgr.T([]string{"tr"}, "welcome_user", map[string]any{"Name": "Ali"}) // "Hoş geldin, Ali!"
```

### 4. Fallback ve Eksik Çeviri

```go
msg := mgr.T([]string{"fr", "tr"}, "hello", nil) // "Merhaba!" (fr yoksa tr'ye düşer)
msg := mgr.T([]string{"tr"}, "not_exist", nil)    // "not_exist" (eksikse anahtar döner)
```

### 5. Context ile Dil Yönetimi

```go
ctx := localization.WithLang(context.Background(), "en")
lang := localization.LangFromCtx(ctx, "tr") // "en"
msg := mgr.T([]string{lang}, "hello", nil)
```

### 6. Localizer ile Gelişmiş Kullanım

```go
loc := mgr.Localizer("tr", "en")
msg, err := loc.Localize(&i18n.LocalizeConfig{MessageID: "bye"})
```

---

## HTTP / Fiber ile Kullanım (Önerilen)

```go
import (
    "github.com/gofiber/fiber/v2"
    "github.com/mustafacaglarkara/webdev/pkg/web"
    "github.com/mustafacaglarkara/webdev/pkg/localization"
)

// Dil belirleme (Accept-Language + fallback)
func Handler(c *fiber.Ctx) error {
    langs := web.FiberLangs(c, "tr")
    msg := localization.TDefault(langs, "home.title", nil)
    return c.SendString(msg)
}

// Dil değiştirme (session'a kaydet)
func SwitchLang(c *fiber.Ctx) error {
    code := c.Params("code") // "tr" veya "en"
    _ = web.FiberSetPreferredLang(c, code)
    return c.Redirect("/", fiber.StatusFound)
}
```

---

## JSON Dil Dosyası Örneği (locales/active.tr.json)

```json
[
  { "id": "hello", "translation": "Merhaba!" },
  { "id": "welcome_user", "translation": "Hoş geldin, {{.Name}}!" },
  { "id": "bye", "translation": "Güle güle!" }
]
```

---

## Edge-Case ve Test Senaryoları

- Eksik dil dosyası: mgr.T(["fr"], "hello", nil) → fallback veya anahtar döner.
- Eksik parametre: mgr.T(["tr"], "welcome_user", nil) → "Hoş geldin, <no value>!"
- Context yoksa: LangFromCtx(context.Background(), "tr") → "tr"
- JSON dosyası bozuksa: LoadFS hata döner.

---

## Notlar
- JSON dosyaları UTF-8 ve array formatında olmalı.
- Parametreli mesajlar için map[string]any ile veri geçebilirsiniz.
- Context ile dil yönetimi, handler ve middleware'lerde kolaylık sağlar.
- Daha fazla detay ve gelişmiş kullanım için go-i18n dökümantasyonuna bakınız.
