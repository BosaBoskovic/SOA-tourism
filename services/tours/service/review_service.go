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
	repo         *repository.ReviewRepository
	tourRepo     *repository.TourRepository
	purchaseRepo *repository.PurchaseRepository
	execRepo     *repository.TourExecutionRepository
}

func NewReviewService(repo *repository.ReviewRepository, tourRepo *repository.TourRepository, purchaseRepo *repository.PurchaseRepository, execRepo *repository.TourExecutionRepository) *ReviewService {
	return &ReviewService{repo: repo, tourRepo: tourRepo, purchaseRepo: purchaseRepo, execRepo: execRepo}
}

// Create adds a review from the verified caller. touristName is a display
// name only (not an identity) - it still comes from the request, but
// touristId is always the verified caller, never client-supplied.
func (s *ReviewService) Create(req *model.CreateReviewRequest, callerUsername string) (*model.Review, error) {
	tourOID, err := bson.ObjectIDFromHex(req.TourID)
	if err != nil {
		return nil, errors.New("invalid tourId")
	}

	if req.Rating < 1 || req.Rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}
	req.TouristID = callerUsername
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

	// Samo turisti koji su kupili ili odradili turu mogu ostaviti recenziju
	purchased, err := s.purchaseRepo.HasToken(req.TouristID, req.TourID)
	if err != nil {
		return nil, err
	}
	if !purchased {
		executed, err := s.execRepo.ExistsByTouristAndTour(req.TouristID, tourOID)
		if err != nil {
			return nil, err
		}
		if !executed {
			return nil, ErrForbidden
		}
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

func (s *ReviewService) Delete(id string, callerUsername, callerRole string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid review ID")
	}

	review, err := s.repo.FindByID(oid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return errors.New("review not found")
	}
	if err != nil {
		return err
	}
	if !isOwner(review.TouristID, callerUsername, callerRole) {
		return ErrForbidden
	}

	err = s.repo.Delete(oid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return errors.New("review not found")
	}
	return err
}
