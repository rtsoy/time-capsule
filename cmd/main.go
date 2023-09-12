package main

import (
	"log"

	"time-capsule/config"
	"time-capsule/internal/app"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to initilize a config: %v", err)
	}

	app.Run(cfg)
}
