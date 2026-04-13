package handler

import (
	"net/http"
	"stakeholders/model"
	"stakeholders/service"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	svc     *service.ProfileService
	authSvc *service.AuthService
}

func NewProfileHandler(svc *service.ProfileService, authSvc *service.AuthService) *ProfileHandler {
	return &ProfileHandler{svc: svc, authSvc: authSvc}
}

func (h *ProfileHandler) RegisterRoutes(r *gin.Engine) {
	profile := r.Group("/stakeholders/profile")
	profile.Use(h.authMiddleware())
	profile.GET("", h.getProfile)
	profile.PUT("", h.updateProfile)
}

func (h *ProfileHandler) getProfile(c *gin.Context) {
	username := c.GetString("username")

	resp, err := h.svc.GetProfile(c.Request.Context(), username)
	if err != nil {
		switch err.Error() {
		case "profile_not_found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Profil nije pronadjen"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri citanju profila"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"profile": resp})
}

func (h *ProfileHandler) updateProfile(c *gin.Context) {
	username := c.GetString("username")

	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.UpdateProfile(c.Request.Context(), username, req)
	if err != nil {
		switch err.Error() {
		case "profile_not_found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Profil nije pronadjen"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri azuriranju profila"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profil uspesno azuriran", "profile": resp})
}

func (h *ProfileHandler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Nedostaje Authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Neispravan Authorization header"})
			c.Abort()
			return
		}

		claims, err := h.authSvc.ParseClaims(strings.TrimSpace(parts[1]))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Neispravan ili istekao token"})
			c.Abort()
			return
		}

		c.Set("username", claims.Subject)
		c.Set("role", claims.Role)
		c.Next()
	}
}
