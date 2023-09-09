package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"time-capsule/internal/domain"
	mock_repository "time-capsule/internal/repository/mocks"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_CreateUser(t *testing.T) {
	type mockBehavior func(r *mock_repository.MockUserRepository, ctx context.Context,
		input domain.CreateUserDTO)

	wayBack := time.Unix(0, 0)
	patches := gomonkey.ApplyFunc(time.Now, func() time.Time { return wayBack })
	defer patches.Reset()

	hashPatch := gomonkey.ApplyFunc(bcrypt.GenerateFromPassword, func(password []byte, cost int) ([]byte, error) {
		if len(password) > 72 {
			return nil, errors.New("failed to hash")
		}
		return []byte("hash"), nil
	})
	defer hashPatch.Reset()

	tests := []struct {
		name          string
		mockBehavior  mockBehavior
		input         domain.CreateUserDTO
		expectedError error
	}{
		{
			name: "OK",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, input domain.CreateUserDTO) {
				hash, _ := hashPassword(input.Password)

				r.EXPECT().InsertUser(ctx, &domain.User{
					Username:     input.Username,
					Email:        input.Email,
					PasswordHash: hash,
					RegisteredAt: time.Now().UTC(),
				}).Return(&domain.User{}, nil).Times(1)
			},
			input: domain.CreateUserDTO{
				Username: "username123",
				Email:    "foo@example.com",
				Password: "Qwerty123",
			},
			expectedError: nil,
		},
		{
			name:         "Invalid-Username-Special-Characters",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, input domain.CreateUserDTO) {},
			input: domain.CreateUserDTO{
				Username: "!@!@",
			},
			expectedError: ErrInvalidUsername,
		},
		{
			name:         "Invalid-Username-Too-Short",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, input domain.CreateUserDTO) {},
			input: domain.CreateUserDTO{
				Username: "aa",
			},
			expectedError: ErrInvalidUsername,
		},
		{
			name:         "Invalid-Username-Not-English",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, input domain.CreateUserDTO) {},
			input: domain.CreateUserDTO{
				Username: "привет",
			},
			expectedError: ErrInvalidUsername,
		},
		{
			name:         "Invalid-Email",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, input domain.CreateUserDTO) {},
			input: domain.CreateUserDTO{
				Username: "username123",
				Email:    "bad-email",
			},
			expectedError: ErrInvalidEmail,
		},
		{
			name:         "Invalid-Password-Too-Short",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, input domain.CreateUserDTO) {},
			input: domain.CreateUserDTO{
				Username: "username123",
				Email:    "foo@example.com",
				Password: "abc123",
			},
			expectedError: ErrInvalidPassword,
		},
		{
			name:         "Invalid-Password-No-Upper-Case",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, input domain.CreateUserDTO) {},
			input: domain.CreateUserDTO{
				Username: "username123",
				Email:    "foo@example.com",
				Password: "qwerty123",
			},
			expectedError: ErrInvalidPassword,
		},
		{
			name:         "Invalid-Password-No-Digit",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, input domain.CreateUserDTO) {},
			input: domain.CreateUserDTO{
				Username: "username123",
				Email:    "foo@example.com",
				Password: "qwertyqwe",
			},
			expectedError: ErrInvalidPassword,
		},
		{
			name:         "Hash-Failure",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, input domain.CreateUserDTO) {},
			input: domain.CreateUserDTO{
				Username: "username123",
				Email:    "foo@example.com",
				Password: strings.Repeat("a1A", 25),
			},
			expectedError: ErrPasswordHashFailure,
		},
		{
			name: "Creating-DB-Duplicate-Email",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, input domain.CreateUserDTO) {
				hash, _ := hashPassword(input.Password)

				r.EXPECT().InsertUser(ctx, &domain.User{
					Username:     input.Username,
					Email:        input.Email,
					PasswordHash: hash,
					RegisteredAt: time.Now().UTC(),
				}).Return(&domain.User{}, errors.New("duplicate key error collection: time-capsule.users index: email_1 dup key: { email: \"foo@example.com\" })"))
			},
			input: domain.CreateUserDTO{
				Username: "username123",
				Email:    "foo@example.com",
				Password: "Qwerty123",
			},
			expectedError: ErrEmailDuplicate,
		},
		{
			name: "Creating-DB-Duplicate-Username",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, input domain.CreateUserDTO) {
				hash, _ := hashPassword(input.Password)

				r.EXPECT().InsertUser(ctx, &domain.User{
					Username:     input.Username,
					Email:        input.Email,
					PasswordHash: hash,
					RegisteredAt: time.Now().UTC(),
				}).Return(&domain.User{}, errors.New("duplicate key error collection: time-capsule.users index: username_1 dup key: { username: \"username123\" }"))
			},
			input: domain.CreateUserDTO{
				Username: "username123",
				Email:    "foo@example.com",
				Password: "Qwerty123",
			},
			expectedError: ErrUsernameDuplicate,
		},
		{
			name: "Creating-DB-Failure",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, input domain.CreateUserDTO) {
				hash, _ := hashPassword(input.Password)

				r.EXPECT().InsertUser(ctx, &domain.User{
					Username:     input.Username,
					Email:        input.Email,
					PasswordHash: hash,
					RegisteredAt: time.Now().UTC(),
				}).Return(nil, errors.New("some error")).Times(1)
			},
			input: domain.CreateUserDTO{
				Username: "username123",
				Email:    "foo@example.com",
				Password: "Qwerty123",
			},
			expectedError: ErrDBFailure,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.name == "Creating-DB-Duplicate-Email" || test.name == "Creating-DB-Duplicate-Username" {
				duplicatePatch := gomonkey.ApplyFunc(mongo.IsDuplicateKeyError, func(err error) bool { return true })
				defer duplicatePatch.Reset()
			}

			c := gomock.NewController(t)
			defer c.Finish()

			var (
				rpstry = mock_repository.NewMockUserRepository(c)
				svc    = NewUserService(rpstry)
				ctx    = context.Background()
			)

			test.mockBehavior(rpstry, ctx, test.input)

			_, err := svc.CreateUser(ctx, test.input)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestUserService_GenerateToken(t *testing.T) {
	type mockBehavior func(r *mock_repository.MockUserRepository, ctx context.Context,
		email, password string)

	tests := []struct {
		name          string
		mockBehavior  mockBehavior
		email         string
		password      string
		expectedError error
	}{
		{
			name: "OK",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, email, password string) {
				hash, _ := hashPassword(password)

				r.EXPECT().GetUser(ctx, bson.M{"email": email}).Return(&domain.User{PasswordHash: hash}, nil).Times(1)
			},
			email:         "foo@example.com",
			password:      "Qwerty123",
			expectedError: nil,
		},
		{
			name: "Retrieving-DB-Failure",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, email, password string) {
				r.EXPECT().GetUser(ctx, bson.M{"email": email}).Return(nil, errors.New("some error")).Times(1)
			},
			email:         "foo@example.com",
			password:      "Qwerty123",
			expectedError: ErrDBFailure,
		},
		{
			name: "Invalid-Credentials-Email",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, email, password string) {
				r.EXPECT().GetUser(ctx, bson.M{"email": email}).Return(nil, mongo.ErrNoDocuments).Times(1)
			},
			email:         "foo@example.com",
			password:      "Qwerty123",
			expectedError: ErrInvalidCredentials,
		},
		{
			name: "Invalid-Credentials-Password",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, email, password string) {
				r.EXPECT().GetUser(ctx, bson.M{"email": email}).Return(&domain.User{PasswordHash: "some_password"}, nil).Times(1)
			},
			email:         "foo@example.com",
			password:      "Qwerty123",
			expectedError: ErrInvalidCredentials,
		},
		{
			name: "Token-Generation-Failure",
			mockBehavior: func(r *mock_repository.MockUserRepository, ctx context.Context, email, password string) {
				hash, _ := hashPassword(password)

				r.EXPECT().GetUser(ctx, bson.M{"email": email}).Return(&domain.User{PasswordHash: hash}, nil).Times(1)
			},
			email:         "foo@example.com",
			password:      "Qwerty123",
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			var (
				rpstry = mock_repository.NewMockUserRepository(c)
				svc    = NewUserService(rpstry)
				ctx    = context.Background()
			)

			test.mockBehavior(rpstry, ctx, test.email, test.password)

			_, err := svc.GenerateToken(ctx, test.email, test.password)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestUserService_ParseToken(t *testing.T) {
	tests := []struct {
		name          string
		accessToken   func() string
		expectedError error
	}{
		{
			name: "OK",
			accessToken: func() string {
				claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"userID": primitive.NilObjectID,
					"exp":    time.Now().UTC().Add(tokenTTL).Unix(),
				})

				secret := os.Getenv("JWT_SECRET")

				token, _ := claims.SignedString([]byte(secret))

				return token
			},
			expectedError: nil,
		},
		{
			name: "Expired-Token",
			accessToken: func() string {
				claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"userID": primitive.NilObjectID,
					"exp":    time.Now().UTC().Add(-1 * time.Minute).Unix(),
				})

				secret := os.Getenv("JWT_SECRET")

				token, _ := claims.SignedString([]byte(secret))

				return token
			},
			expectedError: ErrTokenExpired,
		},
		{
			name: "Invalid-Token",
			accessToken: func() string {
				return "12312312321"
			},
			expectedError: ErrInvalidToken,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc := NewUserService(nil)

			_, err := svc.ParseToken(test.accessToken())
			assert.Equal(t, test.expectedError, err)
		})
	}
}
