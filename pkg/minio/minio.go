package minio

import (
	"fmt"

	"time-capsule/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func New(cfg *config.Config) (*minio.Client, error) {
	m, err := minio.New(fmt.Sprintf("%s:%s", cfg.MinioHost, cfg.MinioPort), &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("minio connection failed: %s", err)
	}

	return m, nil
}
