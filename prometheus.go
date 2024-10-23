package main

import (
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/CloudDetail/node-agent/netanaly"
	"github.com/CloudDetail/node-agent/proc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	networkPacketLossRTT = prometheus.NewDesc(
		"kindling_network_packet_loss_rtt",
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
		"Process Start Time (NS)",
		[]string{
			"pid",
			"node_name",
			"node_ip",
			"container_id",
		},
		nil,
	)
)

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

var processTime = false
var nodeName = ""
var nodeIp = ""

func init() {
	if os.Getenv("PROCESS_TIME") == "true" {
		processTime = true
	}
	nodeName = os.Getenv("MY_NODE_NAME")
	nodeIp = os.Getenv("MY_NODE_IP")

	rc := &RttCollector{}
	prometheus.MustRegister(rc)
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Println("Prometheus metrics server started on :9408")
		http.ListenAndServe(":9408", nil)
	}()
}

func createRttMetric(
	value float64,
	src_ip, dst_ip, pid, level,
	src_pod, src_namespace, src_node,
	dst_pod, dst_namespace, dst_node,
	node_name, node_ip, container_id string,
) prometheus.Metric {
	if math.Abs(value-1.0) < 1e-9 {
		return prometheus.MustNewConstMetric(
			networkPacketLossRTT, prometheus.GaugeValue, value,
			src_ip, dst_ip, pid, level,
			src_pod, src_namespace, src_node,
			dst_pod, dst_namespace, dst_node,
			node_name, node_ip, container_id,
		)
	} else {
		return prometheus.MustNewConstMetric(
			networkRTT, prometheus.GaugeValue, value,
			src_ip, dst_ip, pid, level,
			src_pod, src_namespace, src_node,
			dst_pod, dst_namespace, dst_node,
			node_name, node_ip, container_id,
		)
	}
}
