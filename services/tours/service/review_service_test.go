package service

import (
	"testing"
	"tours/model"
)

// A zero-value ReviewService (nil repos) is safe here as long as the test
// requests fail one of the pure validation checks that run before Create
// ever touches tourRepo/purchaseRepo/execRepo/repo.
const validTourIDForReviewTests = "507f1f77bcf86cd799439011"

func TestReviewService_Create_RejectsOutOfRangeRating(t *testing.T) {
	svc := &ReviewService{}

	for _, rating := range []int{-1, 0, 6, 100} {
		req := &model.CreateReviewRequest{
			TourID:      validTourIDForReviewTests,
			Rating:      rating,
			TouristName: "Ana",
		}
		if _, err := svc.Create(req, "ana"); err == nil {
			t.Errorf("rating %d: expected Create to reject it, got no error", rating)
		}
	}
}

func TestReviewService_Create_RejectsInvalidTourId(t *testing.T) {
	svc := &ReviewService{}
	req := &model.CreateReviewRequest{TourID: "not-an-object-id", Rating: 5, TouristName: "Ana"}

	if _, err := svc.Create(req, "ana"); err == nil {
		t.Error("expected an invalid tourId to be rejected")
	}
}

// Regression test for deriving tourist identity from the verified caller
// instead of trusting whatever touristId the client sent in the body.
func TestReviewService_Create_DerivesTouristIdentityFromCaller(t *testing.T) {
	svc := &ReviewService{}
	req := &model.CreateReviewRequest{
		TourID:        validTourIDForReviewTests,
		Rating:        5,
		TouristID:     "attacker-supplied-id",
		TouristName:   "Ana",
		TourVisitDate: "not-a-date", // fails validation right after the override, before any repo is touched
	}

	if _, err := svc.Create(req, "verified-caller"); err == nil {
		t.Fatal("expected the invalid tourVisitDate to produce an error")
	}
	if req.TouristID != "verified-caller" {
		t.Errorf("expected TouristID to be overwritten with the verified caller, got %q", req.TouristID)
	}
}

func TestReviewService_Create_RejectsMalformedVisitDate(t *testing.T) {
	svc := &ReviewService{}
	req := &model.CreateReviewRequest{
		TourID:        validTourIDForReviewTests,
		Rating:        5,
		TouristName:   "Ana",
		TourVisitDate: "01/05/2024", // not the required YYYY-MM-DD
	}

	if _, err := svc.Create(req, "ana"); err == nil {
		t.Error("expected a non-ISO tourVisitDate to be rejected")
	}
}
