package service

import (
	"context"

	"time-capsule/internal/domain"
	"time-capsule/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	UserService
	CapsuleService
}

func NewService(repository *repository.Repository) *Service {
	return &Service{
		UserService:    nil,
		CapsuleService: NewCapsuleService(repository.CapsuleRepository),
	}
}

type UserService interface {
}

type CapsuleService interface {
	CreateCapsule(ctx context.Context, userID primitive.ObjectID, capsule domain.CreateCapsuleDTO) (*domain.Capsule, error)
	GetAllCapsules(ctx context.Context, userID primitive.ObjectID) ([]*domain.Capsule, error)
	GetCapsuleByID(ctx context.Context, id primitive.ObjectID) (*domain.Capsule, error)
	UpdateCapsule(ctx context.Context, id primitive.ObjectID, update domain.UpdateCapsuleDTO) error
	DeleteCapsule(ctx context.Context, id primitive.ObjectID) error
}
