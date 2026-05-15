package repository

import (
	"context"
	"time"
	"tours/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ReviewRepository struct {
	collection *mongo.Collection
}

func NewReviewRepository(db *mongo.Database) *ReviewRepository {
	return &ReviewRepository{
		collection: db.Collection("reviews"),
	}
}

func (r *ReviewRepository) Create(review *model.Review) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, review)
	return err
}

func (r *ReviewRepository) FindByID(id bson.ObjectID) (*model.Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var review model.Review
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&review)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ReviewRepository) FindByTour(tourID bson.ObjectID) ([]model.Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"tourId": tourID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reviews []model.Review
	if err = cursor.All(ctx, &reviews); err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewRepository) ExistsByTourAndTourist(tourID bson.ObjectID, touristID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := r.collection.CountDocuments(ctx, bson.M{
		"tourId":    tourID,
		"touristId": touristID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ReviewRepository) Delete(id bson.ObjectID) error {
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
