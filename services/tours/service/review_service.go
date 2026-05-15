package service

import (
	"errors"
	"time"
	"tours/model"
	"tours/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ReviewService struct {
	repo     *repository.ReviewRepository
	tourRepo *repository.TourRepository
}

func NewReviewService(repo *repository.ReviewRepository, tourRepo *repository.TourRepository) *ReviewService {
	return &ReviewService{repo: repo, tourRepo: tourRepo}
}

func (s *ReviewService) Create(req *model.CreateReviewRequest) (*model.Review, error) {
	tourOID, err := bson.ObjectIDFromHex(req.TourID)
	if err != nil {
		return nil, errors.New("invalid tourId")
	}

	if req.Rating < 1 || req.Rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}
	if req.TouristID == "" || req.TouristName == "" {
		return nil, errors.New("touristId and touristName are required")
	}

	visitDate, err := time.Parse("2006-01-02", req.TourVisitDate)
	if err != nil {
		return nil, errors.New("tourVisitDate must be in format YYYY-MM-DD")
	}

	// Proveri da tura postoji
	if _, err := s.tourRepo.FindByID(tourOID); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("tour not found")
		}
		return nil, err
	}

	// Turista moze ostaviti samo jednu recenziju po turi
	exists, err := s.repo.ExistsByTourAndTourist(tourOID, req.TouristID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("tourist has already reviewed this tour")
	}

	images := req.Images
	if images == nil {
		images = []string{}
	}

	review := &model.Review{
		ID:            bson.NewObjectID(),
		TourID:        tourOID,
		TouristID:     req.TouristID,
		TouristName:   req.TouristName,
		Rating:        req.Rating,
		Comment:       req.Comment,
		Images:        images,
		TourVisitDate: visitDate,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.Create(review); err != nil {
		return nil, err
	}
	return review, nil
}

func (s *ReviewService) GetByID(id string) (*model.Review, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid review ID")
	}

	review, err := s.repo.FindByID(oid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("review not found")
	}
	return review, err
}

func (s *ReviewService) GetByTour(tourID string) ([]model.Review, error) {
	oid, err := bson.ObjectIDFromHex(tourID)
	if err != nil {
		return nil, errors.New("invalid tourId")
	}

	reviews, err := s.repo.FindByTour(oid)
	if err != nil {
		return nil, err
	}
	if reviews == nil {
		reviews = []model.Review{}
	}
	return reviews, nil
}

func (s *ReviewService) Delete(id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid review ID")
	}

	err = s.repo.Delete(oid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return errors.New("review not found")
	}
	return err
}
