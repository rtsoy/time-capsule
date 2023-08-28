package service

import (
	"context"
	"errors"
	"log"
	"time"

	"time-capsule/internal/domain"
	"time-capsule/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrDBFailure   = errors.New("something went wrong... try again later :(")
	ErrNotFound    = errors.New("not found")
	ErrInvalidTime = errors.New("opening time cannot be before than now")
)

type capsuleService struct {
	repository repository.CapsuleRepository
}

func NewCapsuleService(repository repository.CapsuleRepository) CapsuleService {
	return &capsuleService{
		repository: repository,
	}
}

// Todo: min openTime 1 day ? update capsule only 30 min after creation ?

func (s *capsuleService) CreateCapsule(ctx context.Context, userID primitive.ObjectID, capsule domain.CreateCapsuleDTO) (*domain.Capsule, error) {
	toInsert := &domain.Capsule{
		UserID:    userID,
		Message:   capsule.Message,
		Images:    capsule.Images,
		OpenAt:    capsule.OpenAt.UTC(),
		CreatedAt: time.Now().UTC(),
	}

	if toInsert.Images == nil {
		toInsert.Images = []string{}
	}

	res, err := s.repository.InsertCapsule(ctx, toInsert)
	if err != nil {
		log.Println("CreateCapsule", err)
		return nil, ErrDBFailure
	}

	return res, nil
}

func (s *capsuleService) GetAllCapsules(ctx context.Context, userID primitive.ObjectID) ([]*domain.Capsule, error) {
	capsules, err := s.repository.GetCapsules(ctx, bson.M{"userID": userID})
	if err != nil {
		log.Println("GetAllCapsules", err)
		return nil, ErrDBFailure
	}

	return capsules, nil
}

func (s *capsuleService) GetCapsuleByID(ctx context.Context, id primitive.ObjectID) (*domain.Capsule, error) {
	capsule, err := s.repository.GetCapsule(ctx, bson.M{"_id": id})

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}

		log.Println("GetCapsuleByID", err)
		return nil, ErrDBFailure
	}

	return capsule, nil
}

func (s *capsuleService) UpdateCapsule(ctx context.Context, id primitive.ObjectID, update domain.UpdateCapsuleDTO) error {
	updateArgs := bson.M{}

	if update.Message != "" {
		updateArgs["message"] = update.Message
	}

	if len(update.Images) > 0 {
		updateArgs["images"] = update.Images
	}

	if !update.OpenAt.IsZero() {
		if update.OpenAt.Before(time.Now().UTC()) || update.OpenAt.Equal(time.Now().UTC()) {
			return ErrInvalidTime
		}

		updateArgs["openAt"] = update.OpenAt
	}

	if err := s.repository.UpdateCapsule(ctx, id, bson.M{"$set": updateArgs}); err != nil {
		log.Println("UpdateCapsule", err)
		return ErrDBFailure
	}

	return nil
}

func (s *capsuleService) DeleteCapsule(ctx context.Context, id primitive.ObjectID) error {
	if err := s.repository.DeleteCapsule(ctx, id); err != nil {
		log.Println("DeleteCapsule", err)
		return ErrDBFailure
	}

	return nil
}
