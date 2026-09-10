package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		requestID, _ := c.Get(RequestIDHeader)

		logger.Info(
			"http request",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"route", c.FullPath(),
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
		)
	}
}
