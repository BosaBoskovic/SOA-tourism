package repository

import (
	"context"
	"time"
	"tours/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type TourRepository struct {
	collection *mongo.Collection
}

func NewTourRepository(db *mongo.Database) *TourRepository {
	return &TourRepository{
		collection: db.Collection("tours"),
	}
}

func (r *TourRepository) Create(tour *model.Tour) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, tour)
	return err
}

func (r *TourRepository) FindByID(id bson.ObjectID) (*model.Tour, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var tour model.Tour
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&tour)
	if err != nil {
		return nil, err
	}
	return &tour, nil
}

func (r *TourRepository) FindByStringID(id string) (*model.Tour, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	return r.FindByID(oid)
}

func (r *TourRepository) FindByAuthor(authorID string) ([]model.Tour, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"authorId": authorID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tours []model.Tour
	if err = cursor.All(ctx, &tours); err != nil {
		return nil, err
	}
	return tours, nil
}

func (r *TourRepository) Update(id bson.ObjectID, update bson.M) error {
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

func (r *TourRepository) FindAllPublished() ([]model.Tour, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"status": model.StatusPublished})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tours []model.Tour
	if err = cursor.All(ctx, &tours); err != nil {
		return nil, err
	}

	if tours == nil {
		tours = []model.Tour{}
	}

	return tours, nil
}

func (r *TourRepository) UpdateStatus(id bson.ObjectID, status model.TourStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"status":    status,
				"updatedAt": time.Now(),
			},
		},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *TourRepository) UpdateStatusWithTimestamps(id bson.ObjectID, status model.TourStatus, publishedAt *time.Time, archivedAt *time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	setFields := bson.M{
		"status":    status,
		"updatedAt": time.Now(),
	}
	if publishedAt != nil {
		setFields["publishedAt"] = *publishedAt
	}
	if archivedAt != nil {
		setFields["archivedAt"] = *archivedAt
	}

	update := bson.M{"$set": setFields}

	unsetFields := bson.M{}
	if publishedAt == nil {
		unsetFields["publishedAt"] = ""
	}
	if archivedAt == nil {
		unsetFields["archivedAt"] = ""
	}
	if len(unsetFields) > 0 {
		update["$unset"] = unsetFields
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
