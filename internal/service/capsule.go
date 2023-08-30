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
	ErrInvalidTime  = errors.New("opening time cannot be before than now")
	ErrShortMessage = errors.New("message must be at least 5 characters long")
	ErrForbidden    = errors.New("not allowed")
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

func (s *capsuleService) CreateCapsule(ctx context.Context, userID primitive.ObjectID, input domain.CreateCapsuleDTO) (*domain.Capsule, error) {
	if len(input.Message) < 5 {
		return nil, ErrShortMessage
	}

	if input.OpenAt.UTC().Before(time.Now().UTC()) {
		return nil, ErrInvalidTime
	}

	toInsert := &domain.Capsule{
		UserID:    userID,
		Message:   input.Message,
		Images:    []string{},
		OpenAt:    input.OpenAt.UTC(),
		CreatedAt: time.Now().UTC(),
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

func (s *capsuleService) GetCapsuleByID(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) (*domain.Capsule, error) {
	capsule, err := s.repository.GetCapsule(ctx, bson.M{"_id": id})

	if capsule.UserID != userID {
		return nil, ErrForbidden
	}

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}

		log.Println("GetCapsuleByID", err)
		return nil, ErrDBFailure
	}

	return capsule, nil
}

func (s *capsuleService) UpdateCapsule(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, update domain.UpdateCapsuleDTO) error {
	if _, err := s.GetCapsuleByID(ctx, userID, id); err != nil {
		return err
	}

	updateArgs := bson.M{}

	if update.Message != "" {
		updateArgs["message"] = update.Message
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

func (s *capsuleService) DeleteCapsule(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) error {
	if _, err := s.GetCapsuleByID(ctx, userID, id); err != nil {
		return err
	}

	if err := s.repository.DeleteCapsule(ctx, id); err != nil {
		log.Println("DeleteCapsule", err)
		return ErrDBFailure
	}

	return nil
}

func (s *capsuleService) AddImage(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, image string) error {
	if _, err := s.GetCapsuleByID(ctx, userID, id); err != nil {
		return err
	}

	return s.repository.UpdateCapsule(ctx, id, bson.M{
		"$push": bson.M{
			"images": image,
		},
	})
}

func (s *capsuleService) RemoveImage(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, image string) error {
	if _, err := s.GetCapsuleByID(ctx, userID, id); err != nil {
		return err
	}

	return s.repository.UpdateCapsule(ctx, id, bson.M{
		"$pull": bson.M{
			"images": image,
		},
	})
}
