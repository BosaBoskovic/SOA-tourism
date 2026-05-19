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
	repo         *repository.TourRepository
	keyPointRepo *repository.KeyPointRepository
}

func NewTourService(repo *repository.TourRepository, keyPointRepo *repository.KeyPointRepository) *TourService {
	return &TourService{
		repo: repo,
		keyPointRepo: keyPointRepo,
	}
}

func (s *TourService) Create(req *model.CreateTourRequest) (*model.Tour, error) {
	if req.AuthorID == "" || req.Name == "" || req.Description == "" || req.Difficulty == "" {
		return nil, errors.New("authorId, name, description and difficulty are required")
	}
	if err := validateDurations(req.Durations); err != nil {
		return nil, err
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
		LengthKm:    0,
		Durations:   req.Durations,
		Price:       0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if tour.Tags == nil {
		tour.Tags = []string{}
	}
	if tour.Durations == nil {
		tour.Durations = []model.TourDuration{}
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
	if err := validateDurations(req.Durations); err != nil {
		return nil, err
	}

	update := bson.M{
		"name":        req.Name,
		"description": req.Description,
		"difficulty":  req.Difficulty,
		"tags":        req.Tags,
		"durations":   req.Durations,
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

func (s *TourService) GetPublished() ([]model.TourPreview, error) {
	tours, err := s.repo.FindAllPublished()
	if err != nil {
		return nil, err
	}

	previews := make([]model.TourPreview, 0, len(tours))
	for _, tour := range tours {
		preview := model.TourPreview{
			ID:          tour.ID,
			AuthorID:    tour.AuthorID,
			Name:        tour.Name,
			Description: tour.Description,
			Difficulty:  tour.Difficulty,
			Tags:        tour.Tags,
			LengthKm:    tour.LengthKm,
			Price:       tour.Price,
			PublishedAt: tour.PublishedAt,
		}

		kp, err := s.keyPointRepo.FindFirstByTour(tour.ID)
		if err == nil {
			preview.FirstKeyPoint = kp
		} else if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}

		previews = append(previews, preview)
	}

	return previews, nil
}

func (s *TourService) Publish(id string) (*model.Tour, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid tour ID")
	}

	tour, err := s.repo.FindByID(oid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("tour not found")
	}
	if err != nil {
		return nil, err
	}
	if tour.Status != model.StatusDraft {
		return nil, errors.New("tour is not in draft status")
	}
	if tour.Name == "" || tour.Description == "" || tour.Difficulty == "" || len(tour.Tags) == 0 {
		return nil, errors.New("tour must have name, description, difficulty, and tags before publishing")
	}
	if len(tour.Durations) == 0 {
		return nil, errors.New("tour must have at least one duration before publishing")
	}
	if err := validateDurations(tour.Durations); err != nil {
		return nil, err
	}

	count, err := s.keyPointRepo.CountByTourID(id)
	if err != nil {
		return nil, err
	}

	if count < 2 {
		return nil, errors.New("Tura mora imati najmanje 2 ključne tačke da bi bila objavljena.")
	}

	now := time.Now()
	err = s.repo.UpdateStatusWithTimestamps(oid, model.StatusPublished, &now, nil)
	if err != nil {
		return nil, err
	}

	return s.repo.FindByID(oid)
}

func (s *TourService) Archive(id string) (*model.Tour, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid tour ID")
	}

	tour, err := s.repo.FindByID(oid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("tour not found")
	}
	if err != nil {
		return nil, err
	}
	if tour.Status != model.StatusPublished {
		return nil, errors.New("only published tours can be archived")
	}

	now := time.Now()
	err = s.repo.UpdateStatusWithTimestamps(oid, model.StatusArchived, tour.PublishedAt, &now)
	if err != nil {
		return nil, err
	}

	return s.repo.FindByID(oid)
}

func (s *TourService) Activate(id string) (*model.Tour, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid tour ID")
	}

	tour, err := s.repo.FindByID(oid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("tour not found")
	}
	if err != nil {
		return nil, err
	}
	if tour.Status != model.StatusArchived {
		return nil, errors.New("only archived tours can be activated")
	}

	now := time.Now()
	err = s.repo.UpdateStatusWithTimestamps(oid, model.StatusPublished, &now, nil)
	if err != nil {
		return nil, err
	}

	return s.repo.FindByID(oid)
}

func validateDurations(durations []model.TourDuration) error {
	for _, duration := range durations {
		switch duration.Transport {
		case model.TransportWalk, model.TransportBike, model.TransportCar:
			// ok
		default:
			return errors.New("invalid transport type")
		}
		if duration.Minutes <= 0 {
			return errors.New("duration minutes must be greater than zero")
		}
	}
	return nil
}
