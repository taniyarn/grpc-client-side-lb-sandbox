package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/tanihata/grpc-client-side-lb-sandbox/proto"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedHelloServiceServer
	port int
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	message := fmt.Sprintf("Hello, %s! (from server port %d)", req.GetName(), s.port)
	return &pb.HelloResponse{Message: message}, nil
}

func main() {
	port := flag.Int("port", 50051, "The server port")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterHelloServiceServer(s, &server{port: *port})
	log.Printf("Server listening on :%d", *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
} 