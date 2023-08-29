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
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"userID" bson:"userID"`
	Message   string             `json:"message" bson:"message"`
	Images    []string           `json:"images" bson:"images"`
	OpenAt    time.Time          `json:"openAt" bson:"openAt"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}
