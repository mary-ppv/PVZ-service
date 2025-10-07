package controllers

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler godoc
// @Summary Метрики приложения
// @Description Prometheus метрики приложения
// @Tags Metrics
// @Produce text/plain
// @Success 200 {string} string "Prometheus metrics"
// @Router /metrics [get]
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
