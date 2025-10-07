// Package grpcx gRPC ile ilgili işleri kolaylaştıran yardımcı fonksiyonlar ve yapılandırmalar sunar.
package grpcx

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// DialOptions, gRPC bağlantısı için kolaylaştırıcı ayarlar.
type DialOptions struct {
	Address           string
	Timeout           time.Duration
	WithInsecure      bool
	UnaryInterceptors []grpc.UnaryClientInterceptor
}

// Dial: Kolay gRPC bağlantısı kurar.
func Dial(opts DialOptions) (*grpc.ClientConn, error) {
	var dialOpts []grpc.DialOption
	if opts.WithInsecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	if len(opts.UnaryInterceptors) > 0 {
		chain := grpc.WithChainUnaryInterceptor(opts.UnaryInterceptors...)
		dialOpts = append(dialOpts, chain)
	}
	if opts.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
		defer cancel()
		return grpc.DialContext(ctx, opts.Address, dialOpts...)
	}
	return grpc.Dial(opts.Address, dialOpts...)
}

// Stream interceptor desteği: gRPC client için stream interceptor zinciri ile bağlantı kurma.
type StreamDialOptions struct {
	Address            string
	Timeout            time.Duration
	WithInsecure       bool
	StreamInterceptors []grpc.StreamClientInterceptor
}

// DialStream: Stream interceptor zinciri ile gRPC bağlantısı kurar.
func DialStream(opts StreamDialOptions) (*grpc.ClientConn, error) {
	var dialOpts []grpc.DialOption
	if opts.WithInsecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	if len(opts.StreamInterceptors) > 0 {
		chain := grpc.WithChainStreamInterceptor(opts.StreamInterceptors...)
		dialOpts = append(dialOpts, chain)
	}
	if opts.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
		defer cancel()
		return grpc.DialContext(ctx, opts.Address, dialOpts...)
	}
	return grpc.Dial(opts.Address, dialOpts...)
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
