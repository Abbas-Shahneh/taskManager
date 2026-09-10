package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/Abbas-Shahneh/taskManager/internal/metrics"
)

func TestMetrics_RecordsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appMetrics := metrics.New()

	registry := prometheus.NewRegistry()

	require.NoError(
		t,
		appMetrics.Register(registry),
	)

	router := gin.New()

	router.Use(Metrics(appMetrics))

	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(
		t,
		http.StatusCreated,
		recorder.Code,
	)

	families, err := registry.Gather()

	require.NoError(t, err)

	require.Len(t, families, 2)
}
