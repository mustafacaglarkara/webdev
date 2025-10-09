# i18ncheck

Statik analiz aracı: template'lerde kullanılan i18n anahtarlarını locale JSON dosyaları ile karşılaştırır.

Özellikler
- Missing/unused anahtar tespiti
- Usage map (anahtar -> kullanan dosyalar)
- Gelişmiş .gitignore desteği (negation `!`, `**`, anchored `/`, `dir/`)
- Locale şema doğrulaması (gerekli alanlar, strict mod, duplicate ID)
- Prefix bazlı duplicate toleransı (örn: `admin.` ile başlayanlar göz ardı edilebilir)
- Paralel tarama (worker pool)
- JSON raporu dosyaya yazma (`-out`)

Kullanım

```bash
# Temel
go run ./cmd/i18ncheck

# JSON çıktı + usage map
go run ./cmd/i18ncheck -format json -show-usage

# Sadece eksik anahtarların usage'ı
go run ./cmd/i18ncheck -show-usage -usage-filter=missing

# .gitignore ve exclude desenleri
go run ./cmd/i18ncheck -gitignore .gitignore -exclude "partials/**,vendor/**"

# Locale şeması (zorunlu alanlar + strict)
go run ./cmd/i18ncheck -locale-required translation -locale-strict

# Duplicate toleransı
go run ./cmd/i18ncheck -locale-allow-dup-prefixes "admin.,site."

# CI'de fail koşulları
go run ./cmd/i18ncheck -fail-on-unused -fail-on-locale-errors

# JSON raporunu dosyaya yaz (stdout yine seçtiğiniz formatta basılır)
# Aşağıdaki örnek, raporu reports/i18n_report.json dosyasına yazar ve konsola text basar:
go run ./cmd/i18ncheck -out reports/i18n_report.json -format text

# Konsola da JSON basmak istiyorsanız -format json ile beraber kullanın:
go run ./cmd/i18ncheck -out reports/i18n_report.json -format json
```

Exit Kodları
- 0: Başarılı
- 2: Missing anahtar(lar) var
- 3: (fail-on-unused) Kullanılmayan anahtar(lar) var
- 4: (fail-on-locale-errors) Locale şema/duplicate hataları var
- 1: Çeşitli çalışma zamanı hataları (IO, regex, vs.)

Sık Bayraklar
- `-templates` Template kök dizini (vars: `cmd/crm/templates`)
- `-locales` Locale glob (vars: `cmd/crm/locales/*.json`)
- `-pattern` Anahtar regex deseni (vars: `\bt\((?:ctx\s*,\s*)?"([^"]+)"`)
- `-ext` Taranacak uzantılar (örn: `.jet,.html`)
- `-ignore` Raporlardan hariç tutulacak anahtar ön ekleri
- `-exclude` Template köküne göre hariç tutulacak desenler (glob)
- `-gitignore` `.gitignore` dosyası yolu (templates root relatif olabilir)
- `-show-usage` Usage map üret
- `-usage-filter` `all|missing`
- `-workers` Paralel worker sayısı (0=CPU)
- `-locale-required` `id` dışında zorunlu alanlar (vars: `translation`)
- `-locale-strict` Zorunlu alanlar dışında alanları hata say
- `-locale-allow-dup-prefixes` Duplicate toleransı için izinli ön ekler
- `-fail-on-unused`, `-fail-on-locale-errors`
- `-out` JSON raporu dosya yoluna yaz (dizin otomatik oluşturulur)
