package main

import (
	"log"

	"time-capsule/config"
	"time-capsule/internal/app"
)

// @title TimeCapsule
// @version 1.0
// @description API Server for TimeCapsule Application

// @host localhost:8080
// @Basepath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to initilize a config: %v", err)
	}

	app.Run(cfg)
}
