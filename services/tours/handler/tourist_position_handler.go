package handler

import (
	"encoding/json"
	"net/http"
	"tours/model"
	"tours/service"

	"github.com/gorilla/mux"
)

type TouristPositionHandler struct {
	service *service.TouristPositionService
}

func NewTouristPositionHandler(service *service.TouristPositionService) *TouristPositionHandler {
	return &TouristPositionHandler{service: service}
}

// PUT /tourist-position
func (h *TouristPositionHandler) Update(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireIdentity(w, r)
	if !ok {
		return
	}

	var req model.UpdateTouristPositionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	// Always the verified caller's own position - never client-supplied.
	req.TouristID = caller.Username

	position, err := h.service.Update(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, position)
}

// GET /tourist-position/{touristId}
// A live GPS position is sensitive - only the tourist themselves (or an
// admin) may read it.
func (h *TouristPositionHandler) GetByTouristID(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireIdentity(w, r)
	if !ok {
		return
	}
	touristID := mux.Vars(r)["touristId"]
	if caller.Role != "admin" && caller.Username != touristID {
		respondError(w, http.StatusForbidden, "You can only view your own position")
		return
	}

	position, err := h.service.GetByTouristID(touristID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, position)
}