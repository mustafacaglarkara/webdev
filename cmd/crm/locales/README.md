# cmd/crm/locales — dil dosyaları

Uygulamanın çevirileri. Dosyalar `locales_embed.go` ile ikiliye gömülür ve başlangıçta
`localization.InitDefault(language.Turkish, localesFS, "locales/*.json")` ile yüklenir.
Varsayılan (son çare) dil Türkçedir.

## Dosya adları

Dil, go-i18n kuralıyla dosya adından alınır: `<grup>.<dil>.json`.

| Dosya | İçerik |
|---|---|
| `active.tr.json`, `active.en.json` | Genel arayüz: menü, giriş, profil, form demosu, hata sayfaları, flash mesajları |
| `admin.tr.json`, `admin.en.json` | Yalnızca yönetim sayfası (`admin.*` anahtarları) |

Aynı dile ait iki dosyada (ör. `active.tr.json` ve `admin.tr.json`) **aynı anahtar
bulunmamalıdır**; `admin.*` anahtarları yalnızca `admin.<dil>.json` dosyalarındadır.
Her `tr` dosyasının `en` eşi aynı anahtarlara sahip olmalıdır (`TestLocaleParity` denetler).

## Biçim

Düz nesne haritası (go-i18n v2 biçimi); anahtarlar noktalı ad alanı kullanır:

```json
{
  "home.title": "Hoş geldiniz",
  "home.welcome_user": "Merhaba {{.Name}}, oturumunuz açık."
}
```

- Değer, go-i18n şablonudur: parametreler `{{.Ad}}` biçimindedir.
- Dosyalar UTF-8'dir; boş dosya bırakmayın (en az `{}`), aksi halde i18ncheck hata verir.

## Kullanım

Şablonda (Jet) — anahtar **çift tırnaklı sabit metin** olmalı ve bağlam adı `ctx` olmalı ki
i18ncheck anahtarı görebilsin:

```jet
{{ t(ctx, "home.title") }}
{{ t(ctx, "home.welcome_user", dict("Name", UserName)) }}
```

Dil sırası: oturumdaki tercih (`POST /lang`) → `Accept-Language` → `tr`.

Handler'da (flash mesajları, hata sayfaları):

```go
h.T(c, "login.success", map[string]any{"Name": u.Name})
```

## Yeni anahtar ekleme

1. Anahtarı `active.tr.json` **ve** `active.en.json`'a (yönetim sayfası içinse
   `admin.tr.json` ve `admin.en.json`'a) ekleyin.
2. Şablonda `t(ctx, "anahtar")` ile kullanın.
3. Denetleyin:

```bash
go run ./cmd/i18ncheck                 # şablon anahtarları ↔ locale dosyaları; eksik varsa çıkış 2
go test ./cmd/crm -run 'Locale|Menu'   # tr/en eşliği ve menü etiketleri
```

## i18ncheck çıktısındaki "Unused" listesi

i18ncheck yalnızca şablonlardaki sabit `t(...)` çağrılarını görür. Şu anahtarlar Go
kodundan veya değişkenle kullanıldığı için "Unused" listesinde görünür; bu beklenen
durumdur (varsayılan bayraklarla çıkış kodu 0'dır):

- `menu.*` — `base.jet` / `demo_menu.jet` içinde `t(ctx, item.LabelKey)`
- `about.pkg.*`, `login.*` flash mesajları, `logout.success`, `lang.*`,
  `formdemo.success`, `formdemo.failed`, `formdemo.err.*`, `error.*` — handler'larda `h.T(...)`

Bu anahtarların varlığı `TestLocaleParity` ve `TestMenuLabelsTranslated` testleriyle
denetlenir. `-fail-on-unused` bu proje için kullanılmaz.
