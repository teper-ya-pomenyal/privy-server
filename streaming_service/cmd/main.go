package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	client "github.com/teper-ya-pomenyal/privy_stream/streaming_service/internal/client/grpc"
	"github.com/teper-ya-pomenyal/privy_stream/streaming_service/internal/config"
	delivery "github.com/teper-ya-pomenyal/privy_stream/streaming_service/internal/delivery/http"
	"github.com/teper-ya-pomenyal/privy_stream/streaming_service/internal/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := config.NewConfig()
	streamer := usecase.NewStreamer(cfg.TrackStoragePath)
	conn, err := grpc.NewClient(cfg.CatalogServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	catalog := client.NewCatalogClient(conn)

	handlers := delivery.NewHTTPHandler(streamer, catalog)

	router := handlers.NewRouter()

	srv := http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		fmt.Printf("streaming_service listening on: %s\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("streaming_service остановлени по таймауту: %s", err)
	} else {
		log.Println("streaming_service остановлен gracefully")
	}
}
