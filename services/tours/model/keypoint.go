package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type KeyPoint struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TourID      bson.ObjectID `bson:"tourId"        json:"tourId"`
	Name        string        `bson:"name"          json:"name"`
	Description string        `bson:"description"   json:"description"`
	Latitude    float64       `bson:"latitude"      json:"latitude"`
	Longitude   float64       `bson:"longitude"     json:"longitude"`
	ImageURL    string        `bson:"imageUrl"      json:"imageUrl"`
	Order       int           `bson:"order"         json:"order"`
	CreatedAt   time.Time     `bson:"createdAt"     json:"createdAt"`
}

type CreateKeyPointRequest struct {
	TourID      string  `json:"tourId"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ImageURL    string  `json:"imageUrl"`
	Order       int     `json:"order"`
}

type UpdateKeyPointRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ImageURL    string  `json:"imageUrl"`
	Order       int     `json:"order"`
}
