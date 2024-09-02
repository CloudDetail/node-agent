package proc

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const procPath = "/proc"

var invalidProcess = []string{
	"/usr/lib/systemd",
	"sh",
	"bash",
	"-bash",
	"/bin/sh",
	"/bin/bash",
	"sshd",
}

var processType = []string{}
var UserHZ = 100

var GlobalPidMutex = &sync.RWMutex{}
var GlobalNeedMonitorPid = make(map[uint32]*ProcessInfo)

type ProcessInfo struct {
	StartTime time.Time
	ContainId string
}

func init() {
	types := os.Getenv("PROCESS_TYPE")
	if types != "" {
		processType = strings.Split(types, ",")
	}
}

func GetPid() {
	pids := listPids()
	for pid, info := range pids {
		GlobalNeedMonitorPid[pid] = info
	}
}

func UpdatePid() {
	pids := listPids()
	newSet := make(map[uint32]struct{})
	for pid := range pids {
		newSet[pid] = struct{}{}
	}

	GlobalPidMutex.Lock()
	for pid := range GlobalNeedMonitorPid {
		if _, ok := newSet[pid]; !ok {
			delete(GlobalNeedMonitorPid, pid)
		}
	}

	for pid := range newSet {
		if _, ok := GlobalNeedMonitorPid[pid]; !ok {
			GlobalNeedMonitorPid[pid] = pids[pid]
		}
	}
	GlobalPidMutex.Unlock()
}

func listPids() map[uint32]*ProcessInfo {
	pids := make(map[uint32]*ProcessInfo)
	d, err := os.Open(procPath)
	if err != nil {
		return pids
	}
	defer d.Close()

	procDirs, err := d.Readdirnames(-1)
	if err != nil {
		return pids
	}

	for _, procDir := range procDirs {
		pid, err := strconv.ParseInt(procDir, 10, 64)
		if err != nil {
			continue
		}
		if pid == 1 || pid == 2 {
			continue
		}
		intpid := uint32(pid)
		command := getCommand(intpid)
		if command == "" || filterProcess(command) {
			continue
		}
		startTime, err := getProcessStartTime(intpid)
		if err != nil {
			continue
		}
		pids[intpid] = &ProcessInfo{
			StartTime: *startTime,
			ContainId: getContainerId(intpid),
		}
	}
	return pids
}

// 过滤进程，先黑名单，后白名单
func filterProcess(command string) bool {
	for _, invalid := range invalidProcess {
		if strings.HasPrefix(command, invalid) {
			return true
		}
	}
	if len(processType) == 0 {
		return false
	}
	for _, t := range processType {
		if strings.Contains(command, t) {
			return false
		}
	}
	return true
}
