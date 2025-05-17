package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	pb "github.com/tanihata/grpc-client-side-lb-sandbox/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedHelloServiceServer
	port     int
	hostname string
	ipAddr   string
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	message := fmt.Sprintf("Hello, %s! (from server %s, IP: %s, port: %d)", 
		req.GetName(), s.hostname, s.ipAddr, s.port)
	return &pb.HelloResponse{Message: message}, nil
}

func getOutboundIP() string {
	// ダミーの接続を作成して自分のIPアドレスを取得
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "unknown"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func main() {
	// デフォルトのポート番号
	defaultPort := 50051
	if portStr := os.Getenv("SERVER_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			defaultPort = port
		}
	}

	port := flag.Int("port", defaultPort, "The server port")
	flag.Parse()

	// ホスト名とIPアドレスを取得
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	ipAddr := getOutboundIP()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterHelloServiceServer(s, &server{
		port:     *port,
		hostname: hostname, 
		ipAddr:   ipAddr,
	})
	// リフレクションAPIを有効にする
	reflection.Register(s)
	
	log.Printf("Server listening on %s (%s):%d", hostname, ipAddr, *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
} 