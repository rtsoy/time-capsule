package service

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"time-capsule/internal/domain"
	"time-capsule/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 10
	tokenTTL   = 24 * time.Hour * 7
)

var (
	ErrPasswordHashFailure = errors.New("failed to hash a password")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidToken        = errors.New("invalid token")
)

type jwtClaims struct {
	jwt.Claims
	userID primitive.ObjectID
	exp    time.Time
}

type userService struct {
	repository repository.UserRepository
}

func NewUserService(repository repository.UserRepository) UserService {
	return &userService{
		repository: repository,
	}
}

// Todo: Validation (username, email uniqueness, length etc)

func (s *userService) CreateUser(ctx context.Context, input domain.CreateUserDTO) (*domain.User, error) {
	toInsert := &domain.User{
		Username:     input.Username,
		Email:        input.Email,
		RegisteredAt: time.Now().UTC(),
	}

	hash, err := hashPassword(input.Password)
	if err != nil {
		log.Println("hashPassword", err)
		return nil, ErrPasswordHashFailure
	}
	toInsert.PasswordHash = hash

	res, err := s.repository.InsertUser(ctx, toInsert)
	if err != nil {
		log.Println("CreateUser", err)
		return nil, ErrDBFailure
	}

	return res, nil
}

func (s *userService) GenerateToken(ctx context.Context, email, password string) (string, error) {
	user, err := s.repository.GetUser(ctx, bson.M{"email": email})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", ErrInvalidCredentials
		}

		log.Println("GetUser", err)
		return "", ErrDBFailure
	}

	if !comparePasswords(password, user.PasswordHash) {
		return "", ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		exp:    time.Now().UTC().Add(tokenTTL),
		userID: user.ID,
	})

	secret := os.Getenv("JWT_SECRET")

	return token.SignedString(secret)
}

func (s *userService) ParseToken(accessToken string) (*jwtClaims, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		secret := os.Getenv("JWT_SECRET")
		return []byte(secret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func comparePasswords(pw, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(pw), []byte(hash)) == nil
}

func hashPassword(pw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcryptCost)
	return string(hash), err
}
