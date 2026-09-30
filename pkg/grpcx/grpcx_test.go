package grpcx

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

func startHealthServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer()
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s, hs)
	go func() { _ = s.Serve(ln) }()
	t.Cleanup(s.Stop)
	return ln.Addr().String()
}

func TestDial_UnaryAndStreamInterceptorsOnOneConn(t *testing.T) {
	addr := startHealthServer(t)
	var unary, stream atomic.Int32
	conn, err := Dial(DialOptions{
		Address:      addr,
		WithInsecure: true,
		Timeout:      5 * time.Second,
		UnaryInterceptors: []grpc.UnaryClientInterceptor{
			func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				unary.Add(1)
				return invoker(ctx, method, req, reply, cc, opts...)
			},
		},
		StreamInterceptors: []grpc.StreamClientInterceptor{
			func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				stream.Add(1)
				return streamer(ctx, desc, cc, method, opts...)
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := healthpb.NewHealthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Check(ctx, &healthpb.HealthCheckRequest{})
	if err != nil || resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		t.Fatalf("check: %v %v", resp, err)
	}
	ws, err := client.Watch(ctx, &healthpb.HealthCheckRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ws.Recv(); err != nil {
		t.Fatal(err)
	}
	if unary.Load() != 1 || stream.Load() != 1 {
		t.Fatalf("interceptors: unary=%d stream=%d", unary.Load(), stream.Load())
	}
}

func TestDial_TimeoutIsEffective(t *testing.T) {
	// Dinlenmeyen bir port: kapatılmış listener adresi.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()
	start := time.Now()
	_, err = Dial(DialOptions{Address: addr, WithInsecure: true, Timeout: 300 * time.Millisecond})
	if err == nil {
		t.Fatal("expected error when server is unreachable and Timeout is set")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline error, got %v", err)
	}
	if el := time.Since(start); el > 3*time.Second {
		t.Fatalf("timeout not respected: %v", el)
	}
	// Timeout yoksa tembel bağlantı hemen döner.
	conn, err := Dial(DialOptions{Address: addr, WithInsecure: true})
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestDial_RequiresTransportSecurity(t *testing.T) {
	if _, err := Dial(DialOptions{Address: "127.0.0.1:1"}); !errors.Is(err, ErrNoTransportSecurity) {
		t.Fatalf("expected ErrNoTransportSecurity, got %v", err)
	}
	conn, err := Dial(DialOptions{Address: "127.0.0.1:1", TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12}})
	if err != nil {
		t.Fatalf("TLS config should be accepted: %v", err)
	}
	conn.Close()
}

func TestDialStream_BackwardCompatible(t *testing.T) {
	addr := startHealthServer(t)
	conn, err := DialStream(StreamDialOptions{Address: addr, WithInsecure: true, Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestMetadataHelpers(t *testing.T) {
	ctx := NewContextWithMetadata(context.Background(), map[string]string{"x-id": "1"})
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok || md.Get("x-id")[0] != "1" {
		t.Fatalf("outgoing md: %v", md)
	}
	in := metadata.NewIncomingContext(context.Background(), metadata.Pairs("a", "b"))
	if ExtractMetadata(in)["a"][0] != "b" {
		t.Fatal("incoming md")
	}
	if ExtractMetadata(context.Background()) != nil {
		t.Fatal("expected nil")
	}
}
