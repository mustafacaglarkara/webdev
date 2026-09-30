# grpcx

gRPC istemci bağlantısı ve metadata yardımcıları.

```go
import "github.com/mustafacaglarkara/webdev/pkg/grpcx"
```

## Bağlantı

```go
conn, err := grpcx.Dial(grpcx.DialOptions{
    Address:   "api.internal:443",
    TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12},
    Timeout:   5 * time.Second, // READY olana kadar bekle
    UnaryInterceptors:  []grpc.UnaryClientInterceptor{logUnary},
    StreamInterceptors: []grpc.StreamClientInterceptor{logStream},
})
if err != nil {
    return err
}
defer conn.Close()
```

`DialOptions` alanları:

| Alan | Açıklama |
|---|---|
| `Address` | Şemasız (`host:port`) ise `passthrough` çözümleyici kullanılır; DNS için `dns:///host:port` verin |
| `Timeout` | `> 0` ise bağlantı `READY` olana kadar en fazla bu süre beklenir; aşılırsa bağlantı kapatılıp hata döner. `0` ise tembel bağlantı, hemen döner |
| `TransportCredentials` | Özel kimlik bilgisi (ör. `credentials.NewClientTLSFromFile`) |
| `TLSConfig` | TLS yapılandırması |
| `WithInsecure` | Şifresiz bağlantı (yalnızca yerel/test) |
| `PerRPCCredentials` | Her çağrıya eklenen kimlik (ör. OAuth) |
| `UnaryInterceptors`, `StreamInterceptors` | Aynı bağlantıda zincirlenir |
| `ExtraOptions` | `grpc.NewClient`'a doğrudan geçen ek `grpc.DialOption`'lar |

Taşıma güvenliği önceliği: `TransportCredentials` > `TLSConfig` > `WithInsecure`.
Hiçbiri verilmezse `ErrNoTransportSecurity` döner.

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
conn, err := grpcx.DialContext(ctx, grpcx.DialOptions{Address: "localhost:50051", WithInsecure: true})

// Tembel bağlantıyı daha sonra beklemek için:
err = grpcx.WaitForReady(ctx, conn)
```

`DialStream(StreamDialOptions{...})` geriye dönük uyumluluk için korunur
(Deprecated); yeni kodda `DialOptions.StreamInterceptors` kullanın.

Bağlantılar `grpc.NewClient` ile oluşturulur (`grpc.Dial` kullanılmaz).

## Metadata

```go
// İstemci
ctx := grpcx.NewContextWithMetadata(context.Background(), map[string]string{"x-request-id": "123"})

// Sunucu
md := grpcx.ExtractMetadata(ctx) // map[string][]string, yoksa nil
_ = grpcx.SendHeader(ctx, map[string]string{"x-server": "a"})
_ = grpcx.SetTrailer(ctx, map[string]string{"x-took-ms": "12"})
```

## Güvenlik notları

- Taşıma güvenliği açıkça seçilmelidir; paket sessizce şifresiz bağlantıya düşmez.
- `WithInsecure` yalnızca güvenilir ağlarda/testlerde kullanılmalıdır.
- `PerRPCCredentials` kullanan kimlik bilgilerinin çoğu (ör. OAuth) TLS gerektirir.
