# helpers/text

Bu klasör; metin ve dil işlemleri ile ilgili yardımcı fonksiyonları toplamak için oluşturuldu.

Taşınacak dosyalar (plan):
- metin.go (ör: ToSlug, TrimSpace, NormalizeWhitespace ...)
- stringx.go (ör: Left, Right, ContainsAny, ReverseString ...)
- slice.go (string slice manipülasyonu ile alakalı ortak kısmi fonksiyonlar buraya ayrılabilir veya `collection` altına taşınabilir)

Geçiş Stratejisi:
1. Orijinal `pkg/helpers/metin.go` ve `stringx.go` içeriği `package text` olarak taşınacak.
2. `pkg/helpers` altında geriye dönük uyumluluk için ince wrapper fonksiyonlar bırakılacak; dış kullanım kırılmayacak.
3. Testler: Basit smoke test + önceki fonksiyon imzalarını koruma testi.

Not: Henüz kod taşınmadı, yalnızca klasör yapısı hazırlandı.

# text paketi

Metin ve dil işlemleri için yardımcı fonksiyonlar içerir. Türkçe karakter desteği, slug, ters çevirme, büyük/küçük harf, boşluk normalizasyonu, truncate, coalesce gibi işlemler için kullanılır.

## Fonksiyonlar ve Detaylı Kullanım Örnekleri

### ToSlug
Türkçe karakterleri sadeleştirip, küçük harfe çevirir ve a-z0-9 ile '-' karakterlerinden oluşan bir slug üretir.

```go
package main
import (
    "fmt"
    "your/module/path/pkg/text"
)
func main() {
    fmt.Println(text.ToSlug("Çalışma Alanı - 2025!")) // calisma-alani-2025
    fmt.Println(text.ToSlug("Güzel Günler"))         // guzel-gunler
}
```

### ToSlugForFile
Dosya adları için güvenli slug üretir, son uzantıyı korur.

```go
fmt.Println(text.ToSlugForFile("Çılgın Fotoğraf(1).JPG")) // cilgin-fotograf-1.jpg
fmt.Println(text.ToSlugForFile("rapor.v1.2.PDF"))         // rapor-v1-2.pdf
```

### ReverseString
Unicode (rune-safe) olarak string'i ters çevirir.

```go
fmt.Println(text.ReverseString("merhaba"))  // abahrem
fmt.Println(text.ReverseString("İstanbul")) // lubnatsİ
```

### ToUpper / ToLower

```go
fmt.Println(text.ToUpper("merhaba"))  // MERHABA
fmt.Println(text.ToLower("İSTANBUL")) // istanbul
```

### IsBlank
Sadece whitespace ise true döner.

```go
fmt.Println(text.IsBlank("   \t\n")) // true
fmt.Println(text.IsBlank("  x  "))    // false
```

### Coalesce
Verilen stringler arasında boş olmayan ilkini döner.

```go
fmt.Println(text.Coalesce("", "", "ilk")) // ilk
fmt.Println(text.Coalesce("", ""))         // ""
```

### Truncate
Unicode (rune-safe) olarak string'i n karaktere keser.

```go
fmt.Println(text.Truncate("merhaba dünya", 7)) // merhaba
fmt.Println(text.Truncate("abc", 10))         // abc
fmt.Println(text.Truncate("", 5))             // ""
```

### NormalizeSpace
Çoklu boşlukları tek boşluğa indirger, baştaki ve sondaki boşlukları kırpar.

```go
fmt.Println(text.NormalizeSpace("  merhaba   dünya   ")) // "merhaba dünya"
fmt.Println(text.NormalizeSpace("a   b\tc\nd"))        // "a b c d"
```

### UnescapeHTML
HTML entity'lerini gerçek karakterlere çözer.

```go
fmt.Println(text.UnescapeHTML("&Ccedil;alışma &amp; Deneme")) // Çalışma & Deneme
```

### FixTurkishMojibake
Yanlış encoding sonucu oluşan yaygın Türkçe karakter bozulmalarını düzeltir.

```go
s := "GÃ¼nÃ¼n Ã¶zeti: ÃaÄdaÅlÄ±k"
fmt.Println(text.FixTurkishMojibake(s)) // Günün özeti: Çağdaşlık
```

### SanitizeHTML / SanitizeHTMLWith
Kullanıcı girdisini güvenli HTML'e çevirir (bluemonday UGCPolicy varsayılan).

```go
unsafe := "<script>alert('x')</script><b>kalın</b> <a href='http://ex.com' onclick='x()'>link</a>"
safe := text.SanitizeHTML(unsafe) // script/onclick kaldırılır, uygun etiketler korunur

// Özel policy
p := text.HTMLPolicyUGC()
p.AllowAttrs("class").OnElements("span", "div")
safe2 := text.SanitizeHTMLWith(p, unsafe)
```

## Edge-Case ve Hata Yönetimi
- ToSlug: Sadece özel karakterlerden oluşan string için "" döner.
- ToSlugForFile: Birden fazla noktayı tek uzantıya indirger, uzantı yoksa sadece güvenli ad döner.
- Truncate: n<=0 ise "" döner, n>len(s) ise orijinal string döner.
- Coalesce: Tüm argümanlar boşsa "" döner.
- IsBlank: Sadece boşluk karakterleri varsa true, en az bir harf/rakam varsa false.
- SanitizeHTML: script/style event handler'ları (onclick vb.) kaldırılır.

## Notlar
- Tüm fonksiyonlar Unicode/rune-safe çalışır.
- Türkçe karakter desteği ToSlug ve ToSlugForFile fonksiyonlarında mevcuttur.
- HTML sanitize işlemi için `github.com/microcosm-cc/bluemonday` kullanılır.
