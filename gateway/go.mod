module github.com/teper-ya-pomenyal/privy_stream/gateway

go 1.26.5

require (
	github.com/go-chi/chi/v5 v5.3.2
	github.com/go-chi/cors v1.2.2
	github.com/google/uuid v1.6.0
	github.com/redis/go-redis/v9 v9.22.0
	github.com/teper-ya-pomenyal/privy_stream/jwtmanager v0.0.0-20260902211619-8d2a303aa899
	github.com/teper-ya-pomenyal/privy_stream/proto v0.0.0-20260902211619-8d2a303aa899
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa
	google.golang.org/grpc v1.83.2
	google.golang.org/protobuf v1.36.12
)

replace github.com/teper-ya-pomenyal/privy_stream/jwtmanager => ../jwtmanager

replace github.com/teper-ya-pomenyal/privy_stream/proto => ../proto

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
)
