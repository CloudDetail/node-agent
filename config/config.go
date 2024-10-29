package config

import (
	"os"
	"strconv"
)

type Config struct {
	ProcessType  []string
	K8SNameSpace []string
	ProcessTime  bool

	LRUCacheSize int
	PingSpan     int
	PidSpan      int

	NodeIP   string
	NodeName string

	FetchSourceAddr string
	AuthType        string
	KubeConfig      string
}

func (c *Config) SetDefault() {
	if c.PingSpan == 0 {
		c.PingSpan = 5
	}
	if c.PidSpan == 0 {
		c.PidSpan = 1
	}
	if c.LRUCacheSize == 0 {
		c.LRUCacheSize = 50000
	}
}

var GlobalCfg = newConfig()

func newConfig() *Config {
	cfg := &Config{}
	if value, err := strconv.Atoi(os.Getenv("PING_SPAN")); err == nil {
		cfg.PingSpan = value
	}
	if value, err := strconv.Atoi(os.Getenv("PID_SPAN")); err == nil {
		cfg.PidSpan = value
	}
	if value, err := strconv.Atoi(os.Getenv("LRU_CACHE_SIZE")); err == nil {
		cfg.LRUCacheSize = value
	}
	if os.Getenv("PROCESS_TIME") == "true" {
		cfg.ProcessTime = true
	}

	cfg.FetchSourceAddr = os.Getenv("FETCH_SOURCE_ADDR")
	cfg.AuthType = os.Getenv("AUTH_TYPE")
	cfg.KubeConfig = os.Getenv("KUBE_CONFIG")

	cfg.NodeName = os.Getenv("MY_NODE_NAME")
	cfg.NodeIP = os.Getenv("MY_NODE_IP")

	cfg.SetDefault()
	return cfg
}
