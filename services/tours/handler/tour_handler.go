package handler

import (
	"encoding/json"
	"net/http"
	"tours/model"
	"tours/service"

	"github.com/gorilla/mux"
)

type TourHandler struct {
	service *service.TourService
}

func NewTourHandler(service *service.TourService) *TourHandler {
	return &TourHandler{service: service}
}

// POST /tours
func (h *TourHandler) Create(w http.ResponseWriter, r *http.Request) {
	id, ok := requireIdentity(w, r)
	if !ok {
		return
	}

	var req model.CreateTourRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tour, err := h.service.Create(&req, id.Username, id.Role)
	if err != nil {
		respondServiceError(w, err, http.StatusBadRequest)
		return
	}

	respondJSON(w, http.StatusCreated, tour)
}

// GET /tours/{id}?touristId=xxx
func (h *TourHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	touristID := r.URL.Query().Get("touristId")

	tour, err := h.service.GetByIDForTourist(id, touristID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, tour)
}

// GET /tours/author/{authorId}
func (h *TourHandler) GetByAuthor(w http.ResponseWriter, r *http.Request) {
	authorID := mux.Vars(r)["authorId"]

	tours, err := h.service.GetByAuthor(authorID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, tours)
}

// PUT /tours/{id}
func (h *TourHandler) Update(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireIdentity(w, r)
	if !ok {
		return
	}
	tourID := mux.Vars(r)["id"]

	var req model.UpdateTourRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tour, err := h.service.Update(tourID, &req, caller.Username, caller.Role)
	if err != nil {
		respondServiceError(w, err, http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, tour)
}

// GET /tours
func (h *TourHandler) GetPublished(w http.ResponseWriter, r *http.Request) {
	tours, err := h.service.GetPublished()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, tours)
}

// PUT /tours/{id}/publish
func (h *TourHandler) Publish(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireIdentity(w, r)
	if !ok {
		return
	}
	tourID := mux.Vars(r)["id"]

	tour, err := h.service.Publish(tourID, caller.Username, caller.Role)
	if err != nil {
		respondServiceError(w, err, http.StatusBadRequest)
		return
	}

	respondJSON(w, http.StatusOK, tour)
}

// PUT /tours/{id}/archive
func (h *TourHandler) Archive(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireIdentity(w, r)
	if !ok {
		return
	}
	tourID := mux.Vars(r)["id"]

	tour, err := h.service.Archive(tourID, caller.Username, caller.Role)
	if err != nil {
		respondServiceError(w, err, http.StatusBadRequest)
		return
	}

	respondJSON(w, http.StatusOK, tour)
}

// PUT /tours/{id}/activate
func (h *TourHandler) Activate(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireIdentity(w, r)
	if !ok {
		return
	}
	tourID := mux.Vars(r)["id"]

	tour, err := h.service.Activate(tourID, caller.Username, caller.Role)
	if err != nil {
		respondServiceError(w, err, http.StatusBadRequest)
		return
	}

	respondJSON(w, http.StatusOK, tour)
}
