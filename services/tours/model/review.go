package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Review struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TourID        bson.ObjectID `bson:"tourId"        json:"tourId"`
	TouristID     string        `bson:"touristId"     json:"touristId"`
	TouristName   string        `bson:"touristName"   json:"touristName"`
	Rating        int           `bson:"rating"        json:"rating"` // 1-5
	Comment       string        `bson:"comment"       json:"comment"`
	Images        []string      `bson:"images"        json:"images"`
	TourVisitDate time.Time     `bson:"tourVisitDate" json:"tourVisitDate"`
	CreatedAt     time.Time     `bson:"createdAt"     json:"createdAt"`
}

type CreateReviewRequest struct {
	TourID        string   `json:"tourId"`
	TouristID     string   `json:"touristId"`
	TouristName   string   `json:"touristName"`
	Rating        int      `json:"rating"`
	Comment       string   `json:"comment"`
	Images        []string `json:"images"`
	TourVisitDate string   `json:"tourVisitDate"` // "2024-05-01"
}
