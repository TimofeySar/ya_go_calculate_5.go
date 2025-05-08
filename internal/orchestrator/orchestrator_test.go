package orchestrator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TimofeySar/ya_go_calculate.go/internal/calculation"
)

func TestExpressionCalculateResult(t *testing.T) {
	expr := NewExpression("test", "user1", "2+2")
	tasksChan := make(chan *calculation.Task, 1)
	expr.Start(tasksChan)

	for task := range tasksChan {
		expr.UpdateTaskResult(task.ID, 4)
		close(tasksChan)
		break
	}

	if expr.Result != 4 {
		t.Errorf("Expected result 4, got %f", expr.Result)
	}
	if expr.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", expr.Status)
	}
}

func TestServerHandleRegister(t *testing.T) {
	srv := NewServer()
	srv.db.Exec("DELETE FROM users WHERE login = ?", "testuser")

	req, _ := http.NewRequest("POST", "/api/v1/register", strings.NewReader(`{"login": "testuser", "password": "testpass"}`))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestServerHandleLogin(t *testing.T) {
	srv := NewServer()
	srv.db.Exec("DELETE FROM users WHERE login = ?", "testuser")
	req, _ := http.NewRequest("POST", "/api/v1/register", strings.NewReader(`{"login": "testuser", "password": "testpass"}`))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	req, _ = http.NewRequest("POST", "/api/v1/login", strings.NewReader(`{"login": "testuser", "password": "testpass"}`))
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestServerHandleCalculate(t *testing.T) {
	srv := NewServer()
	srv.db.Exec("DELETE FROM users WHERE login = ?", "testuser")
	req, _ := http.NewRequest("POST", "/api/v1/register", strings.NewReader(`{"login": "testuser", "password": "testpass"}`))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	req, _ = http.NewRequest("POST", "/api/v1/login", strings.NewReader(`{"login": "testuser", "password": "testpass"}`))
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	var tokenResp map[string]string
	json.NewDecoder(rr.Body).Decode(&tokenResp)
	token := tokenResp["token"]

	req, _ = http.NewRequest("POST", "/api/v1/calculate", strings.NewReader(`{"expression": "2+2"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}
}
func TestServerHandleGetExpressions(t *testing.T) {
	srv := NewServer()
	req, _ := http.NewRequest("POST", "/api/v1/register", strings.NewReader(`{"login": "testuser", "password": "testpass"}`))
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	req, _ = http.NewRequest("POST", "/api/v1/login", strings.NewReader(`{"login": "testuser", "password": "testpass"}`))
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	var tokenResp map[string]string
	json.NewDecoder(rr.Body).Decode(&tokenResp)
	token := tokenResp["token"]

	req, _ = http.NewRequest("GET", "/api/v1/expressions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
func TestExpressionComplexCalculateResult(t *testing.T) {
	expr := NewExpression("test", "user1", "(2+2)*3")
	tasksChan := make(chan *calculation.Task, 2)
	expr.Start(tasksChan)

	var intermediateResult float64
	taskCount := 0
	for task := range tasksChan {
		if task == nil {
			t.Fatalf("Received nil task")
		}
		var result float64
		t.Logf("Task %d: Operation=%s, Arg1=%f, Arg2=%f", taskCount+1, task.Operation, task.Arg1, task.Arg2)
		switch task.Operation {
		case "+":
			result = task.Arg1 + task.Arg2
			intermediateResult = result
			t.Logf("Calculating %f + %f = %f", task.Arg1, task.Arg2, result)
		case "*":
			result = intermediateResult * task.Arg2
			t.Logf("Calculating %f * %f = %f", intermediateResult, task.Arg2, result)
		}
		t.Logf("Updating task %s with result %f", task.ID, result)
		if !expr.UpdateTaskResult(task.ID, result) {
			t.Fatalf("Failed to update task %s", task.ID)
		}
		taskCount++
		if taskCount == len(expr.TaskOrder) {
			break
		}
	}
	close(tasksChan)

	t.Logf("Final e.Results: %v", expr.Results)
	t.Logf("Final e.Result: %f", expr.Result)
	if expr.Result != 12 {
		t.Errorf("Expected result 12, got %f", expr.Result)
	}
	if expr.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", expr.Status)
	}
}

func TestExpressionMultiplyWithAddition(t *testing.T) {
	expr := NewExpression("test2", "user1", "2*(3+4)")
	tasksChan := make(chan *calculation.Task, 2)
	expr.Start(tasksChan)

	var intermediateResult float64
	taskCount := 0
	for task := range tasksChan {
		if task == nil {
			t.Fatalf("Received nil task")
		}
		var result float64
		t.Logf("Task %d: Operation=%s, Arg1=%f, Arg2=%f", taskCount+1, task.Operation, task.Arg1, task.Arg2)
		switch task.Operation {
		case "+":
			result = task.Arg1 + task.Arg2
			intermediateResult = result
			t.Logf("Calculating %f + %f = %f", task.Arg1, task.Arg2, result)
		case "*":
			result = task.Arg1 * intermediateResult
			t.Logf("Calculating %f * %f = %f", task.Arg1, intermediateResult, result)
		}
		t.Logf("Updating task %s with result %f", task.ID, result)
		if !expr.UpdateTaskResult(task.ID, result) {
			t.Fatalf("Failed to update task %s", task.ID)
		}
		taskCount++
		if taskCount == len(expr.TaskOrder) {
			break
		}
	}
	close(tasksChan)

	t.Logf("Final e.Results: %v", expr.Results)
	t.Logf("Final e.Result: %f", expr.Result)
	if expr.Result != 14 {
		t.Errorf("Expected result 14, got %f", expr.Result)
	}
	if expr.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", expr.Status)
	}
}

func TestExpressionMultipleOperations(t *testing.T) {
	expr := NewExpression("test3", "user1", "(5-2)*3+1")
	tasksChan := make(chan *calculation.Task, 3)
	expr.Start(tasksChan)

	var intermediateResult float64
	taskCount := 0
	for task := range tasksChan {
		if task == nil {
			t.Fatalf("Received nil task")
		}
		var result float64
		t.Logf("Task %d: Operation=%s, Arg1=%f, Arg2=%f", taskCount+1, task.Operation, task.Arg1, task.Arg2)
		switch task.Operation {
		case "-":
			result = task.Arg1 - task.Arg2
			intermediateResult = result
			t.Logf("Calculating %f - %f = %f", task.Arg1, task.Arg2, result)
		case "*":
			result = intermediateResult * task.Arg2
			intermediateResult = result
			t.Logf("Calculating %f * %f = %f", intermediateResult, task.Arg2, result)
		case "+":
			result = intermediateResult + task.Arg2
			t.Logf("Calculating %f + %f = %f", intermediateResult, task.Arg2, result)
		}
		t.Logf("Updating task %s with result %f", task.ID, result)
		if !expr.UpdateTaskResult(task.ID, result) {
			t.Fatalf("Failed to update task %s", task.ID)
		}
		taskCount++
		if taskCount == len(expr.TaskOrder) {
			break
		}
	}
	close(tasksChan)

	t.Logf("Final e.Results: %v", expr.Results)
	t.Logf("Final e.Result: %f", expr.Result)
	if expr.Result != 10 {
		t.Errorf("Expected result 10, got %f", expr.Result)
	}
	if expr.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", expr.Status)
	}
}
