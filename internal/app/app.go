package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"time-capsule/config"
	"time-capsule/internal/handler"
	"time-capsule/internal/repository"
	"time-capsule/internal/service"
	"time-capsule/internal/storage"
	"time-capsule/internal/worker"
	"time-capsule/pkg/httpserver"
	"time-capsule/pkg/minio"
	"time-capsule/pkg/mongodb"
)

func Run(cfg *config.Config) {
	ctx := context.Background()

	db, err := mongodb.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create a mongodb connection: %v", err)
	}

	minioStorage, err := minio.New(cfg)
	if err != nil {
		log.Fatalf("failed to create a minio connection: %v", err)
	}

	var (
		rpstry = repository.NewRepository(db)
		strge  = storage.NewMinioStorage(minioStorage, cfg.MinioBucketName)
		svc    = service.NewService(rpstry, strge)
		hndlr  = handler.NewHandler(svc, strge)
		srvr   = httpserver.NewServer()
	)

	go worker.Run(ctx, cfg, rpstry)

	go func() {
		if err = srvr.Run(cfg, hndlr.Router()); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to listen on tcp network: %v", err)
		}
	}()

	log.Printf("time capsule service is up and running on port %s\n", cfg.HttpAddr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)
	<-quit

	log.Println("received shutdown signal. initiating graceful shutdown...")

	if err = srvr.Shutdown(ctx); err != nil {
		log.Printf("error occurred while shutting down http server: %v\n", err)
	}

	if err = db.Client().Disconnect(ctx); err != nil {
		log.Printf("error occurred while disconneting from mongodb: %v\n", err)
	}

	log.Println("have a nice day!")
}
