package messaging

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// promauto registers against the default Prometheus registry, so these
// show up on the existing GET /metrics endpoint (see
// services/stakeholders/observability.go's metricsHandler) with no extra
// wiring needed.
var (
	deliveryAttemptsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notification_delivery_attempts_total",
		Help: "Terminal outcome of a notification delivery (after all in-process retries), labeled by result.",
	}, []string{"result"})

	deliveryDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "notification_delivery_duration_seconds",
		Help:    "Time from consuming a notification-delivery message to its terminal outcome (spans all retries).",
		Buckets: prometheus.DefBuckets,
	})

	deadLetteredTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "notification_delivery_dead_lettered_total",
		Help: "Notifications that exhausted all delivery attempts and were moved to notification-delivery-dead.",
	})
)
