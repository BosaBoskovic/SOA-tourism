package repo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"stakeholders/model"
)

type NotificationRepo struct {
	driver   neo4j.DriverWithContext
	database string
}

func NewNotificationRepo(driver neo4j.DriverWithContext, database string) *NotificationRepo {
	return &NotificationRepo{driver: driver, database: database}
}

func newNotificationID() string {
	buf := make([]byte, 12)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

// Create adds a notification for username and returns the stored record,
// including its generated id/createdAt and default (pending) delivery
// status - the caller needs this back to publish it to the async delivery
// pipeline (see services/stakeholders/messaging). Best-effort by design:
// callers (followers/blog/tours, or stakeholders itself) should log and
// move on if this fails, not treat it as fatal to whatever triggered it.
func (r *NotificationRepo) Create(ctx context.Context, n model.Notification) (model.Notification, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	id := newNotificationID()
	createdAt, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx,
			`CREATE (n:Notification {
				id: $id,
				username: $username,
				type: $type,
				message: $message,
				relatedUsername: $relatedUsername,
				createdAt: datetime(),
				read: false,
				deliveryStatus: $deliveryStatus,
				deliveryAttempts: 0
			})
			RETURN n.createdAt AS createdAt`,
			map[string]any{
				"id":              id,
				"username":        n.Username,
				"type":            n.Type,
				"message":         n.Message,
				"relatedUsername": n.RelatedUsername,
				"deliveryStatus":  model.DeliveryStatusPending,
			},
		)
		if err != nil {
			return nil, err
		}
		record, err := res.Single(ctx)
		if err != nil {
			return nil, err
		}
		value, _ := record.Get("createdAt")
		return value, nil
	})
	if err != nil {
		return model.Notification{}, err
	}

	n.ID = id
	n.CreatedAt = fmt.Sprintf("%v", createdAt)
	n.DeliveryStatus = model.DeliveryStatusPending
	n.DeliveryAttempts = 0
	return n, nil
}

// ListForUser returns username's most recent notifications, newest first.
func (r *NotificationRepo) ListForUser(ctx context.Context, username string, limit int) ([]model.Notification, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx,
			`MATCH (n:Notification {username: $username})
			 RETURN n.id AS id, n.type AS type, n.message AS message,
			        n.relatedUsername AS relatedUsername, n.createdAt AS createdAt, n.read AS read,
			        COALESCE(n.deliveryStatus, $defaultStatus) AS deliveryStatus,
			        COALESCE(n.deliveryAttempts, 0) AS deliveryAttempts,
			        n.deliveredAt AS deliveredAt
			 ORDER BY n.createdAt DESC
			 LIMIT $limit`,
			// deliveryStatus/deliveryAttempts are COALESCEd because this
			// service has no migration mechanism - notification nodes
			// created before the delivery pipeline existed simply don't
			// have these properties yet.
			map[string]any{"username": username, "limit": int64(limit), "defaultStatus": model.DeliveryStatusPending},
		)
		if err != nil {
			return nil, err
		}
		records, err := res.Collect(ctx)
		if err != nil {
			return nil, err
		}

		notifications := make([]model.Notification, 0, len(records))
		for _, rec := range records {
			id, _ := rec.Get("id")
			typ, _ := rec.Get("type")
			message, _ := rec.Get("message")
			relatedUsername, _ := rec.Get("relatedUsername")
			createdAt, _ := rec.Get("createdAt")
			read, _ := rec.Get("read")
			deliveryStatus, _ := rec.Get("deliveryStatus")
			deliveryAttempts, _ := rec.Get("deliveryAttempts")
			deliveredAt, _ := rec.Get("deliveredAt")

			notifications = append(notifications, model.Notification{
				ID:               asString(id),
				Username:         username,
				Type:             asString(typ),
				Message:          asString(message),
				RelatedUsername:  asString(relatedUsername),
				CreatedAt:        fmt.Sprintf("%v", createdAt),
				Read:             asBool(read),
				DeliveryStatus:   asString(deliveryStatus),
				DeliveryAttempts: asInt(deliveryAttempts),
				DeliveredAt:      asOptionalString(deliveredAt),
			})
		}
		return notifications, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]model.Notification), nil
}

func (r *NotificationRepo) MarkRead(ctx context.Context, username, id string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx,
			`MATCH (n:Notification {id: $id, username: $username}) SET n.read = true`,
			map[string]any{"id": id, "username": username},
		)
		return nil, err
	})
	return err
}

func (r *NotificationRepo) MarkAllRead(ctx context.Context, username string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx,
			`MATCH (n:Notification {username: $username, read: false}) SET n.read = true`,
			map[string]any{"username": username},
		)
		return nil, err
	})
	return err
}

// UpdateDeliveryStatus records the outcome of an async delivery attempt -
// either a terminal "delivered"/"failed", driven by the messaging consumer
// worker after it either succeeds or exhausts its retries (see
// services/stakeholders/messaging/consumer.go).
func (r *NotificationRepo) UpdateDeliveryStatus(ctx context.Context, id, status string, attempts int, deliveredAt *time.Time) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: r.database})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("cannot close neo4j session: %v", err)
		}
	}()

	var deliveredAtParam any
	if deliveredAt != nil {
		deliveredAtParam = *deliveredAt
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx,
			`MATCH (n:Notification {id: $id})
			 SET n.deliveryStatus = $status, n.deliveryAttempts = $attempts, n.deliveredAt = $deliveredAt`,
			map[string]any{
				"id":          id,
				"status":      status,
				"attempts":    attempts,
				"deliveredAt": deliveredAtParam,
			},
		)
		return nil, err
	})
	return err
}
