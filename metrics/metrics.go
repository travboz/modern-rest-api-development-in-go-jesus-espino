package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	RequestCounter  *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	metrics := Metrics{}

	metrics.RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of hTTP requests received",
		},
		[]string{"method", "path"},
	)

	metrics.RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Histogram of response durations",
		},
		[]string{"method", "path"},
	)

	prometheus.MustRegister(metrics.RequestCounter)
	prometheus.MustRegister(metrics.RequestDuration)

	return &metrics
}
