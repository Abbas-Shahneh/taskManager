package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		requestID, _ := c.Get(RequestIDHeader)

		spanContext := trace.SpanContextFromContext(c.Request.Context())

		logger.Info(
			"http request",
			"request_id", requestID,
			"trace_id", spanContext.TraceID().String(),
			"span_id", spanContext.SpanID().String(),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"route", c.FullPath(),
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
		)
	}
}
