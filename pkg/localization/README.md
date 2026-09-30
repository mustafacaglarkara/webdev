# localization

[nicksnyder/go-i18n](https://github.com/nicksnyder/go-i18n) tabanlı çoklu dil
yardımcıları: JSON dil dosyaları, parametreli mesajlar, varsayılan dile düşme,
eksik çeviri kancası ve `Accept-Language` ayrıştırma.

```go
import "github.com/mustafacaglarkara/webdev/pkg/localization"
```

## Manager

```go
//go:embed locales/*.json
var locales embed.FS

mgr := localization.New(language.Turkish) // varsayılan dil
if err := mgr.LoadFS(locales, "locales/*.json"); err != nil {
	log.Fatal(err)
}

mgr.T([]string{"tr"}, "hello", nil)                                   // "Merhaba!"
mgr.T([]string{"en"}, "hello", nil)                                   // "Hello!"
mgr.T([]string{"tr"}, "welcome_user", map[string]any{"Name": "Ali"})  // "Hoş geldin, Ali!"
mgr.T([]string{"fr", "tr"}, "hello", nil)                             // "Merhaba!"
mgr.T([]string{"tr"}, "not_exist", nil)                               // "not_exist"
```

`T` davranışı:

- Mesaj istenen dillerde yoksa **varsayılan dildeki çeviri** döner.
- Hiçbir dilde yoksa `msgID` döner.
- Her iki durumda da `OnMissing` kancası (tanımlıysa) çağrılır.

Gelişmiş kullanım için doğrudan go-i18n localizer'ı:

```go
loc := mgr.Localizer("tr", "en")
msg, err := loc.Localize(&i18n.LocalizeConfig{MessageID: "bye"})
```

## Eksik çeviri kancası

```go
mgr.OnMissing(func(langs []string, msgID string, err error) {
	slog.Warn("eksik çeviri", "id", msgID, "langs", langs, "err", err)
})
mgr.OnMissing(nil) // kancayı kaldırır
```

## Varsayılan manager

```go
if err := localization.InitDefault(language.Turkish, locales, "locales/*.json"); err != nil {
	log.Fatal(err)
}
msg := localization.TDefault([]string{"en"}, "hello", nil) // başlatılmadıysa "hello"

localization.SetDefault(mgr)     // hazır bir manager'ı varsayılan yap
m := localization.Default()      // başlatılmadıysa nil
_, _ = msg, m
```

Varsayılan manager `atomic.Pointer` ile tutulur; eşzamanlı okuma/yazma güvenlidir.
Başarısız `InitDefault` mevcut varsayılanı değiştirmez. `Manager.T` eşzamanlı
çağrılabilir; `LoadFS` yazma kilidi alır.

## Context ile dil

```go
ctx := localization.WithLang(r.Context(), "en")
lang := localization.LangFromCtx(ctx, "tr") // "en" (yoksa/boşsa "tr")
```

## ParseAcceptLanguage

```go
localization.ParseAcceptLanguage("en;q=0.5, tr;q=0.9", "tr") // [tr en]
localization.ParseAcceptLanguage("de;q=0, en-US, en;q=0.8", "tr") // [en-US en tr]
localization.ParseAcceptLanguage("", "tr")                    // [tr]
```

- q değerine göre azalan, eşitlikte başlıktaki sırayı koruyan (kararlı) sıralama.
- `q=0` olan diller dışlanır; geçersiz q değerli girdiler ve `*` atlanır; `q>1` 1'e kırpılır.
- `en ; q=0.8` gibi boşluklu biçimler ve `Q=` desteklenir.
- Aynı dil (büyük/küçük harf duyarsız) bir kez yer alır.
- `fallback` boş değilse ve listede yoksa sona eklenir.

Fiber için `fiberweb.Langs(c, "tr")` (`pkg/web/fiberweb`) bu fonksiyonu kullanır ve oturumdaki dil tercihini listenin başına koyar.

## Dil dosyası örneği (`locales/active.tr.json`)

```json
[
  { "id": "hello", "translation": "Merhaba!" },
  { "id": "welcome_user", "translation": "Hoş geldin, {{.Name}}!" },
  { "id": "bye", "translation": "Güle güle!" }
]
```

Dosya adı dil etiketini içermelidir (`active.tr.json`, `en.json` ...). Bozuk
JSON veya hatalı glob deseninde `LoadFS` hata döner. Eksik şablon verisi
`<no value>` olarak basılır (go-i18n davranışı).
