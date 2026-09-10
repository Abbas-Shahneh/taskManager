package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
}

func New() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "task_manager_http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{
				"method",
				"route",
				"status",
			},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "task_manager_http_request_duration_seconds",
				Help: "HTTP request duration in seconds.",
			},
			[]string{
				"method",
				"route",
			},
		),
	}
}

func (m *Metrics) Register(registry prometheus.Registerer) error {
	if err := registry.Register(m.HTTPRequestsTotal); err != nil {
		return err
	}

	if err := registry.Register(m.HTTPRequestDuration); err != nil {
		return err
	}

	return nil
}

func (m *Metrics) ObserveHTTPRequest(
	method string,
	route string,
	status int,
	duration time.Duration,
) {
	m.HTTPRequestsTotal.WithLabelValues(
		method,
		route,
		strconv.Itoa(status),
	).Inc()

	m.HTTPRequestDuration.WithLabelValues(
		method,
		route,
	).Observe(duration.Seconds())
}
