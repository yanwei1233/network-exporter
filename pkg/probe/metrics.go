package probe

import "github.com/prometheus/client_golang/prometheus"

var (
	//是4种探针通用的 2个metric
	M_PROBE_RESULT = prometheus.NewDesc(
		"network_probe_probe_result",
		"network_probe_probe_result",
		[]string{
			"probe_node",
			"probe_ip",
			"run_type",
			"target",
			"func",
			"err_reason",
		},
		nil)
	M_PROBE_DURATION = prometheus.NewDesc(
		"network_probe_duration_seconds",
		"network_probe_probe_result",
		[]string{
			"probe_node",
			"probe_ip",
			"run_type",
			"func",
			"target",
		},
		nil)
)
