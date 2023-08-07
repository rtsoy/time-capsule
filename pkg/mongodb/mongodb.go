package mongodb

import (
	"context"
	"fmt"

	"time-capsule/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func New(ctx context.Context, cfg *config.Config) (*mongo.Database, error) {
	isAuth := false

	mongoURI := fmt.Sprintf("mongodb://%s:%s", cfg.MongoHost, cfg.MongoPort)
	if cfg.MongoUsername != "" && cfg.MongoPassword != "" {
		mongoURI = fmt.Sprintf("mongodb://%s:%s@%s:%s", cfg.MongoUsername, cfg.MongoPassword, cfg.MongoHost, cfg.MongoPort)
		isAuth = true
	}

	clientOpts := options.Client().ApplyURI(mongoURI)
	if isAuth {
		clientOpts.SetAuth(options.Credential{
			Username: cfg.MongoUsername,
			Password: cfg.MongoPassword,
		})
	}

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("mongoDB connection failed: %s", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongoDB: %s", err)
	}

	return client.Database(cfg.MongoDBName), nil
}
