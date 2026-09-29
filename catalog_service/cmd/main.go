package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/config"
	handler "github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/delivery/grpc"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/repository/postgres"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/usecase"
	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadConfig()

	repo, err := postgres.NewPostgresCatalog(cfg)
	if err != nil {
		log.Fatal(err)
	}

	trackUseCase := usecase.NewTrackUserCase(repo)
	artistUseCase := usecase.NewArtistRepository(repo)
	albumUseCase := usecase.NewAlbumUseCase(repo)
	healthUseCase := usecase.NewHealthUseCase(repo)

	catalogHandler := handler.NewCatalogGRPCHandler(trackUseCase, artistUseCase, albumUseCase, healthUseCase)
	catalogWriteHandler := handler.NewCatalogWriteGRPCHandler(trackUseCase, artistUseCase, albumUseCase)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatal(err)
	}
	writeLis, err := net.Listen("tcp", ":"+cfg.WritePort)
	if err != nil {
		log.Fatal(err)
	}

	interceptors := grpc.ChainUnaryInterceptor(handler.RecoveryInterceptor(), handler.LoggingInterceptor())
	grpcServer := grpc.NewServer(interceptors)
	catalogv1.RegisterCatalogServiceServer(grpcServer, catalogHandler)

	grpcWriteServer := grpc.NewServer(interceptors)
	catalogv1.RegisterCatalogWriteServiceServer(grpcWriteServer, catalogWriteHandler)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	go func() {
		if err := grpcWriteServer.Serve(writeLis); err != nil {
			log.Fatalf("failed to serve write: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	grpcDone := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		grpcWriteServer.GracefulStop()
		close(grpcDone)
	}()

	select {
	case <-grpcDone:
		log.Println("gRPC сервер остановлен gracefully")
	case <-time.After(10 * time.Second):
		log.Println("gRPC сервер остановлен по таймауту")
	}

	repo.Close()
}
