package agent

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// TestDuration measures the duration of a test.
	TestDuration = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nextcloud_test_duration_seconds",
			Help: "Duration of the Nextcloud performance test in seconds.",
		},
		[]string{"service", "instance", "type"},
	)

	// TestSuccess indicates if a test was successful.
	TestSuccess = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nextcloud_test_success",
			Help: "Indicates if the Nextcloud performance test was successful.",
		},
		[]string{"service", "instance", "type", "error_code"},
	)

	// TestSpeedMbytesPerSec measures the speed of a test in Megabytes per second.
	TestSpeedMbytesPerSec = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nextcloud_test_speed_mbytes_per_sec",
			Help: "Speed of the Nextcloud performance test in MB/s.",
		},
		[]string{"service", "instance", "type"},
	)

	// NEW METRICS FOR ENHANCED DASHBOARD

	// TestErrors counts the total number of test errors.
	TestErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nextcloud_test_errors_total",
			Help: "Total number of Nextcloud performance test errors.",
		},
		[]string{"service", "instance", "type", "error_type"},
	)

	// ChunksUploaded counts the total number of chunks uploaded.
	ChunksUploaded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nextcloud_chunks_uploaded_total",
			Help: "Total number of chunks uploaded.",
		},
		[]string{"service", "instance"},
	)

	// ChunkRetries counts the total number of chunk upload retries.
	ChunkRetries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nextcloud_chunk_retries_total",
			Help: "Total number of chunk upload retries.",
		},
		[]string{"service", "instance", "chunk_number"},
	)

	// ChunkUploadDuration measures the duration of individual chunk uploads.
	ChunkUploadDuration = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nextcloud_chunk_upload_duration_seconds",
			Help: "Duration of individual chunk uploads in seconds.",
		},
		[]string{"service", "instance", "chunk_number"},
	)

	// NetworkLatency measures the network latency to the instance.
	NetworkLatency = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nextcloud_network_latency_ms",
			Help: "Network latency to the instance in milliseconds.",
		},
		[]string{"service", "instance"},
	)

	// ConnectionTimeouts counts the total number of connection timeouts.
	ConnectionTimeouts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nextcloud_connection_timeouts_total",
			Help: "Total number of connection timeouts.",
		},
		[]string{"service", "instance"},
	)

	// CircuitBreakerState indicates the current state of the circuit breaker.
	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nextcloud_circuit_breaker_state",
			Help: "Current state of the circuit breaker (0=closed, 1=open, 2=half-open).",
		},
		[]string{"service", "instance"},
	)

	// TestDurationHistogram provides histogram data for test durations.
	TestDurationHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "nextcloud_test_duration_seconds_bucket",
			Help:    "Histogram of Nextcloud performance test durations.",
			Buckets: prometheus.ExponentialBuckets(1, 2, 10), // 1s, 2s, 4s, 8s, ..., 512s
		},
		[]string{"service", "instance", "type"},
	)

	// ChunkSize tracks the size of uploaded chunks.
	ChunkSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nextcloud_chunk_size_bytes",
			Help: "Size of uploaded chunks in bytes.",
		},
		[]string{"service", "instance"},
	)

	// HiDrive Legacy specific metrics
	hidriveLegacyTestDuration = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hidrive_legacy_test_duration_seconds",
			Help: "Duration of HiDrive Legacy performance test in seconds.",
		},
		[]string{"service", "instance", "type", "status"},
	)

	hidriveLegacyTestSpeed = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hidrive_legacy_test_speed_mbytes_per_sec",
			Help: "Speed of HiDrive Legacy performance test in MB/s.",
		},
		[]string{"service", "instance", "type", "status"},
	)

	// NEW: Historical Performance Averages
	DailyAverageUploadSpeed = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nextcloud_daily_average_upload_speed_mbytes_per_sec",
			Help: "Daily average upload speed in MB/s for the last 24 hours.",
		},
		[]string{"service", "instance"},
	)

	DailyAverageDownloadSpeed = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nextcloud_daily_average_download_speed_mbytes_per_sec",
			Help: "Daily average download speed in MB/s for the last 24 hours.",
		},
		[]string{"service", "instance"},
	)

	MonthlyAverageUploadSpeed = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nextcloud_monthly_average_upload_speed_mbytes_per_sec",
			Help: "Monthly average upload speed in MB/s for the last 30 days.",
		},
		[]string{"service", "instance"},
	)

	MonthlyAverageDownloadSpeed = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nextcloud_monthly_average_download_speed_mbytes_per_sec",
			Help: "Monthly average download speed in MB/s for the last 30 days.",
		},
		[]string{"service", "instance"},
	)

	// Daily and Monthly Test Counts
	DailyTestCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nextcloud_daily_test_count",
			Help: "Number of tests completed in the last 24 hours.",
		},
		[]string{"service", "instance", "type"},
	)

	MonthlyTestCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nextcloud_monthly_test_count",
			Help: "Number of tests completed in the last 30 days.",
		},
		[]string{"service", "instance", "type"},
	)

	// Success Rate Averages
	DailySuccessRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nextcloud_daily_success_rate_percent",
			Help: "Daily success rate percentage for the last 24 hours.",
		},
		[]string{"service", "instance"},
	)

	MonthlySuccessRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nextcloud_monthly_success_rate_percent",
			Help: "Monthly success rate percentage for the last 30 days.",
		},
		[]string{"service", "instance"},
	)
)
