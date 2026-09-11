package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"stakeholders/model"
	"stakeholders/repo"
)

const notificationListLimit = 50

// DeliveryPublisher hands a freshly-created notification off to the async
// delivery pipeline (see services/stakeholders/messaging). It's an
// interface, not a concrete *messaging.Publisher, so this package doesn't
// need to depend on amqp091-go, and so tests can construct a
// NotificationService without a broker.
type DeliveryPublisher interface {
	Publish(ctx context.Context, n model.Notification) error
}

type NotificationService struct {
	repo      *repo.NotificationRepo
	publisher DeliveryPublisher
	logger    *slog.Logger
}

// NewNotificationService wires the notification repo with (optionally) an
// async delivery publisher. publisher may be nil - Create simply skips
// publishing in that case, which keeps this service usable in tests/tools
// that don't stand up a broker. logger may also be nil, in which case a
// publish failure is dropped rather than logged (still never fatal to the
// notification that was already created).
func NewNotificationService(repo *repo.NotificationRepo, publisher DeliveryPublisher, logger *slog.Logger) *NotificationService {
	return &NotificationService{repo: repo, publisher: publisher, logger: logger}
}

func (s *NotificationService) Create(ctx context.Context, req model.CreateNotificationRequest) error {
	username := strings.TrimSpace(req.Username)
	if username == "" || strings.TrimSpace(req.Message) == "" {
		return errors.New("username and message are required")
	}
	switch req.Type {
	case model.NotificationTypeFollow, model.NotificationTypeComment, model.NotificationTypeReview:
	default:
		return errors.New("invalid notification type")
	}

	created, err := s.repo.Create(ctx, model.Notification{
		Username:        username,
		Type:            req.Type,
		Message:         req.Message,
		RelatedUsername: req.RelatedUsername,
	})
	if err != nil {
		return err
	}

	if s.publisher != nil {
		// Best-effort, matching this codebase's existing philosophy for
		// cross-service/cross-boundary calls (e.g. tours' notify package):
		// a delivery-pipeline hiccup should never fail the notification
		// that's already been created.
		if err := s.publisher.Publish(ctx, created); err != nil && s.logger != nil {
			s.logger.Warn("notification delivery publish failed", "id", created.ID, "error", err)
		}
	}
	return nil
}

func (s *NotificationService) ListForUser(ctx context.Context, username string) ([]model.Notification, error) {
	return s.repo.ListForUser(ctx, username, notificationListLimit)
}

func (s *NotificationService) MarkRead(ctx context.Context, username, id string) error {
	return s.repo.MarkRead(ctx, username, id)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, username string) error {
	return s.repo.MarkAllRead(ctx, username)
}
