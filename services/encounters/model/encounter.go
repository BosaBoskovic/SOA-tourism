package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Encounter is a proximity-based challenge: reach a real-world location
// within RadiusMeters to claim it. TourID/KeyPointID are optional -
// standalone encounters (not tied to any tour) are allowed too.
type Encounter struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name         string        `bson:"name" json:"name"`
	Description  string        `bson:"description" json:"description"`
	TourID       string        `bson:"tourId,omitempty" json:"tourId,omitempty"`
	KeyPointID   string        `bson:"keyPointId,omitempty" json:"keyPointId,omitempty"`
	Latitude     float64       `bson:"latitude" json:"latitude"`
	Longitude    float64       `bson:"longitude" json:"longitude"`
	RadiusMeters float64       `bson:"radiusMeters" json:"radiusMeters"`
	Reward       string        `bson:"reward" json:"reward"`
	CreatedBy    string        `bson:"createdBy" json:"createdBy"`
	CreatedAt    time.Time     `bson:"createdAt" json:"createdAt"`
}

type CreateEncounterRequest struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	TourID       string  `json:"tourId"`
	KeyPointID   string  `json:"keyPointId"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	RadiusMeters float64 `json:"radiusMeters"`
	Reward       string  `json:"reward"`
}

type ClaimEncounterRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// EncounterProgress records that a tourist has claimed an encounter -
// claiming is idempotent, so at most one row exists per (touristId, encounterId).
type EncounterProgress struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TouristID   string        `bson:"touristId" json:"touristId"`
	EncounterID bson.ObjectID `bson:"encounterId" json:"encounterId"`
	CompletedAt time.Time     `bson:"completedAt" json:"completedAt"`
}

// EncounterWithProgress is what list endpoints return - an encounter plus
// whether/when the requesting tourist has already claimed it.
type EncounterWithProgress struct {
	Encounter   `bson:",inline"`
	Claimed     bool       `json:"claimed"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}
