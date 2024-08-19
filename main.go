package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	"github.com/CloudDetail/node-agent/netanaly"
	"github.com/CloudDetail/node-agent/nettool"
	"github.com/CloudDetail/node-agent/proc"
	"github.com/vishvananda/netns"
	"inet.af/netaddr"
)

var pingSpan = 5

func init() {
	go netanaly.InitMetaData()
	envVar := os.Getenv("PING_SPAN")
	if value, err := strconv.Atoi(envVar); err == nil {
		pingSpan = value
	}
}

func main() {

	go func() {
		log.Println(http.ListenAndServe(":6061", nil))
	}()
	KernelBlow317()
	done := make(chan struct{})

	go func() {
		selfNs, err := proc.GetSelfNetNs()
		if err != nil {
			log.Println(err)
		}
		netanaly.GetPid()
		rttMap := make(map[string]map[string]netanaly.Result)
		for mPid := range netanaly.GlobalNeedMonitorPid {
			netanaly.GetNeedPingsIp(uint32(mPid), selfNs, rttMap)
		}
		log.Println(rttMap)
		netanaly.GlobalRttMap = rttMap
		selfNs.Close()
		go netanaly.UpdatePid()

		ticker := time.NewTicker(10 * time.Minute)
		for {
			select {
			case <-ticker.C:
				selfNs, err := proc.GetSelfNetNs()
				if err != nil {
					log.Fatalln(err)
				}
				var needPid []uint32
				netanaly.GlobalPidMutex.RLock()
				for pid := range netanaly.GlobalNeedMonitorPid {
					needPid = append(needPid, uint32(pid))
				}
				netanaly.GlobalPidMutex.RUnlock()

				rttMap := make(map[string]map[string]netanaly.Result)
				log.Printf("needPid: %v", needPid)
				for _, mPid := range needPid {
					netanaly.GetNeedPingsIp(mPid, selfNs, rttMap)
				}

				log.Println(rttMap)
				netanaly.GlobalRttMutex.Lock()
				netanaly.GlobalRttMap = rttMap
				netanaly.GlobalRttMutex.Unlock()
				selfNs.Close()
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Duration(pingSpan) * time.Second)
		for {

			select {
			case <-ticker.C:
				PingIp()
			}
		}

	}()

	<-done

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

func PingIp() {
	selfNs, _ := proc.GetSelfNetNs()
	defer selfNs.Close()
	netanaly.GlobalRttMutex.Lock()
	defer netanaly.GlobalRttMutex.Unlock()
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
		rttMap, _ := Ping(tmpNs, selfNs, targets, 300*time.Millisecond)
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
