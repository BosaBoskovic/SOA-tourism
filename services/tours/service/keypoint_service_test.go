package service

import (
	"testing"
	"tours/model"
)

// These only exercise the validation guards that run before KeyPointService
// ever touches its repos, so a zero-value service (nil repo/tourRepo) is
// safe to use - the point is that these requests never get that far.
func TestKeyPointService_Create_Validation(t *testing.T) {
	svc := &KeyPointService{}
	validTourID := "507f1f77bcf86cd799439011"

	tests := []struct {
		name string
		req  *model.CreateKeyPointRequest
	}{
		{
			"invalid tourId is rejected",
			&model.CreateKeyPointRequest{TourID: "not-a-valid-object-id", Name: "Trg", Latitude: 44.8, Longitude: 20.4},
		},
		{
			"empty name is rejected",
			&model.CreateKeyPointRequest{TourID: validTourID, Name: "", Latitude: 44.8, Longitude: 20.4},
		},
		{
			"(0,0) coordinates are rejected as missing, not a real location",
			&model.CreateKeyPointRequest{TourID: validTourID, Name: "Trg", Latitude: 0, Longitude: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := svc.Create(tt.req, "guide1", "guide"); err == nil {
				t.Errorf("expected Create to reject %+v, got no error", tt.req)
			}
		})
	}
}
