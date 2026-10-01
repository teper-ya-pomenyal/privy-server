package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/teper-ya-pomenyal/privy_stream/gateway/internal/clients"
	"github.com/teper-ya-pomenyal/privy_stream/gateway/internal/config"
	"github.com/teper-ya-pomenyal/privy_stream/gateway/internal/handlers"
	"github.com/teper-ya-pomenyal/privy_stream/gateway/internal/middlewares"
	"github.com/teper-ya-pomenyal/privy_stream/gateway/internal/ratelimit"
	"github.com/teper-ya-pomenyal/privy_stream/gateway/internal/storage"
	"github.com/teper-ya-pomenyal/privy_stream/jwtmanager"
)

func main() {
	cfg := config.LoadConfig()
	userClient, err := clients.NewUserClient(cfg.UserServiceAddress)
	if err != nil {
		log.Fatal(err)
	}

	catalogClient, err := clients.NewCatalogClient(cfg.CatalogServiceAddress, cfg.CatalogWriteServiceAddress)
	if err != nil {
		log.Fatal(err)
	}

	publicKey, err := jwtmanager.LoadPublicKey(cfg.PubKeyAddress)
	if err != nil {
		log.Fatal(err)
	}
	verifier := jwtmanager.NewVerifier(publicKey)
	mw := middlewares.NewMiddleWares(verifier)

	trackStorage := storage.NewTrackStorage(cfg.TrackStoragePath)

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPassword,
	})
	defer rdb.Close()
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 3*time.Second)
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		log.Printf("redis %s unavailable, listen rate limit is disabled until it recovers: %v", cfg.RedisAddress, err)
	}
	pingCancel()
	listenLimiter := ratelimit.NewListenLimiter(rdb, cfg.ListenRateLimitWindow)

	userHandler := handlers.NewUserHandler(userClient)
	catalogHandler := handlers.NewCatalogHandler(catalogClient, trackStorage, listenLimiter)
	adminLogs := handlers.NewAdminLogRing(500)
	adminHandler := handlers.NewAdminHandler(userClient, catalogClient, adminLogs, cfg.StreamingServiceAddress, cfg.TrackStoragePath)
	streamHandler := handlers.NewStreamingHandlers(userClient)

	streamingRouter := streamHandler.NewStreamingRouter(cfg.StreamingServiceAddress, mw)
	router := userHandler.NewRouter(mw, cfg.CORSAllowedOrigins)
	catalogHandler.MountRoutes(router, mw)
	adminHandler.MountRoutes(router, mw)
	router.Mount("/", streamingRouter)
	// Журнал gateway для GET /admin/logs: оборачиваем весь роутер после
	// монтирования, чтобы буфер видел каждый запрос, включая /admin.
	var root http.Handler = handlers.AdminAccessLog(adminLogs)(router)

	srv := http.Server{
		Addr:    ":" + cfg.Port,
		Handler: root,
	}

	go func() {
		log.Printf("gateway listening on :%s\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("gateway остановлен по таймауту: %v", err)
	} else {
		log.Print("gateway остановлен gracefully")
	}
}
