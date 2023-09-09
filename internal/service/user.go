package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"time-capsule/internal/domain"
	"time-capsule/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 10
	tokenTTL   = 24 * time.Hour * 7

	usernameRegex = `^[A-Za-z0-9]{3,30}$`
	emailRegex    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
)

var (
	ErrTokenCreationFailed = errors.New("failed to create a token")
	ErrPasswordHashFailure = errors.New("failed to hash a password")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidToken        = errors.New("invalid token")
	ErrEmailDuplicate      = errors.New("email already in use")
	ErrUsernameDuplicate   = errors.New("username already in use")
	ErrInvalidEmail        = errors.New("use a valid email address")
	ErrInvalidUsername     = errors.New("username must be between 3 and 30 characters long and can only contain english alphabet letters (both lowercase and uppercase) and digits")
	ErrInvalidPassword     = errors.New("password must be at least 8 characters long and include at least one uppercase letter and one digit")
	ErrTokenExpired        = errors.New("token expired")
)

type userService struct {
	repository repository.UserRepository
}

func NewUserService(repository repository.UserRepository) UserService {
	return &userService{
		repository: repository,
	}
}

func (s *userService) CreateUser(ctx context.Context, input domain.CreateUserDTO) (*domain.User, error) {
	if !usernameValidation(input.Username) {
		return nil, ErrInvalidUsername
	}

	if !emailValidation(input.Email) {
		return nil, ErrInvalidEmail
	}

	if !passwordValidation(input.Password) {
		return nil, ErrInvalidPassword
	}

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
		if mongo.IsDuplicateKeyError(err) {
			if strings.Contains(err.Error(), "email") {
				return nil, ErrEmailDuplicate
			} else if strings.Contains(err.Error(), "username") {
				return nil, ErrUsernameDuplicate
			}
		}

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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": user.ID,
		"exp":    time.Now().UTC().Add(tokenTTL).Unix(),
	})

	secret := os.Getenv("JWT_SECRET")

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Println("GenerateToken", err)
		return "", ErrTokenCreationFailed
	}

	return signed, nil
}

func (s *userService) ParseToken(accessToken string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		secret := os.Getenv("JWT_SECRET")
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}

		fmt.Println("ParseToken1", err)
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	fmt.Println("ParseToken2", err)
	return nil, ErrInvalidToken
}

func passwordValidation(pw string) bool {
	if len(pw) < 8 {
		return false
	}

	hasUppercase := false
	hasDigit := false

	for _, char := range pw {
		if strings.ContainsRune("ABCDEFGHIJKLMNOPQRSTUVWXYZ", char) {
			hasUppercase = true
		} else if strings.ContainsRune("0123456789", char) {
			hasDigit = true
		}
	}

	return hasUppercase && hasDigit
}

func emailValidation(email string) bool {
	res, _ := regexp.Match(emailRegex, []byte(email))
	return res
}

func usernameValidation(username string) bool {
	res, _ := regexp.Match(usernameRegex, []byte(username))
	return res
}

func comparePasswords(pw, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

func hashPassword(pw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcryptCost)
	return string(hash), err
}
