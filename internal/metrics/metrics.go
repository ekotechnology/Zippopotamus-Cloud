package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	ns  string = "zippopotamus"
	sub string = "api"
)

var (
	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "request_duration",
		Help:      "Duration of all requests from when it is received by the middleware stack until its returned upstream",
		Buckets:   []float64{1, 2, 3, 5, 8, 13, 21, 34, 50, 75, 100},
	}, []string{"path", "code", "apiversion", "method"})

	PayloadBytes = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "response_payload_bytes",
		Help:      "Bytes returned upstream as a result of API requests",
		Buckets:   []float64{150, 250, 300, 500, 800, 1100, 1500, 2000, 2250, 2500, 3000},
	}, []string{"path", "code", "apiversion", "method"})
)
