// Package grpcx gRPC ile ilgili işleri kolaylaştıran yardımcı fonksiyonlar ve yapılandırmalar sunar.
package grpcx

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// ErrNoTransportSecurity hiçbir taşıma güvenliği seçeneği verilmediğinde döner.
var ErrNoTransportSecurity = errors.New("grpcx: no transport security configured (set TransportCredentials, TLSConfig or WithInsecure)")

// DialOptions gRPC istemci bağlantısı için tek yapılandırma.
//
// Taşıma güvenliği önceliği: TransportCredentials > TLSConfig > WithInsecure.
// Hiçbiri verilmezse ErrNoTransportSecurity döner (sessizce şifresiz bağlanılmaz).
type DialOptions struct {
	// Address hedef adres. Şema içermiyorsa ("host:port") geriye dönük uyumluluk
	// için "passthrough" çözümleyici kullanılır; DNS çözümleme ve yük dengeleme
	// için "dns:///host:port" verin.
	Address string
	// Timeout > 0 ise Dial, bağlantı READY durumuna gelene kadar en fazla bu süre
	// bekler; süre dolarsa bağlantı kapatılıp hata döner. 0 ise bağlantı tembel
	// (lazy) kurulur ve Dial hemen döner.
	Timeout time.Duration
	// WithInsecure şifresiz bağlantı (yalnızca yerel/test ortamları için).
	WithInsecure bool
	// TLSConfig verilirse TLS kullanılır.
	TLSConfig *tls.Config
	// TransportCredentials özel kimlik bilgileri (ör. credentials.NewClientTLSFromFile).
	TransportCredentials credentials.TransportCredentials
	// PerRPCCredentials her çağrıya eklenecek kimlik bilgileri (ör. OAuth token).
	PerRPCCredentials credentials.PerRPCCredentials

	UnaryInterceptors  []grpc.UnaryClientInterceptor
	StreamInterceptors []grpc.StreamClientInterceptor

	// ExtraOptions doğrudan grpc.NewClient'a eklenen ek seçenekler.
	ExtraOptions []grpc.DialOption
}

// buildDialOptions DialOptions'ı grpc.DialOption listesine çevirir.
func (o DialOptions) buildDialOptions() ([]grpc.DialOption, error) {
	var dialOpts []grpc.DialOption
	switch {
	case o.TransportCredentials != nil:
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(o.TransportCredentials))
	case o.TLSConfig != nil:
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(o.TLSConfig)))
	case o.WithInsecure:
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	default:
		return nil, ErrNoTransportSecurity
	}
	if o.PerRPCCredentials != nil {
		dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(o.PerRPCCredentials))
	}
	if len(o.UnaryInterceptors) > 0 {
		dialOpts = append(dialOpts, grpc.WithChainUnaryInterceptor(o.UnaryInterceptors...))
	}
	if len(o.StreamInterceptors) > 0 {
		dialOpts = append(dialOpts, grpc.WithChainStreamInterceptor(o.StreamInterceptors...))
	}
	dialOpts = append(dialOpts, o.ExtraOptions...)
	return dialOpts, nil
}

func normalizeTarget(addr string) string {
	if strings.Contains(addr, "://") || strings.HasPrefix(addr, "unix:") {
		return addr
	}
	return "passthrough:///" + addr
}

// Dial DialContext(context.Background(), opts) ile aynıdır.
func Dial(opts DialOptions) (*grpc.ClientConn, error) {
	return DialContext(context.Background(), opts)
}

// DialContext grpc.NewClient ile bağlantı oluşturur. opts.Timeout > 0 ise veya
// ctx bir son tarih taşıyorsa bağlantının READY olması beklenir.
func DialContext(ctx context.Context, opts DialOptions) (*grpc.ClientConn, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if opts.Address == "" {
		return nil, errors.New("grpcx: empty address")
	}
	dialOpts, err := opts.buildDialOptions()
	if err != nil {
		return nil, err
	}
	conn, err := grpc.NewClient(normalizeTarget(opts.Address), dialOpts...)
	if err != nil {
		return nil, err
	}
	_, hasDeadline := ctx.Deadline()
	if opts.Timeout <= 0 && !hasDeadline {
		return conn, nil
	}
	waitCtx := ctx
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}
	if err := WaitForReady(waitCtx, conn); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

// WaitForReady bağlantı READY olana kadar (veya ctx bitene kadar) bekler.
func WaitForReady(ctx context.Context, conn *grpc.ClientConn) error {
	conn.Connect()
	for {
		s := conn.GetState()
		switch s {
		case connectivity.Ready:
			return nil
		case connectivity.Shutdown:
			return errors.New("grpcx: connection shut down")
		}
		if !conn.WaitForStateChange(ctx, s) {
			return fmt.Errorf("grpcx: connection not ready (last state %s): %w", s, ctx.Err())
		}
	}
}

// StreamDialOptions geriye dönük uyumluluk için korunur; yeni kodda
// DialOptions.StreamInterceptors kullanın.
type StreamDialOptions struct {
	Address            string
	Timeout            time.Duration
	WithInsecure       bool
	StreamInterceptors []grpc.StreamClientInterceptor
}

// DialStream stream interceptor zinciri ile bağlantı kurar.
//
// Deprecated: Dial ile DialOptions{StreamInterceptors: ...} kullanın; tek
// bağlantıda hem unary hem stream interceptor tanımlanabilir.
func DialStream(opts StreamDialOptions) (*grpc.ClientConn, error) {
	return Dial(DialOptions{
		Address:            opts.Address,
		Timeout:            opts.Timeout,
		WithInsecure:       opts.WithInsecure,
		StreamInterceptors: opts.StreamInterceptors,
	})
}

// NewContextWithMetadata: Metadata ekleyerek context üretir.
func NewContextWithMetadata(ctx context.Context, md map[string]string) context.Context {
	pairs := make([]string, 0, len(md)*2)
	for k, v := range md {
		pairs = append(pairs, k, v)
	}
	return metadata.NewOutgoingContext(ctx, metadata.Pairs(pairs...))
}

// ExtractMetadata: gRPC context'ten metadata okur.
func ExtractMetadata(ctx context.Context) map[string][]string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	return md
}

// SendHeader: Sunucu tarafında response header göndermek için yardımcı.
func SendHeader(ctx context.Context, pairs map[string]string) error {
	md := make([]string, 0, len(pairs)*2)
	for k, v := range pairs {
		md = append(md, k, v)
	}
	return grpc.SendHeader(ctx, metadata.Pairs(md...))
}

// SetTrailer: Sunucu tarafında response trailer göndermek için yardımcı.
func SetTrailer(ctx context.Context, pairs map[string]string) error {
	md := make([]string, 0, len(pairs)*2)
	for k, v := range pairs {
		md = append(md, k, v)
	}
	return grpc.SetTrailer(ctx, metadata.Pairs(md...))
}
