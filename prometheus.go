package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/CloudDetail/node-agent/netanaly"
	"github.com/CloudDetail/node-agent/proc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/utils/lru"
)

var (
	counterLabel    *lru.Cache
	packetLossCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "originx_packet_loss_count",
			Help: "Network packet loss count",
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
	networkRTT = prometheus.NewDesc(
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
	processStartTime = prometheus.NewDesc(
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
)

var processTime = false
var nodeName = ""
var nodeIp = ""
var cacheSize = 50000

func init() {
	if os.Getenv("PROCESS_TIME") == "true" {
		processTime = true
	}
	nodeName = os.Getenv("MY_NODE_NAME")
	nodeIp = os.Getenv("MY_NODE_IP")
	envVar := os.Getenv("LRU_CACHE_SIZE")
	if value, err := strconv.Atoi(envVar); err == nil {
		cacheSize = value
	}

	counterLabel = lru.NewWithEvictionFunc(cacheSize, func(key lru.Key, value interface{}) {
		labelstr := key.(string)
		labels := strings.Split(labelstr, ":")
		packetLossCount.DeleteLabelValues(labels...)
	})

	rc := &RttCollector{}
	prometheus.MustRegister(rc)
	prometheus.MustRegister(packetLossCount)
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Println("Prometheus metrics server started on :9408")
		http.ListenAndServe(":9408", nil)
	}()
}

type RttCollector struct {
}

// Describe implements prometheus.Collector.
func (rc *RttCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- networkRTT
}

func (rc *RttCollector) Collect(ch chan<- prometheus.Metric) {
	pid2cid := make(map[uint32]string)
	if processTime {
		proc.GlobalPidMutex.RLock()
		for pid, processInfo := range proc.GlobalNeedMonitorPid {
			ch <- prometheus.MustNewConstMetric(
				processStartTime, prometheus.GaugeValue,
				float64(processInfo.StartTime.Unix()),
				strconv.FormatUint(uint64(pid), 10),
				nodeName,
				nodeIp,
				processInfo.ContainId,
			)
			pid2cid[pid] = processInfo.ContainId
		}
		proc.GlobalPidMutex.RUnlock()
	}

	netanaly.RttResultMapMutex.Lock()
	for tuple, statistic := range netanaly.GlobalRttResultMap {
		for _, n := range statistic.Pids {
			pid := strconv.Itoa(int(n))
			containerId := pid2cid[n]
			rtt := statistic.SumLatency / float64(statistic.Count)
			srcPod := ""
			srcNamespace := ""
			srcNode := ""
			if pod, ok := netanaly.GetPodByIp(tuple.SrcIp); ok {
				srcPod = pod.Name
				srcNamespace = pod.NS()
				srcNode = pod.NodeName()
			} else if node, ok := netanaly.GetNodeByIP(tuple.SrcIp); ok {
				srcNode = node.Name
			}
			if tuple.ServiceIp != "" {
				if math.Abs(rtt-1.0) < 1e-9 {
					getOrCreateCounter(
						tuple.SrcIp, tuple.ServiceIp, pid, "service",
						srcPod, srcNamespace, srcNode,
						"", "", "",
						nodeName, nodeIp, containerId,
					).Inc()
					continue
				}
				ch <- createRttMetric(rtt,
					tuple.SrcIp, tuple.ServiceIp, pid, "service",
					srcPod, srcNamespace, srcNode,
					"", "", "",
					nodeName, nodeIp, containerId,
				)
			} else {
				dstPod := ""
				dstNamespace := ""
				dstNode := ""
				if pod, ok := netanaly.GetPodByIp(tuple.DstIp); ok {
					dstPod = pod.Name
					dstNamespace = pod.NS()
					dstNode = pod.NodeName()
				} else if node, ok := netanaly.GetNodeByIP(tuple.DstIp); ok {
					dstNode = node.Name
				}
				if math.Abs(rtt-1.0) < 1e-9 {
					getOrCreateCounter(
						tuple.SrcIp, tuple.DstIp, pid, "instance",
						srcPod, srcNamespace, srcNode,
						dstPod, dstNamespace, dstNode,
						nodeName, nodeIp, containerId,
					).Inc()
					continue
				}
				ch <- createRttMetric(rtt,
					tuple.SrcIp, tuple.DstIp, pid, "instance",
					srcPod, srcNamespace, srcNode,
					dstPod, dstNamespace, dstNode,
					nodeName, nodeIp, containerId,
				)
			}
		}
		// 删除元素
		delete(netanaly.GlobalRttResultMap, tuple)
	}
	// 解锁 RttResultMapMutex
	netanaly.RttResultMapMutex.Unlock()
}

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
