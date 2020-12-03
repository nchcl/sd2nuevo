package main

import (
	"fmt"
	"log"
	"net"
    "strconv"

	"github.com/nchcl/sd/chat"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Server on")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9002))
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
 
func generar_propuesta(partes int) string {
    var propuesta string
    var nodo int = 1
    for i := 0; i < partes; i++ {
        if i == partes-1 {
            propuesta += strconv.Itoa(nodo)
            break
        }
        
        propuesta += strconv.Itoa(nodo)+","
        if nodo == 3 {
            nodo = 1
        } else {
            nodo++
        }
    }
    
    return propuesta
}
