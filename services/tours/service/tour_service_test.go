package service

import (
	"testing"
	"tours/model"
)

func TestValidateDurations(t *testing.T) {
	tests := []struct {
		name      string
		durations []model.TourDuration
		wantErr   bool
	}{
		{"empty is valid", nil, false},
		{"walk is a valid transport", []model.TourDuration{{Transport: model.TransportWalk, Minutes: 30}}, false},
		{"bike is a valid transport", []model.TourDuration{{Transport: model.TransportBike, Minutes: 15}}, false},
		{"car is a valid transport", []model.TourDuration{{Transport: model.TransportCar, Minutes: 10}}, false},
		{"unknown transport is rejected", []model.TourDuration{{Transport: "scooter", Minutes: 10}}, true},
		{"zero minutes is rejected", []model.TourDuration{{Transport: model.TransportWalk, Minutes: 0}}, true},
		{"negative minutes is rejected", []model.TourDuration{{Transport: model.TransportWalk, Minutes: -5}}, true},
		{
			"one bad entry among good ones still fails the whole batch",
			[]model.TourDuration{
				{Transport: model.TransportWalk, Minutes: 30},
				{Transport: model.TransportBike, Minutes: 0},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDurations(tt.durations)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDurations(%+v) error = %v, wantErr %v", tt.durations, err, tt.wantErr)
			}
		})
	}
}
