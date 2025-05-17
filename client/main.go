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

  ctx := context.Background()

  // Dial オプション
  opts := []grpc.DialOption{
    grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
      d := &net.Dialer{
        Timeout:   10 * time.Second,
        KeepAlive: 30 * time.Second,
        DualStack: true,
      }
      return d.DialContext(ctx, "tcp", addr)
    }),
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithKeepaliveParams(keepalive.ClientParameters{
      Time:                10 * time.Second,
      Timeout:             3 * time.Second,
      PermitWithoutStream: true,
    }),
    grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
    grpc.WithDefaultServiceConfig(`
    {
      "loadBalancingConfig":[
        { "round_robin": {} }
      ]
    }`),
  }

  // dns:/// を付ける
  target := fmt.Sprintf("dns:///%s", *serverAddr)
  conn, err := grpc.DialContext(ctx, target, opts...)
  if err != nil {
    log.Fatalf("failed to dial: %v", err)
  }
  defer conn.Close()

  client := pb.NewHelloServiceClient(conn)

	// プロキシサーバーを作成
	proxy := &proxyServer{
	 client: client,
	}

	log.Printf("NEEEEEEEEw :%d", *proxyPort)

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
