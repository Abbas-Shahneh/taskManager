package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID_GeneratesIDWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID())

	router.GET("/test", func(c *gin.Context) {
		value, exists := c.Get(RequestIDHeader)

		require.True(t, exists)

		requestID, ok := value.(string)
		require.True(t, ok)

		assert.NotEmpty(t, requestID)

		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	responseRequestID := recorder.Header().Get(RequestIDHeader)

	assert.NotEmpty(t, responseRequestID)
}

func TestRequestID_PreservesProvidedID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const requestID = "test-request-id-123"

	router := gin.New()
	router.Use(RequestID())

	router.GET("/test", func(c *gin.Context) {
		value, exists := c.Get(RequestIDHeader)

		require.True(t, exists)
		assert.Equal(t, requestID, value)

		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	request.Header.Set(
		RequestIDHeader,
		requestID,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	assert.Equal(
		t,
		requestID,
		recorder.Header().Get(RequestIDHeader),
	)
}

func TestRequestID_RejectsWhitespaceOnlyID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID())

	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	request.Header.Set(RequestIDHeader, "   ")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	responseRequestID := recorder.Header().Get(RequestIDHeader)

	assert.NotEmpty(t, responseRequestID)
	assert.NotEqual(t, "   ", responseRequestID)
}
