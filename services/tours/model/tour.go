package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TourStatus string
type DifficultyLevel string

const (
	StatusDraft     TourStatus = "draft"
	StatusPublished TourStatus = "published"
	StatusArchived  TourStatus = "archived"

	DifficultyEasy   DifficultyLevel = "easy"
	DifficultyMedium DifficultyLevel = "medium"
	DifficultyHard   DifficultyLevel = "hard"
)

type Tour struct {
	ID          bson.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	AuthorID    string          `bson:"authorId"      json:"authorId"`
	Name        string          `bson:"name"          json:"name"`
	Description string          `bson:"description"   json:"description"`
	Difficulty  DifficultyLevel `bson:"difficulty"    json:"difficulty"`
	Tags        []string        `bson:"tags"          json:"tags"`
	Status      TourStatus      `bson:"status"        json:"status"`
	Price       float64         `bson:"price"         json:"price"`
	CreatedAt   time.Time       `bson:"createdAt"     json:"createdAt"`
	UpdatedAt   time.Time       `bson:"updatedAt"     json:"updatedAt"`
}

type CreateTourRequest struct {
	AuthorID    string          `json:"authorId"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Difficulty  DifficultyLevel `json:"difficulty"`
	Tags        []string        `json:"tags"`
}

type UpdateTourRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Difficulty  DifficultyLevel `json:"difficulty"`
	Tags        []string        `json:"tags"`
}
