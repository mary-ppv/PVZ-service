package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
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

func RegisterMetrics() {
	prometheus.MustRegister(RequestCount)
	prometheus.MustRegister(PVZCreated)
	prometheus.MustRegister(ReceptionCreated)
	prometheus.MustRegister(ProductAdded)
}
