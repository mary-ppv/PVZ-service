package middleware

import (
	"PVZ/pkg/metrics"
	"strconv"

	"github.com/gin-gonic/gin"
)

func PrometheusMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		endpoint := c.FullPath()

		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		metrics.RequestCount.WithLabelValues(method, endpoint, status).Inc()
	}
}
