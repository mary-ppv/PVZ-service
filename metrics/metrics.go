package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Initialisation of counters
var (
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	ResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_time_seconds",
			Help:    "HTTP response time in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5},
		},
		[]string{"method", "endpoint"},
	)

	PVZCreated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "pvz_created_total",
			Help: "Total number of created PVZs",
		},
	)

	ReceptionCreated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "reception_created_total",
			Help: "Total number of created receptions",
		},
	)

	ProductAdded = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "product_added_total",
			Help: "Total number of added products",
		},
	)
)

// Register metrics
func RegisterMetrics() {
	prometheus.MustRegister(RequestCount)
	prometheus.MustRegister(ResponseTime)
	prometheus.MustRegister(PVZCreated)
	prometheus.MustRegister(ReceptionCreated)
	prometheus.MustRegister(ProductAdded)
}

// Middleware for collecting technical metrics
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()

		RequestCount.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rw.status)).Inc()

		ResponseTime.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

// Structure for recording the response status
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Handler for providing metrics
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
