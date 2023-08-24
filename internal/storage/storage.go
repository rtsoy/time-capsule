package storage

import (
	"context"

	"time-capsule/internal/domain"
)

type Storage interface {
	Upload(ctx context.Context, file domain.File) error
	Get(ctx context.Context, fileName string) (*domain.File, error)
	Delete(ctx context.Context, fileName string) error
}
