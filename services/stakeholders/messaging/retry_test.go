package messaging

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
)

func TestAttemptDelivery_SucceedsAfterTransientFailures(t *testing.T) {
	tracer := otel.Tracer("test")
	calls := 0
	deliver := func(ctx context.Context) error {
		calls++
		if calls < 3 {
			return errors.New("transient failure")
		}
		return nil
	}

	var slept []time.Duration
	sleepFunc := func(d time.Duration) { slept = append(slept, d) }

	attempts, err := attemptDelivery(context.Background(), tracer, deliver, 4, 500*time.Millisecond, sleepFunc)
	if err != nil {
		t.Fatalf("expected eventual success, got error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	if calls != 3 {
		t.Fatalf("expected deliver to be called 3 times, got %d", calls)
	}
	if len(slept) != 2 {
		t.Fatalf("expected 2 backoff sleeps (after attempts 1 and 2), got %d", len(slept))
	}
	if slept[0] != 500*time.Millisecond || slept[1] != time.Second {
		t.Fatalf("expected backoff to double (500ms, 1s), got %v", slept)
	}
}

func TestAttemptDelivery_ReturnsLastErrorAfterExhaustingAttempts(t *testing.T) {
	tracer := otel.Tracer("test")
	wantErr := errors.New("always fails")
	calls := 0
	deliver := func(ctx context.Context) error {
		calls++
		return wantErr
	}

	var slept []time.Duration
	sleepFunc := func(d time.Duration) { slept = append(slept, d) }

	attempts, err := attemptDelivery(context.Background(), tracer, deliver, 4, 500*time.Millisecond, sleepFunc)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected the last error to be returned, got %v", err)
	}
	if attempts != 4 {
		t.Fatalf("expected exactly maxAttempts (4) attempts, got %d", attempts)
	}
	if calls != 4 {
		t.Fatalf("expected deliver to be called 4 times, got %d", calls)
	}
	if len(slept) != 3 {
		t.Fatalf("expected 3 backoff sleeps (no sleep after the final attempt), got %d", len(slept))
	}
}

func TestBackoffDuration_DoublesEachAttempt(t *testing.T) {
	base := 500 * time.Millisecond
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 500 * time.Millisecond},
		{2, time.Second},
		{3, 2 * time.Second},
	}
	for _, c := range cases {
		if got := backoffDuration(c.attempt, base); got != c.want {
			t.Errorf("backoffDuration(%d, base) = %v, want %v", c.attempt, got, c.want)
		}
	}
}
