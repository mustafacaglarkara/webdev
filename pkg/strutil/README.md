# strutil

Küçük, bağımsız string yardımcıları paketi.

Bu paket, projede sık kullanılan küçük string operasyonlarını merkezi bir yerde toplar. `pkg/text` daha yüksek seviyeli metin işleme (sanitize vb.) için kullanılmaya devam edebilir; ancak temel dönüşümler burada saklanır.

Kısa kullanım örnekleri:

```go
import "github.com/mustafacaglarkara/webdev/pkg/strutil"

// SplitAndTrim
parts := strutil.SplitAndTrim("a, b, c") // ["a","b","c"]

// ToSlug
slug := strutil.ToSlug("Çağrı & Örnek!") // "cagri-ornek"

// NormalizeSpace
s := strutil.NormalizeSpace("  çok   boşluk  ") // "çok boşluk"

// ToSlugForFile
fn := strutil.ToSlugForFile("Örnek Dosya.JPG") // "ornek-dosya.jpg"
```

Testler:
- `go test ./pkg/strutil -v` çalıştırarak birim testleri çalıştırabilirsin.

Tasarım tercihleri:
- Küçük, bağımsız fonksiyonlar; dışarıya çıta (wrapper) bırakmak yerine doğrudan kullanılabilir.
- Unicode/rune güvenli operasyonlar (ör. `ReverseString`, `Truncate`).

