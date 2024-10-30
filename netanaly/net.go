package netanaly

import (
	"context"
	"log"
	"time"

	"github.com/CloudDetail/node-agent/config"
	"github.com/CloudDetail/node-agent/proc"
)

func UpdateNetConnectInfo(ctx context.Context) {
	selfNs, err := proc.GetSelfNetNs()
	if err != nil {
		log.Println(err)
	}
	proc.GetPid()
	rttMap := make(map[string]map[string]Result)
	for mPid := range proc.GlobalNeedMonitorPid {
		GetNeedPingsIp(mPid, selfNs, rttMap)
	}
	log.Println(rttMap)
	GlobalRttMap = rttMap
	selfNs.Close()

	ticker := time.NewTicker(time.Duration(config.GlobalCfg.PidSpan) * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			proc.UpdatePid()
			selfNs, err := proc.GetSelfNetNs()
			if err != nil {
				log.Fatalln(err)
			}

			var needPid []uint32
			proc.GlobalPidMutex.RLock()
			for pid := range proc.GlobalNeedMonitorPid {
				needPid = append(needPid, pid)
			}
			proc.GlobalPidMutex.RUnlock()

			rttMap := make(map[string]map[string]Result)
			log.Printf("Moniter Pid: %v", needPid)
			for _, mPid := range needPid {
				GetNeedPingsIp(mPid, selfNs, rttMap)
			}
			log.Println(rttMap)

			GlobalRttMutex.Lock()
			GlobalRttMap = rttMap
			GlobalRttMutex.Unlock()

			selfNs.Close()
		}
	}
}
