package messaging

import (
	"context"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TestAMQPHeaders_RoundTripsTraceContext(t *testing.T) {
	// Without an explicit propagator, otel's global default is a no-op -
	// this test sets the same W3C composite propagator main() sets at
	// startup, otherwise Inject/Extract below would silently do nothing.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	traceID, err := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	if err != nil {
		t.Fatalf("invalid test trace id: %v", err)
	}
	spanID, err := trace.SpanIDFromHex("00f067aa0ba902b7")
	if err != nil {
		t.Fatalf("invalid test span id: %v", err)
	}
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	headers := InjectAMQPHeaders(ctx, amqp.Table{})
	if _, ok := headers["traceparent"]; !ok {
		t.Fatal("expected InjectAMQPHeaders to set a traceparent header")
	}

	extracted := ExtractAMQPContext(context.Background(), headers)
	got := trace.SpanContextFromContext(extracted)
	if got.TraceID() != traceID {
		t.Fatalf("trace id did not round-trip: got %s, want %s", got.TraceID(), traceID)
	}
	if got.SpanID() != spanID {
		t.Fatalf("span id did not round-trip: got %s, want %s", got.SpanID(), spanID)
	}
}

func TestExtractAMQPContext_NilHeaders_ReturnsInputContext(t *testing.T) {
	ctx := context.Background()
	got := ExtractAMQPContext(ctx, nil)
	if got != ctx {
		t.Fatal("expected nil headers to return the same context unchanged")
	}
}
