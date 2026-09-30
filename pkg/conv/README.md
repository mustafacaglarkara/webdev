# conv

```go
import "github.com/mustafacaglarkara/webdev/pkg/conv"
```

String ve sayısal tipler arasında güvenli dönüşüm, rastgele sayı ve yuvarlama
yardımcıları. Örnek çıktıları `example_test.go` ile doğrulanır.

## Varsayılan değerli dönüşümler

Girdi baştaki/sondaki boşluklar kırpılarak ayrıştırılır; ayrıştırılamazsa
verilen varsayılan döner (hata sessizce yutulur — hatayı görmek için
`ParseInt`/`ParseFloat` kullanın).

```go
fmt.Println(conv.ToInt("42", 0), conv.ToInt(" 42 ", 0), conv.ToInt("abc", 5)) // 42 42 5
fmt.Println(conv.ToInt64("123456789012", 0))                                   // 123456789012
fmt.Println(conv.ToFloat64("3.14", 0), conv.ToFloat64("3,14", -1))             // 3.14 -1
fmt.Println(conv.ToBool("true", false), conv.ToBool("evet", false))            // true false
fmt.Println(conv.ToDuration("2h45m", time.Minute), conv.ToDuration("x", time.Minute)) // 2h45m0s 1m0s
```

- `ToFloat64` ondalık ayırıcı olarak yalnızca `.` kabul eder (`"3,14"` geçersiz).
- `ToBool` `strconv.ParseBool` kurallarını kullanır: `1 t T TRUE true True 0 f F FALSE false False`.

## Hata dönen dönüşümler

```go
v, err := conv.ParseInt("12a")
fmt.Println(v, err) // 0 strconv.ParseInt: parsing "12a": invalid syntax
f, err := conv.ParseFloat("3.14")
```

## Rastgele sayı

```go
// Sözde rastgele (math/rand/v2). max < min ise uçlar yer değiştirir; tüm int
// aralığı dahil hiçbir girdide panik atmaz.
r := conv.RandomInt(10, 20) // 10..20 (uçlar dahil)

// Kriptografik olarak güvenli (crypto/rand), düzgün dağılımlı.
n, err := conv.SecureRandomInt(100000, 999999) // ör. 6 haneli OTP
```

**Güvenlik:** `RandomInt` tahmin edilebilir; token, OTP, parola, kupon kodu
gibi değerler için **mutlaka** `SecureRandomInt` (veya `pkg/id.RandomString`)
kullanın. `SecureRandomInt` rastgele kaynak okunamazsa `ErrRandomSource`'u
saran bir hata döner.

## RoundFloat

Yarım değerleri sıfırdan uzağa yuvarlar. Negatif `precision` onlar/yüzler
basamağına yuvarlar; `NaN`/`±Inf` aynen döner; taşma durumunda değer
değiştirilmez.

```go
fmt.Println(conv.RoundFloat(3.14159, 2), conv.RoundFloat(1234.5, -2), conv.RoundFloat(-2.5, 0)) // 3.14 1200 -3
```

Not: float64 ikili gösterim nedeniyle `RoundFloat(1.005, 2)` sonucu `1` olur
(1.005 aslında 1.00499...). Para hesapları için tam sayı kuruş veya ondalık
tip kullanın.
