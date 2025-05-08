package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/TimofeySar/ya_go_calculate.go/internal/orchestrator"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	srv := orchestrator.NewServer()
	grpcServer := grpc.NewServer()
	orchestrator.RegisterTaskServiceServer(grpcServer, srv)
	go func() {
		fmt.Println("gRPC Orchestrator running at :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	fmt.Println("HTTP API running at :8080")
	log.Fatal(http.ListenAndServe(":8080", srv.Router))
}
