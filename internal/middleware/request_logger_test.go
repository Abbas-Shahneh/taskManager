package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLogger_LogsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buffer bytes.Buffer

	logger := slog.New(
		slog.NewTextHandler(
			&buffer,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		),
	)

	router := gin.New()

	router.Use(RequestID())
	router.Use(RequestLogger(logger))

	router.GET("/tasks/:id", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	const requestID = "test-request-id"

	request := httptest.NewRequest(
		http.MethodGet,
		"/tasks/123",
		nil,
	)

	request.Header.Set(
		RequestIDHeader,
		requestID,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	logOutput := buffer.String()

	assert.Contains(t, logOutput, "http request")
	assert.Contains(t, logOutput, "request_id="+requestID)
	assert.Contains(t, logOutput, "method=GET")
	assert.Contains(t, logOutput, "path=/tasks/123")
	assert.Contains(t, logOutput, "route=/tasks/:id")
	assert.Contains(t, logOutput, "status=200")
	assert.True(t, strings.Contains(logOutput, "duration_ms="))
}
