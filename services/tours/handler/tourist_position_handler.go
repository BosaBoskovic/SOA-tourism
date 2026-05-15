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
	var req model.UpdateTouristPositionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	position, err := h.service.Update(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, position)
}

// GET /tourist-position/{touristId}
func (h *TouristPositionHandler) GetByTouristID(w http.ResponseWriter, r *http.Request) {
	touristID := mux.Vars(r)["touristId"]

	position, err := h.service.GetByTouristID(touristID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, position)
}