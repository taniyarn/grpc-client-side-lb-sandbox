package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
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
	serverAddr := flag.String("server-ports", "server:50051", "Server address (host:port)")
	flag.Parse()

	// Create proxy server
	proxy := &proxyServer{
		clients: make([]pb.HelloServiceClient, 0),
	}

	// Resolve server address
	host, port, err := net.SplitHostPort(*serverAddr)
	if err != nil {
		log.Fatalf("invalid server address: %v", err)
	}

	// Lookup all IP addresses for the host
	ips, err := net.LookupIP(host)
	if err != nil {
		log.Fatalf("failed to lookup host: %v", err)
	}

	// Connect to all backend servers
	for _, ip := range ips {
		addr := fmt.Sprintf("%s:%s", ip.String(), port)
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("failed to connect to server %s: %v", addr, err)
			continue
		}
		proxy.clients = append(proxy.clients, pb.NewHelloServiceClient(conn))
		log.Printf("Connected to server on %s", addr)
	}

	if len(proxy.clients) == 0 {
		log.Fatal("no servers available")
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