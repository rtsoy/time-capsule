package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LogInUserDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserDTO struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username     string             `json:"username"`
	Email        string             `json:"email"`
	PasswordHash string             `json:"-"`
	RegisteredAt time.Time          `json:"registeredAt"`
}
