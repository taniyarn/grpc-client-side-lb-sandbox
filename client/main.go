package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	pb "github.com/tanihata/grpc-client-side-lb-sandbox/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type proxyServer struct {
	pb.UnimplementedHelloServiceServer
	clients []pb.HelloServiceClient
	mu      sync.Mutex
	index   int // for round-robin
}

func (s *proxyServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	s.mu.Lock()
	client := s.clients[s.index]
	s.index = (s.index + 1) % len(s.clients)
	s.mu.Unlock()

	return client.SayHello(ctx, req)
}

func main() {
	proxyPort := flag.Int("proxy-port", 50050, "The proxy server port")
	serverPorts := flag.String("server-ports", "50051,50052,50053", "Comma-separated list of server ports")
	flag.Parse()

	// Create proxy server
	proxy := &proxyServer{
		clients: make([]pb.HelloServiceClient, 0),
	}

	// Connect to all backend servers
	for _, port := range strings.Split(*serverPorts, ",") {
		conn, err := grpc.Dial("localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("failed to connect to server %s: %v", port, err)
		}
		proxy.clients = append(proxy.clients, pb.NewHelloServiceClient(conn))
		log.Printf("Connected to server on port %s", port)
	}

	// Start proxy server
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