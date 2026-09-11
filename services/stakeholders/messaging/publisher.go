package messaging

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"

	"stakeholders/model"
)

// DeliveryMessage is what's published to notification-delivery: just
// enough for a worker to simulate sending it and to update the stored
// notification afterward. It's a separate wire type from model.Notification
// because Notification.Username is `json:"-"` (must never leak into the
// GET /notifications API response) but the delivery worker still needs it.
type DeliveryMessage struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	Type            string `json:"type"`
	Message         string `json:"message"`
	RelatedUsername string `json:"relatedUsername,omitempty"`
	CreatedAt       string `json:"createdAt"`
}

// Publisher hands a freshly-created notification to the notification-delivery
// queue. It satisfies service.DeliveryPublisher structurally (that
// interface is declared in the service package to avoid an import cycle,
// since this package also depends on stakeholders/repo for the consumer).
type Publisher struct {
	broker *Broker
}

func NewPublisher(broker *Broker) *Publisher {
	return &Publisher{broker: broker}
}

func (p *Publisher) Publish(ctx context.Context, n model.Notification) error {
	body, err := json.Marshal(DeliveryMessage{
		ID:              n.ID,
		Username:        n.Username,
		Type:            n.Type,
		Message:         n.Message,
		RelatedUsername: n.RelatedUsername,
		CreatedAt:       n.CreatedAt,
	})
	if err != nil {
		return err
	}

	headers := InjectAMQPHeaders(ctx, amqp.Table{})
	return p.broker.Publish(ctx, queueDelivery, headers, body)
}
