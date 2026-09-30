# crypto

Hash, parola özetleme, AES-GCM şifreleme, HMAC ve token yardımcıları.

```go
import "github.com/mustafacaglarkara/webdev/pkg/crypto"
```

## Hash ve Base64

```go
h := crypto.SHA256Hash("merhaba") // hex
m := crypto.MD5Hash("merhaba")    // hex — yalnızca sağlama/uyumluluk için
b := crypto.Base64Encode("veri")
s, err := crypto.Base64Decode(b)
```

> `MD5Hash` ve `SHA256Hash` parola saklamak için **kullanılmamalıdır**.

## Parola (bcrypt)

```go
hash, err := crypto.HashPassword("s3cr3t")
ok := crypto.CheckPassword(hash, "s3cr3t")
```

## Rastgele token

```go
tok, err := crypto.GenerateBearerToken(32) // 32 bayt, base64url; opak oturum/API anahtarı
```

## AES-GCM

```go
ct, err := crypto.EncryptAESGCM("gizli veri", "uzun-bir-anahtar")
pt, err := crypto.DecryptAESGCM(ct, "uzun-bir-anahtar")
```

- Anahtar, her şifrelemede yeni rastgele tuz ile **Argon2id** (t=2, m=19 MiB,
  p=1) kullanılarak 32 bayta türetilir; AES-256-GCM ile şifrelenir.
- Çıktı biçimi sürümlüdür: `v2.` + base64url(tuz | nonce | şifreli metin + etiket).
- `DecryptAESGCM` eski (öneksiz, `SHA-256(anahtar)`) biçimdeki verileri de çözer;
  mevcut veriler kaybolmaz. Yeniden şifrelemek için çözüp `EncryptAESGCM` ile tekrar yazın.
- KDF bilinçli olarak yavaştır (çağrı başına onlarca ms); sıcak yolda her istekte çağırmayın.

## HMAC-SHA256

```go
sig := crypto.HMACSign("mesaj", "anahtar")        // hex
ok := crypto.HMACVerify("mesaj", "anahtar", sig)  // sabit zamanlı
```

## İmzalı token (önerilen)

Parola taşımayan, kimlik/claim tabanlı, süreli token. Çıktı standart bir JWT'dir
(HS256), başka kütüphanelerle de doğrulanabilir.

```go
key := []byte(os.Getenv("TOKEN_KEY")) // en az 32 bayt

tok, err := crypto.GenerateSignedToken(key, "user-42",
    map[string]any{"role": "admin"}, 15*time.Minute)

claims, err := crypto.ParseSignedToken(r.Header.Get("Authorization"), key) // "Bearer " öneki atılır
switch {
case errors.Is(err, crypto.ErrTokenExpired):
case errors.Is(err, crypto.ErrInvalidToken):
case err == nil:
    fmt.Println(claims.Subject, claims.Custom["role"], claims.ExpiresAt)
}
```

Seçenekler: `crypto.WithLeeway(30*time.Second)`, `crypto.WithClock(fn)`,
`crypto.AllowNoExpiry()` (süresiz token üretimi/kabulü — önerilmez).

Hatalar: `ErrInvalidToken`, `ErrTokenExpired`, `ErrNoExpiry`,
`ErrTokenNotYetValid`, `ErrWeakKey`.

## Kimlik bilgisi taşıyan eski token'lar (Deprecated)

`GenerateBearerTokenFromCredentials`, `ParseBearerToken` ve
`GenerateBasicBearer` geriye dönük uyumluluk için korunur ama **Deprecated**
olarak işaretlidir: token içinde parola taşırlar.

```go
tok, err := crypto.GenerateBearerTokenFromCredentials("alice", "s3cr3t",
    crypto.WithHMACKey("hmac-anahtari"), crypto.WithPrefix("Bearer"),
    crypto.WithExpiry(time.Hour))
user, pass, ok, err := crypto.ParseBearerToken(tok,
    crypto.WithHMACKey("hmac-anahtari"), crypto.WithPrefix("Bearer"))

basic := crypto.GenerateBasicBearer("bob", "pwd") // "Bearer base64(bob:pwd)"
```

Modlar:

| Seçenek | Biçim | Koruma |
|---|---|---|
| `WithAESKey` | AES-GCM (v2) şifreli JSON | Gizlilik + bütünlük |
| `WithHMACKey` | `base64(json).hmacHex` | Yalnızca bütünlük — parola base64 içinde **okunabilir** |
| (anahtarsız) | `base64(user:pass)` | **Hiçbiri** — HTTP Basic ile eşdeğer |

`WithExpiry` verilmezse token süresizdir; anahtarsız modda süre hiç uygulanmaz.

## Güvenlik notları

- **Token modları (fail-closed)**: `ParseBearerToken`, `WithAESKey` verildiğinde
  yalnızca AES biçimini, `WithHMACKey` verildiğinde yalnızca imzalı biçimi kabul
  eder; biçime uymayan token `ErrInvalidToken` ile reddedilir. İmzasız
  `base64(user:pass)` yalnızca hiçbir anahtar verilmediğinde kabul edilir.
  (Önceki sürümde HMAC anahtarı tanımlıyken imzasız token kabul ediliyordu.)
- Yeni kodda `GenerateSignedToken` / `ParseSignedToken` kullanın; token'a parola koymayın.
- İmzalı token payload'u şifreli değildir; gizli bilgi koymayın.
- İmza doğrulaması sabit zamanlıdır (`hmac.Equal`); `alg` başlığı yalnızca `HS256` kabul edilir.
