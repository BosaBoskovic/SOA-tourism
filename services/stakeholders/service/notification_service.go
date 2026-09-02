package service

import (
	"context"
	"errors"
	"strings"

	"stakeholders/model"
	"stakeholders/repo"
)

const notificationListLimit = 50

type NotificationService struct {
	repo *repo.NotificationRepo
}

func NewNotificationService(repo *repo.NotificationRepo) *NotificationService {
	return &NotificationService{repo: repo}
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

	return s.repo.Create(ctx, model.Notification{
		Username:        username,
		Type:            req.Type,
		Message:         req.Message,
		RelatedUsername: req.RelatedUsername,
	})
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
