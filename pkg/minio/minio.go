package minio

import (
	"context"
	"fmt"

	"time-capsule/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func New(cfg *config.Config) (*minio.Client, error) {
	m, err := minio.New(fmt.Sprintf("%s:%s", cfg.MinioHost, cfg.MinioPort), &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioUsername, cfg.MinioPassword, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("minio connection failed: %v", err)
	}

	// Ping
	if _, err = m.ListBuckets(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to verify minio connection: %v", err)
	}

	return m, nil
}
