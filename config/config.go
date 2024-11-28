package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	MY_NODE_NAME = "MY_NODE_NAME"
	MY_NODE_IP   = "MY_NODE_IP"
)

type Config struct {
	Metric      MetricConfig         `yaml:"metric"`
	WhiteList   WhiteListConfig      `yaml:"whitelist"`
	K8SMetaData K8SMetaDataConfig    `yaml:"k8s_metadata"`
	Middleware  MiddlewarePortConfig `yaml:"middleware_port"`
	NodeIP      string
	NodeName    string
	Kernel317   bool
}

type MetricConfig struct {
	PingSpan     int  `yaml:"ping_span"`
	PidSpan      int  `yaml:"pid_span"`
	LRUCacheSize int  `yaml:"lru_cache_size"`
	ProcessTime  bool `yaml:"process_time"`
}

func (m *MetricConfig) setDefault() {
	if m.PingSpan == 0 {
		m.PingSpan = 5
	}
	if m.PidSpan == 0 {
		m.PidSpan = 1
	}
	if m.LRUCacheSize == 0 {
		m.LRUCacheSize = 50000
	}
	m.ProcessTime = true
}

type WhiteListConfig struct {
	ProcessType  []string `yaml:"process_type"`
	K8SNameSpace []string `yaml:"k8s_namespace"`
}

type K8SMetaDataConfig struct {
	FetchSourceAddr string
	AuthType        string
	KubeConfig      string
}

func (c *Config) checkAndSetDefault() {
	c.Metric.setDefault()
	c.Middleware.setDefault()
	c.NodeName = os.Getenv(MY_NODE_NAME)
	c.NodeIP = os.Getenv(MY_NODE_IP)
	c.Kernel317 = KernelBlow317()
}

var GlobalCfg = newConfig()

func newConfig() *Config {
	cfg := &Config{}
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("read config.yaml failed: %v", err)
		return cfg
	}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		log.Fatalf("unmarshal config.yaml failed: %v", err)
	}
	cfg.checkAndSetDefault()
	return cfg
}
