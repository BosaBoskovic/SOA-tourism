package service

import (
	"errors"
	"time"
	"tours/model"
	"tours/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type TouristPositionService struct {
	repo *repository.TouristPositionRepository
}

func NewTouristPositionService(repo *repository.TouristPositionRepository) *TouristPositionService {
	return &TouristPositionService{repo: repo}
}

func (s *TouristPositionService) Update(req *model.UpdateTouristPositionRequest) (*model.TouristPosition, error) {
	if req.TouristID == "" {
		return nil, errors.New("touristId is required")
	}

	if req.Latitude < -90 || req.Latitude > 90 {
		return nil, errors.New("invalid latitude")
	}

	if req.Longitude < -180 || req.Longitude > 180 {
		return nil, errors.New("invalid longitude")
	}

	position := &model.TouristPosition{
		ID:        bson.NewObjectID(),
		TouristID: req.TouristID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Upsert(position); err != nil {
		return nil, err
	}

	return s.repo.FindByTouristID(req.TouristID)
}

func (s *TouristPositionService) GetByTouristID(touristID string) (*model.TouristPosition, error) {
	if touristID == "" {
		return nil, errors.New("touristId is required")
	}

	position, err := s.repo.FindByTouristID(touristID)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("tourist position not found")
	}

	return position, err
}