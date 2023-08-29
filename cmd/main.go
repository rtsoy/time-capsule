package main

import (
	"context"
	"log"
	"net/http"

	"time-capsule/config"
	"time-capsule/internal/handler"
	"time-capsule/internal/repository"
	"time-capsule/internal/service"
	"time-capsule/internal/storage"
	"time-capsule/pkg/minio"
	"time-capsule/pkg/mongodb"
)

func main() {
	ctx := context.Background()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to initiliaze a config: %v", err)
	}

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
		svc    = service.NewService(rpstry)
		hndlr  = handler.NewHandler(svc, strge)
	)

	hndlr.InitRoutes()
	log.Fatal(http.ListenAndServe(":8080", hndlr.Router()))
}
