package repository

import (
	"context"

	"time-capsule/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const usersCollection = "users"

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(db *mongo.Database) UserRepository {
	return &MongoUserRepository{
		collection: db.Collection(usersCollection),
	}
}

func (r *MongoUserRepository) InsertUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	res, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	user.ID = res.InsertedID.(primitive.ObjectID)

	return user, nil
}

func (r *MongoUserRepository) GetUser(ctx context.Context, filter bson.M) (*domain.User, error) {
	var user domain.User

	if err := r.collection.FindOne(ctx, filter).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
