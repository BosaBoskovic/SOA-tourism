package handler

import (
	"errors"
	"net/http/httptest"
	"testing"

	"tours/service"
)

func TestParseOptionalFloat(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantNil bool
		want    float64
	}{
		{"empty string is nil", "", true, 0},
		{"unparseable string is nil", "not-a-number", true, 0},
		{"valid integer", "42", false, 42},
		{"valid decimal", "19.99", false, 19.99},
		{"negative value", "-5", false, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOptionalFloat(tt.raw)
			if tt.wantNil {
				if got != nil {
					t.Errorf("parseOptionalFloat(%q) = %v, want nil", tt.raw, *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("parseOptionalFloat(%q) = nil, want %v", tt.raw, tt.want)
			}
			if *got != tt.want {
				t.Errorf("parseOptionalFloat(%q) = %v, want %v", tt.raw, *got, tt.want)
			}
		})
	}
}

func TestRespondServiceError_MapsForbiddenTo403(t *testing.T) {
	w := httptest.NewRecorder()
	respondServiceError(w, service.ErrForbidden, 400)

	if w.Code != 403 {
		t.Errorf("expected ErrForbidden to map to 403, got %d", w.Code)
	}
}

func TestRespondServiceError_FallsBackToDefaultStatus(t *testing.T) {
	w := httptest.NewRecorder()
	respondServiceError(w, errors.New("tour not found"), 404)

	if w.Code != 404 {
		t.Errorf("expected a generic error to use the caller-supplied default status, got %d", w.Code)
	}
}
