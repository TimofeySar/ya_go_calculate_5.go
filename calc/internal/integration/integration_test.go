package integration

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TimofeySar/ya_go_calculate.go/internal/agent"
	"github.com/TimofeySar/ya_go_calculate.go/internal/orchestrator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestIntegration(t *testing.T) {
	srv := orchestrator.NewServer()
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		t.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	orchestrator.RegisterTaskServiceServer(grpcServer, srv)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("gRPC server failed: %v", err)
		}
	}()
	defer grpcServer.Stop()

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	go agent.Run(1, conn)

	time.Sleep(100 * time.Millisecond)

	req, err := http.NewRequest("POST", "/api/v1/register", bytes.NewBuffer([]byte(`{"login":"123123","password":"123123"}`)))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	req, err = http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer([]byte(`{"login":"123123","password":"123123"}`)))
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	var loginResp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&loginResp); err != nil {
		t.Fatal(err)
	}
	token := loginResp["token"]

	req, err = http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer([]byte(`{"expression":"2+2"}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
	var calcResp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&calcResp); err != nil {
		t.Fatal(err)
	}
	exprID := calcResp["id"]

	time.Sleep(2 * time.Second)
	req, err = http.NewRequest("GET", "/api/v1/expressions/"+exprID, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	var exprResp map[string]orchestrator.Expression
	if err := json.NewDecoder(rr.Body).Decode(&exprResp); err != nil {
		t.Fatal(err)
	}
	expr := exprResp["expression"]
	if expr.Result != 4 {
		t.Errorf("expected result 4, got %f", expr.Result)
	}
}
