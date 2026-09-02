package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"stakeholders/model"
	"stakeholders/service"
)

type NotificationHandler struct {
	svc     *service.NotificationService
	authSvc *service.AuthService
}

func NewNotificationHandler(svc *service.NotificationService, authSvc *service.AuthService) *NotificationHandler {
	return &NotificationHandler{svc: svc, authSvc: authSvc}
}

func (h *NotificationHandler) RegisterRoutes(r *gin.Engine) {
	// Called service-to-service by followers/blog/tours right after a
	// follow/comment/review succeeds - only reachable from other containers
	// on soa-network (stakeholders publishes no host port), same trust
	// model as followers' is-following/visible-authors.
	r.POST("/stakeholders/notifications/internal", h.createInternal)

	notifications := r.Group("/stakeholders/notifications")
	notifications.Use(h.authMiddleware())
	notifications.GET("", h.list)
	notifications.PATCH("/read-all", h.markAllRead)
	notifications.PATCH("/:id/read", h.markRead)
}

func (h *NotificationHandler) createInternal(c *gin.Context) {
	var req model.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Create(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func (h *NotificationHandler) list(c *gin.Context) {
	username := c.GetString("username")
	notifications, err := h.svc.ListForUser(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri citanju obavestenja"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

func (h *NotificationHandler) markRead(c *gin.Context) {
	username := c.GetString("username")
	id := c.Param("id")
	if err := h.svc.MarkRead(c.Request.Context(), username, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri azuriranju obavestenja"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *NotificationHandler) markAllRead(c *gin.Context) {
	username := c.GetString("username")
	if err := h.svc.MarkAllRead(c.Request.Context(), username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Greska pri azuriranju obavestenja"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *NotificationHandler) authMiddleware() gin.HandlerFunc {
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
		c.Next()
	}
}
