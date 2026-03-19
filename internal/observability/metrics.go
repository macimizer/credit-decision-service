package observability

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	httpRequestsTotal   *prometheus.CounterVec
	httpDurationSeconds *prometheus.HistogramVec
	creditDecisions     *prometheus.CounterVec
	creditDuration      prometheus.Histogram
}

var (
	defaultMetrics *Metrics
	metricsOnce    sync.Once
)

func NewMetrics() *Metrics {
	metricsOnce.Do(func() {
		defaultMetrics = &Metrics{
			httpRequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests handled by the service.",
			}, []string{"method", "route", "status"}),
			httpDurationSeconds: promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Latency distribution for HTTP requests.",
				Buckets: prometheus.DefBuckets,
			}, []string{"method", "route"}),
			creditDecisions: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "credit_decisions_total",
				Help: "Total credit decisions emitted by the service.",
			}, []string{"status", "credit_type"}),
			creditDuration: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "credit_decision_duration_seconds",
				Help:    "Latency distribution for credit creation and decisioning.",
				Buckets: prometheus.DefBuckets,
			}),
		}
	})

	return defaultMetrics
}

func (m *Metrics) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)

		routePattern := "unknown"
		if routeContext := chi.RouteContext(r.Context()); routeContext != nil {
			if currentPattern := routeContext.RoutePattern(); currentPattern != "" {
				routePattern = currentPattern
			}
		}

		m.httpRequestsTotal.WithLabelValues(r.Method, routePattern, strconv.Itoa(ww.status)).Inc()
		m.httpDurationSeconds.WithLabelValues(r.Method, routePattern).Observe(time.Since(start).Seconds())
	})
}

func (m *Metrics) ObserveCreditDecision(duration time.Duration, status string, creditType string) {
	m.creditDuration.Observe(duration.Seconds())
	m.creditDecisions.WithLabelValues(status, creditType).Inc()
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
