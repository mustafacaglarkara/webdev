# crypto paketi

Bu paket, Go ile kriptografi işlemleri için yardımcı fonksiyonlar sunar. Şifreleme, hash, base64, HMAC, AES-GCM, bcrypt ve token üretimi gibi işlemler için kullanılır.

## Kurulum

Modülünüze ekleyin:

```
go get <sizin-modul-adiniz>/pkg/crypto
```

## Fonksiyonlar ve Kullanım Örnekleri

### MD5 ve SHA256 Hash
```go
import "your/module/path/pkg/crypto"

hash := crypto.MD5Hash("merhaba")
sha := crypto.SHA256Hash("merhaba")
```

### Base64 Encode/Decode
```go
b64 := crypto.Base64Encode("test")
dec, _ := crypto.Base64Decode(b64)
```

### Bcrypt ile Şifre Hashleme
```go
hash, _ := crypto.HashPassword("sifre123")
valid := crypto.CheckPassword(hash, "sifre123") // true
```

### Rastgele Bearer Token
```go
tok, _ := crypto.GenerateBearerToken(32)
```

### AES-GCM ile Şifreleme/Çözme
```go
enc, _ := crypto.EncryptAESGCM("gizli veri", "anahtar123")
dec, _ := crypto.DecryptAESGCM(enc, "anahtar123")
```

### HMAC-SHA256 İmzalama/Doğrulama
```go
sig := crypto.HMACSign("mesaj", "hmac-key")
valid := crypto.HMACVerify("mesaj", "hmac-key", sig) // true
```

### Token Üretimi (AES/HMAC ile)
```go
// AES ile
bearer, _ := crypto.GenerateBearerTokenFromCredentials("kullanici", "sifre", crypto.WithAESKey("anahtar"), crypto.WithExpiry(time.Hour))
// HMAC ile
bearer, _ := crypto.GenerateBearerTokenFromCredentials("kullanici", "sifre", crypto.WithHMACKey("hmac-key"), crypto.WithExpiry(time.Hour))
```

### Token Çözümleme
```go
user, pass, valid, err := crypto.ParseBearerToken(bearer, crypto.WithAESKey("anahtar"))
```

### Basit Bearer Token
```go
b := crypto.GenerateBasicBearer("kullanici", "sifre")
```

## Notlar
- AES anahtarı en az 16 karakter olmalıdır.
- HMAC anahtarı gizli tutulmalıdır.
- Token süresi (expiry) ayarlanabilir.

Daha fazla detay için kodu inceleyebilirsiniz.

