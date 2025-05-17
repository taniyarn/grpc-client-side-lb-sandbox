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
	serverAddr := flag.String("server-ports", "server:50051", "Server address (host:port)")
	flag.Parse()

	const (
		timeout           = 10 * time.Second
		keepAliveRetry    = 3
		keepAliveInterval = 10 * time.Second
		keepAliveIdle     = 30 * time.Second
	)
	
	// gRPC クライアント接続オプションの設定
	opts := []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
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
			conn, err := d.DialContext(ctx, "tcp", s)
			if err != nil {
				return nil, err
			}
			return conn, nil
		}),
		// WaitForReady を使用して、サーバーが利用可能になるまで待機
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		// ラウンドロビンロードバランシングポリシーを設定
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		// 非暗号化通信を使用
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// バックエンドサーバーへの接続を確立
	conn, err := grpc.NewClient(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("failed to connect to backend server: %v", err)
	}
	defer conn.Close()

	// gRPC クライアントを作成
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
	log.Printf("Proxy server listening on :%d", *proxyPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
} 