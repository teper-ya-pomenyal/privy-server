package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	userv1 "github.com/teper-ya-pomenyal/privy_stream/proto/user/v1"
	handler "github.com/teper-ya-pomenyal/privy_stream/user_service/internal/delivery/grpc"
	grpc "google.golang.org/grpc"

	"github.com/teper-ya-pomenyal/privy_stream/jwtmanager"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/config"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/repository/postgres"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/repository/redis"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/usecase"
)

func main() {

	cfg := config.LoadConfig()

	userRepo, err := postgres.NewUsersPostgresRepository(cfg.UserDB)
	if err != nil {
		log.Fatal(err)
	}

	userCache, err := redis.NewRedisSessionStore(cfg.UserCacheDB, cfg.TTLRefresh)
	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := jwtmanager.LoadPrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	tokenManager := jwtmanager.NewManager(privateKey, cfg.TTLAccess)

	login := usecase.NewLoginUseCase(userRepo, tokenManager, userCache)
	register := usecase.NewRegisterUseCase(userRepo, tokenManager, userCache)
	refresh := usecase.NewRefreshUseCase(userRepo, tokenManager, userCache)
	logout := usecase.NewLogoutUseCase(userCache)
	listSessions := usecase.NewListSessionsUseCase(userCache)
	revokeSessions := usecase.NewRevokeSessionsUseCase(userCache)
	listUsers := usecase.NewListUsersUseCase(userRepo)
	setUserBlocked := usecase.NewSetUserBlockedUseCase(userRepo, userCache)
	health := usecase.NewHealthUseCase(userRepo, userCache)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatal(err)
	}

	interceptors := grpc.ChainUnaryInterceptor(handler.RecoveryInterceptor(), handler.LoggingInterceptor())
	grpcServer := grpc.NewServer(interceptors)
	userHandler := handler.NewUserGRPCHandler(login, register, refresh, logout, listSessions, revokeSessions, listUsers, setUserBlocked, health)

	userv1.RegisterUserServiceServer(grpcServer, userHandler)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	grpcDone := make(chan struct{})

	go func() {
		grpcServer.GracefulStop()
		close(grpcDone)
	}()

	select {
	case <-grpcDone:
		log.Println("gRPC сервер остановлен gracefully")
	case <-time.After(10 * time.Second):
		log.Println("gRPC сервер остановлен по таймауту")
	}

	userRepo.Close()

}
