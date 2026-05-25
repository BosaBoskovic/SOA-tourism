package repository

import (
	"context"
	"time"
	"tours/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type TourExecutionRepository struct {
	collection *mongo.Collection
}

func NewTourExecutionRepository(db *mongo.Database) *TourExecutionRepository {
	return &TourExecutionRepository{
		collection: db.Collection("tour_executions"),
	}
}

func (r *TourExecutionRepository) Create(exec *model.TourExecution) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.collection.InsertOne(ctx, exec)
	return err
}

func (r *TourExecutionRepository) FindByID(id bson.ObjectID) (*model.TourExecution, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var exec model.TourExecution
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&exec)
	if err != nil {
		return nil, err
	}
	return &exec, nil
}

func (r *TourExecutionRepository) FindActiveByTourist(touristID string) (*model.TourExecution, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var exec model.TourExecution
	err := r.collection.FindOne(ctx, bson.M{
		"touristId": touristID,
		"status":    model.ExecutionActive,
	}).Decode(&exec)
	if err != nil {
		return nil, err
	}
	return &exec, nil
}

func (r *TourExecutionRepository) FindByTourist(touristID string) ([]model.TourExecution, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := r.collection.Find(ctx, bson.M{"touristId": touristID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var execs []model.TourExecution
	if err = cursor.All(ctx, &execs); err != nil {
		return nil, err
	}
	if execs == nil {
		execs = []model.TourExecution{}
	}
	return execs, nil
}

func (r *TourExecutionRepository) Update(id bson.ObjectID, update bson.M) error {
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

// HasTouristPurchasedTour checks if a tourist has a purchase token for a tour
// This is a simple check - purchase tokens are stored in the purchases collection
func (r *TourExecutionRepository) ExistsByTouristAndTour(touristID string, tourID bson.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"touristId": touristID,
		"tourId":    tourID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
