package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type PurchaseRepository struct {
	paymentsURL string
	client      *http.Client
	tokens      map[string]bool
}

func NewPurchaseRepository() *PurchaseRepository {
	paymentsURL := os.Getenv("PAYMENTS_URL")
	if paymentsURL == "" {
		paymentsURL = "http://localhost:8086"
	}

	return &PurchaseRepository{
		paymentsURL: paymentsURL,
		client:      &http.Client{Timeout: 5 * time.Second},
		tokens:      make(map[string]bool),
	}
}

func (r *PurchaseRepository) SaveToken(touristID, tourID string) {
	key := touristID + "_" + tourID
	r.tokens[key] = true
}

func (r *PurchaseRepository) HasToken(touristID, tourID string) (bool, error) {
	key := touristID + "_" + tourID

	// 1. prvo proveri RabbitMQ lokalnu kopiju
	if r.tokens[key] {
		return true, nil
	}

	// 2. fallback na stari HTTP, da aplikacija radi kao pre
	url := fmt.Sprintf("%s/checkout/%s/has-purchased/%s", r.paymentsURL, touristID, tourID)

	resp, err := r.client.Get(url)
	if err != nil {
		return false, nil
	}
	defer resp.Body.Close()

	var result struct {
		HasPurchased bool `json:"hasPurchased"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, nil
	}

	if result.HasPurchased {
		r.SaveToken(touristID, tourID)
	}

	return result.HasPurchased, nil
}