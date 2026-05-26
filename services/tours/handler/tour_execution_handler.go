package handler

import (
	"encoding/json"
	"net/http"
	"tours/model"
	"tours/service"

	"github.com/gorilla/mux"
)

type TourExecutionHandler struct {
	service *service.TourExecutionService
}

func NewTourExecutionHandler(service *service.TourExecutionService) *TourExecutionHandler {
	return &TourExecutionHandler{service: service}
}

// POST /executions
func (h *TourExecutionHandler) Start(w http.ResponseWriter, r *http.Request) {
	var req model.StartExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Uzmi touristId iz X-Username headera (gateway ga postavlja)
	if req.TouristID == "" {
		req.TouristID = r.Header.Get("X-Username")
	}

	exec, err := h.service.Start(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, exec)
}

// GET /executions/{id}
func (h *TourExecutionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	exec, err := h.service.GetByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, exec)
}

// GET /executions/tourist/{touristId}
func (h *TourExecutionHandler) GetByTourist(w http.ResponseWriter, r *http.Request) {
	touristID := mux.Vars(r)["touristId"]
	execs, err := h.service.GetByTourist(touristID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, execs)
}

// POST /executions/{id}/check-keypoint
func (h *TourExecutionHandler) CheckKeyPoint(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req model.CheckKeyPointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.service.CheckKeyPoint(id, req.Latitude, req.Longitude)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

// PUT /executions/{id}/complete
func (h *TourExecutionHandler) Complete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	exec, err := h.service.Complete(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, exec)
}

// PUT /executions/{id}/abandon
func (h *TourExecutionHandler) Abandon(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	exec, err := h.service.Abandon(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, exec)
}
