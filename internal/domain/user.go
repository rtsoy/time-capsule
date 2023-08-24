package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateUserDTO struct {
	Username string
	Email    string
	Password string
}

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Username     string
	Email        string
	PasswordHash string
	RegisteredAt time.Time
}
