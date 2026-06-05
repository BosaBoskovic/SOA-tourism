package messaging

import (
	"encoding/json"
	"log"
	"os"
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

func StartPurchaseCompletedConsumer(purchaseRepo *repository.PurchaseRepository) {
	host := os.Getenv("RABBITMQ_HOST")
	if host == "" {
		host = "localhost"
	}

	conn, err := amqp.Dial("amqp://guest:guest@" + host + ":5672/")
	if err != nil {
		log.Println("RabbitMQ connection error:", err)
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Println(err)
		return
	}

	q, _ := ch.QueueDeclare("purchase-completed", true, false, false, false, nil)

	msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

	go func() {
		for msg := range msgs {
			var event PurchaseCompletedEvent
			json.Unmarshal(msg.Body, &event)

			for _, item := range event.Items {
				purchaseRepo.SaveToken(event.TouristId, item.TourId)
			}

			log.Println("Purchase completed event saved in tours service")
		}
	}()
}