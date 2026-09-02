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
	r.POST("/stakeholders/refresh", h.refresh)
	r.POST("/stakeholders/logout", h.logout)
	r.POST("/stakeholders/password-reset/request", h.requestPasswordReset)
	r.POST("/stakeholders/password-reset/confirm", h.confirmPasswordReset)

	authed := r.Group("/stakeholders")
	authed.Use(h.authMiddleware())
	authed.PUT("/password", h.changePassword)

	admin := r.Group("/stakeholders")
	admin.Use(h.adminOnlyMiddleware())
	admin.GET("/accounts", h.getAllAccounts)
	admin.PATCH("/accounts/:username/block", h.blockAccount)
	admin.PATCH("/accounts/:username/unblock", h.unblockAccount)
	admin.DELETE("/accounts/:username", h.deleteAccount)
}

func tokenPairResponse(message string, pair service.TokenPair, account gin.H) gin.H {
	return gin.H{
		"message":               message,
		"accessToken":           pair.AccessToken,
		"tokenType":             "Bearer",
		"expiresIn":             int64(15 * time.Minute / time.Second),
		"expiresAt":             pair.AccessTokenExpiresAt.Format(time.RFC3339),
		"refreshToken":          pair.RefreshToken,
		"refreshTokenExpiresAt": pair.RefreshTokenExpiresAt.Format(time.RFC3339),
		"account":               account,
	}
}

func (h *AuthHandler) register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, pair, err := h.svc.Register(c.Request.Context(), req)
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

	c.JSON(http.StatusCreated, tokenPairResponse("Uspesna registracija", pair, gin.H{
		"username": resp.Username,
		"email":    resp.Email,
		"role":     resp.Role,
	}))
}

func (h *AuthHandler) login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pair, acc, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		switch err.Error() {
		case "invalid_credentials":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Pogresni kredencijali"})
		case "account_blocked":
			c.JSON(http.StatusForbidden, gin.H{"error": "Nalog je blokiran"})
		case "too_many_attempts":
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Previse pokusaja prijave, pokusajte ponovo kasnije"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri prijavi"})
		}
		return
	}

	c.JSON(http.StatusOK, tokenPairResponse("Uspesna prijava", pair, gin.H{
		"username": acc.Username,
		"email":    acc.Email,
		"role":     acc.Role,
	}))
}

func (h *AuthHandler) refresh(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pair, acc, err := h.svc.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch err.Error() {
		case "account_blocked":
			c.JSON(http.StatusForbidden, gin.H{"error": "Nalog je blokiran"})
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Neispravan ili istekao refresh token"})
		}
		return
	}

	c.JSON(http.StatusOK, tokenPairResponse("Token osvezen", pair, gin.H{
		"username": acc.Username,
		"email":    acc.Email,
		"role":     acc.Role,
	}))
}

func (h *AuthHandler) logout(c *gin.Context) {
	var req model.RefreshTokenRequest
	// Logout is best-effort - a missing/malformed body still returns 200,
	// the client is clearing its local session either way.
	_ = c.ShouldBindJSON(&req)
	_ = h.svc.Logout(c.Request.Context(), req.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"message": "Uspesna odjava"})
}

func (h *AuthHandler) changePassword(c *gin.Context) {
	username := c.GetString("username")

	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), username, req); err != nil {
		switch err.Error() {
		case "invalid_current_password":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Trenutna lozinka nije tacna"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri promeni lozinke"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lozinka je uspesno promenjena"})
}

func (h *AuthHandler) requestPasswordReset(c *gin.Context) {
	var req model.RequestPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, expiresAt, err := h.svc.RequestPasswordReset(c.Request.Context(), req.UsernameOrEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri kreiranju zahteva"})
		return
	}

	// This stack has no email service. The token is returned directly
	// instead of emailed - that's fine for a demo (the frontend shows it
	// as a "reset link"/code inline) but is NOT how this would work with
	// real users; a real deployment must swap this for an actual email send
	// and drop the token from the response body.
	resp := gin.H{
		"message": "Ako nalog postoji, kreiran je zahtev za reset lozinke",
	}
	if token != "" {
		resp["resetToken"] = token
		resp["expiresAt"] = expiresAt.Format(time.RFC3339)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) confirmPasswordReset(c *gin.Context) {
	var req model.ConfirmPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.ConfirmPasswordReset(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Neispravan ili istekao token za reset lozinke"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lozinka je uspesno promenjena"})
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

func (h *AuthHandler) unblockAccount(c *gin.Context) {
	username := strings.TrimSpace(c.Param("username"))
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username je obavezan"})
		return
	}

	if err := h.svc.UnblockAccount(c.Request.Context(), username); err != nil {
		switch err.Error() {
		case "account_not_found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Nalog nije pronadjen"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri deblokiranju naloga"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nalog je uspesno deblokiran"})
}

func (h *AuthHandler) deleteAccount(c *gin.Context) {
	username := strings.TrimSpace(c.Param("username"))
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username je obavezan"})
		return
	}

	if err := h.svc.DeleteAccount(c.Request.Context(), username); err != nil {
		switch err.Error() {
		case "account_not_found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Nalog nije pronadjen"})
		case "cannot_delete_admin":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Admin nalog ne moze biti obrisan"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri brisanju naloga"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nalog je uspesno obrisan"})
}

func (h *AuthHandler) authMiddleware() gin.HandlerFunc {
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

		claims, err := h.svc.ParseClaims(strings.TrimSpace(parts[1]))
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
