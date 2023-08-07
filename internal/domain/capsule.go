package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Capsule struct {
	ID        primitive.ObjectID
	UserID    primitive.ObjectID
	Message   string
	Images    []string
	OpenAt    time.Time
	CreatedAt time.Time
}
