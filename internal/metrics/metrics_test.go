package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestMetrics_RegisterAndObserveHTTPRequest(t *testing.T) {
	metrics := New()

	registry := prometheus.NewRegistry()

	require.NoError(
		t,
		metrics.Register(registry),
	)

	metrics.ObserveHTTPRequest(
		"GET",
		"/tasks/:id",
		200,
		100*time.Millisecond,
	)

	families, err := registry.Gather()

	require.NoError(t, err)

	require.Len(t, families, 2)

	var requestCounterFound bool
	var durationHistogramFound bool

	for _, family := range families {
		switch family.GetName() {
		case "task_manager_http_requests_total":
			requestCounterFound = true
			require.Len(t, family.GetMetric(), 1)

			metric := family.GetMetric()[0]

			require.Equal(
				t,
				float64(1),
				metric.GetCounter().GetValue(),
			)

		case "task_manager_http_request_duration_seconds":
			durationHistogramFound = true
			require.Len(t, family.GetMetric(), 1)

			metric := family.GetMetric()[0]

			require.Equal(
				t,
				uint64(1),
				metric.GetHistogram().GetSampleCount(),
			)
		}
	}

	require.True(t, requestCounterFound)
	require.True(t, durationHistogramFound)
}
