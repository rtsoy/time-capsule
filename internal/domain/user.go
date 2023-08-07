package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID
	Username     string
	Email        string
	PasswordHash string
	RegisteredAt time.Time
}
