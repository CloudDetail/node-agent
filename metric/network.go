package metric

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/utils/lru"
)

var counterLabel *lru.Cache
var packetLossCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "originx_ping_failed_count",
		Help: "Network ping failed count",
	},
	[]string{
		"src_ip",
		"dst_ip",
		"pid",
		"level",
		"src_pod",
		"src_namespace",
		"src_node",
		"dst_pod",
		"dst_namespace",
		"dst_node",
		"node_name",
		"node_ip",
		"container_id",
	},
)

func getOrCreateCounter(src_ip, dst_ip, pid, level,
	src_pod, src_namespace, src_node,
	dst_pod, dst_namespace, dst_node,
	node_name, node_ip, container_id string,
) prometheus.Counter {
	key := fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s",
		src_ip, dst_ip, pid, level,
		src_pod, src_namespace, src_node,
		dst_pod, dst_namespace, dst_node,
		node_name, node_ip, container_id)

	if _, found := counterLabel.Get(key); !found {
		counterLabel.Add(key, struct{}{})
	}

	return packetLossCount.WithLabelValues(
		src_ip, dst_ip, pid, level,
		src_pod, src_namespace, src_node,
		dst_pod, dst_namespace, dst_node,
		node_name, node_ip, container_id,
	)
}

var networkRTT = prometheus.NewDesc(
	"kindling_network_rtt",
	"Network Round-Trip Time (RTT)",
	[]string{
		"src_ip",
		"dst_ip",
		"pid",
		"level",
		"src_pod",
		"src_namespace",
		"src_node",
		"dst_pod",
		"dst_namespace",
		"dst_node",
		"node_name",
		"node_ip",
		"container_id",
	},
	nil,
)

func createRttMetric(
	value float64,
	src_ip, dst_ip, pid, level,
	src_pod, src_namespace, src_node,
	dst_pod, dst_namespace, dst_node,
	node_name, node_ip, container_id string,
) prometheus.Metric {
	return prometheus.MustNewConstMetric(
		networkRTT, prometheus.GaugeValue, value,
		src_ip, dst_ip, pid, level,
		src_pod, src_namespace, src_node,
		dst_pod, dst_namespace, dst_node,
		node_name, node_ip, container_id,
	)
}
