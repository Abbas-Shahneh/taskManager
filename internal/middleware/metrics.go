package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Abbas-Shahneh/taskManager/internal/metrics"
)

func Metrics(m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}

		m.ObserveHTTPRequest(
			c.Request.Method,
			route,
			c.Writer.Status(),
			time.Since(start),
		)
	}
}
