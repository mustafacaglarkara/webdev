# text

```go
import "github.com/mustafacaglarkara/webdev/pkg/text"
```

Metin işlemleri için kanonik paket: slug üretimi, güvenli dosya adı, Türkçe
duyarlı büyük/küçük harf dönüşümü, mojibake onarımı, boşluk normalizasyonu ve
bluemonday tabanlı HTML temizleme.

`pkg/strutil` bu paketin ince bir sarmalayıcısıdır; yeni kodda doğrudan
`pkg/text` kullanın.

Aşağıdaki örneklerin çıktıları `example_test.go` içinde `go test` ile
doğrulanır.

## Slug

### ToSlug

Türkçe harfleri açıkça çevirir (`İ`/`I` dahil), diğer aksanlı harfleri Unicode
NFD normalizasyonu ile sadeleştirir, `a-z0-9` dışındaki her şeyi tek `-` yapar.

```go
fmt.Println(text.ToSlug("Çalışma Alanı - 2025!")) // calisma-alani-2025
fmt.Println(text.ToSlug("İstanbul IŞIK"))         // istanbul-isik
fmt.Println(text.ToSlug("Kâğıt café crème"))      // kagit-cafe-creme
fmt.Printf("%q\n", text.ToSlug("日本語"))          // ""
```

Latin alfabesine çevrilemeyen metinler (CJK, emoji, yalnızca noktalama) için
**boş string** döner; URL üretirken bu durumu kontrol edin.

### ToSlugForFile

Kullanıcıdan gelen dosya adlarından güvenli ad üretir. Hiçbir girdide panik
atmaz (fuzz testi: `FuzzToSlugForFile`). Sonuç her zaman tek bir yol
bileşenidir: yalnızca `[a-z0-9-]` ve en fazla bir `.` içerir, boş değildir,
`.`/`..` olamaz, dizin ayracı, boşluk veya kontrol karakteri içermez.

```go
fmt.Println(text.ToSlugForFile("Çılgın Fotoğraf(1).JPG")) // cilgin-fotograf-1.jpg
fmt.Println(text.ToSlugForFile("rapor.v1.2.PDF"))         // rapor-v1-2.pdf
fmt.Println(text.ToSlugForFile("../../etc/passwd"))       // etc-passwd
fmt.Println(text.ToSlugForFile(".jpg"))                   // file.jpg
fmt.Println(text.ToSlugForFile(""))                       // file
fmt.Println(text.ToSlugForFile("belge."))                 // belge
```

Kurallar:

- Yalnızca son uzantı korunur; uzantı küçük harfe çevrilir, yalnızca ASCII
  harf/rakam bırakılır ve en fazla 16 karakterdir. Temizlendikten sonra boş
  kalan uzantı (sondaki nokta, `x.$$$`) atılır.
- Ad kısmı `ToSlug` ile çevrilir, en fazla 200 bayttır; boşa düşerse `file`
  kullanılır.
- Uzantı yalnızca ada göre temizlenir, içerik doğrulanmaz. Yüklenen dosyanın
  türünü içerikten (ör. MIME sniffing) ayrıca kontrol edin.

## Büyük/küçük harf

`ToUpper`/`ToLower` yerel ayardan bağımsızdır (`strings.ToUpper/ToLower`).
Türkçe metinlerde `ToUpperTR`, `ToLowerTR`, `TitleTR` kullanın.

```go
fmt.Println(text.ToUpper("istanbul ılık"))      // ISTANBUL ILIK
fmt.Println(text.ToUpperTR("istanbul ılık"))    // İSTANBUL ILIK
fmt.Println(text.ToLowerTR("IŞIK İSTANBUL"))    // ışık istanbul
fmt.Println(text.TitleTR("iSTANBUL ılık"))      // İstanbul Ilık
```

Not: `text.ToLower("IŞIK")` sonucu `"işik"`tir (`I` -> `i`); Türkçe için
`ToLowerTR` kullanın (`"ışık"`).

## FixTurkishMojibake

UTF-8 olarak kaydedilmiş ancak Latin-1 (ISO-8859-1), Windows-1252,
Windows-1254 veya ISO-8859-9 olarak çözülmüş metni onarır. `ş`/`Ş`, `ğ`/`Ğ`,
`ı`/`İ` dahil tüm Türkçe harfleri ayırt eder, doğru metne dokunmaz ve
deterministiktir.

