package netanaly

import (
	"log"
	"sync"

	"github.com/CloudDetail/metadata/configs"
	"github.com/CloudDetail/metadata/model/cache"
	"github.com/CloudDetail/metadata/source"
	"github.com/CloudDetail/node-agent/config"
	"github.com/CloudDetail/node-agent/middleware"
	"github.com/CloudDetail/node-agent/nettool"
	"github.com/CloudDetail/node-agent/proc"
	"github.com/shirou/gopsutil/net"
	"github.com/vishvananda/netns"
)

func init() {
	cfg := &configs.MetaSourceConfig{
		Querier: &configs.QuerierConfig{
			IsSingleCluster: true,
		},
	}
	kubeAuth := config.GlobalCfg.K8SMetaData.AuthType
	if len(kubeAuth) > 0 {
		kubeConfig := config.GlobalCfg.K8SMetaData.KubeConfig
		cfg.KubeSource = &configs.KubeSourceConfig{
			KubeAuthType:      kubeAuth,
			KubeAuthConfig:    kubeConfig,
			IsEndpointsNeeded: true,
		}
	}

	sourceAddr := config.GlobalCfg.K8SMetaData.FetchSourceAddr
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
	go func() {
		meta := source.CreateMetaSourceFromConfig(cfg)
		err := meta.Run()
		if err != nil {
			log.Printf("failed to get service meta from k8s: %s", err)
		}
	}()

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

func GetNeedPingsIp(pid uint32, selfNs netns.NsHandle, rttMap map[string]map[string]Result, middleNet map[uint32]map[string]middleware.MiddlewareInfo) {
	ns, err := proc.GetNetNs(pid)
	if err != nil {
		log.Println(err)
		return
	}
	defer ns.Close()

	if ns.Equal(proc.HOST_NET_NS) {
		connets, err := net.ConnectionsPid("tcp", int32(pid))
		if err != nil {
			log.Println(err)
			return
		}
		for _, conn := range connets {
			if conn.Status == "ESTABLISHED" {
				AddPing(conn.Laddr.IP, conn.Raddr.IP, pid, rttMap)
				middleware.AddMiddlewareNetwork(conn.Raddr.IP, pid, uint16(conn.Raddr.Port), middleNet)
			}
		}
		return
	}

	err = proc.ExecuteInNetNs(ns, selfNs, func() error {
		c, _ := nettool.New()
		tuples4, _ := c.ListTcp4Conns()
		tuples6, _ := c.ListTcp6Conns()
		c.CloseConnection()
		for _, tuple := range tuples4 {
			AddPing(tuple.SrcIp, tuple.DstIp, pid, rttMap)
			middleware.AddMiddlewareNetwork(tuple.DstIp, pid, tuple.DstPort, middleNet)
		}
		for _, tuple := range tuples6 {
			AddPing(tuple.SrcIp, tuple.DstIp, pid, rttMap)
			middleware.AddMiddlewareNetwork(tuple.DstIp, pid, tuple.DstPort, middleNet)
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
		log.Printf("src: %s dst: %s endpoints: %v\n", ip4LaString, ip4ReString, podsIp)
		for _, podIp := range podsIp {
			addResult(ip4LaString, podIp, pid, ip4ReString, rttMap)
		}
	} else {
		addResult(ip4LaString, ip4ReString, pid, "", rttMap)
	}
}
