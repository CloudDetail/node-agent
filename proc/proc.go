package proc

import (
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var root = "/proc"

func Path(pid uint32, subpath ...string) string {
	return path.Join(append([]string{root, strconv.Itoa(int(pid))}, subpath...)...)
}

func HostPath(p string) string {
	return Path(1, "root", p)
}

func getCommand(pid uint32) string {
	cmdline, err := os.ReadFile(Path(pid, "cmdline"))
	if err != nil {
		return ""
	}
	return strings.Replace(string(cmdline), "\x00", " ", -1)
}
func getContainerId(pid uint32) string {
	data, err := os.ReadFile(Path(pid, "cgroup"))
	if err != nil {
		return ""
	}
	subsystems := make(map[string]string)
	var cgroupContent string

	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		for _, cgType := range strings.Split(parts[1], ",") {
			subsystems[cgType] = path.Join("", parts[2])
		}
	}
	if p := subsystems["name=systemd"]; p != "" {
		cgroupContent = p
	} else if p = subsystems["cpu"]; p != "" {
		cgroupContent = p
	} else {
		cgroupContent = subsystems[""]
	}
	if ContainerId, err := containerByCgroup(cgroupContent); err != nil && len(ContainerId) >= 12 {
		return ContainerId[:12]
	}
	return ""
}
func getProcessStartTime(pid uint32) (*time.Time, error) {
	data, err := os.ReadFile(Path(pid, "stat"))
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
