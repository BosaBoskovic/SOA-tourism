package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"stakeholders/model"
	"stakeholders/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/stakeholders/register", h.register)
	r.POST("/stakeholders/login", h.login)

	admin := r.Group("/stakeholders")
	admin.Use(h.adminOnlyMiddleware())
	admin.GET("/accounts", h.getAllAccounts)
	admin.PATCH("/accounts/:username/block", h.blockAccount)
}

func (h *AuthHandler) register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		switch err.Error() {
		case "role must be guide or tourist":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Dozvoljene uloge za registraciju su vodic i turista"})
		case "username_or_email_exists":
			c.JSON(http.StatusConflict, gin.H{"error": "Korisnicko ime ili email vec postoje"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri cuvanju naloga"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Uspesna registracija", "account": resp})
}

func (h *AuthHandler) login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, expiresAt, acc, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		switch err.Error() {
		case "invalid_credentials":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Pogresni kredencijali"})
		case "account_blocked":
			c.JSON(http.StatusForbidden, gin.H{"error": "Nalog je blokiran"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri prijavi"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Uspesna prijava",
		"accessToken": token,
		"tokenType":   "Bearer",
		"expiresIn":   int64((15 * time.Minute).Seconds()),
		"expiresAt":   expiresAt.Format(time.RFC3339),
		"account": gin.H{
			"username": acc.Username,
			"email":    acc.Email,
			"role":     acc.Role,
		},
	})
}

func (h *AuthHandler) getAllAccounts(c *gin.Context) {
	accounts, err := h.svc.GetAllAccounts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri citanju naloga"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"accounts": accounts})
}

func (h *AuthHandler) blockAccount(c *gin.Context) {
	username := strings.TrimSpace(c.Param("username"))
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username je obavezan"})
		return
	}

	err := h.svc.BlockAccount(c.Request.Context(), username)
	if err != nil {
		switch err.Error() {
		case "account_not_found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Nalog nije pronadjen"})
		case "cannot_block_admin":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Admin nalog ne moze biti blokiran"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri blokiranju naloga"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nalog je uspesno blokiran"})
}

func (h *AuthHandler) adminOnlyMiddleware() gin.HandlerFunc {
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

		claims, err := h.svc.ParseAdminClaims(strings.TrimSpace(parts[1]))
		if err != nil {
			switch err.Error() {
			case "forbidden":
				c.JSON(http.StatusForbidden, gin.H{"error": "Samo admin ima pristup"})
			default:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Neispravan ili istekao token"})
			}
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}
