package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/TimofeySar/ya_go_calculate.go/internal/agent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	power, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if power <= 0 {
		power = 1
	}
	fmt.Printf("Agent starting with %d workers\n", power)
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("Failed to connect to orchestrator:", err)
		return
	}
	defer conn.Close()
	agent.Run(power, conn)
}
