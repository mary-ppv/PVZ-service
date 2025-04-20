package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

// Helper function to get the value of a counter
func getCounterValue(counter prometheus.Counter) float64 {
	var metric dto.Metric
	counter.Write(&metric)
	return metric.GetCounter().GetValue()
}

// Helper function to get the sample count in a histogram
func getHistogramCount(histogram *prometheus.HistogramVec, labels ...string) float64 {
	hist, err := histogram.GetMetricWithLabelValues(labels...)
	if err != nil {
		panic(err)
	}

	var metric dto.Metric
	hist.(prometheus.Metric).Write(&metric) // Use the Write method of the metric
	return float64(metric.GetHistogram().GetSampleCount())
}

// Testing counters
func TestCounters(t *testing.T) {
	// Register metrics
	RegisterMetrics()

	// Increment counters
	PVZCreated.Inc()
	ReceptionCreated.Inc()
	ProductAdded.Inc()

	// Check counter values
	assert.Equal(t, 1.0, getCounterValue(PVZCreated), "Expected PVZCreated counter to increment")
	assert.Equal(t, 1.0, getCounterValue(ReceptionCreated), "Expected ReceptionCreated counter to increment")
	assert.Equal(t, 1.0, getCounterValue(ProductAdded), "Expected ProductAdded counter to increment")
}

// Testing Middleware for collecting request metrics
func TestMetricsMiddleware(t *testing.T) {
	// Register metrics
	RegisterMetrics()

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap the handler with MetricsMiddleware
	middleware := MetricsMiddleware(handler)

	// Create a test request
	req := httptest.NewRequest("GET", "/test-endpoint", nil)
	rec := httptest.NewRecorder()

	// Execute the request
	middleware.ServeHTTP(rec, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200")

	// Check the RequestCount counter
	counterValue := getCounterValue(RequestCount.WithLabelValues("GET", "/test-endpoint", "200"))
	assert.Equal(t, 1.0, counterValue, "Expected RequestCount counter to increment")

	// Check the ResponseTime histogram
	histogramCount := getHistogramCount(ResponseTime, "GET", "/test-endpoint")
	assert.GreaterOrEqual(t, histogramCount, 1.0, "Expected ResponseTime histogram to record response time")
}
