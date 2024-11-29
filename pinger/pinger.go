package pinger

import (
	"context"
	"time"

	"github.com/CloudDetail/node-agent/config"
	"github.com/CloudDetail/node-agent/netanaly"
	"github.com/CloudDetail/node-agent/nettool"
	"github.com/CloudDetail/node-agent/proc"
	"github.com/vishvananda/netns"
	"inet.af/netaddr"
)

func Ping(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(config.GlobalCfg.Metric.PingSpan) * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingIp()
		}
	}
}

func updateRttResultMap(tuple nettool.Tuple, f float64, pid map[uint32]struct{}) {
	rs, ok := netanaly.GlobalRttResultMap[tuple]
	if ok {
		rs.Count++
		rs.SumLatency += f
	} else {
		pids := make([]uint32, 0)
		for pid := range pid {
			pids = append(pids, pid)
		}
		rs = netanaly.RttStatistic{
			Count:      1,
			SumLatency: f,
			Pids:       pids,
		}
	}
	netanaly.GlobalRttResultMap[tuple] = rs
}

func pingIp() {
	selfNs, _ := proc.GetSelfNetNs()
	defer selfNs.Close()
	netanaly.GlobalRttMutex.RLock()
	defer netanaly.GlobalRttMutex.RUnlock()
	for srcIp, reMap := range netanaly.GlobalRttMap {
		i := 0
		var tmpNs netns.NsHandle
		targets := make([]netaddr.IP, len(reMap))
		tmpPid := uint32(0)
		for s, result := range reMap {
			targets[i], _ = netaddr.ParseIP(s)
			i++
			for pid := range result.Pid {
				tmpPid = pid
				break
			}
		}
		if tmpPid == 0 {
			continue
		}
		tmpNs, _ = proc.GetNetNs(tmpPid)
		rttMap, _ := ping(tmpNs, selfNs, targets, 300*time.Millisecond)
		tmpNs.Close()
		netanaly.RttResultMapMutex.Lock()
		for ip, f := range rttMap {
			res, _ := reMap[ip.String()]
			if res.ServiceIp != "" {
				serviceTuple := nettool.Tuple{
					SrcIp:     srcIp,
					ServiceIp: res.ServiceIp,
					DstIp:     "",
				}
				updateRttResultMap(serviceTuple, f, res.Pid)
			}
			tuple := nettool.Tuple{
				SrcIp:     srcIp,
				ServiceIp: "",
				DstIp:     ip.String(),
			}
			updateRttResultMap(tuple, f, res.Pid)
		}
		netanaly.RttResultMapMutex.Unlock()

	}

}
