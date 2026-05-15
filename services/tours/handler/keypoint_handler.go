package handler

import (
	"encoding/json"
	"net/http"
	"tours/model"
	"tours/service"

	"github.com/gorilla/mux"
)

type KeyPointHandler struct {
	service *service.KeyPointService
}

func NewKeyPointHandler(service *service.KeyPointService) *KeyPointHandler {
	return &KeyPointHandler{service: service}
}

// POST /keypoints
func (h *KeyPointHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateKeyPointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	kp, err := h.service.Create(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, kp)
}

// GET /keypoints/{id}
func (h *KeyPointHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	kp, err := h.service.GetByID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, kp)
}

// GET /keypoints/tour/{tourId}
func (h *KeyPointHandler) GetByTour(w http.ResponseWriter, r *http.Request) {
	tourID := mux.Vars(r)["tourId"]

	kps, err := h.service.GetByTour(tourID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, kps)
}

// PUT /keypoints/{id}
func (h *KeyPointHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req model.UpdateKeyPointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	kp, err := h.service.Update(id, &req)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, kp)
}

// DELETE /keypoints/{id}
func (h *KeyPointHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if err := h.service.Delete(id); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Key point deleted"})
}
