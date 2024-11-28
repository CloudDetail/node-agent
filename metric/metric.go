package metric

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/CloudDetail/node-agent/config"
	"github.com/CloudDetail/node-agent/middleware"
	"github.com/CloudDetail/node-agent/netanaly"
	"github.com/CloudDetail/node-agent/proc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/utils/lru"
)

func init() {
	counterLabel = lru.NewWithEvictionFunc(config.GlobalCfg.Metric.LRUCacheSize, func(key lru.Key, value interface{}) {
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
	cfg := config.GlobalCfg
	pid2cid := make(map[uint32]string)
	if cfg.Metric.ProcessTime {
		proc.GlobalPidMutex.RLock()
		for pid, processInfo := range proc.GlobalNeedMonitorPid {
			ch <- prometheus.MustNewConstMetric(
				processStartTime, prometheus.GaugeValue,
				float64(processInfo.StartTime.Unix()),
				strconv.FormatUint(uint64(pid), 10),
				cfg.NodeName,
				cfg.NodeIP,
				processInfo.ContainId,
			)
			pid2cid[pid] = processInfo.ContainId
		}
		proc.GlobalPidMutex.RUnlock()
	}
	middleware.MiddlewareInstance.Mu.Lock()
	for pid, conn := range middleware.MiddlewareInstance.Pid2Connect {
		for _, info := range conn {
			ch <- createMiddlewareMetric(
				strconv.FormatUint(uint64(pid), 10), pid2cid[pid],
				cfg.NodeName, cfg.NodeIP,
				info.PeerIp, strconv.Itoa(int(info.PeerPort)), info.PeerType.String(),
				info.ServiceIp, strconv.Itoa(int(info.ServicePort)),
			)
		}
	}
	middleware.MiddlewareInstance.Mu.Unlock()

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
						cfg.NodeName, cfg.NodeIP, containerId,
					).Inc()
					continue
				}
				ch <- createRttMetric(rtt,
					tuple.SrcIp, tuple.ServiceIp, pid, "service",
					srcPod, srcNamespace, srcNode,
					"", "", "",
					cfg.NodeName, cfg.NodeIP, containerId,
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
						cfg.NodeName, cfg.NodeIP, containerId,
					).Inc()
					continue
				}
				ch <- createRttMetric(rtt,
					tuple.SrcIp, tuple.DstIp, pid, "instance",
					srcPod, srcNamespace, srcNode,
					dstPod, dstNamespace, dstNode,
					cfg.NodeName, cfg.NodeIP, containerId,
				)
			}
		}
		// 删除元素
		delete(netanaly.GlobalRttResultMap, tuple)
	}
	// 解锁 RttResultMapMutex
	netanaly.RttResultMapMutex.Unlock()
}
