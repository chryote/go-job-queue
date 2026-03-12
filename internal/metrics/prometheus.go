package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	JobsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "job_queue_processed_total",
		Help: "Total processed jobs labeled by type and status",
	}, []string{"type", "status"})
)

var (
	// Histogram tracks the duration of jobs in seconds
	JobDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "job_duration_seconds",
		Help:    "Time taken to process a job",
		Buckets: prometheus.DefBuckets, // Default buckets: .005s to 10s
	}, []string{"type"})
)
