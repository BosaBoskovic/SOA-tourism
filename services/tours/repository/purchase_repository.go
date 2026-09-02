package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type PurchaseRepository struct {
	paymentsURL string
	client      *http.Client
	mu          sync.RWMutex
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

// SaveToken and HasToken are both called concurrently - SaveToken from the
// RabbitMQ consumer goroutine, HasToken from HTTP request goroutines. Plain
// map access under concurrent read/write is a fatal, unrecoverable crash in
// Go, not just a data race.
func (r *PurchaseRepository) SaveToken(touristID, tourID string) {
	key := touristID + "_" + tourID
	r.mu.Lock()
	r.tokens[key] = true
	r.mu.Unlock()
}

func (r *PurchaseRepository) HasToken(touristID, tourID string) (bool, error) {
	key := touristID + "_" + tourID

	// 1. prvo proveri RabbitMQ lokalnu kopiju
	r.mu.RLock()
	cached := r.tokens[key]
	r.mu.RUnlock()
	if cached {
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