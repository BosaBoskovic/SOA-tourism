package service

import (
	"errors"
	"time"
	"tours/model"
	"tours/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type KeyPointService struct {
	repo     *repository.KeyPointRepository
	tourRepo *repository.TourRepository
}

func NewKeyPointService(repo *repository.KeyPointRepository, tourRepo *repository.TourRepository) *KeyPointService {
	return &KeyPointService{repo: repo, tourRepo: tourRepo}
}

func (s *KeyPointService) Create(req *model.CreateKeyPointRequest) (*model.KeyPoint, error) {
	tourOID, err := bson.ObjectIDFromHex(req.TourID)
	if err != nil {
		return nil, errors.New("invalid tourId")
	}

	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.Latitude == 0 || req.Longitude == 0 {
		return nil, errors.New("latitude and longitude are required")
	}

	// Proveri da tura postoji
	tour, err := s.tourRepo.FindByID(tourOID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("tour not found")
		}
		return nil, err
	}
	if tour.Status != model.StatusDraft {
		return nil, errors.New("key points can be managed only while tour is in draft status")
	}

	kp := &model.KeyPoint{
		ID:          bson.NewObjectID(),
		TourID:      tourOID,
		Name:        req.Name,
		Description: req.Description,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		ImageURL:    req.ImageURL,
		Order:       req.Order,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Create(kp); err != nil {
		return nil, err
	}
	if err := s.updateTourLength(tourOID, req.LengthKm); err != nil {
		return nil, err
	}
	return kp, nil
}

func (s *KeyPointService) GetByID(id string) (*model.KeyPoint, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid key point ID")
	}

	kp, err := s.repo.FindByID(oid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("key point not found")
	}
	return kp, err
}

func (s *KeyPointService) GetByTour(tourID string) ([]model.KeyPoint, error) {
	oid, err := bson.ObjectIDFromHex(tourID)
	if err != nil {
		return nil, errors.New("invalid tourId")
	}

	kps, err := s.repo.FindByTour(oid)
	if err != nil {
		return nil, err
	}
	if kps == nil {
		kps = []model.KeyPoint{}
	}
	return kps, nil
}

func (s *KeyPointService) Update(id string, req *model.UpdateKeyPointRequest) (*model.KeyPoint, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid key point ID")
	}

	kp, err := s.repo.FindByID(oid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("key point not found")
	}
	if err != nil {
		return nil, err
	}

	tour, err := s.tourRepo.FindByID(kp.TourID)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("tour not found")
	}
	if err != nil {
		return nil, err
	}
	if tour.Status != model.StatusDraft {
		return nil, errors.New("key points can be managed only while tour is in draft status")
	}

	update := bson.M{
		"name":        req.Name,
		"description": req.Description,
		"latitude":    req.Latitude,
		"longitude":   req.Longitude,
		"imageUrl":    req.ImageURL,
		"order":       req.Order,
	}

	err = s.repo.Update(oid, update)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("key point not found")
	}
	if err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(oid)
	if err != nil {
		return nil, err
	}
	if err := s.updateTourLength(updated.TourID, req.LengthKm); err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *KeyPointService) Delete(id string, lengthKm *float64) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid key point ID")
	}

	kp, err := s.repo.FindByID(oid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return errors.New("key point not found")
	}
	if err != nil {
		return err
	}

	tour, err := s.tourRepo.FindByID(kp.TourID)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return errors.New("tour not found")
	}
	if err != nil {
		return err
	}
	if tour.Status != model.StatusDraft {
		return errors.New("key points can be managed only while tour is in draft status")
	}

	err = s.repo.Delete(oid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return errors.New("key point not found")
	}
	if err != nil {
		return err
	}

	if err := s.normalizeKeyPointOrder(kp.TourID); err != nil {
		return err
	}

	return s.updateTourLength(kp.TourID, lengthKm)
}

func (s *KeyPointService) updateTourLength(tourID bson.ObjectID, lengthKm *float64) error {
	if lengthKm == nil {
		return nil
	}

	update := bson.M{
		"lengthKm": *lengthKm,
		"updatedAt": time.Now(),
	}

	return s.tourRepo.Update(tourID, update)
}

func (s *KeyPointService) normalizeKeyPointOrder(tourID bson.ObjectID) error {
	kps, err := s.repo.FindByTour(tourID)
	if err != nil {
		return err
	}

	for i, kp := range kps {
		newOrder := i + 1
		if kp.Order == newOrder {
			continue
		}
		update := bson.M{"order": newOrder}
		if err := s.repo.Update(kp.ID, update); err != nil {
			return err
		}
	}

	return nil
}
