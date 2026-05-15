package repository

import (
	"context"
	"time"
	"tours/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type TouristPositionRepository struct {
	collection *mongo.Collection
}

func NewTouristPositionRepository(db *mongo.Database) *TouristPositionRepository {
	return &TouristPositionRepository{
		collection: db.Collection("tourist_positions"),
	}
}

func (r *TouristPositionRepository) Upsert(position *model.TouristPosition) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"touristId": position.TouristID}

	update := bson.M{
		"$set": bson.M{
			"touristId": position.TouristID,
			"latitude":  position.Latitude,
			"longitude": position.Longitude,
			"updatedAt": position.UpdatedAt,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *TouristPositionRepository) FindByTouristID(touristID string) (*model.TouristPosition, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var position model.TouristPosition
	err := r.collection.FindOne(ctx, bson.M{"touristId": touristID}).Decode(&position)
	if err != nil {
		return nil, err
	}

	return &position, nil
}