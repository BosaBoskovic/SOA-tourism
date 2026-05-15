package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TouristPosition struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TouristID string        `bson:"touristId" json:"touristId"`
	Latitude  float64       `bson:"latitude" json:"latitude"`
	Longitude float64       `bson:"longitude" json:"longitude"`
	UpdatedAt time.Time     `bson:"updatedAt" json:"updatedAt"`
}

type UpdateTouristPositionRequest struct {
	TouristID string  `json:"touristId"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}