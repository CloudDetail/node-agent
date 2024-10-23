package netanaly

import (
	"log"
	"os"
	"sync"

	"github.com/CloudDetail/metadata/configs"
	"github.com/CloudDetail/metadata/model/cache"
	"github.com/CloudDetail/metadata/source"
	"github.com/CloudDetail/node-agent/nettool"
	"github.com/CloudDetail/node-agent/proc"
	"github.com/shirou/gopsutil/net"
	"github.com/vishvananda/netns"
)

func InitMetaData() {
	cfg := &configs.MetaSourceConfig{
		Querier: &configs.QuerierConfig{
			IsSingleCluster: true,
		},
	}
	kubeAuth := os.Getenv("AUTH_TYPE")
	if len(kubeAuth) > 0 {
		kubeConfig := os.Getenv("KUBE_CONFIG")
		cfg.KubeSource = &configs.KubeSourceConfig{
			KubeAuthType:      kubeAuth,
			KubeAuthConfig:    kubeConfig,
			IsEndpointsNeeded: true,
		}
	}

	sourceAddr := os.Getenv("FETCH_SOURCE_ADDR")
	if len(sourceAddr) > 0 {
		cfg.FetchSource = &configs.FetchSourceConfig{
			SourceAddr:   sourceAddr,
			FetchedTypes: "service",
		}
	}

	if cfg.KubeSource == nil && cfg.FetchSource == nil {
		// 未启用K8s信息获取
		log.Println("kubernetes info fetch disable")
		return
	}
	meta := source.CreateMetaSourceFromConfig(cfg)

	err := meta.Run()
	if err != nil {
		log.Printf("failed to get service meta from k8s: %s", err)
	}
}

func GetPodByIp(podIp string) (*cache.Pod, bool) {
	return cache.Querier.GetPodByIP("", podIp)
}

func GetNodeByIP(nodeIP string) (*cache.Node, bool) {
	return cache.Querier.GetNodeByIP("", nodeIP)
}

type Result struct {
	// 定义 Result 类型的字段
	ServiceIp string
	Latency   int
	Pid       map[uint32]struct{}
}

var GlobalRttMap = make(map[string]map[string]Result)

var GlobalRttMutex = &sync.RWMutex{}

func GetNeedPingsIp(pid uint32, selfNs netns.NsHandle, rttMap map[string]map[string]Result) {
	ns, err := proc.GetNetNs(pid)
	if err != nil {
		log.Println(err)
		return
	}

	if ns.Equal(proc.HOST_NET_NS) {
		connets, err := net.ConnectionsPid("tcp", int32(pid))
		if err != nil {
			log.Println(err)
			return
		}
		for _, conn := range connets {
			if conn.Status == "ESTABLISHED" {
				AddPing(conn.Laddr.IP, conn.Raddr.IP, pid, rttMap)
			}
		}
		return
	}

	defer ns.Close()
	err = proc.ExecuteInNetNs(ns, selfNs, func() error {
		c, _ := nettool.New()
		tuples4, _ := c.ListTcp4Conns()
		tuples6, _ := c.ListTcp6Conns()
		c.CloseConnection()
		for _, tuple := range tuples4 {
			AddPing(tuple.SrcIp, tuple.DstIp, pid, rttMap)
		}
		for _, tuple := range tuples6 {
			AddPing(tuple.SrcIp, tuple.DstIp, pid, rttMap)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func addResult(ip4LaString string, ip4ReString string, pid uint32, serviceIp string, rttMap map[string]map[string]Result) {
	reMap, isFound := rttMap[ip4LaString]
	if isFound {
		re, isFoundRe := reMap[ip4ReString]
		if !isFoundRe {
			result := &Result{}
			result.Pid = make(map[uint32]struct{})
			result.Pid[pid] = struct{}{}
			result.ServiceIp = serviceIp
			reMap[ip4ReString] = *result
		} else {
			re.Pid[pid] = struct{}{}
			reMap[ip4ReString] = re
		}
	} else {
		reMap := make(map[string]Result)
		result := &Result{}
		result.Pid = make(map[uint32]struct{})
		result.Pid[pid] = struct{}{}
		result.ServiceIp = serviceIp
		reMap[ip4ReString] = *result
		rttMap[ip4LaString] = reMap
	}
}

func AddPing(ip4LaString string, ip4ReString string, pid uint32, rttMap map[string]map[string]Result) {
	if ip4LaString == ip4ReString {
		return
	}
	service, ok := cache.Querier.GetServiceByIP("", ip4ReString)
	if ok {
		podsIp := service.EndPoints()
		for _, podIp := range podsIp {
			addResult(ip4LaString, podIp, pid, ip4ReString, rttMap)
		}
	} else {
		addResult(ip4LaString, ip4ReString, pid, "", rttMap)
	}
}
