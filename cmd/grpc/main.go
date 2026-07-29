package main

import (
	"flag"
	"log"
	"net"

	grpcserver "github.com/HonLaderDev/HonLader/frame/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	addr := flag.String("addr", ":1182", "gRPC listen address")
	flag.Parse()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen %s: %v", *addr, err)
	}

	grpcServer := grpc.NewServer()
	grpcserver.Register(grpcServer, grpcserver.New())
	reflection.Register(grpcServer)

	log.Printf("HonLader API gRPC listening on %s", *addr)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("serve grpc: %v", err)
	}
}
