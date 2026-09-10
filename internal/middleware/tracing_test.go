package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestTracing_CreatesSpan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	provider := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(provider)

	router := gin.New()
	router.Use(Tracing())

	router.GET("/health", func(c *gin.Context) {
		span := oteltrace.SpanFromContext(c.Request.Context())

		assert.True(t, span.SpanContext().IsValid())
		assert.True(t, span.SpanContext().IsValid())

		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
}
