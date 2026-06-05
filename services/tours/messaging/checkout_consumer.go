package messaging

import (
	"encoding/json"
	"log"
	"os"
	"tours/repository"

	amqp "github.com/rabbitmq/amqp091-go"
)

type CheckoutRequestedEvent struct {
	SagaId    string             `json:"SagaId"`
	TouristId string             `json:"TouristId"`
	Items     []CheckoutTourItem `json:"Items"`
}

type CheckoutTourItem struct {
	TourId   string  `json:"TourId"`
	TourName string  `json:"TourName"`
	Price    float64 `json:"Price"`
}

type CheckoutApprovedEvent struct {
	SagaId    string `json:"SagaId"`
	TouristId string `json:"TouristId"`
}

type CheckoutRejectedEvent struct {
	SagaId string `json:"SagaId"`
	Reason string `json:"Reason"`
}

func StartCheckoutConsumer(tourRepo *repository.TourRepository) {
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

	q, _ := ch.QueueDeclare("checkout-requested", true, false, false, false, nil)

	msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

	go func() {
		for msg := range msgs {
			var event CheckoutRequestedEvent
			json.Unmarshal(msg.Body, &event)

			approved := true
			reason := ""

			for _, item := range event.Items {
				tour, err := tourRepo.FindByStringID(item.TourId)

				if err != nil || tour.Status != "published" {
					approved = false
					reason = "Tura nije dostupna za kupovinu: " + item.TourName
					break
				}
			}

			if approved {
				publish(ch, "checkout-approved", CheckoutApprovedEvent{
					SagaId:    event.SagaId,
					TouristId: event.TouristId,
				})
			} else {
				publish(ch, "checkout-rejected", CheckoutRejectedEvent{
					SagaId: event.SagaId,
					Reason: reason,
				})
			}
		}
	}()
}

func publish(ch *amqp.Channel, queue string, body any) {
	ch.QueueDeclare(queue, true, false, false, false, nil)

	jsonBody, _ := json.Marshal(body)

	err := ch.Publish(
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonBody,
		},
	)

	if err != nil {
		log.Println("RabbitMQ publish error:", err)
	}
}