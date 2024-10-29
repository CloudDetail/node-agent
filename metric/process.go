package metric

import "github.com/prometheus/client_golang/prometheus"

var processStartTime = prometheus.NewDesc(
	"originx_process_start_time",
	"Process Start Time (Unix Timestamp)",
	[]string{
		"pid",
		"node_name",
		"node_ip",
		"container_id",
	},
	nil,
)
