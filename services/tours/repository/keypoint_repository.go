package repository

import (
	"context"
	"time"
	"tours/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type KeyPointRepository struct {
	collection *mongo.Collection
}

func NewKeyPointRepository(db *mongo.Database) *KeyPointRepository {
	return &KeyPointRepository{
		collection: db.Collection("keypoints"),
	}
}

func (r *KeyPointRepository) Create(kp *model.KeyPoint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, kp)
	return err
}

func (r *KeyPointRepository) FindByID(id bson.ObjectID) (*model.KeyPoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var kp model.KeyPoint
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&kp)
	if err != nil {
		return nil, err
	}
	return &kp, nil
}

func (r *KeyPointRepository) FindByTour(tourID bson.ObjectID) ([]model.KeyPoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "order", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"tourId": tourID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var kps []model.KeyPoint
	if err = cursor.All(ctx, &kps); err != nil {
		return nil, err
	}
	return kps, nil
}

func (r *KeyPointRepository) Update(id bson.ObjectID, update bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *KeyPointRepository) Delete(id bson.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
