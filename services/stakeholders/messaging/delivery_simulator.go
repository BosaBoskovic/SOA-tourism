package messaging

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// simulateDelivery stands in for actually calling an email/push provider -
// this pipeline's purpose is to give the existing tracing/logging/metrics
// stack a real retry/backoff/dead-letter flow to show, not to integrate a
// real provider. DELIVERY_FAILURE_RATE / DELIVERY_SIMULATED_LATENCY_MS let
// an operator dial the failure rate up or down (e.g. =1 to force every
// message all the way to the dead-letter queue, for demoing that path).
func simulateDelivery(ctx context.Context) error {
	select {
	case <-time.After(envLatency("DELIVERY_SIMULATED_LATENCY_MS", 300*time.Millisecond)):
	case <-ctx.Done():
		return ctx.Err()
	}
	return decideOutcome(envFailureRate("DELIVERY_FAILURE_RATE", 0.4), rand.Float64)
}

// decideOutcome is the pure decision at the core of simulateDelivery,
// pulled out so tests can assert rate=0 never fails and rate=1 always
// fails, without a real sleep or depending on which random draw happens
// to come up.
func decideOutcome(rate float64, randFloat func() float64) error {
	if randFloat() < rate {
		return errors.New("simulated delivery failure")
	}
	return nil
}

func envFailureRate(key string, fallback float64) float64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < 0 || v > 1 {
		return fallback
	}
	return v
}

func envLatency(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	ms, err := strconv.Atoi(raw)
	if err != nil || ms < 0 {
		return fallback
	}
	return time.Duration(ms) * time.Millisecond
}
