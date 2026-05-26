package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ExecutionStatus string

const (
	ExecutionActive    ExecutionStatus = "active"
	ExecutionCompleted ExecutionStatus = "completed"
	ExecutionAbandoned ExecutionStatus = "abandoned"
)

type CompletedKeyPoint struct {
	KeyPointID bson.ObjectID `bson:"keyPointId" json:"keyPointId"`
	ReachedAt  time.Time     `bson:"reachedAt"  json:"reachedAt"`
}

type TourExecution struct {
	ID                 bson.ObjectID       `bson:"_id,omitempty"         json:"id,omitempty"`
	TouristID          string              `bson:"touristId"             json:"touristId"`
	TourID             bson.ObjectID       `bson:"tourId"                json:"tourId"`
	Status             ExecutionStatus     `bson:"status"                json:"status"`
	StartLocation      GeoPoint            `bson:"startLocation"         json:"startLocation"`
	CompletedKeyPoints []CompletedKeyPoint `bson:"completedKeyPoints"    json:"completedKeyPoints"`
	StartedAt          time.Time           `bson:"startedAt"             json:"startedAt"`
	LastActivityAt     time.Time           `bson:"lastActivityAt"        json:"lastActivityAt"`
	CompletedAt        *time.Time          `bson:"completedAt,omitempty" json:"completedAt,omitempty"`
	AbandonedAt        *time.Time          `bson:"abandonedAt,omitempty" json:"abandonedAt,omitempty"`
}

type GeoPoint struct {
	Latitude  float64 `bson:"latitude"  json:"latitude"`
	Longitude float64 `bson:"longitude" json:"longitude"`
}

type StartExecutionRequest struct {
	TouristID string  `json:"touristId"`
	TourID    string  `json:"tourId"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type CheckKeyPointRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type CheckKeyPointResponse struct {
	KeyPointReached bool           `json:"keyPointReached"`
	KeyPoint        *KeyPoint      `json:"keyPoint,omitempty"`
	Execution       *TourExecution `json:"execution"`
}
