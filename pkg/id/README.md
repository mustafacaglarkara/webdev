# id

```go
import "github.com/mustafacaglarkara/webdev/pkg/id"
```

UUID ve rastgele kimlik/token üretimi. Tüm fonksiyonlar `crypto/rand`
kullanır ve kriptografik olarak güvenlidir.

## UUIDv4 / MustUUIDv4

RFC 9562 (eski RFC 4122) uyumlu, sürüm 4 rastgele UUID; küçük harfli
`xxxxxxxx-xxxx-4xxx-[89ab]xxx-xxxxxxxxxxxx` biçiminde 36 karakter.

```go
u, err := id.UUIDv4()
if err != nil {
    return err
}
fmt.Println(u) // ör. 3f2504e0-4f89-41d3-9a0c-0305e82c3301

u2 := id.MustUUIDv4() // rastgele kaynak okunamazsa panik atar
```

## RandomString

`[a-zA-Z0-9]` alfabesinden `n` karakterlik, düzgün dağılımlı (modulo sapması
yok) rastgele metin. Oturum/şifre sıfırlama token'ı, API anahtarı, kısa kimlik
ve referans kodu üretimi için uygundur. Karakter başına ~5.95 bit entropi
vardır: 128 bit güvenlik için en az 22 karakter kullanın. `n <= 0` ise `""`.

```go
tok, err := id.RandomString(32)
if err != nil {
    return err
}
fmt.Println(len(tok)) // 32
```

Not: Token'ları veritabanında saklarken ham değer yerine özetini (ör.
SHA-256) saklamanız önerilir.
