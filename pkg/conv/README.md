# conv paketi

Tip ve sayı dönüşüm yardımcıları.

Bu paket, string ve sayısal tipler arasında güvenli dönüşüm işlemleri için fonksiyonlar sunar. Ayrıca rastgele sayı üretimi ve ondalık yuvarlama gibi yardımcılar da içerir.

## Fonksiyonlar ve Kullanım Örnekleri

### ToInt
Bir string'i int'e çevirir, hata olursa varsayılan değeri döner.
```go
v := conv.ToInt("42", 0) // 42
v2 := conv.ToInt("abc", 5) // 5
```

### ToInt64
Bir string'i int64'e çevirir, hata olursa varsayılan değeri döner.
```go
v := conv.ToInt64("123456789", 0) // 123456789
```

### ToFloat64
Bir string'i float64'e çevirir, hata olursa varsayılan değeri döner.
```go
f := conv.ToFloat64("3.14", 0) // 3.14
```

### ToBool
Bir string'i bool'a çevirir, hata olursa varsayılan değeri döner.
```go
b := conv.ToBool("true", false) // true
```

### ToDuration
Bir string'i time.Duration'a çevirir, hata olursa varsayılan değeri döner.
```go
d := conv.ToDuration("2h45m", time.Minute) // 2 saat 45 dakika
```

### ParseInt
Bir string'i int64'e çevirir, hata döner.
```go
v, err := conv.ParseInt("123")
```

### ParseFloat
Bir string'i float64'e çevirir, hata döner.
```go
f, err := conv.ParseFloat("3.14")
```

### RandomInt
Belirtilen aralıkta rastgele int üretir (her çağrıda farklı değer döner).
```go
r := conv.RandomInt(10, 20) // 10 ile 20 arasında bir sayı
```

### RoundFloat
Bir ondalık sayıyı belirtilen basamağa yuvarlar.
```go
rf := conv.RoundFloat(3.14159, 2) // 3.14
```

## Notlar
- Tüm fonksiyonlar hatalı girişte güvenli şekilde varsayılan değer döner.
- RandomInt fonksiyonu [min, max] aralığında değer üretir.
- Yuvarlama işlemi için RoundFloat kullanılır.

Daha fazla detay için kodu inceleyebilirsiniz.
