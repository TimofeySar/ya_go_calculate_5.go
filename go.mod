module github.com/TimofeySar/ya_go_calculate.go

go 1.23.1

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/google/uuid v1.6.0 
	github.com/gorilla/mux v1.8.1
	github.com/mattn/go-sqlite3 v1.14.22
	golang.org/x/crypto v0.26.0
	google.golang.org/grpc v1.67.0
	google.golang.org/protobuf v1.34.2
)

replace github.com/TimofeySar/ya_go_calculate.go => ./

require (
	golang.org/x/net v0.28.0 
	golang.org/x/sys v0.24.0 
	golang.org/x/text v0.17.0 
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 
)
