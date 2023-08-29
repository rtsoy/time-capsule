package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateCapsuleDTO struct {
	Message string    `json:"message"`
	OpenAt  time.Time `json:"openAt"`
}

type UpdateCapsuleDTO struct {
	Message string    `json:"message"`
	OpenAt  time.Time `json:"openAt"`
}

type Capsule struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"userID"`
	Message   string             `bson:"message"`
	Images    []string           `bson:"images"`
	OpenAt    time.Time          `bson:"openAt"`
	CreatedAt time.Time          `bson:"createdAt"`
}
