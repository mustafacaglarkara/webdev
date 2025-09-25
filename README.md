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
Veritabanı işlemleri için yardımcılar.

```go
import "your/module/path/pkg/db"
// Bağlantı, sorgu ve cache işlemleri için kullanılır.
```

---

### 7. router
HTTP yönlendirme ve router adapterleri.

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

## Diğer Paketler
- httpx: HTTP yardımcıları
- logx: Loglama
- resilience: Retry, circuit breaker vb.
- middleware: Ara katmanlar
- migrate: Veritabanı migrasyonları
- sqlutil: SQL yardımcıları
- web: Web yardımcıları

## Katkı

Katkı yapmak için PR veya issue açabilirsiniz. Kod stiline ve test kapsamına dikkat edin.

## Lisans

Projenin lisansı burada belirtilmelidir.

