package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// PurchaseRepository checks if a tourist has purchased a tour.
// Purchase tokens are written by the payments service but stored in the same
// MongoDB instance (tourServiceDB) so the tours service can query them directly.
type PurchaseRepository struct {
	collection *mongo.Collection
}

func NewPurchaseRepository(db *mongo.Database) *PurchaseRepository {
	return &PurchaseRepository{
		collection: db.Collection("purchase_tokens"),
	}
}

// HasToken returns true if touristId has a valid purchase token for tourId.
func (r *PurchaseRepository) HasToken(touristID, tourID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"touristId": touristID,
		"tourId":    tourID,
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// SaveToken stores a purchase token (called by payments service or via shared DB).
// Tours service uses this only for testing; the payments service writes the tokens.
func (r *PurchaseRepository) SaveToken(touristID, tourID, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := bson.M{
		"_id":         bson.NewObjectID(),
		"touristId":   touristID,
		"tourId":      tourID,
		"token":       token,
		"purchasedAt": time.Now(),
	}
	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

// FindByTourist returns all purchase tokens for a tourist.
func (r *PurchaseRepository) FindByTourist(touristID string) ([]bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"touristId": touristID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []bson.M
	if err = cursor.All(ctx, &tokens); err != nil {
		return nil, err
	}
	if tokens == nil {
		tokens = []bson.M{}
	}
	return tokens, nil
}

// DeleteByTouristAndTour removes a token (e.g. if purchase is refunded).
func (r *PurchaseRepository) DeleteByTouristAndTour(touristID, tourID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{
		"touristId": touristID,
		"tourId":    tourID,
	})
	return err
}

// FindTokensByTourAndTourist finds tokens for a specific tour and tourist
func (r *PurchaseRepository) FindAll() ([]bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []bson.M
	if err = cursor.All(ctx, &tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

// MongoObjectID returns a new ObjectID — helper for tests
func MongoObjectID() bson.ObjectID {
	return bson.NewObjectID()
}

// PurchaseToken is a lightweight struct for serialization
type PurchaseToken struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TouristID   string        `bson:"touristId"     json:"touristId"`
	TourID      string        `bson:"tourId"        json:"tourId"`
	Token       string        `bson:"token"         json:"token"`
	PurchasedAt time.Time     `bson:"purchasedAt"   json:"purchasedAt"`
}

// FindTokensForTourist returns typed tokens for a tourist
func (r *PurchaseRepository) FindTokensForTourist(touristID string) ([]PurchaseToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"touristId": touristID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []PurchaseToken
	if err = cursor.All(ctx, &tokens); err != nil {
		return nil, err
	}
	if tokens == nil {
		tokens = []PurchaseToken{}
	}
	return tokens, nil
}
