package netanaly

import (
	"fmt"
	"log"
	"os"
	"pinger/nettool"
	"pinger/proc"
	"sync"

	"github.com/CloudDetail/originx-module/meta/configs"
	"github.com/CloudDetail/originx-module/meta/model/cache"
	"github.com/CloudDetail/originx-module/meta/source"
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

func GetNeedPingsIp(pid uint32, selfNs netns.NsHandle) {
	ns, err := proc.GetNetNs(pid)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ns.Close()
	proc.ExecuteInNetNs(ns, selfNs, func() error {
		c, _ := nettool.New()
		tuples4, _ := c.ListTcp4Conns()
		tuples6, _ := c.ListTcp6Conns()
		c.CloseConnection()
		for _, tuple := range tuples4 {
			AddPing(tuple.SrcIp, tuple.DstIp, pid)
		}
		for _, tuple := range tuples6 {
			AddPing(tuple.SrcIp, tuple.DstIp, pid)
		}
		return nil
	})
}

func addResult(ip4LaString string, ip4ReString string, pid uint32, serviceIp string) {
	reMap, isFound := GlobalRttMap[ip4LaString]
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
		GlobalRttMap[ip4LaString] = reMap
	}
}

func AddPing(ip4LaString string, ip4ReString string, pid uint32) {
	if ip4LaString == ip4ReString {
		return
	}
	service, ok := cache.Querier.GetServiceByIP("", ip4ReString)
	GlobalRttMutex.Lock()
	if ok {
		podsIp := service.EndPoints()
		for _, podIp := range podsIp {
			addResult(ip4LaString, podIp, pid, ip4ReString)
		}
	} else {
		addResult(ip4LaString, ip4ReString, pid, "")
	}
	GlobalRttMutex.Unlock()
}
