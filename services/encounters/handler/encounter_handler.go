package handler

import (
	"errors"
	"net/http"
	"strconv"

	"encounters/auth"
	"encounters/model"
	"encounters/service"

	"github.com/gin-gonic/gin"
)

type EncounterHandler struct {
	svc *service.EncounterService
}

func NewEncounterHandler(svc *service.EncounterService) *EncounterHandler {
	return &EncounterHandler{svc: svc}
}

func (h *EncounterHandler) RegisterRoutes(r *gin.Engine, verifier *auth.Verifier) {
	group := r.Group("/encounters")
	group.Use(verifier.RequireAuth())
	group.POST("", h.create)
	group.GET("", h.list)
	group.GET("/nearby", h.nearby)
	group.GET("/mine", h.mine)
	group.POST("/:id/claim", h.claim)
	group.DELETE("/:id", h.delete)
}

func (h *EncounterHandler) create(c *gin.Context) {
	identity := auth.CurrentIdentity(c)
	if identity.Role != "guide" && identity.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Samo vodici mogu kreirati izazove"})
		return
	}

	var req model.CreateEncounterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	encounter, err := h.svc.Create(&req, identity.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, encounter)
}

func (h *EncounterHandler) list(c *gin.Context) {
	identity := auth.CurrentIdentity(c)
	encounters, err := h.svc.ListForTourist(identity.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri citanju izazova"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"encounters": encounters})
}

func (h *EncounterHandler) nearby(c *gin.Context) {
	identity := auth.CurrentIdentity(c)
	lat, latErr := strconv.ParseFloat(c.Query("lat"), 64)
	lon, lonErr := strconv.ParseFloat(c.Query("lon"), 64)
	if latErr != nil || lonErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat i lon su obavezni"})
		return
	}
	radiusKm, err := strconv.ParseFloat(c.DefaultQuery("radiusKm", "5"), 64)
	if err != nil || radiusKm <= 0 {
		radiusKm = 5
	}

	encounters, err := h.svc.Nearby(identity.Username, lat, lon, radiusKm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri pretrazi izazova"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"encounters": encounters})
}

func (h *EncounterHandler) mine(c *gin.Context) {
	identity := auth.CurrentIdentity(c)
	encounters, err := h.svc.ListForTourist(identity.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri citanju izazova"})
		return
	}
	claimed := make([]model.EncounterWithProgress, 0)
	for _, e := range encounters {
		if e.Claimed {
			claimed = append(claimed, e)
		}
	}
	c.JSON(http.StatusOK, gin.H{"encounters": claimed})
}

func (h *EncounterHandler) claim(c *gin.Context) {
	identity := auth.CurrentIdentity(c)
	id := c.Param("id")

	var req model.ClaimEncounterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	progress, err := h.svc.Claim(id, identity.Username, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTooFarAway):
			c.JSON(http.StatusConflict, gin.H{"error": "Predaleko si da bi preuzeo/la ovaj izazov"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, progress)
}

func (h *EncounterHandler) delete(c *gin.Context) {
	identity := auth.CurrentIdentity(c)
	id := c.Param("id")

	if err := h.svc.Delete(id, identity.Username, identity.Role); err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "Nije tvoj izazov"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Izazov obrisan"})
}
