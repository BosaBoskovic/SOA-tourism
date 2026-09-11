package messaging

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"stakeholders/model"
	"stakeholders/repo"
)

const numDeliveryWorkers = 3

// StartDeliveryConsumer launches a small worker pool consuming
// notification-delivery, each with its own channel off broker's shared
// connection (Qos(1) + manual ack each) - one slow/retrying message only
// blocks its own worker, not the other two.
func StartDeliveryConsumer(broker *Broker, notificationRepo *repo.NotificationRepo, logger *slog.Logger) {
	for i := 0; i < numDeliveryWorkers; i++ {
		go runDeliveryWorker(broker, notificationRepo, logger, i)
	}
}

func runDeliveryWorker(broker *Broker, notificationRepo *repo.NotificationRepo, logger *slog.Logger, workerID int) {
	for {
		ch, err := broker.Channel()
		if err != nil {
			logger.Warn("delivery worker waiting for rabbitmq", "worker", workerID, "error", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if err := ch.Qos(1, 0, false); err != nil {
			logger.Warn("delivery worker: qos failed, retrying", "worker", workerID, "error", err)
			_ = ch.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		msgs, err := ch.Consume(queueDelivery, "", false, false, false, false, nil)
		if err != nil {
			logger.Warn("delivery worker: consume failed, retrying", "worker", workerID, "error", err)
			_ = ch.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		closed := ch.NotifyClose(make(chan *amqp.Error, 1))
	consuming:
		for {
			select {
			case d, ok := <-msgs:
				if !ok {
					break consuming
				}
				handleDelivery(broker, notificationRepo, logger, d)
			case <-closed:
				break consuming
			}
		}
		_ = ch.Close()
		time.Sleep(2 * time.Second) // avoid a tight reconnect spin if rabbitmq is flapping
	}
}

// handleDelivery processes one message inside its own consumer span,
// running the bounded retry/backoff loop and ending in either a delivered
// notification or a dead-lettered one - see attemptDelivery for the retry
// mechanics and simulateDelivery for what "delivering" means here.
func handleDelivery(broker *Broker, notificationRepo *repo.NotificationRepo, logger *slog.Logger, d amqp.Delivery) {
	ctx := ExtractAMQPContext(context.Background(), d.Headers)
	tracer := otel.Tracer("stakeholders-messaging")
	ctx, span := tracer.Start(ctx, "notification.delivery", trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()
	span.SetAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.destination", queueDelivery),
	)

	var msg DeliveryMessage
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		logger.Error("notification-delivery: malformed message, dropping", "error", err)
		span.RecordError(err)
		_ = d.Nack(false, false) // never requeue an unparseable message
		return
	}
	span.SetAttributes(
		attribute.String("notification.id", msg.ID),
		attribute.String("notification.type", msg.Type),
	)

	start := time.Now()
	attempts, deliverErr := attemptDelivery(ctx, tracer, simulateDelivery, maxDeliveryAttempts, baseBackoff, time.Sleep)
	deliveryDuration.Observe(time.Since(start).Seconds())
	traceID := span.SpanContext().TraceID().String()

	if deliverErr == nil {
		deliveryAttemptsTotal.WithLabelValues("success").Inc()
		now := time.Now()
		if err := notificationRepo.UpdateDeliveryStatus(ctx, msg.ID, model.DeliveryStatusDelivered, attempts, &now); err != nil {
			logger.Error("failed to persist delivered status", "id", msg.ID, "error", err)
		}
		logger.Info("notification_delivered", "id", msg.ID, "attempts", attempts, "trace_id", traceID)
		_ = d.Ack(false)
		return
	}

	// Exhausted retries: record the failure, move the message to the
	// dead-letter queue, and don't requeue the original.
	deliveryAttemptsTotal.WithLabelValues("failure").Inc()
	deadLetteredTotal.Inc()
	span.RecordError(deliverErr)

	if err := notificationRepo.UpdateDeliveryStatus(ctx, msg.ID, model.DeliveryStatusFailed, attempts, nil); err != nil {
		logger.Error("failed to persist failed status", "id", msg.ID, "error", err)
	}

	deadHeaders := InjectAMQPHeaders(ctx, amqp.Table{"x-delivery-attempts": attempts})
	if err := broker.Publish(ctx, queueDead, deadHeaders, d.Body); err != nil {
		logger.Error("failed to publish to dead-letter queue", "id", msg.ID, "error", err)
	}

	logger.Warn("notification_dead_lettered", "id", msg.ID, "attempts", attempts, "error", deliverErr.Error(), "trace_id", traceID)
	_ = d.Nack(false, false)
}
