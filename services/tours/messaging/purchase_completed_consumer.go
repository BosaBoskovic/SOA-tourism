package messaging

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"
	"tours/repository"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type PurchaseCompletedEvent struct {
	SagaId    string              `json:"SagaId"`
	TouristId string              `json:"TouristId"`
	Items     []PurchasedTourItem `json:"Items"`
}

type PurchasedTourItem struct {
	TourId   string  `json:"TourId"`
	TourName string  `json:"TourName"`
	Price    float64 `json:"Price"`
}

// StartPurchaseCompletedConsumer subscribes to "purchase-completed" and
// keeps retrying with backoff if the connection is ever lost, or never
// comes up in the first place. HasToken() also falls back to an HTTP call
// to payments, so this consumer is a warm-cache refresh, not a hard
// dependency - it just shouldn't silently stop forever after one outage.
func StartPurchaseCompletedConsumer(purchaseRepo *repository.PurchaseRepository) {
	go func() {
		for {
			if err := consumePurchaseCompleted(purchaseRepo); err != nil {
				log.Printf("purchase-completed consumer stopped, retrying in 5s: %v", err)
			}
			time.Sleep(5 * time.Second)
		}
	}()
}

func consumePurchaseCompleted(purchaseRepo *repository.PurchaseRepository) error {
	host := os.Getenv("RABBITMQ_HOST")
	if host == "" {
		host = "localhost"
	}

	conn, err := amqp.Dial("amqp://guest:guest@" + host + ":5672/")
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("purchase-completed", true, false, false, false, nil)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	closed := ch.NotifyClose(make(chan *amqp.Error, 1))

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return nil // delivery channel closed - let the caller reconnect
			}
			handlePurchaseCompleted(purchaseRepo, msg)
		case err := <-closed:
			if err != nil {
				return err
			}
			return nil
		}
	}
}

// handlePurchaseCompleted processes one delivery inside its own span
// (needed because a bare defer inside the for/select loop above would only
// fire once, at goroutine exit, not per message). Extracting the trace
// context from the delivery's headers means this span joins the same
// trace as the checkout request in payments that published it, instead of
// starting a disconnected one - see RabbitMqPublisher.InjectTraceContext
// on the publishing side.
func handlePurchaseCompleted(purchaseRepo *repository.PurchaseRepository, msg amqp.Delivery) {
	ctx := ExtractAMQPContext(context.Background(), msg.Headers)
	tracer := otel.Tracer("tours-messaging")
	_, span := tracer.Start(ctx, "purchase-completed process", trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()
	span.SetAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.destination", "purchase-completed"),
	)

	var event PurchaseCompletedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("purchase-completed: skipping malformed message: %v", err)
		span.RecordError(err)
		return
	}
	for _, item := range event.Items {
		purchaseRepo.SaveToken(event.TouristId, item.TourId)
	}
	log.Println("Purchase completed event saved in tours service")
}
