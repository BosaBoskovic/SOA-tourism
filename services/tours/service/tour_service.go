package service

import (
	"errors"
	"time"
	"tours/model"
	"tours/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type TourService struct {
	repo *repository.TourRepository
}

func NewTourService(repo *repository.TourRepository) *TourService {
	return &TourService{repo: repo}
}

func (s *TourService) Create(req *model.CreateTourRequest) (*model.Tour, error) {
	if req.AuthorID == "" || req.Name == "" || req.Description == "" || req.Difficulty == "" {
		return nil, errors.New("authorId, name, description and difficulty are required")
	}

	now := time.Now()
	tour := &model.Tour{
		ID:          bson.NewObjectID(),
		AuthorID:    req.AuthorID,
		Name:        req.Name,
		Description: req.Description,
		Difficulty:  req.Difficulty,
		Tags:        req.Tags,
		Status:      model.StatusDraft,
		Price:       0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if tour.Tags == nil {
		tour.Tags = []string{}
	}

	if err := s.repo.Create(tour); err != nil {
		return nil, err
	}
	return tour, nil
}

func (s *TourService) GetByID(id string) (*model.Tour, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid tour ID")
	}

	tour, err := s.repo.FindByID(oid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("tour not found")
	}
	return tour, err
}

func (s *TourService) GetByAuthor(authorID string) ([]model.Tour, error) {
	tours, err := s.repo.FindByAuthor(authorID)
	if err != nil {
		return nil, err
	}
	if tours == nil {
		tours = []model.Tour{}
	}
	return tours, nil
}

func (s *TourService) Update(id string, req *model.UpdateTourRequest) (*model.Tour, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid tour ID")
	}

	update := bson.M{
		"name":        req.Name,
		"description": req.Description,
		"difficulty":  req.Difficulty,
		"tags":        req.Tags,
		"updatedAt":   time.Now(),
	}

	err = s.repo.Update(oid, update)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("tour not found")
	}
	if err != nil {
		return nil, err
	}

	return s.repo.FindByID(oid)
}
