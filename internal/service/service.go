package service

import (
	"context"
	"errors"

	"time-capsule/internal/domain"
	"time-capsule/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrDBFailure = errors.New("something went wrong... try again later :(")
	ErrNotFound  = errors.New("not found")
)

type Service struct {
	UserService
	CapsuleService
}

func NewService(repository *repository.Repository) *Service {
	return &Service{
		UserService:    NewUserService(repository.UserRepository),
		CapsuleService: NewCapsuleService(repository.CapsuleRepository),
	}
}

type UserService interface {
	CreateUser(ctx context.Context, input domain.CreateUserDTO) (*domain.User, error)
	GenerateToken(ctx context.Context, email, password string) (string, error)
	ParseToken(accessToken string) (*jwtClaims, error)
}

type CapsuleService interface {
	CreateCapsule(ctx context.Context, userID primitive.ObjectID, capsule domain.CreateCapsuleDTO) (*domain.Capsule, error)
	GetAllCapsules(ctx context.Context, userID primitive.ObjectID) ([]*domain.Capsule, error)
	GetCapsuleByID(ctx context.Context, id primitive.ObjectID) (*domain.Capsule, error)
	UpdateCapsule(ctx context.Context, id primitive.ObjectID, update domain.UpdateCapsuleDTO) error
	DeleteCapsule(ctx context.Context, id primitive.ObjectID) error
}
