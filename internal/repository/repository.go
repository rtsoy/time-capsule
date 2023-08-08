package repository

import (
	"context"

	"time-capsule/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	UserRepository
	CapsuleRepository
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		UserRepository:    NewMongoUserRepository(db),
		CapsuleRepository: NewMongoCapsuleRepository(db),
	}
}

type UserRepository interface {
	InsertUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUser(ctx context.Context, filter bson.M) (*domain.User, error)
}

type CapsuleRepository interface {
	InsertCapsule(ctx context.Context, capsule *domain.Capsule) (*domain.Capsule, error)
	GetCapsule(ctx context.Context, filter bson.M) (*domain.Capsule, error)
	GetCapsules(ctx context.Context, filter bson.M) ([]*domain.Capsule, error)
	UpdateCapsule(ctx context.Context, id primitive.ObjectID, update bson.M) error
	DeleteCapsule(ctx context.Context, id primitive.ObjectID) error
}
