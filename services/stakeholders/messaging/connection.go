package messaging

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	queueDelivery = "notification-delivery"
	queueDead     = "notification-delivery-dead"
)

// Broker owns one long-lived AMQP connection (plus a dedicated publish
// channel) for stakeholders' notification-delivery pipeline, reconnecting
// on its own if the connection drops - the "never fatal, just retry"
// approach tours' purchase-completed consumer already uses for its own
// connection. Consumer workers (see consumer.go) each open their own
// channel off this same connection via Channel().
type Broker struct {
	mu        sync.RWMutex
	conn      *amqp.Connection
	pubCh     *amqp.Channel
	connected bool

	logger *slog.Logger
}

// NewBroker starts connecting in the background and returns immediately -
// it never blocks service startup on RabbitMQ being reachable yet.
func NewBroker(logger *slog.Logger) *Broker {
	b := &Broker{logger: logger}
	go b.connectLoop()
	return b
}

func (b *Broker) connectLoop() {
	for {
		if err := b.connectOnce(); err != nil {
			b.logger.Warn("rabbitmq connect failed, retrying in 5s", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		b.logger.Info("rabbitmq connected", "queue", queueDelivery)
		closeCh := make(chan *amqp.Error, 1)
		b.conn.NotifyClose(closeCh)
		err := <-closeCh // blocks until the connection drops (or is closed cleanly)

		b.setConnected(false)
		if err != nil {
			b.logger.Warn("rabbitmq connection closed, reconnecting", "error", err)
		}
	}
}

func (b *Broker) connectOnce() error {
	host := os.Getenv("RABBITMQ_HOST")
	if host == "" {
		host = "localhost"
	}

	conn, err := amqp.Dial("amqp://guest:guest@" + host + ":5672/")
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return err
	}
	if _, err := ch.QueueDeclare(queueDelivery, true, false, false, false, nil); err != nil {
		_ = conn.Close()
		return err
	}
	if _, err := ch.QueueDeclare(queueDead, true, false, false, false, nil); err != nil {
		_ = conn.Close()
		return err
	}

	b.mu.Lock()
	b.conn, b.pubCh, b.connected = conn, ch, true
	b.mu.Unlock()
	return nil
}

func (b *Broker) setConnected(v bool) {
	b.mu.Lock()
	b.connected = v
	b.mu.Unlock()
}

// Channel opens a fresh channel off the shared connection, for a consumer
// worker's exclusive use (its own Qos + manual ack) - one shared
// connection, one channel per worker.
func (b *Broker) Channel() (*amqp.Channel, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if !b.connected {
		return nil, errors.New("rabbitmq: not connected")
	}
	return b.conn.Channel()
}

// Publish sends body to the default exchange, routed by routingKey (i.e.
// straight to the queue of that name) - matching the payments/tours
// convention of never declaring a named exchange - carrying headers
// (trace context, retry metadata, etc.).
func (b *Broker) Publish(ctx context.Context, routingKey string, headers amqp.Table, body []byte) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if !b.connected {
		return errors.New("rabbitmq: not connected")
	}
	return b.pubCh.PublishWithContext(ctx, "", routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Headers:     headers,
		Body:        body,
	})
}
