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
)
