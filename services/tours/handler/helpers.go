package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"tours/auth"
	"tours/service"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}

// requireIdentity writes 401 and returns ok=false if the request has no
// verified caller (see auth.Verifier.Middleware).
func requireIdentity(w http.ResponseWriter, r *http.Request) (*auth.Identity, bool) {
	id, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return nil, false
	}
	return id, true
}

// respondServiceError maps service-layer sentinel errors to the right HTTP
// status; anything else falls back to the caller-supplied default.
func respondServiceError(w http.ResponseWriter, err error, defaultStatus int) {
	if errors.Is(err, service.ErrForbidden) {
		respondError(w, http.StatusForbidden, "You do not have permission to perform this action")
		return
	}
	respondError(w, defaultStatus, err.Error())
}