```go
fmt.Println(text.FixTurkishMojibake("GÃ¼nÃ¼n Ã¶zeti: Ã‡aÄŸdaÅŸlÄ±k")) // Günün özeti: Çağdaşlık
fmt.Println(text.FixTurkishMojibake("ÅŸ Åž Ä± Ä°"))                    // ş Ş ı İ
fmt.Println(text.FixTurkishMojibake("KÂR"))                            // KÂR
```

Yöntem: metin önce tek baytlık kodlamayla bayta geri çevrilir; sonuç geçerli
UTF-8 ve makul (Latin/Türkçe harf, noktalama) ise kabul edilir. Olmazsa sabit
sıralı bir Türkçe harf tablosu uygulanır (doğru ve bozuk karakterlerin karışık
olduğu metinler için). Sınırlama: bozulma sırasında bayt kaybolmuşsa (ör.
Windows-1254'te tanımsız `0x9E` nedeniyle `Ş`/`Ğ` yerine `�` oluşmuşsa veya
C1 kontrol karakterleri silinmişse, yalnız kalan `Ã`, `Å` gibi) onarım mümkün
değildir; bu karakterler olduğu gibi bırakılır.

## HTML temizleme

`SanitizeHTML` bluemonday UGC politikasını, `SanitizeHTMLStrict` tüm etiketleri
kaldıran politikayı kullanır. Politikalar paket yüklenirken bir kez kurulur ve
eşzamanlı kullanım için güvenlidir.

```go
unsafe := `<script>alert('x')</script><b>kalın</b> <a href="http://ex.com" onclick="x()">link</a>`
fmt.Println(text.SanitizeHTML(unsafe))
// <b>kalın</b> <a href="http://ex.com" rel="nofollow">link</a>
fmt.Println(text.SanitizeHTMLStrict(unsafe))
// kalın link

// Özel politika: HTMLPolicyUGC / HTMLPolicyStrict her çağrıda YENİ bir kopya
// döner; özelleştirmeler varsayılanı etkilemez. Politikayı bir kez kurup
// saklayın (ör. paket değişkeni), her istekte yeniden oluşturmayın.
p := text.HTMLPolicyUGC()
p.AllowAttrs("class").OnElements("span")
fmt.Println(text.SanitizeHTMLWith(p, `<span class="not">x</span>`))
// <span class="not">x</span>
```

`SanitizeHTMLWith(nil, s)` varsayılan UGC politikasını kullanır.

Güvenlik notu: temizlenmiş HTML'i şablonda kaçışsız basarken yalnızca bu
fonksiyonların çıktısını kullanın; temizlik sonrası metne tekrar ekleme
yapmayın.

## Diğer yardımcılar

```go
fmt.Println(text.Truncate("merhaba dünya", 9))                  // merhaba d
fmt.Println(text.NormalizeSpace("  çok   fazla\t boşluk \n"))   // çok fazla boşluk
fmt.Println(text.ReverseString("İstanbul"))                     // lubnatsİ
fmt.Println(text.Coalesce("", "", "ilk"))                       // ilk
fmt.Println(text.IsBlank(" \t\n"))                              // true
fmt.Println(text.UnescapeHTML("&Ccedil;alışma &amp; Deneme"))   // Çalışma & Deneme
fmt.Println(text.SplitAndTrim(" a , b ,, c "))                  // [a b c]
```

- `Truncate(s, n)`: rune sayısına göre keser; `n <= 0` ise `""`.
- `NormalizeSpace`: tüm Unicode boşluklarını (NBSP dahil) tek boşluğa indirir.
- `Coalesce`: boş olmayan ilk metni döner (yalnızca boşluktan oluşan metin boş
  sayılmaz; gerekirse `IsBlank` ile birleştirin).
- `SplitAndTrim`: virgülle böler, kırpar, boşları atar; boş girdide `nil`.

## API özeti

| Fonksiyon | Açıklama |
|---|---|
| `ToSlug(s) string` | URL slug'ı; çevrilemeyen girdi için `""` |
| `ToSlugForFile(name) string` | Güvenli dosya adı; asla boş/panik değil |
| `ToUpper`, `ToLower` | Yerel ayardan bağımsız |
| `ToUpperTR`, `ToLowerTR`, `TitleTR` | Türkçe kurallar |
| `FixTurkishMojibake(s) string` | Latin-1/1252/1254/8859-9 onarımı |
| `SanitizeHTML`, `SanitizeHTMLStrict`, `SanitizeHTMLWith` | bluemonday |
| `HTMLPolicyUGC`, `HTMLPolicyStrict` | Özelleştirilebilir politika kopyası |
| `ReverseString`, `Truncate`, `NormalizeSpace`, `IsBlank`, `Coalesce`, `UnescapeHTML`, `SplitAndTrim` | Genel |
