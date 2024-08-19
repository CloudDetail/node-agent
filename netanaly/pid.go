package netanaly

import (
	"fmt"
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
var GlobalNeedMonitorPid = make(map[int]*ProcessInfo)

type ProcessInfo struct {
	StartTime time.Time
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
	newSet := make(map[int]struct{})
	for pid, _ := range pids {
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

func listPids() map[int]*ProcessInfo {
	pids := make(map[int]*ProcessInfo)
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
		cmdlinePath := fmt.Sprintf("%s/%d/cmdline", procPath, pid)
		cmdline, err := os.ReadFile(cmdlinePath)
		if err != nil {
			continue
		}
		command := strings.Replace(string(cmdline), "\x00", " ", -1)
		if command == "" || filterProcess(command) {
			continue
		}
		intpid := int(pid)
		startTime, err := getProcessStartTime(intpid)
		if err != nil {
			continue
		}
		pids[intpid] = &ProcessInfo{
			StartTime: *startTime,
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

func getProcessStartTime(pid int) (*time.Time, error) {
	statFilePath := fmt.Sprintf("%s/%d/stat", procPath, pid)
	data, err := os.ReadFile(statFilePath)
	if err != nil {
		return nil, err
	}

	stats := strings.Fields(string(data))

	startTimeJiffies, err := strconv.ParseUint(stats[21], 10, 64)
	if err != nil {
		return nil, err
	}

	uptimeData, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return nil, err
	}

	uptimeFields := strings.Fields(string(uptimeData))
	uptimeSeconds, err := strconv.ParseFloat(uptimeFields[0], 64)
	if err != nil {
		return nil, err
	}

	startTimeSeconds := float64(startTimeJiffies) / float64(UserHZ)
	bootTime := time.Now().Add(-time.Duration(uptimeSeconds) * time.Second)
	processStartTime := bootTime.Add(time.Duration(startTimeSeconds) * time.Second)

	return &processStartTime, nil
}
