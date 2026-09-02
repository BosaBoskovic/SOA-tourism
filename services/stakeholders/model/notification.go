package model

// Notification types this service creates. followers/blog/tours call the
// internal endpoint with one of these after a triggering action succeeds.
const (
	NotificationTypeFollow  = "follow"
	NotificationTypeComment = "comment"
	NotificationTypeReview  = "review"
)

type Notification struct {
	ID              string `json:"id"`
	Username        string `json:"-"`
	Type            string `json:"type"`
	Message         string `json:"message"`
	RelatedUsername string `json:"relatedUsername,omitempty"`
	CreatedAt       string `json:"createdAt"`
	Read            bool   `json:"read"`
}

// CreateNotificationRequest is what followers/blog/tours POST to the
// internal notification endpoint.
type CreateNotificationRequest struct {
	Username        string `json:"username" binding:"required"`
	Type            string `json:"type" binding:"required"`
	Message         string `json:"message" binding:"required"`
	RelatedUsername string `json:"relatedUsername"`
}
