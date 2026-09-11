package messaging

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

// amqpHeaderCarrier adapts amqp.Table (map[string]interface{}) to
// propagation.TextMapCarrier so W3C trace context can travel in AMQP
// headers the same way it already travels in HTTP headers everywhere else
// in this platform (see every service's otel.SetTextMapPropagator call).
type amqpHeaderCarrier amqp.Table

func (c amqpHeaderCarrier) Get(key string) string {
	v, ok := c[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func (c amqpHeaderCarrier) Set(key, value string) {
	c[key] = value
}

func (c amqpHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

// InjectAMQPHeaders writes the trace context carried by ctx into headers
// (creating a new amqp.Table if headers is nil), for use as
// amqp.Publishing.Headers.
func InjectAMQPHeaders(ctx context.Context, headers amqp.Table) amqp.Table {
	if headers == nil {
		headers = amqp.Table{}
	}
	otel.GetTextMapPropagator().Inject(ctx, amqpHeaderCarrier(headers))
	return headers
}

// ExtractAMQPContext returns a context carrying the trace context found in
// an inbound delivery's headers, so a consumer span can be started as a
// child of the producer's span instead of starting a disconnected trace.
func ExtractAMQPContext(ctx context.Context, headers amqp.Table) context.Context {
	if headers == nil {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, amqpHeaderCarrier(headers))
}
