package main

import (
	"fmt"
	"log"
	"net"

	"github.com/nchcl/sd/chat"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Server on")
    
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9001))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := chat.Server{}
	grpcServer := grpc.NewServer()
	chat.RegisterChatServiceServer(grpcServer, &s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

