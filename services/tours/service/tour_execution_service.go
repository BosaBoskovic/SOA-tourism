package service

import (
	"errors"
	"math"
	"time"
	"tours/model"
	"tours/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const keyPointRadiusMeters = 200.0 // turista je "blizu" ako je unutar 200m

type TourExecutionService struct {
	execRepo     *repository.TourExecutionRepository
	tourRepo     *repository.TourRepository
	keyPointRepo *repository.KeyPointRepository
	purchaseRepo *repository.PurchaseRepository
}

func NewTourExecutionService(
	execRepo *repository.TourExecutionRepository,
	tourRepo *repository.TourRepository,
	keyPointRepo *repository.KeyPointRepository,
	purchaseRepo *repository.PurchaseRepository,
) *TourExecutionService {
	return &TourExecutionService{
		execRepo:     execRepo,
		tourRepo:     tourRepo,
		keyPointRepo: keyPointRepo,
		purchaseRepo: purchaseRepo,
	}
}

func (s *TourExecutionService) Start(req *model.StartExecutionRequest) (*model.TourExecution, error) {
	if req.TouristID == "" || req.TourID == "" {
		return nil, errors.New("touristId and tourId are required")
	}

	tourOID, err := bson.ObjectIDFromHex(req.TourID)
	if err != nil {
		return nil, errors.New("invalid tourId")
	}

	// Provjeri da tura postoji i da je published ili archived
	tour, err := s.tourRepo.FindByID(tourOID)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("tour not found")
	}
	if err != nil {
		return nil, err
	}
	if tour.Status != model.StatusPublished && tour.Status != model.StatusArchived {
		return nil, errors.New("only published or archived tours can be started")
	}

	// Provjeri da je turista kupio turu
	purchased, err := s.purchaseRepo.HasToken(req.TouristID, req.TourID)
	if err != nil {
		return nil, err
	}
	if !purchased {
		return nil, errors.New("tourist has not purchased this tour")
	}

	now := time.Now()
	exec := &model.TourExecution{
		ID:        bson.NewObjectID(),
		TouristID: req.TouristID,
		TourID:    tourOID,
		Status:    model.ExecutionActive,
		StartLocation: model.GeoPoint{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
		},
		CompletedKeyPoints: []model.CompletedKeyPoint{},
		StartedAt:          now,
		LastActivityAt:     now,
	}

	if err := s.execRepo.Create(exec); err != nil {
		return nil, err
	}
	return exec, nil
}

func (s *TourExecutionService) CheckKeyPoint(executionID string, lat, lon float64) (*model.CheckKeyPointResponse, error) {
	oid, err := bson.ObjectIDFromHex(executionID)
	if err != nil {
		return nil, errors.New("invalid executionId")
	}

	exec, err := s.execRepo.FindByID(oid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("execution not found")
	}
	if err != nil {
		return nil, err
	}
	if exec.Status != model.ExecutionActive {
		return nil, errors.New("execution is not active")
	}

	// Azuriraj lastActivityAt uvijek
	now := time.Now()

	keyPoints, err := s.keyPointRepo.FindByTour(exec.TourID)
	if err != nil {
		return nil, err
	}

	// Provjeri koje su ključne tačke već kompletovane
	completedIDs := map[bson.ObjectID]bool{}
	for _, ckp := range exec.CompletedKeyPoints {
		completedIDs[ckp.KeyPointID] = true
	}

	var reachedKP *model.KeyPoint
	for i := range keyPoints {
		kp := &keyPoints[i]
		if completedIDs[kp.ID] {
			continue // već kompletovana
		}
		dist := haversineMeters(lat, lon, kp.Latitude, kp.Longitude)
		if dist <= keyPointRadiusMeters {
			reachedKP = kp
			break
		}
	}

	update := bson.M{"lastActivityAt": now}

	if reachedKP != nil {
		exec.CompletedKeyPoints = append(exec.CompletedKeyPoints, model.CompletedKeyPoint{
			KeyPointID: reachedKP.ID,
			ReachedAt:  now,
		})
		update["completedKeyPoints"] = exec.CompletedKeyPoints

		// Provjeri da li je turista kompletovao sve ključne tačke
		if len(exec.CompletedKeyPoints) == len(keyPoints) {
			exec.Status = model.ExecutionCompleted
			exec.CompletedAt = &now
			update["status"] = model.ExecutionCompleted
			update["completedAt"] = now
		}
	}

	if err := s.execRepo.Update(oid, update); err != nil {
		return nil, err
	}

	exec.LastActivityAt = now
	return &model.CheckKeyPointResponse{
		KeyPointReached: reachedKP != nil,
		KeyPoint:        reachedKP,
		Execution:       exec,
	}, nil
}

func (s *TourExecutionService) Complete(executionID string) (*model.TourExecution, error) {
	return s.finishExecution(executionID, model.ExecutionCompleted)
}

func (s *TourExecutionService) Abandon(executionID string) (*model.TourExecution, error) {
	return s.finishExecution(executionID, model.ExecutionAbandoned)
}

func (s *TourExecutionService) finishExecution(executionID string, status model.ExecutionStatus) (*model.TourExecution, error) {
	oid, err := bson.ObjectIDFromHex(executionID)
	if err != nil {
		return nil, errors.New("invalid executionId")
	}

	exec, err := s.execRepo.FindByID(oid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("execution not found")
	}
	if err != nil {
		return nil, err
	}
	if exec.Status != model.ExecutionActive {
		return nil, errors.New("execution is not active")
	}

	now := time.Now()
	update := bson.M{
		"status":         status,
		"lastActivityAt": now,
	}
	if status == model.ExecutionCompleted {
		update["completedAt"] = now
		exec.CompletedAt = &now
	} else {
		update["abandonedAt"] = now
		exec.AbandonedAt = &now
	}

	if err := s.execRepo.Update(oid, update); err != nil {
		return nil, err
	}

	exec.Status = status
	exec.LastActivityAt = now
	return exec, nil
}

func (s *TourExecutionService) GetByID(executionID string) (*model.TourExecution, error) {
	oid, err := bson.ObjectIDFromHex(executionID)
	if err != nil {
		return nil, errors.New("invalid executionId")
	}
	exec, err := s.execRepo.FindByID(oid)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.New("execution not found")
	}
	return exec, err
}

func (s *TourExecutionService) GetByTourist(touristID string) ([]model.TourExecution, error) {
	return s.execRepo.FindByTourist(touristID)
}

// haversineMeters returns distance in meters between two coordinates
func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000.0 // Earth radius in meters
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
