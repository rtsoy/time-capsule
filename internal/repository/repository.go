package repository

import (
	"context"

	"time-capsule/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Repository struct {
	UserRepository
	CapsuleRepository
}

type UserRepository interface {
	InsertUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUser(ctx context.Context, filter bson.M) (*domain.User, error)
}

type CapsuleRepository interface {
	InsertCapsule(ctx context.Context, capsule *domain.Capsule) (*domain.Capsule, error)
	GetCapsule(ctx context.Context, filter bson.M) (*domain.Capsule, error)
	GetCapsules(ctx context.Context, filter bson.M) ([]*domain.Capsule, error)
	UpdateCapsule(ctx context.Context, id primitive.ObjectID) error
	DeleteCapsule(ctx context.Context, id primitive.ObjectID) error
}
