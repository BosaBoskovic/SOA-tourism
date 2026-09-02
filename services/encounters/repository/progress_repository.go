package repository

import (
	"context"
	"time"

	"encounters/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ProgressRepository struct {
	collection *mongo.Collection
}

func NewProgressRepository(db *mongo.Database) *ProgressRepository {
	return &ProgressRepository{collection: db.Collection("encounter_progress")}
}

// Claim is idempotent (upsert): claiming the same encounter twice just
// returns the original completion record instead of erroring or duplicating.
func (r *ProgressRepository) Claim(touristID string, encounterID bson.ObjectID) (*model.EncounterProgress, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"touristId": touristID, "encounterId": encounterID}
	update := bson.M{
		"$setOnInsert": bson.M{
			"touristId":   touristID,
			"encounterId": encounterID,
			"completedAt": time.Now(),
		},
	}
	opts := options.UpdateOne().SetUpsert(true)
	if _, err := r.collection.UpdateOne(ctx, filter, update, opts); err != nil {
		return nil, err
	}

	var progress model.EncounterProgress
	if err := r.collection.FindOne(ctx, filter).Decode(&progress); err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *ProgressRepository) FindByTourist(touristID string) ([]model.EncounterProgress, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"touristId": touristID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var progress []model.EncounterProgress
	if err := cursor.All(ctx, &progress); err != nil {
		return nil, err
	}
	if progress == nil {
		progress = []model.EncounterProgress{}
	}
	return progress, nil
}
