package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"time-capsule/internal/domain"
	"time-capsule/internal/repository"
	"time-capsule/internal/storage"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	minMessageLength  = 5
	minOpenAtInterval = 24 * time.Hour
	maxUpdateInterval = 30 * time.Minute
)

var (
	ErrInvalidTime      = errors.New("opening time cannot be before than now")
	ErrShortMessage     = errors.New("message must be at least 5 characters long")
	ErrForbidden        = errors.New("not allowed")
	ErrStorageFailure   = errors.New("something went wrong... try again later :(")
	ErrEmptyUpdate      = errors.New("no changes to apply")
	ErrOpenTimeTooEarly = fmt.Errorf("opening time must be at least %d hours from creation date", minOpenAtInterval/time.Hour)
	ErrUpdateTooLate    = fmt.Errorf("updating the capsule is not allowed after %d minutes from creation", maxUpdateInterval/time.Minute)
)

type capsuleService struct {
	repository repository.CapsuleRepository
	storage    storage.Storage
}

func NewCapsuleService(repository repository.CapsuleRepository, storage storage.Storage) CapsuleService {
	return &capsuleService{
		repository: repository,
		storage:    storage,
	}
}

func (s *capsuleService) CreateCapsule(ctx context.Context, userID primitive.ObjectID, input domain.CreateCapsuleDTO) (*domain.Capsule, error) {
	if len(input.Message) < minMessageLength {
		return nil, ErrShortMessage
	}

	if input.OpenAt.UTC().Before(time.Now().UTC()) || input.OpenAt.UTC().Equal(time.Now().UTC()) {
		return nil, ErrInvalidTime
	}

	if input.OpenAt.UTC().Sub(time.Now().UTC()) < minOpenAtInterval {
		return nil, ErrOpenTimeTooEarly
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

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}

		log.Println("GetCapsuleByID", err)
		return nil, ErrDBFailure
	}

	if capsule.UserID != userID {
		return nil, ErrForbidden
	}

	return capsule, nil
}

func (s *capsuleService) UpdateCapsule(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, update domain.UpdateCapsuleDTO) error {
	if reflect.DeepEqual(update, domain.UpdateCapsuleDTO{}) {
		return ErrEmptyUpdate
	}

	capsule, err := s.GetCapsuleByID(ctx, userID, id)
	if err != nil {
		return err
	}

	if time.Now().UTC().Sub(capsule.CreatedAt) > maxUpdateInterval {
		return ErrUpdateTooLate
	}

	updateArgs := bson.M{}

	if update.Message != "" {
		if len(update.Message) < minMessageLength {
			return ErrShortMessage
		}

		updateArgs["message"] = update.Message
	}

	if !update.OpenAt.IsZero() {
		if update.OpenAt.UTC().Before(time.Now().UTC()) || update.OpenAt.Equal(time.Now().UTC()) {
			return ErrInvalidTime
		}

		if update.OpenAt.UTC().Sub(capsule.CreatedAt) < minOpenAtInterval {
			return ErrOpenTimeTooEarly
		}

		updateArgs["openAt"] = update.OpenAt
	}

	if err = s.repository.UpdateCapsule(ctx, id, bson.M{"$set": updateArgs}); err != nil {
		log.Println("UpdateCapsule", err)
		return ErrDBFailure
	}

	return nil
}

func (s *capsuleService) DeleteCapsule(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) error {
	capsule, err := s.GetCapsuleByID(ctx, userID, id)
	if err != nil {
		return err
	}

	for _, img := range capsule.Images {
		if err = s.storage.Delete(ctx, img); err != nil {
			log.Println("DeleteCapsule", err)
			return ErrStorageFailure
		}
	}

	if err = s.repository.DeleteCapsule(ctx, id); err != nil {
		log.Println("DeleteCapsule", err)
		return ErrDBFailure
	}

	return nil
}

func (s *capsuleService) AddImage(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, image string) error {
	if _, err := s.GetCapsuleByID(ctx, userID, id); err != nil {
		return err
	}

	if err := s.repository.UpdateCapsule(ctx, id, bson.M{
		"$push": bson.M{
			"images": image,
		},
	}); err != nil {
		log.Println("AddImage", err)
		return ErrDBFailure
	}

	return nil
}

func (s *capsuleService) RemoveImage(ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, image string) error {
	if _, err := s.GetCapsuleByID(ctx, userID, id); err != nil {
		return err
	}

	if err := s.repository.UpdateCapsule(ctx, id, bson.M{
		"$pull": bson.M{
			"images": image,
		},
	}); err != nil {
		log.Println("RemoveImage", err)
		return ErrDBFailure
	}

	return nil
}
