package messaging

import (
	"encoding/json"
	"log"
	"os"
	"time"
	"tours/repository"

	amqp "github.com/rabbitmq/amqp091-go"
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
			var event PurchaseCompletedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("purchase-completed: skipping malformed message: %v", err)
				continue
			}
			for _, item := range event.Items {
				purchaseRepo.SaveToken(event.TouristId, item.TourId)
			}
			log.Println("Purchase completed event saved in tours service")
		case err := <-closed:
			if err != nil {
				return err
			}
			return nil
		}
	}
}
