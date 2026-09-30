# i18ncheck

```go
import "github.com/mustafacaglarkara/webdev/pkg/i18ncheck"
```

Statik analiz aracı: şablonlarda kullanılan i18n anahtarlarını go-i18n JSON
locale dosyalarıyla karşılaştırır. Komut satırı aracı `cmd/i18ncheck`'tir.

Özellikler:

- Eksik (şablonda var, locale'de yok) ve kullanılmayan (locale'de var,
  şablonda yok) anahtar tespiti
- Usage map (anahtar -> kullanan dosyalar)
- `.gitignore` benzeri hariç tutma (negation `!`, `**`, anchored `/`, `dir/`)
  ve `-exclude` glob desenleri
- Locale şema doğrulaması, dil bazında tekrar (duplicate) tespiti
- Paralel tarama, JSON raporu (`-out`)

## Şablon anahtar sözdizimi

Varsayılan desen (`i18ncheck.DefaultPattern`):

```
\bt\(\s*(?:ctx\s*,\s*)?"([^"]+)"
```

Yani anahtar, `t(` çağrısının ilk (veya `ctx`'ten sonraki) argümanı olan
**çift tırnaklı bir metin** olmalıdır. Yakalananlar:

```jet
{{ t("home.title") }}
{{ t(ctx, "home.welcome", map("Name", user.Name)) }}
{{ t( ctx , "nav.about") }}
```

Yakalanmayanlar: tek tırnak/backtick (`t('x')`), değişkenle anahtar
(`t(key)`), `ctx` dışında bir adla bağlam (`t(c, "x")`). Farklı bir sözdizimi
için `-pattern` verin; desenin 1. yakalama grubu anahtar olmalıdır. Varsayılan
olarak yalnızca `.jet` uzantılı dosyalar taranır (`-ext .jet,.html`).

## Locale dosya biçimleri

Dosyalar `-locales` glob'u ile bulunur (varsayılan `cmd/crm/locales/*.json`).
Dil, go-i18n kuralıyla dosya adından alınır: `tr.json`, `active.tr.json`,
`admin.en-US.json` -> `tr`, `tr`, `en-US`. Her iki go-i18n biçimi desteklenir
ve dosyalar karışık biçimde olabilir.

### 1. Nesne dizisi (go-i18n v1 / "translation" biçimi)

```json
[
  {"id": "home.title", "translation": "Ana Sayfa"},
  {"id": "home.welcome", "translation": "Hoş geldin {{.Name}}"}
]
```

- Her öğe bir nesne olmalı, boş olmayan `"id"` içermelidir.
- `-locale-required` (varsayılan `translation`) alanları boş olmamalıdır.
- `-locale-strict` ile `id` ve zorunlu alanlar dışındaki alanlar hatadır.

### 2. Nesne haritası (go-i18n v2 biçimi), düz veya iç içe

```json
{
  "home": {
    "title": "Ana Sayfa",
    "welcome": "Hoş geldin {{.Name}}"
  },
  "nav.about": "Hakkımızda",
  "items": {"one": "{{.Count}} öğe", "other": "{{.Count}} öğe"},
  "greet": {"description": "Karşılama", "other": "Merhaba"}
}
```

Bu dosya `home.title`, `home.welcome`, `nav.about`, `items`, `greet`
anahtarlarını üretir. Kurallar (go-i18n v2 ile aynı):

- Değer metinse bir mesajdır; iç içe anahtarlar `.` ile birleştirilir.
- Nesne, go-i18n ayrılmış alanlarından (`id`, `description`, `hash`,
  `leftdelim`, `rightdelim`, `zero`, `one`, `two`, `few`, `many`, `other`,
  `translation`) en az birini içeriyorsa mesajdır; `id` alanı varsa anahtar
  odur. Hiç içermiyorsa ad alanıdır (içine inilir).
- Ayrılmış ve ayrılmamış anahtarları karıştıran nesne, boş metin, `null`,
  sayı/dizi değerler, boş nesne ve metin taşımayan mesaj
  (`translation`/`other`/`zero`/`one`/`two`/`few`/`many` hepsi boş) geçersiz
  girdi olarak raporlanır (`index: -1`, `key`: noktalı yol).
- `-locale-required` ve `-locale-strict` bu biçime uygulanmaz.

### Hatalar

- **Boş dosya** (0 bayt veya yalnızca boşluk) çalışma hatasıdır (çıkış 1):
  `locale load error: cmd/crm/locales/active.en.json: locale dosyası boş; en az [] veya {} içermeli ...`
  Henüz çevirisi olmayan dosyaya `{}` veya `[]` yazın. (go-i18n boş dosyayı
  sessizce kabul eder; bu araç eksik dosyayı fark ettirmek için reddeder.)
- Geçersiz JSON, kökü dizi/nesne olmayan dosya ve hiçbir dosyayla eşleşmeyen
  glob da çalışma hatasıdır (çıkış 1).

### Tekrarlar

Tekrarlar **dil bazında** aranır: `active.tr.json` ve `active.en.json`'da aynı
anahtarın olması normaldir. Aynı dile ait iki dosyada (ör. `active.tr.json` ve
`admin.tr.json`) veya aynı dosyada iki kez geçen anahtar tekrardır. Dili
tanınamayan dosyalar (`a.json`, `messages.json`) tek bir ortak grupta
değerlendirilir. Her tekrarlanan anahtar raporda bir kez yer alır.
`-locale-allow-dup-prefixes admin.,site.` ile ön ek bazında tolerans verilir.

## Komut satırı

```bash
go run ./cmd/i18ncheck                                   # varsayılan yollar
go run ./cmd/i18ncheck -templates web/templates -locales 'web/locales/*.json'
go run ./cmd/i18ncheck -format json -show-usage
go run ./cmd/i18ncheck -show-usage -usage-filter=missing
go run ./cmd/i18ncheck -gitignore .gitignore -exclude "partials/**,vendor/**"
go run ./cmd/i18ncheck -locale-required translation -locale-strict
go run ./cmd/i18ncheck -locale-allow-dup-prefixes "admin.,site."
go run ./cmd/i18ncheck -fail-on-unused -fail-on-locale-errors   # CI
go run ./cmd/i18ncheck -out reports/i18n_report.json -format text
```

Örnek (yukarıdaki iki locale dosyası ve `footer.copyright` anahtarını da
kullanan bir şablonla, `old.key` yalnızca `active.tr.json`'da):

```
$ i18ncheck -templates templates -locales 'locales/*.json'
== i18n Check ==
Templates keys: 4
Locale keys:    4
Missing (1):
  - footer.copyright
Unused (1):
  - old.key
$ echo $?
2
```

JSON raporda dilim alanları hiçbir zaman `null` değildir:

```json
{
  "templateCount": 4,
  "localeCount": 4,
  "missing": [],
  "unused": ["old.key"],
  "invalidLocale": [],
  "duplicateLocaleIDs": [],
  "readErrors": []
}
```

(`-show-usage` ile `"usage": {"anahtar": ["dosya.jet"]}` alanı eklenir.
`invalidLocale` öğeleri `{"file","index","key","reason"}` biçimindedir.)

### Bayraklar

| Bayrak | Varsayılan | Açıklama |
|---|---|---|
| `-templates` | `cmd/crm/templates` | Şablon kök dizini |
| `-locales` | `cmd/crm/locales/*.json` | Locale glob'u |
| `-pattern` | `DefaultPattern` | Anahtar regex'i (1. grup = anahtar) |
| `-ext` | `.jet` | Virgülle uzantılar |
| `-ignore` | | Rapordan hariç anahtar ön ekleri |
| `-exclude` | | Şablon köküne göre glob desenleri |
| `-gitignore` | | `.gitignore` benzeri dosya (şablon köküne göre veya mutlak) |
| `-format` | `text` | `text` veya `json` |
| `-show-usage`, `-usage-filter` | `all` | Usage map; `all` veya `missing` |
| `-workers` | `0` | İşçi sayısı; `0` = CPU sayısı |
| `-locale-required` | `translation` | Dizi biçiminde zorunlu alanlar |
| `-locale-strict` | `false` | Dizi biçiminde fazla alan hatası |
| `-locale-allow-dup-prefixes` | | Tekrar toleransı ön ekleri |
| `-fail-on-unused`, `-fail-on-locale-errors` | `false` | CI başarısızlık koşulları |
| `-out` | | JSON raporu dosyaya da yaz (dizin oluşturulur) |

### Çıkış kodları

| Kod | Anlam |
|---|---|
| 0 | Başarılı |
| 1 | Çalışma hatası: geçersiz bayrak, regex, okunamayan şablon dizini, locale yükleme hatası (boş/geçersiz dosya, eşleşmeyen glob), **okunamayan şablon dosyası** |
| 2 | Eksik anahtar(lar) var |
| 3 | `-fail-on-unused` ve kullanılmayan anahtar(lar) var |
| 4 | `-fail-on-locale-errors` ve locale şema/tekrar hatası var |

Birden fazla koşul varsa tablodaki sıraya göre ilki döner.

## Kütüphane olarak

```go
cfg, err := i18ncheck.ParseFlags(os.Args[1:])
if err != nil {
    os.Exit(1)
}
os.Exit(i18ncheck.Run(cfg)) // veya RunTo(cfg, stdout, stderr)
```

`Config` doğrudan da oluşturulabilir; `Workers <= 0` ise kütüphane CPU sayısı
kadar işçi kullanır, `Format` boşsa `text` kabul edilir. Diğer API:
`CollectTemplateKeys`, `LoadLocaleKeys`, `BuildReport`, `Report.Write`,
`Report.ExitCode`, `LoadGitignore`, `NewPathExcluder`.
