package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const serviceName = "encounters"

// newLogger returns a JSON structured logger tagged with the service name.
func newLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(handler).With("service", serviceName)
}

// initTracer wires an OpenTelemetry TracerProvider that exports spans to
// Jaeger via OTLP/HTTP, matching the pattern used across every service.
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
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithResource(res))
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

func tracingMiddleware() gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
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
)

func observabilityMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}
		status := c.Writer.Status()

		httpRequestsTotal.WithLabelValues(serviceName, c.Request.Method, route, strconv.Itoa(status)).Inc()
		httpRequestDuration.WithLabelValues(serviceName, c.Request.Method, route).Observe(duration.Seconds())

		attrs := []any{
			"method", c.Request.Method,
			"path", route,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", c.ClientIP(),
		}
		if span := trace.SpanContextFromContext(c.Request.Context()); span.IsValid() {
			attrs = append(attrs, "trace_id", span.TraceID().String(), "span_id", span.SpanID().String())
		}
		logger.Info("http_request", attrs...)
	}
}

func metricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// healthHandler reports liveness plus MongoDB connectivity.
func healthHandler(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		status := "ok"
		mongoStatus := "ok"
		httpStatus := http.StatusOK
		if err := client.Ping(ctx, nil); err != nil {
			status = "degraded"
			mongoStatus = "error: " + err.Error()
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, gin.H{
			"service": serviceName,
			"status":  status,
			"checks": gin.H{
				"mongodb": mongoStatus,
			},
		})
	}
}
