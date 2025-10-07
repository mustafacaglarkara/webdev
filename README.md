# Go Yardımcı Paketler Projesi

Bu proje, Go ile geliştirilmiş çok amaçlı yardımcı paketler koleksiyonudur. Her bir paket, belirli bir ihtiyaca yönelik pratik fonksiyonlar sunar. Tüm paketler `pkg` dizini altında toplanmıştır.

## Kurulum

```bash
git clone <repository-url>
cd <repository-directory>
go mod download
```

## Kullanım

Her paketi doğrudan import ederek kullanabilirsiniz. Aşağıda öne çıkan paketler ve örnek kullanımlar yer almaktadır.

---

### 1. collection
Generic slice işlemleri için yardımcılar.

```go
import "your/module/path/pkg/collection"
arr := []int{1, 2, 3, 2}
collection.Contains(arr, 2) // true
collection.IndexOf(arr, 3)  // 2
collection.Dedup(arr)       // [1 2 3]
```

---

### 2. conv
Tip dönüşüm ve sayı işlemleri.

```go
import "your/module/path/pkg/conv"
conv.ToInt("42", 0)         // 42
conv.ToFloat64("3.14", 0)   // 3.14
conv.RandomInt(10, 20)       // 10-20 arası rastgele sayı
```

---

### 3. config
Ortam değişkeni ve YAML işlemleri.

```go
import "your/module/path/pkg/config"
port := config.GetEnv("PORT", "8080")
max := config.GetEnvInt("MAX_CONN", 10)
cfg := struct{ Name string `yaml:"name"` }{Name: "svc"}
yamlStr, _ := config.ToYAML(cfg)
```

---

### 4. fs
Dosya, arşiv, FTP, resim ve e-posta işlemleri.

```go
import "your/module/path/pkg/fs"
fs.FileExists("dosya.txt")
_ = fs.ZipDir("klasor", "arsiv.zip")
img, _, _ := fs.LoadImage("resim.jpg")
```

---

### 5. crypto
Kriptografi, hash, şifreleme ve token işlemleri.

```go
import "your/module/path/pkg/crypto"
hash := crypto.MD5Hash("merhaba")
tok, _ := crypto.GenerateBearerToken(32)
```

---

### 6. db
Veritabanı işlemleri için yardımcılar. GORM ile uyumlu, bağlantı, sorgu, cache, migration, toplu ekleme/güncelleme, statement cache, context ve loglama desteği sunar.

```go
import "your/module/path/pkg/db"
// Bağlantı, sorgu ve cache işlemleri için kullanılır.
// Ayrıntılı kullanım ve örnekler için pkg/db/README.md dosyasına bakınız.
```

---

### 7. router
HTTP yönlendirme ve router adapterleri. Chi ve Mux desteğiyle kolay router kullanımı sağlar.

```go
import "your/module/path/pkg/router"
// Chi veya Mux ile kolay router kullanımı sağlar.
```

---

### 8. seeder
Veritabanı için başlangıç (seed) verileri ekleme.

```go
import "your/module/path/pkg/seeder"
// Seed fonksiyonları ile örnek veri ekleyebilirsiniz.
```

---

### 9. timex
Zaman ve tarih işlemleri.

```go
import "your/module/path/pkg/timex"
// Zaman hesaplama ve biçimlendirme işlemleri için kullanılır.
```

---

### 10. validate & validation
Veri doğrulama işlemleri.

```go
import "your/module/path/pkg/validate"
import "your/module/path/pkg/validation"
// Form ve veri doğrulama işlemleri için kullanılır.
```

---

### 11. httpx
HTTP istekleri, JSON API, retry/backoff, loglama, response doğrulama, header/body redaksiyonu gibi gelişmiş HTTP yardımcıları.

```go
import "your/module/path/pkg/httpx"
cli := httpx.New()
var resp map[string]any
err := cli.GetJSON(ctx, "/api/v1/foo", &resp)
```
Daha fazla örnek ve detay için pkg/httpx/README.md dosyasına bakınız.
```

---

### 12. logx
Loglama işlemleri için yardımcılar. Structured logging, slog desteği.

```go
import "your/module/path/pkg/logx"
logx.L().Info("mesaj", "key", "value")
```

---

### 13. resilience
Retry, circuit breaker, bulkhead gibi dayanıklılık (resilience) yardımcıları.

```go
import "your/module/path/pkg/resilience"
// Retry ve circuit breaker ile işlemleri güvenli hale getirin.
```

---

### 14. middleware
HTTP için ara katman (middleware) fonksiyonları.

```go
import "your/module/path/pkg/middleware"
// Recover, logging, CORS vb. middleware'ler içerir.
```

---

### 15. migrate
Veritabanı migration işlemleri için yardımcılar.

```go
import "your/module/path/pkg/migrate"
// Migration işlemlerini kolaylaştırır.
```

---

### 16. sqlutil
SQL template, çoklu sorgu, dinamik query üretimi, GORM ile uyumlu SQL yardımcıları.

```go
import "your/module/path/pkg/sqlutil"
loader := sqlutil.NewSQLLoader(os.DirFS("queries"), nil)
sql, err := loader.LoadNamed("product.sql", "getProductById", map[string]any{"id": 1})
// GORM ile: db.Raw(sql).Scan(&result)
```
- SQL Server stored procedure OUT parametrelerini otomatik toplamak için `spOutDecls` ve `spOutVals` fonksiyonlarını kullanabilirsiniz. Detaylar pkg/sqlutil/README.md içinde.

Daha fazla örnek ve detay için pkg/sqlutil/README.md dosyasına bakınız.
```

---

### 17. web
Web uygulamaları için yardımcılar: auth decorator, flash mesaj, redirect, response, template, tag işlemleri.

```go
import "your/module/path/pkg/web"
// Web uygulamalarında pratik yardımcılar.
```

---

### 18. ratelimit
Token bucket tabanlı basit bir rate limiter. E-posta göndermek, harici API çağrılarını sınırlamak gibi kullanım senaryoları için idealdir.

```go
import "your/module/path/pkg/ratelimit"
lim, _ := ratelimit.NewLimiter(30, time.Minute, 10) // dakikada 30, burst 10
_ = lim.Do(ctx, func() error {
    // sınırlandırılmış iş
    return nil
})
```
Daha fazla örnek ve detay için pkg/ratelimit/README.md dosyasına bakınız.

---

### 19. grpcx
gRPC istemci/sunucu işlemlerini kolaylaştıran yardımcılar: kolay Dial, metadata yönetimi, unary/stream interceptor zinciri, header/trailer yardımcıları.

```go
import "your/module/path/pkg/grpcx"
conn, err := grpcx.Dial(grpcx.DialOptions{Address: "localhost:50051", WithInsecure: true})
```
Detaylar pkg/grpcx/README.md içinde.

---

## Diğer Paketler
- id: UUID ve kısa id üretimi
- policy: Policy ve izin kontrol yardımcıları
- q: Sorgu string yardımcıları
- scheduler: Zamanlayıcı ve cron işlemleri
- security: Güvenlik yardımcıları
- signals: OS signal yönetimi
- socialite: Sosyal giriş yardımcıları
- text: Metin işlemleri
- localization: Çoklu dil desteği
- migrate: Migration işlemleri
- ...ve diğerleri

## Katkı

Katkı yapmak için PR veya issue açabilirsiniz. Kod stiline ve test kapsamına dikkat edin.

## Lisans

Projenin lisansı burada belirtilmelidir.
