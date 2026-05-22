package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TourStatus string
type DifficultyLevel string
type TransportType string

const (
	StatusDraft     TourStatus = "draft"
	StatusPublished TourStatus = "published"
	StatusArchived  TourStatus = "archived"

	DifficultyEasy   DifficultyLevel = "easy"
	DifficultyMedium DifficultyLevel = "medium"
	DifficultyHard   DifficultyLevel = "hard"

	TransportWalk TransportType = "walk"
	TransportBike TransportType = "bike"
	TransportCar  TransportType = "car"
)

type TourDuration struct {
	Transport TransportType `bson:"transport" json:"transport"`
	Minutes   int           `bson:"minutes"   json:"minutes"`
}

type Tour struct {
	ID          bson.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	AuthorID    string          `bson:"authorId"      json:"authorId"`
	Name        string          `bson:"name"          json:"name"`
	Description string          `bson:"description"   json:"description"`
	Difficulty  DifficultyLevel `bson:"difficulty"    json:"difficulty"`
	Tags        []string        `bson:"tags"          json:"tags"`
	Status      TourStatus      `bson:"status"        json:"status"`
	LengthKm    float64         `bson:"lengthKm"      json:"lengthKm"`
	Durations   []TourDuration  `bson:"durations"     json:"durations"`
	Price       float64         `bson:"price"         json:"price"`
	CreatedAt   time.Time       `bson:"createdAt"     json:"createdAt"`
	UpdatedAt   time.Time       `bson:"updatedAt"     json:"updatedAt"`
	PublishedAt *time.Time      `bson:"publishedAt"   json:"publishedAt,omitempty"`
	ArchivedAt  *time.Time      `bson:"archivedAt"    json:"archivedAt,omitempty"`
}

type TourPreview struct {
	ID            bson.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	AuthorID      string          `bson:"authorId"      json:"authorId"`
	Name          string          `bson:"name"          json:"name"`
	Description   string          `bson:"description"   json:"description"`
	Difficulty    DifficultyLevel `bson:"difficulty"    json:"difficulty"`
	Tags          []string        `bson:"tags"          json:"tags"`
	LengthKm      float64         `bson:"lengthKm"      json:"lengthKm"`
	Price         float64         `bson:"price"         json:"price"`
	PublishedAt   *time.Time      `bson:"publishedAt"   json:"publishedAt,omitempty"`
	FirstKeyPoint *KeyPoint       `json:"firstKeyPoint,omitempty"`
}

type CreateTourRequest struct {
	AuthorID    string          `json:"authorId"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Difficulty  DifficultyLevel `json:"difficulty"`
	Tags        []string        `json:"tags"`
	Durations   []TourDuration  `json:"durations"`
}

type UpdateTourRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Difficulty  DifficultyLevel `json:"difficulty"`
	Tags        []string        `json:"tags"`
	Durations   []TourDuration  `json:"durations"`
	Price       float64         `json:"price"`
}
