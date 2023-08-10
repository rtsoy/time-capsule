package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateCapsuleDTO struct {
	UserID  primitive.ObjectID
	Message string
	Images  []string
	OpenAt  time.Time
}

type UpdateCapsuleDTO struct {
	Message string
	Images  []string
	OpenAt  time.Time
}

type Capsule struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"userID"`
	Message   string             `bson:"message"`
	Images    []string           `bson:"images"`
	OpenAt    time.Time          `bson:"openAt"`
	CreatedAt time.Time          `bson:"createdAt"`
}
