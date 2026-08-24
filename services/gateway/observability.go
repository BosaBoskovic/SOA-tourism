package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc/connectivity"
)

const serviceName = "gateway"

// newLogger returns a JSON structured logger tagged with the service name.
// Consumed by Loki/Promtail (service + level become indexed labels) and,
// later, by a dedicated monitoring app reading the same log stream.
func newLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(handler).With("service", serviceName)
}

// initTracer wires an OpenTelemetry TracerProvider that exports spans to
// Jaeger via OTLP/HTTP. Every service in the platform calls an equivalent
// function so a single request can be followed end-to-end in Jaeger.
func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	name := os.Getenv("OTEL_SERVICE_NAME")
	if name == "" {
		name = serviceName
	}

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4318"
	}

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint+"/v1/traces"))
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(name)))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// Without an explicit propagator, otel's default is a no-op: trace
	// context would never actually cross the wire to other services. W3C
	// tracecontext is what every service in this platform standardizes on.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// --- Prometheus metrics -----------------------------------------------

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests handled, labeled by route and status code.",
	}, []string{"service", "method", "route", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "method", "route"})

	httpRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Number of HTTP requests currently being served.",
		ConstLabels: prometheus.Labels{
			"service": serviceName,
		},
	})
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// observabilityMiddleware records RED metrics (rate/errors/duration) and
// emits one structured access-log line per request, including the trace ID
// of the active OpenTelemetry span so logs and traces can be correlated.
func observabilityMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		route := r.URL.Path

		httpRequestsTotal.WithLabelValues(serviceName, r.Method, route, strconv.Itoa(rec.status)).Inc()
		httpRequestDuration.WithLabelValues(serviceName, r.Method, route).Observe(duration.Seconds())

		attrs := []any{
			"method", r.Method,
			"path", route,
			"status", rec.status,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", r.RemoteAddr,
		}
		if span := trace.SpanContextFromContext(r.Context()); span.IsValid() {
			attrs = append(attrs, "trace_id", span.TraceID().String(), "span_id", span.SpanID().String())
		}
		logger.Info("http_request", attrs...)
	})
}

// --- Health -------------------------------------------------------------

type healthCheck struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// healthHandler reports gateway liveness plus the connectivity state of the
// gRPC channels it holds open to the downstream services, so a monitoring
// app can tell "gateway is up" apart from "gateway is up but blind".
func healthHandler(conns map[string]interface {
	GetState() connectivity.State
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		overall := "ok"
		checks := make(map[string]healthCheck, len(conns))
		for name, conn := range conns {
			state := conn.GetState()
			c := healthCheck{Status: state.String()}
			if state != connectivity.Ready && state != connectivity.Idle {
				overall = "degraded"
				c.Error = "grpc channel not ready"
			}
			checks[name] = c
		}

		status := http.StatusOK
		if overall != "ok" {
			status = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"service": serviceName,
			"status":  overall,
			"checks":  checks,
		})
	}
}

func metricsHandler() http.Handler {
	return promhttp.Handler()
}
