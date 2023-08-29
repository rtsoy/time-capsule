package main

import (
	"context"
	"log"
	"net/http"

	"time-capsule/config"
	"time-capsule/internal/handler"
	"time-capsule/internal/repository"
	"time-capsule/internal/service"
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

	var (
		rpstry = repository.NewRepository(db)
		svc    = service.NewService(rpstry)
		hndlr  = handler.NewHandler(svc)
	)

	hndlr.InitRoutes()
	log.Fatal(http.ListenAndServe(":8080", hndlr.Router()))
}
