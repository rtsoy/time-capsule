package service

import (
	"context"
	"errors"
	"fmt"
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
		log.Printf("failed to hash a password: %s\n", err)
		return nil, errors.New("failed to hash a password")
	}
	toInsert.PasswordHash = hash

	res, err := s.repository.InsertUser(ctx, toInsert)
	if err != nil {
		log.Printf("failed to insert a user: %s\n", err)
		return nil, errors.New("failed to insert a user")
	}

	return res, nil
}

func (s *userService) GenerateToken(ctx context.Context, email, password string) (string, error) {
	user, err := s.repository.GetUser(ctx, bson.M{"email": email})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", errors.New("invalid credentials")
		}

		log.Printf("failed to find a user in db: %s\n", err)
		return "", errors.New("failed to find a user in db")
	}

	if !comparePasswords(password, user.PasswordHash) {
		return "", errors.New("invalid credentials")
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
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		secret := os.Getenv("JWT_SECRET")
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func comparePasswords(pw, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(pw), []byte(hash)) == nil
}

func hashPassword(pw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcryptCost)
	return string(hash), err
}
