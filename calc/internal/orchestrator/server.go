package orchestrator

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TimofeySar/ya_go_calculate.go/internal/calculation"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	UnimplementedTaskServiceServer
	Router      *mux.Router
	expressions map[string]*Expression
	tasks       chan *calculation.Task
	mu          sync.Mutex
	db          *sql.DB
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

func NewServer() *Server {
	db, err := sql.Open("sqlite3", "expressions.db")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        login TEXT UNIQUE,
        password TEXT
    )`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS expressions (
        id TEXT PRIMARY KEY,
        user_id INTEGER,
        expression TEXT,
        status TEXT,
        result REAL,
        FOREIGN KEY(user_id) REFERENCES users(id)
    )`)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	srv := &Server{
		Router:      router,
		expressions: make(map[string]*Expression),
		tasks:       make(chan *calculation.Task, 100),
		db:          db,
	}
	router.HandleFunc("/api/v1/register", srv.handleRegister).Methods("POST")
	router.HandleFunc("/api/v1/login", srv.handleLogin).Methods("POST")
	router.HandleFunc("/api/v1/calculate", srv.handleCalculate).Methods("POST")
	router.HandleFunc("/api/v1/expressions", srv.handleGetExpressions).Methods("GET")
	router.HandleFunc("/api/v1/expressions/{id}", srv.handleGetExpression).Methods("GET")
	return srv
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/register" || r.URL.Path == "/api/v1/login" {
			next.ServeHTTP(w, r)
			return
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) { return []byte("secret"), nil })
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		userID, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
			return
		}
		r.Header.Set("X-User-ID", fmt.Sprintf("%.0f", userID))
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleCalculate(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusUnprocessableEntity)
		return
	}
	if req.Expression == "" || strings.ContainsAny(req.Expression, "!@#$%^&") {
		http.Error(w, "Invalid expression", http.StatusUnprocessableEntity)
		return
	}

	id := generateID()
	expr := NewExpression(id, userID, req.Expression)
	s.mu.Lock()
	s.expressions[id] = expr
	s.mu.Unlock()

	uid, _ := strconv.Atoi(userID)
	_, err := s.db.Exec("INSERT INTO expressions (id, user_id, status, expression) VALUES (?, ?, ?, ?)", id, uid, "pending", req.Expression)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	go expr.Start(s.tasks)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (s *Server) handleGetExpressions(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}
	userID, ok := claims["user_id"].(float64)
	if !ok {
		http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
		return
	}

	rows, err := s.db.Query("SELECT id, expression, status, result FROM expressions WHERE user_id = ?", int(userID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var expressions []struct {
		ID         string  `json:"id"`
		Expression string  `json:"expression"`
		Status     string  `json:"status"`
		Result     float64 `json:"result"`
	}
	for rows.Next() {
		var expr struct {
			ID         string  `json:"id"`
			Expression string  `json:"expression"`
			Status     string  `json:"status"`
			Result     float64 `json:"result"`
		}
		if err := rows.Scan(&expr.ID, &expr.Expression, &expr.Status, &expr.Result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		expressions = append(expressions, expr)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(expressions); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusUnprocessableEntity)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	_, err = s.db.Exec("INSERT INTO users (login, password) VALUES (?, ?)", req.Login, hashedPassword)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "Login already exists", http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusUnprocessableEntity)
		return
	}
	var passwordHash []byte
	var userID int
	err := s.db.QueryRow("SELECT id, password FROM users WHERE login = ?", req.Login).Scan(&userID, &passwordHash)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword(passwordHash, []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func (s *Server) handleListExpressions(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	s.mu.Lock()
	defer s.mu.Unlock()
	resp := struct {
		Expressions []Expression `json:"expressions"`
	}{}
	rows, err := s.db.Query("SELECT id, status, result, expr FROM expressions WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var expr Expression
		err := rows.Scan(&expr.ID, &expr.Status, &expr.Result, &expr.Expr)
		if err != nil {
			continue
		}
		resp.Expressions = append(resp.Expressions, expr)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleGetExpression(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	userID := r.Header.Get("X-User-ID")
	s.mu.Lock()
	expr, exists := s.expressions[id]
	s.mu.Unlock()
	if !exists {
		var status, exprStr string
		var result float64
		err := s.db.QueryRow("SELECT status, result, expr FROM expressions WHERE id = ? AND user_id = ?", id, userID).Scan(&status, &result, &exprStr)
		if err != nil {
			http.Error(w, "Expression not found", http.StatusNotFound)
			return
		}
		expr = &Expression{ID: id, Status: status, Result: result, Expr: exprStr, UserID: userID}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]Expression{"expression": *expr})
}

func (s *Server) GetTask(ctx context.Context, _ *Empty) (*Task, error) {
	select {
	case task := <-s.tasks:
		return &Task{
			Id:            task.ID,
			Arg1:          task.Arg1,
			Arg2:          task.Arg2,
			Operation:     task.Operation,
			OperationTime: int32(task.OperationTime),
		}, nil
	default:
		return nil, status.Errorf(codes.NotFound, "No tasks available")
	}
}

func (s *Server) SendResult(ctx context.Context, result *Result) (*Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, expr := range s.expressions {
		if expr.UpdateTaskResult(result.Id, result.Result) {
			_, err := s.db.Exec("UPDATE expressions SET result = ?, status = ? WHERE id = ?", expr.Result, expr.Status, expr.ID)
			if err != nil {
				fmt.Printf("Error updating DB for expr %s: %v\n", expr.ID, err)
				return nil, status.Errorf(codes.Internal, err.Error())
			}
			fmt.Printf("Updated DB for expr %s: result=%f, status=%s\n", expr.ID, expr.Result, expr.Status)
			return &Empty{}, nil
		}
	}
	fmt.Printf("Task %s not found in expressions\n", result.Id)
	return nil, status.Errorf(codes.NotFound, "Task not found")
}

func createTables(db *sql.DB) {
	db.Exec(`CREATE TABLE IF NOT EXISTS users (
            login TEXT PRIMARY KEY,
            password_hash TEXT NOT NULL
        )`)
	db.Exec(`CREATE TABLE IF NOT EXISTS expressions (
            id TEXT PRIMARY KEY,
            user_id TEXT,
            status TEXT,
            result REAL,
            expr TEXT,
            FOREIGN KEY (user_id) REFERENCES users(login)
        )`)
}
