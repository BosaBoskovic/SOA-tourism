package handler

import (
	"encoding/json"
	"net/http"
	"tours/model"
	"tours/service"

	"github.com/gorilla/mux"
)

type ReviewHandler struct {
	service *service.ReviewService
}

func NewReviewHandler(service *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: service}
}

// POST /reviews
func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	review, err := h.service.Create(&req)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "tourist has already reviewed this tour" {
			status = http.StatusConflict
		}
		if err.Error() == "tour not found" {
			status = http.StatusNotFound
		}
		respondError(w, status, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, review)
}

// GET /reviews/{id}
func (h *ReviewHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	review, err := h.service.GetByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, review)
}

// GET /reviews/tour/{tourId}
func (h *ReviewHandler) GetByTour(w http.ResponseWriter, r *http.Request) {
	tourID := mux.Vars(r)["tourId"]

	reviews, err := h.service.GetByTour(tourID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, reviews)
}

// DELETE /reviews/{id}
func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if err := h.service.Delete(id); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Review deleted"})
}
