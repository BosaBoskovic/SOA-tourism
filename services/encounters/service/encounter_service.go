package service

import (
	"errors"
	"math"
	"time"

	"encounters/model"
	"encounters/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrForbidden = errors.New("forbidden")
var ErrTooFarAway = errors.New("too far away to claim this encounter")

type EncounterService struct {
	repo         *repository.EncounterRepository
	progressRepo *repository.ProgressRepository
}

func NewEncounterService(repo *repository.EncounterRepository, progressRepo *repository.ProgressRepository) *EncounterService {
	return &EncounterService{repo: repo, progressRepo: progressRepo}
}

// Create is guide-only, enforced by the handler (role check) before this is
// ever called - kept here as a doc note, not re-checked, since role isn't
// domain state this service owns.
func (s *EncounterService) Create(req *model.CreateEncounterRequest, callerUsername string) (*model.Encounter, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.Latitude == 0 && req.Longitude == 0 {
		return nil, errors.New("latitude and longitude are required")
	}
	if req.RadiusMeters <= 0 {
		req.RadiusMeters = 50 // sane default: "you're basically standing on it"
	}

	e := &model.Encounter{
		ID:           bson.NewObjectID(),
		Name:         req.Name,
		Description:  req.Description,
		TourID:       req.TourID,
		KeyPointID:   req.KeyPointID,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		RadiusMeters: req.RadiusMeters,
		Reward:       req.Reward,
		CreatedBy:    callerUsername,
		CreatedAt:    time.Now(),
	}
	if err := s.repo.Create(e); err != nil {
		return nil, err
	}
	return e, nil
}

// ListForTourist returns every encounter, annotated with whether/when the
// given tourist has already claimed each one.
func (s *EncounterService) ListForTourist(touristID string) ([]model.EncounterWithProgress, error) {
	encounters, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	progress, err := s.progressRepo.FindByTourist(touristID)
	if err != nil {
		return nil, err
	}
	completedAt := make(map[bson.ObjectID]time.Time, len(progress))
	for _, p := range progress {
		completedAt[p.EncounterID] = p.CompletedAt
	}

	result := make([]model.EncounterWithProgress, 0, len(encounters))
	for _, e := range encounters {
		item := model.EncounterWithProgress{Encounter: e}
		if t, ok := completedAt[e.ID]; ok {
			item.Claimed = true
			tCopy := t
			item.CompletedAt = &tCopy
		}
		result = append(result, item)
	}
	return result, nil
}

// Nearby filters to encounters within radiusKm of (lat, lon) - a simple
// proximity search, not a full geospatial index (fine at this data volume).
func (s *EncounterService) Nearby(touristID string, lat, lon, radiusKm float64) ([]model.EncounterWithProgress, error) {
	all, err := s.ListForTourist(touristID)
	if err != nil {
		return nil, err
	}
	radiusMeters := radiusKm * 1000
	nearby := make([]model.EncounterWithProgress, 0, len(all))
	for _, e := range all {
		if haversineMeters(lat, lon, e.Latitude, e.Longitude) <= radiusMeters {
			nearby = append(nearby, e)
		}
	}
	return nearby, nil
}

// Claim records completion if the tourist is actually within the
// encounter's radius. Idempotent - claiming an already-claimed encounter
// just returns the original record rather than erroring.
func (s *EncounterService) Claim(id string, touristID string, req *model.ClaimEncounterRequest) (*model.EncounterProgress, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid encounter id")
	}

	encounter, err := s.repo.FindByID(oid)
	if err != nil {
		return nil, errors.New("encounter not found")
	}

	distance := haversineMeters(req.Latitude, req.Longitude, encounter.Latitude, encounter.Longitude)
	if distance > encounter.RadiusMeters {
		return nil, ErrTooFarAway
	}

	return s.progressRepo.Claim(touristID, oid)
}

func (s *EncounterService) Delete(id string, callerUsername, callerRole string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid encounter id")
	}
	encounter, err := s.repo.FindByID(oid)
	if err != nil {
		return errors.New("encounter not found")
	}
	if encounter.CreatedBy != callerUsername && callerRole != "admin" {
		return ErrForbidden
	}
	return s.repo.Delete(oid)
}

// haversineMeters returns distance in meters between two coordinates - same
// formula tours uses for its own keypoint-proximity check.
func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000.0
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
