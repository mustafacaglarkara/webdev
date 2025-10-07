# grpcx

grpcx paketi, Go ile gRPC istemci ve sunucu işlemlerini kolaylaştırmak için yardımcı fonksiyonlar ve pratik yapılandırmalar sunar.

## Temel Özellikler
- Kolay gRPC bağlantısı (Dial)
- Context ile metadata ekleme/okuma
- Basit interceptor desteği
- Hızlı ve güvenli bağlantı açma (timeout, insecure, interceptor zinciri)
- Client ve server tarafı için pratik context/metadata yönetimi
- Gelişmiş interceptor zinciri ile loglama, retry, trace, auth gibi işlemler

---

## Kullanım

### 1. Bağlantı Kurma
```go
import "your/module/path/pkg/grpcx"

conn, err := grpcx.Dial(grpcx.DialOptions{
    Address: "localhost:50051",
    Timeout: 3 * time.Second,
    WithInsecure: true, // TLS olmadan
})
if err != nil {
    panic(err)
}
defer conn.Close()
```

### 2. Metadata ile Context Oluşturma (Client)
```go
ctx := grpcx.NewContextWithMetadata(context.Background(), map[string]string{
    "authorization": "Bearer ...",
    "trace-id": "abc-123",
})
```

### 3. Context'ten Metadata Okuma (Server)
```go
func (s *MyService) MyMethod(ctx context.Context, req *pb.MyRequest) (*pb.MyResponse, error) {
    md := grpcx.ExtractMetadata(ctx)
    // md["authorization"], md["trace-id"] gibi erişebilirsiniz
    ...
}
```

### 4. Interceptor ile Kullanım (Loglama, Retry, Trace)
```go
import (
    "context"
    "fmt"
    "google.golang.org/grpc"
)

logInterceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
    fmt.Println("gRPC çağrısı:", method)
    return invoker(ctx, method, req, reply, cc, opts...)
}
conn, err := grpcx.Dial(grpcx.DialOptions{
    Address: "localhost:50051",
    WithInsecure: true,
    UnaryInterceptors: []grpc.UnaryClientInterceptor{logInterceptor},
})
```

### 5. Timeout ve Retry Interceptor ile
```go
retryInterceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
    var lastErr error
    for i := 0; i < 3; i++ {
        lastErr = invoker(ctx, method, req, reply, cc, opts...)
        if lastErr == nil {
            return nil
        }
    }
    return lastErr
}
conn, err := grpcx.Dial(grpcx.DialOptions{
    Address: "localhost:50051",
    WithInsecure: true,
    Timeout: 2 * time.Second,
    UnaryInterceptors: []grpc.UnaryClientInterceptor{retryInterceptor},
})
```

### 6. Sunucu Tarafında Metadata Ekleme (Response Metadata)
```go
import "google.golang.org/grpc/metadata"

func (s *MyService) MyMethod(ctx context.Context, req *pb.MyRequest) (*pb.MyResponse, error) {
    header := metadata.Pairs("x-server-version", "1.0.0")
    grpc.SendHeader(ctx, header)
    // ...
    return &pb.MyResponse{}, nil
}
```

### 7. TLS ile Güvenli Bağlantı
```go
import "google.golang.org/grpc/credentials"
creds, _ := credentials.NewClientTLSFromFile("cert.pem", "")
conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
```

---

## Fonksiyonlar ve Parametreler

### DialOptions
- **Address**: Sunucu adresi (host:port)
- **Timeout**: Bağlantı için maksimum bekleme süresi
- **WithInsecure**: TLS olmadan bağlantı (geliştirme/test için)
- **UnaryInterceptors**: []grpc.UnaryClientInterceptor (log, retry, trace, auth vs.)

### Dial
Kolay gRPC bağlantısı kurar. Interceptor ve timeout desteğiyle.

### NewContextWithMetadata
Context'e outgoing metadata ekler (client -> server).

### ExtractMetadata
Context'ten gelen metadata'yı okur (server tarafı).

---

## Sık Kullanılan Senaryolar

- **Loglama/Trace**: Interceptor ile her çağrıyı loglayabilir veya trace id ekleyebilirsiniz.
- **Retry**: Interceptor ile otomatik tekrar mekanizması ekleyebilirsiniz.
- **Auth**: Context'e token ekleyip server tarafında kontrol edebilirsiniz.
- **Timeout**: DialOptions ile bağlantı süresini sınırlandırabilirsiniz.
- **TLS**: WithInsecure=false bırakıp kendi credential'ınızı ekleyebilirsiniz.

---

## Notlar
- Interceptor zinciri ile loglama, retry, trace gibi işlemler kolayca eklenebilir.
- Metadata fonksiyonları hem client hem server tarafında kullanılabilir.
- TLS bağlantısı için WithInsecure=false bırakın ve kendi credential'ınızı ekleyin.
- Daha fazla özellik için katkı yapabilirsiniz.
