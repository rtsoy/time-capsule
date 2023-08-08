package repository

import (
	"context"

	"time-capsule/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const capsulesCollection = "capsules"

type MongoCapsuleRepository struct {
	collection *mongo.Collection
}

func NewMongoCapsuleRepository(db *mongo.Database) CapsuleRepository {
	return &MongoCapsuleRepository{
		collection: db.Collection(capsulesCollection),
	}
}

func (r *MongoCapsuleRepository) InsertCapsule(ctx context.Context, capsule *domain.Capsule) (*domain.Capsule, error) {
	res, err := r.collection.InsertOne(ctx, capsule)
	if err != nil {
		return nil, err
	}

	capsule.ID = res.InsertedID.(primitive.ObjectID)

	return capsule, nil
}

func (r *MongoCapsuleRepository) GetCapsule(ctx context.Context, filter bson.M) (*domain.Capsule, error) {
	var capsule domain.Capsule

	if err := r.collection.FindOne(ctx, filter).Decode(&capsule); err != nil {
		return nil, err
	}

	return &capsule, nil
}

func (r *MongoCapsuleRepository) GetCapsules(ctx context.Context, filter bson.M) ([]*domain.Capsule, error) {
	cur, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var capsules []*domain.Capsule
	if err := cur.All(ctx, &capsules); err != nil {
		return nil, err
	}

	return capsules, nil
}

func (r *MongoCapsuleRepository) UpdateCapsule(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)

	return err
}

func (r *MongoCapsuleRepository) DeleteCapsule(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})

	return err
}
