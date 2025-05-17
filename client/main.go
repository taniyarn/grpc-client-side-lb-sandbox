package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/tanihata/grpc-client-side-lb-sandbox/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

const (
	keepAliveIdle     = 3 * time.Minute
	keepAliveInterval = 30 * time.Second
	timeout           = 5 * time.Second
	keepAliveRetry    = 3
)

// proxyServer はクライアントリクエストをバックエンドサーバーに転送するサーバー
type proxyServer struct {
	pb.UnimplementedHelloServiceServer
	client pb.HelloServiceClient
}

// SayHello はクライアントリクエストをバックエンドサーバーに転送する
func (s *proxyServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	return s.client.SayHello(ctx, req)
}

func main() {
	proxyPort := flag.Int("proxy-port", 50050, "The proxy server port")
	serverAddr := flag.String("server-addr", "server:50051", "Server address (host:port)")
	flag.Parse()

	// Dial オプション
	opts := []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			d := &net.Dialer{
				Timeout: timeout,
				KeepAliveConfig: net.KeepAliveConfig{
					Enable:   true,
					Count:    keepAliveRetry,
					Interval: keepAliveInterval + timeout,
					Idle:     keepAliveIdle,
				},
				DualStack: true,
			}
			return d.DialContext(ctx, "tcp", addr)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                keepAliveInterval,
			Timeout:             timeout,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	}

	conn, err := grpc.NewClient(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewHelloServiceClient(conn)

	// プロキシサーバーを作成
	proxy := &proxyServer{
		client: client,
	}

	// プロキシサーバーの起動
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *proxyPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterHelloServiceServer(s, proxy)

	// リフレクションAPIを有効にする
	reflection.Register(s)

	log.Printf("Proxy server listening on :%d", *proxyPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
