# strutil

```go
import "github.com/mustafacaglarkara/webdev/pkg/strutil"
```

**Geriye dönük uyumluluk paketidir.** Tüm fonksiyonlar
[`pkg/text`](../text/README.md) paketine yönlendiren ince sarmalayıcılardır;
iki import yolu birebir aynı sonucu verir (`TestDelegatesToText` bunu
doğrular). Yeni kodda `github.com/mustafacaglarkara/webdev/pkg/text` kullanın;
davranış, sınırlar ve ayrıntılı örnekler için onun README'sine bakın.

## Fonksiyonlar

| strutil | Yönlendirildiği yer |
|---|---|
| `SplitAndTrim(s) []string` | `text.SplitAndTrim` |
| `ToSlug(s) string` | `text.ToSlug` |
| `ToSlugForFile(name) string` | `text.ToSlugForFile` |
| `ReverseString(s) string` | `text.ReverseString` |
| `ToUpper(s)`, `ToLower(s)` | `text.ToUpper`, `text.ToLower` (yerel ayardan bağımsız) |
| `ToUpperTR(s)`, `ToLowerTR(s)`, `TitleTR(s)` | Türkçe kurallar |
| `IsBlank(s) bool` | `text.IsBlank` |
| `Coalesce(ss...) string` | `text.Coalesce` |
| `Truncate(s, n) string` | `text.Truncate` |
| `NormalizeSpace(s) string` | `text.NormalizeSpace` |
| `UnescapeHTML(s) string` | `text.UnescapeHTML` |
| `FixTurkishMojibake(s) string` | `text.FixTurkishMojibake` |

HTML temizleme (`SanitizeHTML` vb.) yalnızca `pkg/text` içindedir.

## Örnek

```go
parts := strutil.SplitAndTrim("a, b, c")             // [a b c]
slug := strutil.ToSlug("Çağrı & Örnek!")             // cagri-ornek
s := strutil.NormalizeSpace("  çok   boşluk  ")      // çok boşluk
fn := strutil.ToSlugForFile("Örnek Dosya.JPG")       // ornek-dosya.jpg
up := strutil.ToUpperTR("istanbul")                  // İSTANBUL
```

## Önceki sürüme göre davranış değişiklikleri

- `ToSlugForFile` artık hiçbir girdide panik atmaz (eskiden `".Ⱥ"` gibi
  küçük harfi daha uzun baytlı uzantılarda atıyordu); uzantıyı da temizler;
  boş/`.`/`..`/yalnızca uzantı girdileri için `file` tabanlı ad döner (eskiden
  `""` veya `.jpg` gibi gizli dosya adları dönüyordu).
- `ToSlug` aksanlı harfleri sadeleştirir: `"Kâğıt"` -> `kagit` (eskiden
  `k-git`), `"café"` -> `cafe` (eskiden `caf`).
- `FixTurkishMojibake` `ş`/`Ş`'yi ayırt eder, Windows-1252 bozulmasını onarır,
  doğru `Â` harfini silmez ve deterministiktir.
- `NormalizeSpace` NBSP gibi Unicode boşluklarını da tek boşluğa indirir.
