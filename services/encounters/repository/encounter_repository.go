package repository

import (
	"context"
	"time"

	"encounters/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type EncounterRepository struct {
	collection *mongo.Collection
}

func NewEncounterRepository(db *mongo.Database) *EncounterRepository {
	return &EncounterRepository{collection: db.Collection("encounters")}
}

func (r *EncounterRepository) Create(e *model.Encounter) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, e)
	return err
}

func (r *EncounterRepository) FindAll() ([]model.Encounter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var encounters []model.Encounter
	if err := cursor.All(ctx, &encounters); err != nil {
		return nil, err
	}
	if encounters == nil {
		encounters = []model.Encounter{}
	}
	return encounters, nil
}

func (r *EncounterRepository) FindByID(id bson.ObjectID) (*model.Encounter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var e model.Encounter
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EncounterRepository) Delete(id bson.ObjectID) error {
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
