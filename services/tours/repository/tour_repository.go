package repository

import (
	"context"
	"time"
	"tours/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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
	return r.FindPublishedFiltered(model.TourSearchParams{})
}

// FindPublishedFiltered applies FindAllPublished's same "published" gate plus
// whatever optional filters/sort the caller asked for - an empty
// TourSearchParams behaves exactly like FindAllPublished.
func (r *TourRepository) FindPublishedFiltered(params model.TourSearchParams) ([]model.Tour, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"status": model.StatusPublished}

	if params.Difficulty != "" {
		filter["difficulty"] = params.Difficulty
	}
	if len(params.Tags) > 0 {
		filter["tags"] = bson.M{"$in": params.Tags}
	}
	if params.MinPrice != nil || params.MaxPrice != nil {
		priceFilter := bson.M{}
		if params.MinPrice != nil {
			priceFilter["$gte"] = *params.MinPrice
		}
		if params.MaxPrice != nil {
			priceFilter["$lte"] = *params.MaxPrice
		}
		filter["price"] = priceFilter
	}
	if params.MinLengthKm != nil || params.MaxLengthKm != nil {
		lengthFilter := bson.M{}
		if params.MinLengthKm != nil {
			lengthFilter["$gte"] = *params.MinLengthKm
		}
		if params.MaxLengthKm != nil {
			lengthFilter["$lte"] = *params.MaxLengthKm
		}
		filter["lengthKm"] = lengthFilter
	}

	sortField := "publishedAt"
	switch params.SortBy {
	case "price":
		sortField = "price"
	case "length":
		sortField = "lengthKm"
	case "name":
		sortField = "name"
	}
	sortDir := -1
	if params.SortDir == "asc" {
		sortDir = 1
	}

	opts := options.Find().SetSort(bson.D{{Key: sortField, Value: sortDir}})

	cursor, err := r.collection.Find(ctx, filter, opts)
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
