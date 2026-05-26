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
}

func NewPurchaseRepository() *PurchaseRepository {
	paymentsURL := os.Getenv("PAYMENTS_URL")
	if paymentsURL == "" {
		paymentsURL = "http://localhost:8086"
	}
	return &PurchaseRepository{
		paymentsURL: paymentsURL,
		client:      &http.Client{Timeout: 5 * time.Second},
	}
}

func (r *PurchaseRepository) HasToken(touristID, tourID string) (bool, error) {
	url := fmt.Sprintf("%s/checkout/%s/has-purchased/%s", r.paymentsURL, touristID, tourID)

	resp, err := r.client.Get(url)
	if err != nil {
		return false, nil // ako payments nije dostupan, tretiraj kao nekupljeno
	}
	defer resp.Body.Close()

	var result struct {
		HasPurchased bool `json:"hasPurchased"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, nil
	}

	return result.HasPurchased, nil
}
