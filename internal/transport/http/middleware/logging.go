package middleware

import (
	"PVZ/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		c.Next()

		status := c.Writer.Status()
		duration := time.Since(start)
		logger.Log.Printf("[HTTP] %s %s -> %d (%s)", method, path, status, duration)
	}
}
