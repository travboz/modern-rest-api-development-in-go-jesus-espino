package metrics

import (
	"net/http"
	"time"
)

func MetricsMiddleware(ms *Metrics, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start).Seconds()

		ms.RequestDuration.WithLabelValues(
			r.Method, r.URL.Path,
		).Observe(duration)

		ms.RequestCounter.WithLabelValues(
			r.Method, r.URL.Path,
		).Inc()
	})
}
