package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"time-capsule/internal/domain"
	mock_repository "time-capsule/internal/repository/mocks"
	mock_storage "time-capsule/internal/storage/mocks"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/mock/gomock"
)

func TestCapsuleService_CreateCapsule(t *testing.T) {
	type mockBehavior func(r *mock_repository.MockCapsuleRepository, ctx context.Context,
		userID primitive.ObjectID, input domain.CreateCapsuleDTO)

	wayBack := time.Unix(0, 0)
	patches := gomonkey.ApplyFunc(time.Now, func() time.Time { return wayBack })
	defer patches.Reset()

	tests := []struct {
		name          string
		mockBehavior  mockBehavior
		expectedError error
		userID        primitive.ObjectID
		input         domain.CreateCapsuleDTO
	}{
		{
			name: "OK",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, input domain.CreateCapsuleDTO) {
				r.EXPECT().InsertCapsule(ctx, &domain.Capsule{
					Message:   "some message",
					OpenAt:    time.Now().UTC().Add(minOpenAtInterval),
					Images:    []string{},
					CreatedAt: time.Now().UTC(),
				}).Return(&domain.Capsule{}, nil).Times(1)
			},
			expectedError: nil,
			userID:        primitive.NilObjectID,
			input: domain.CreateCapsuleDTO{
				Message: "some message",
				OpenAt:  time.Now().Add(minOpenAtInterval),
			},
		},
		{
			name: "Message-Too-Short",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, input domain.CreateCapsuleDTO) {
			},
			expectedError: ErrShortMessage,
			userID:        primitive.NewObjectID(),
			input: domain.CreateCapsuleDTO{
				Message: strings.Repeat("a", minMessageLength-1),
			},
		},
		{
			name: "OpenAt-Invalid-Time",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, input domain.CreateCapsuleDTO) {
			},
			expectedError: ErrInvalidTime,
			userID:        primitive.NilObjectID,
			input: domain.CreateCapsuleDTO{
				Message: "some message",
				OpenAt:  time.Now().Add(-1 * time.Minute),
			},
		},
		{
			name: "OpenAt-Too-Early",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, input domain.CreateCapsuleDTO) {
			},
			expectedError: ErrOpenTimeTooEarly,
			userID:        primitive.NilObjectID,
			input: domain.CreateCapsuleDTO{
				Message: "some message",
				OpenAt:  time.Now().Add(minOpenAtInterval - 1),
			},
		},
		{
			name: "Creating-DB-Failure",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, input domain.CreateCapsuleDTO) {
				r.EXPECT().InsertCapsule(ctx, &domain.Capsule{
					Message:   "some message",
					OpenAt:    time.Now().UTC().Add(minOpenAtInterval + 1*time.Minute),
					Images:    []string{},
					CreatedAt: time.Now().UTC(),
				}).Return(nil, errors.New("some error")).Times(1)
			},
			expectedError: ErrDBFailure,
			userID:        primitive.NilObjectID,
			input: domain.CreateCapsuleDTO{
				Message: "some message",
				OpenAt:  time.Now().UTC().Add(minOpenAtInterval + 1*time.Minute),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			var (
				rpstry = mock_repository.NewMockCapsuleRepository(c)
				svc    = NewCapsuleService(rpstry, nil)
				ctx    = context.Background()
			)

			test.mockBehavior(rpstry, ctx, test.userID, test.input)

			_, err := svc.CreateCapsule(ctx, test.userID, test.input)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestCapsuleService_GetAllCapsules(t *testing.T) {
	type mockBehavior func(r *mock_repository.MockCapsuleRepository, ctx context.Context,
		userID primitive.ObjectID)

	tests := []struct {
		name          string
		mockBehavior  mockBehavior
		expectedError error
		userID        primitive.ObjectID
	}{
		{
			name: "OK",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID) {
				r.EXPECT().GetCapsules(ctx, bson.M{
					"userID": userID,
				}).Return([]*domain.Capsule{
					{
						Message: "some message",
					},
				}, nil).Times(1)
			},
			expectedError: nil,
			userID:        primitive.NewObjectID(),
		},
		{
			name: "Retrieving-DB-Failure",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID) {
				r.EXPECT().GetCapsules(ctx, bson.M{
					"userID": userID,
				}).Return(nil, errors.New("some error")).Times(1)
			},
			expectedError: ErrDBFailure,
			userID:        primitive.NewObjectID(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			var (
				rpstry = mock_repository.NewMockCapsuleRepository(c)
				svc    = NewCapsuleService(rpstry, nil)
				ctx    = context.Background()
			)

			test.mockBehavior(rpstry, ctx, test.userID)

			_, err := svc.GetAllCapsules(ctx, test.userID)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestCapsuleService_GetCapsuleByID(t *testing.T) {
	type mockBehavior func(r *mock_repository.MockCapsuleRepository, ctx context.Context,
		userID primitive.ObjectID, id primitive.ObjectID)

	tests := []struct {
		name          string
		mockBehavior  mockBehavior
		expectedError error
		userID        primitive.ObjectID
		capsuleID     primitive.ObjectID
	}{
		{
			name: "OK",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: userID}, nil).Times(1)
			},
			expectedError: nil,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
		},
		{
			name: "Forbidden",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: primitive.NewObjectID()}, nil).Times(1)
			},
			expectedError: ErrForbidden,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
		},
		{
			name: "Retrieving-DB-Failure",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(nil, errors.New("some error")).Times(1)
			},
			expectedError: ErrDBFailure,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
		},
		{
			name: "Retrieving-DB-NotFound",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(nil, mongo.ErrNoDocuments).Times(1)
			},
			expectedError: ErrNotFound,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			var (
				rpstry = mock_repository.NewMockCapsuleRepository(c)
				svc    = NewCapsuleService(rpstry, nil)
				ctx    = context.Background()
			)

			test.mockBehavior(rpstry, ctx, test.userID, test.capsuleID)

			_, err := svc.GetCapsuleByID(ctx, test.userID, test.capsuleID)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestCapsuleService_UpdateCapsule(t *testing.T) {
	type mockBehavior func(r *mock_repository.MockCapsuleRepository, ctx context.Context,
		userID primitive.ObjectID, id primitive.ObjectID, update domain.UpdateCapsuleDTO)

	wayBack := time.Unix(0, 0)
	patches := gomonkey.ApplyFunc(time.Now, func() time.Time { return wayBack })
	defer patches.Reset()

	tests := []struct {
		name          string
		mockBehavior  mockBehavior
		expectedError error
		userID        primitive.ObjectID
		capsuleID     primitive.ObjectID
		update        domain.UpdateCapsuleDTO
	}{
		{
			name: "OK",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, update domain.UpdateCapsuleDTO) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: userID, CreatedAt: time.Now().UTC()}, nil).Times(1)

				r.EXPECT().UpdateCapsule(ctx, id, bson.M{"$set": bson.M{
					"message": "some message",
					"openAt":  time.Now().UTC().Add(minOpenAtInterval),
				}}).Return(nil).Times(1)
			},
			expectedError: nil,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
			update: domain.UpdateCapsuleDTO{
				Message: "some message",
				OpenAt:  time.Now().UTC().Add(minOpenAtInterval),
			},
		},
		{
			name: "OK-Only-Message",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, update domain.UpdateCapsuleDTO) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: userID, CreatedAt: time.Now().UTC()}, nil).Times(1)

				r.EXPECT().UpdateCapsule(ctx, id, bson.M{"$set": bson.M{
					"message": "some message",
				}}).Return(nil).Times(1)
			},
			expectedError: nil,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
			update: domain.UpdateCapsuleDTO{
				Message: "some message",
			},
		},
		{
			name: "OK-Only-OpenAt",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, update domain.UpdateCapsuleDTO) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: userID, CreatedAt: time.Now().UTC()}, nil).Times(1)

				r.EXPECT().UpdateCapsule(ctx, id, bson.M{"$set": bson.M{
					"openAt": time.Now().UTC().Add(minOpenAtInterval),
				}}).Return(nil).Times(1)
			},
			expectedError: nil,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
			update: domain.UpdateCapsuleDTO{
				OpenAt: time.Now().UTC().Add(minOpenAtInterval),
			},
		},
		{
			name: "Empty-Input",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, update domain.UpdateCapsuleDTO) {
			},
			expectedError: ErrEmptyUpdate,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
		},
		{
			name: "Forbidden",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, update domain.UpdateCapsuleDTO) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: primitive.NewObjectID(), CreatedAt: time.Now().UTC()}, nil).Times(1)
			},
			expectedError: ErrForbidden,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
			update: domain.UpdateCapsuleDTO{
				Message: "some message",
				OpenAt:  time.Now().UTC().Add(minOpenAtInterval),
			},
		},
		{
			name: "Retrieving-DB-Failure",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, update domain.UpdateCapsuleDTO) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(nil, errors.New("some error")).Times(1)
			},
			expectedError: ErrDBFailure,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
			update: domain.UpdateCapsuleDTO{
				Message: "some message",
				OpenAt:  time.Now().UTC().Add(minOpenAtInterval),
			},
		},
		{
			name: "Retrieving-DB-NotFound",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, update domain.UpdateCapsuleDTO) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(nil, mongo.ErrNoDocuments).Times(1)
			},
			expectedError: ErrNotFound,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
			update: domain.UpdateCapsuleDTO{
				Message: "some message",
				OpenAt:  time.Now().UTC().Add(minOpenAtInterval),
			},
		},
		{
			name: "Update-Too-Late",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, update domain.UpdateCapsuleDTO) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: userID, CreatedAt: time.Now().UTC().Add(-1 * (maxUpdateInterval + 1))}, nil).Times(1)
			},
			expectedError: ErrUpdateTooLate,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
			update: domain.UpdateCapsuleDTO{
				Message: "some message",
				OpenAt:  time.Now().UTC().Add(minOpenAtInterval),
			},
		},
		{
			name: "Message-Too-Short",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, update domain.UpdateCapsuleDTO) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: userID, CreatedAt: time.Now().UTC()}, nil).Times(1)
			},
			expectedError: ErrShortMessage,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
			update: domain.UpdateCapsuleDTO{
				Message: strings.Repeat("a", minMessageLength-1),
			},
		},
		{
			name: "OpenAt-Invalid-Time",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, update domain.UpdateCapsuleDTO) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: userID, CreatedAt: time.Now().UTC()}, nil).Times(1)
			},
			expectedError: ErrInvalidTime,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
			update: domain.UpdateCapsuleDTO{
				OpenAt: time.Now().UTC().Add(-1 * time.Minute),
			},
		},
		{
			name: "OpenAt-Too-Early",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, update domain.UpdateCapsuleDTO) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: userID, CreatedAt: time.Now().UTC()}, nil).Times(1)
			},
			expectedError: ErrOpenTimeTooEarly,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
			update: domain.UpdateCapsuleDTO{
				OpenAt: time.Now().UTC().Add(minOpenAtInterval - 1),
			},
		},
		{
			name: "Updating-DB-Failure",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, update domain.UpdateCapsuleDTO) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: userID, CreatedAt: time.Now().UTC()}, nil).Times(1)

				r.EXPECT().UpdateCapsule(ctx, id, bson.M{"$set": bson.M{
					"message": "some message",
					"openAt":  time.Now().UTC().Add(minOpenAtInterval),
				}}).Return(errors.New("some error")).Times(1)
			},
			expectedError: ErrDBFailure,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
			update: domain.UpdateCapsuleDTO{
				Message: "some message",
				OpenAt:  time.Now().UTC().Add(minOpenAtInterval),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			var (
				rpstry = mock_repository.NewMockCapsuleRepository(c)
				svc    = NewCapsuleService(rpstry, nil)
				ctx    = context.Background()
			)

			test.mockBehavior(rpstry, ctx, test.userID, test.capsuleID, test.update)

			err := svc.UpdateCapsule(ctx, test.userID, test.capsuleID, test.update)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestCapsuleService_DeleteCapsule(t *testing.T) {
	type mockBehavior func(r *mock_repository.MockCapsuleRepository, ctx context.Context,
		userID primitive.ObjectID, id primitive.ObjectID)

	type storageMockBehavior func(s *mock_storage.MockStorage, ctx context.Context, image string)

	tests := []struct {
		name                string
		mockBehavior        mockBehavior
		storageMockBehavior storageMockBehavior
		expectedError       error
		userID              primitive.ObjectID
		capsuleID           primitive.ObjectID
	}{
		{
			name: "OK",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{
					UserID: userID,
					Images: []string{"123.jpg"},
				}, nil).Times(1)

				r.EXPECT().DeleteCapsule(ctx, id).Return(nil).Times(1)
			},
			storageMockBehavior: func(s *mock_storage.MockStorage, ctx context.Context, image string) {
				s.EXPECT().Delete(ctx, image).Return(nil)
			},
			expectedError: nil,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
		},
		{
			name: "Forbidden",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: primitive.NewObjectID()}, nil).Times(1)
			},
			storageMockBehavior: func(s *mock_storage.MockStorage, ctx context.Context, image string) {},
			expectedError:       ErrForbidden,
			userID:              primitive.NewObjectID(),
			capsuleID:           primitive.NewObjectID(),
		},
		{
			name: "Retrieving-DB-Failure",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(nil, errors.New("some error")).Times(1)
			},
			storageMockBehavior: func(s *mock_storage.MockStorage, ctx context.Context, image string) {},
			expectedError:       ErrDBFailure,
			userID:              primitive.NewObjectID(),
			capsuleID:           primitive.NewObjectID(),
		},
		{
			name: "Retrieving-DB-NotFound",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(nil, mongo.ErrNoDocuments).Times(1)
			},
			storageMockBehavior: func(s *mock_storage.MockStorage, ctx context.Context, image string) {},
			expectedError:       ErrNotFound,
			userID:              primitive.NewObjectID(),
			capsuleID:           primitive.NewObjectID(),
		},
		{
			name: "Storage-Failure",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{
					UserID: userID,
					Images: []string{"123.jpg"},
				}, nil).Times(1)
			},
			storageMockBehavior: func(s *mock_storage.MockStorage, ctx context.Context, image string) {
				s.EXPECT().Delete(ctx, image).Return(errors.New("some error"))
			},
			expectedError: ErrStorageFailure,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
		},
		{
			name: "Deleting-DB-Failure",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{
					UserID: userID,
					Images: []string{"123.jpg"},
				}, nil).Times(1)

				r.EXPECT().DeleteCapsule(ctx, id).Return(errors.New("some error")).Times(1)
			},
			storageMockBehavior: func(s *mock_storage.MockStorage, ctx context.Context, image string) {
				s.EXPECT().Delete(ctx, image).Return(nil)
			},
			expectedError: ErrDBFailure,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			var (
				rpstry = mock_repository.NewMockCapsuleRepository(c)
				strge  = mock_storage.NewMockStorage(c)
				svc    = NewCapsuleService(rpstry, strge)
				ctx    = context.Background()
			)

			test.mockBehavior(rpstry, ctx, test.userID, test.capsuleID)
			test.storageMockBehavior(strge, ctx, "123.jpg")

			err := svc.DeleteCapsule(ctx, test.userID, test.capsuleID)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestCapsuleService_AddImage(t *testing.T) {
	type mockBehavior func(r *mock_repository.MockCapsuleRepository, ctx context.Context,
		userID primitive.ObjectID, id primitive.ObjectID, image string)

	tests := []struct {
		name          string
		mockBehavior  mockBehavior
		expectedError error
		userID        primitive.ObjectID
		capsuleID     primitive.ObjectID
		image         string
	}{
		{
			name: "OK",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, image string) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: userID}, nil).Times(1)

				r.EXPECT().UpdateCapsule(ctx, id, bson.M{
					"$push": bson.M{
						"images": image,
					},
				}).Return(nil).Times(1)
			},
			expectedError: nil,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
			image:         "123.jpg",
		},
		{
			name: "Forbidden",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, image string) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: primitive.NewObjectID()}, nil).Times(1)
			},
			expectedError: ErrForbidden,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
		},
		{
			name: "Retrieving-DB-Failure",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, image string) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(nil, errors.New("some error")).Times(1)
			},
			expectedError: ErrDBFailure,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
		},
		{
			name: "Retrieving-DB-NotFound",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, image string) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(nil, mongo.ErrNoDocuments).Times(1)
			},
			expectedError: ErrNotFound,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
		},
		{
			name: "Updating-DB-Failure",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, image string) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: userID}, nil).Times(1)

				r.EXPECT().UpdateCapsule(ctx, id, bson.M{
					"$push": bson.M{
						"images": image,
					},
				}).Return(errors.New("some error")).Times(1)
			},
			expectedError: ErrDBFailure,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			var (
				rpstry = mock_repository.NewMockCapsuleRepository(c)
				svc    = NewCapsuleService(rpstry, nil)
				ctx    = context.Background()
			)

			test.mockBehavior(rpstry, ctx, test.userID, test.capsuleID, test.image)

			err := svc.AddImage(ctx, test.userID, test.capsuleID, test.image)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestCapsuleService_RemoveImage(t *testing.T) {
	type mockBehavior func(r *mock_repository.MockCapsuleRepository, ctx context.Context,
		userID primitive.ObjectID, id primitive.ObjectID, image string)

	tests := []struct {
		name          string
		mockBehavior  mockBehavior
		expectedError error
		userID        primitive.ObjectID
		capsuleID     primitive.ObjectID
		image         string
	}{
		{
			name: "OK",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, image string) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: userID}, nil).Times(1)

				r.EXPECT().UpdateCapsule(ctx, id, bson.M{
					"$pull": bson.M{
						"images": image,
					},
				}).Return(nil).Times(1)
			},
			expectedError: nil,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
			image:         "123.jpg",
		},
		{
			name: "Forbidden",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, image string) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: primitive.NewObjectID()}, nil).Times(1)
			},
			expectedError: ErrForbidden,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
		},
		{
			name: "Retrieving-DB-Failure",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, image string) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(nil, errors.New("some error")).Times(1)
			},
			expectedError: ErrDBFailure,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
		},
		{
			name: "Retrieving-DB-NotFound",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, image string) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(nil, mongo.ErrNoDocuments).Times(1)
			},
			expectedError: ErrNotFound,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
		},
		{
			name: "Updating-DB-Failure",
			mockBehavior: func(r *mock_repository.MockCapsuleRepository, ctx context.Context, userID primitive.ObjectID, id primitive.ObjectID, image string) {
				r.EXPECT().GetCapsule(ctx, bson.M{
					"_id": id,
				}).Return(&domain.Capsule{UserID: userID}, nil).Times(1)

				r.EXPECT().UpdateCapsule(ctx, id, bson.M{
					"$pull": bson.M{
						"images": image,
					},
				}).Return(errors.New("some error")).Times(1)
			},
			expectedError: ErrDBFailure,
			userID:        primitive.NewObjectID(),
			capsuleID:     primitive.NewObjectID(),
			image:         "123.jpg",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			var (
				rpstry = mock_repository.NewMockCapsuleRepository(c)
				svc    = NewCapsuleService(rpstry, nil)
				ctx    = context.Background()
			)

			test.mockBehavior(rpstry, ctx, test.userID, test.capsuleID, test.image)

			err := svc.RemoveImage(ctx, test.userID, test.capsuleID, test.image)
			assert.Equal(t, test.expectedError, err)
		})
	}
}
