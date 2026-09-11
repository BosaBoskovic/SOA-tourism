package model

// Notification types this service creates. followers/blog/tours call the
// internal endpoint with one of these after a triggering action succeeds.
const (
	NotificationTypeFollow  = "follow"
	NotificationTypeComment = "comment"
	NotificationTypeReview  = "review"
)

// Delivery status of a notification's async, out-of-band send (simulated
// email/push) - separate from Read, which tracks the in-app bell item.
// See services/stakeholders/messaging for the RabbitMQ pipeline that
// drives these transitions.
const (
	DeliveryStatusPending   = "pending"
	DeliveryStatusDelivered = "delivered"
	DeliveryStatusFailed    = "failed"
)

type Notification struct {
	ID               string `json:"id"`
	Username         string `json:"-"`
	Type             string `json:"type"`
	Message          string `json:"message"`
	RelatedUsername  string `json:"relatedUsername,omitempty"`
	CreatedAt        string `json:"createdAt"`
	Read             bool   `json:"read"`
	DeliveryStatus   string `json:"deliveryStatus"`
	DeliveryAttempts int    `json:"deliveryAttempts"`
	DeliveredAt      string `json:"deliveredAt,omitempty"`
}

// CreateNotificationRequest is what followers/blog/tours POST to the
// internal notification endpoint.
type CreateNotificationRequest struct {
	Username        string `json:"username" binding:"required"`
	Type            string `json:"type" binding:"required"`
	Message         string `json:"message" binding:"required"`
	RelatedUsername string `json:"relatedUsername"`
}
