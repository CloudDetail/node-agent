package middleware

import (
	"fmt"
	"sync"

	"github.com/CloudDetail/metadata/model/cache"
)

var MiddlewareInstance = &Middleware{
	Pid2Connect: make(map[uint32]map[string]MiddlewareInfo),
}

type Middleware struct {
	Mu          sync.RWMutex
	Pid2Connect map[uint32]map[string]MiddlewareInfo
	// pid -> podip:port -> MiddlewareInfo
}

func (m *Middleware) SetInfo(info map[uint32]map[string]MiddlewareInfo) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.Pid2Connect = info
}

type MiddlewareInfo struct {
	ServiceIp   string
	ServicePort uint16
	PeerIp      string
	PeerPort    uint16
	PeerType    MiddlewareType
}

func addMiddleware(serviceIp, peerIp string, pid uint32, serviceport, peerport uint16, info map[uint32]map[string]MiddlewareInfo) {
	peerType := GetMiddlewareType(peerport)
	if peerType == UNKNOWN {
		return
	}

	conn, ok := info[pid]
	if ok {
		if _, ok := conn[peerIp+":"+fmt.Sprint(peerport)]; ok {
			return
		}
		conn[peerIp+":"+fmt.Sprint(peerport)] = MiddlewareInfo{
			ServiceIp:   serviceIp,
			ServicePort: serviceport,
			PeerIp:      peerIp,
			PeerPort:    peerport,
			PeerType:    peerType,
		}
		info[pid] = conn
	} else {
		conn = make(map[string]MiddlewareInfo)
		conn[peerIp+":"+fmt.Sprint(peerport)] = MiddlewareInfo{
			ServiceIp:   serviceIp,
			ServicePort: serviceport,
			PeerIp:      peerIp,
			PeerPort:    peerport,
			PeerType:    peerType,
		}
		info[pid] = conn
	}
}

func AddMiddlewareNetwork(peerIp string, pid uint32, port uint16, info map[uint32]map[string]MiddlewareInfo) {

	service, ok := cache.Querier.GetServiceByIP("", peerIp)
	if ok {
		podsIp := service.EndPoints()
		ports := service.SvcPorts()
		podport := ports[port]
		for _, podIp := range podsIp {
			addMiddleware(peerIp, podIp, pid, port, podport, info)
		}
	} else {
		addMiddleware("", peerIp, pid, 0, port, info)
	}
}
