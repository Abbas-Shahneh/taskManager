package middleware

import (
	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

const tracingServiceName = "task-manager"

func Tracing() gin.HandlerFunc {
	return otelgin.Middleware(tracingServiceName)
}
