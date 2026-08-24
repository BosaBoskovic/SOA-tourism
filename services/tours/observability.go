package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"go.opentelemetry.io/otel/trace"
)

// newLogger returns a JSON structured logger tagged with the service name.
func newLogger(serviceName string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(handler).With("service", serviceName)
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

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// observabilityMiddleware is a gorilla/mux middleware that records RED
// metrics and emits one structured access-log line per request, tagged
// with the active trace ID. It must be registered via r.Use(...) *after*
// mux has resolved the route so mux.CurrentRoute gives a low-cardinality
// path template (e.g. "/tours/{id}") instead of the raw, high-cardinality URL.
func observabilityMiddleware(serviceName string, logger *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			duration := time.Since(start)

			route := r.URL.Path
			if rt := mux.CurrentRoute(r); rt != nil {
				if tmpl, err := rt.GetPathTemplate(); err == nil {
					route = tmpl
				}
			}

			httpRequestsTotal.WithLabelValues(serviceName, r.Method, route, strconv.Itoa(rec.status)).Inc()
			httpRequestDuration.WithLabelValues(serviceName, r.Method, route).Observe(duration.Seconds())

			attrs := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"route", route,
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
}

func metricsHandler() http.Handler {
	return promhttp.Handler()
}

// --- Health -------------------------------------------------------------

// healthHandler reports liveness plus MongoDB connectivity.
func healthHandler(serviceName string, client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		status := "ok"
		mongoStatus := "ok"
		httpStatus := http.StatusOK
		if err := client.Ping(ctx, nil); err != nil {
			status = "degraded"
			mongoStatus = "error: " + err.Error()
			httpStatus = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"service": serviceName,
			"status":  status,
			"checks": map[string]string{
				"mongodb": mongoStatus,
			},
		})
	}
}
