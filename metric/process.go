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

var middlewareConnect = prometheus.NewDesc(
	"apo_network_middleware_connect",
	"Middleware Connect",
	[]string{
		"pid",
		"container_id",
		"node_name",
		"node_ip",
		"peer_ip",
		"peer_port",
		"peer_type",
		"service_ip",
		"service_port",
	},
	nil,
)

func createMiddlewareMetric(
	pid, container_id, node_name, node_ip,
	peer_ip, peer_port, peer_type,
	service_ip, service_port string,
) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		middlewareConnect, prometheus.GaugeValue, 1,
		pid, container_id, node_name, node_ip,
		peer_ip, peer_port, peer_type,
		service_ip, service_port,
	)
}
