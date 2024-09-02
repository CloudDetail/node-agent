package proc

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	dockerIdRegexp      = regexp.MustCompile(`([a-z0-9]{64})`)
	crioIdRegexp        = regexp.MustCompile(`crio-([a-z0-9]{64})`)
	containerdIdRegexp  = regexp.MustCompile(`cri-containerd[-:]([a-z0-9]{64})`)
	lxcIdRegexp         = regexp.MustCompile(`/lxc/([^/]+)`)
	systemSliceIdRegexp = regexp.MustCompile(`(/(system|runtime)\.slice/([^/]+))`)
)

func containerByCgroup(path string) (string, error) {
	parts := strings.Split(strings.TrimLeft(path, "/"), "/")
	if len(parts) < 2 {
		return "", nil
	}
	prefix := parts[0]
	if prefix == "user.slice" || prefix == "init.scope" {
		return "", nil
	}
	if prefix == "docker" || (prefix == "system.slice" && strings.HasPrefix(parts[1], "docker-")) {
		matches := dockerIdRegexp.FindStringSubmatch(path)
		if matches == nil {
			return "", fmt.Errorf("invalid docker cgroup %s", path)
		}
		return matches[1], nil
	}
	if strings.Contains(path, "kubepods") {
		crioMatches := crioIdRegexp.FindStringSubmatch(path)
		if crioMatches != nil {
			return crioMatches[1], nil
		}
		if strings.Contains(path, "crio-conmon-") {
			return "", nil
		}
		containerdMatches := containerdIdRegexp.FindStringSubmatch(path)
		if containerdMatches != nil {
			return containerdMatches[1], nil
		}
		matches := dockerIdRegexp.FindStringSubmatch(path)
		if matches == nil {
			return "", nil
		}
		return matches[1], nil
	}
	if prefix == "lxc" {
		matches := lxcIdRegexp.FindStringSubmatch(path)
		if matches == nil {
			return "", fmt.Errorf("invalid lxc cgroup %s", path)
		}
		return matches[1], nil
	}
	if prefix == "system.slice" || prefix == "runtime.slice" {
		matches := systemSliceIdRegexp.FindStringSubmatch(path)
		if matches == nil {
			return "", fmt.Errorf("invalid systemd cgroup %s", path)
		}
		return strings.Replace(matches[1], "\\x2d", "-", -1), nil
	}
	return "", fmt.Errorf("unknown container: %s", path)
}
