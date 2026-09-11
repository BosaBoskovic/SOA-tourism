package messaging

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	maxDeliveryAttempts = 4
	baseBackoff         = 500 * time.Millisecond
)

// backoffDuration returns how long to wait before attempt+1, doubling
// each time: attempt 1 -> base, 2 -> 2*base, 3 -> 4*base, ...
func backoffDuration(attempt int, base time.Duration) time.Duration {
	return base * time.Duration(int64(1)<<uint(attempt-1))
}

// attemptDelivery runs deliver up to maxAttempts times with exponential
// backoff between failures, giving each attempt its own child span
// (tagged with its attempt number, and the backoff before the next one)
// so Jaeger shows the retry sequence instead of one opaque span. sleepFunc
// is injected so tests can run this without a real sleep.
func attemptDelivery(ctx context.Context, tracer trace.Tracer, deliver func(context.Context) error, maxAttempts int, base time.Duration, sleepFunc func(time.Duration)) (int, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		attemptCtx, attemptSpan := tracer.Start(ctx, "notification.delivery.attempt", trace.WithSpanKind(trace.SpanKindInternal))
		attemptSpan.SetAttributes(attribute.Int("attempt", attempt))

		lastErr = deliver(attemptCtx)
		if lastErr == nil {
			attemptSpan.End()
			return attempt, nil
		}

		attemptSpan.RecordError(lastErr)
		if attempt < maxAttempts {
			wait := backoffDuration(attempt, base)
			attemptSpan.SetAttributes(attribute.Int64("backoff_ms", wait.Milliseconds()))
			attemptSpan.End()
			sleepFunc(wait)
			continue
		}
		attemptSpan.End()
	}
	return maxAttempts, lastErr
}
