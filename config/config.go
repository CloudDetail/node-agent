package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

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
	PingSpan       int    `yaml:"ping_span"`
	PidSpan        int    `yaml:"pid_span"`
	LRUCacheSize   int    `yaml:"lru_cache_size"`
	ProcessTime    bool   `yaml:"process_time"`
	PrometheusPort uint16 `yaml:"prometheus_port"`
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
	if m.PrometheusPort == 0 {
		m.PrometheusPort = 9408
	}
	m.ProcessTime = true
}

func (m MetricConfig) PrometheusListenAddr() string {
	return ":" + strconv.Itoa(int(m.PrometheusPort))
}

type WhiteListConfig struct {
	ProcessType  []string `yaml:"process_type"`
	K8SNameSpace []string `yaml:"k8s_namespace"`
}

type K8SMetaDataConfig struct {
	FetchSourceAddr string `yaml:"fetch_source_addr"`
	AuthType        string `yaml:"auth_type"`
	KubeConfig      string `yaml:"kube_config"`
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
	data, err := readConfigFile()
	if err != nil {
		log.Fatalf("read config.yaml failed: %v", err)
		return cfg
	}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		log.Fatalf("unmarshal config.yaml failed: %v", err)
	}
	cfg.checkAndSetDefault()
	fmt.Println(cfg)
	return cfg
}

func readConfigFile() ([]byte, error) {
	data, err := os.ReadFile("./config.yaml")
	if err == nil || !os.IsNotExist(err) {
		return data, err
	}

	parentData, parentErr := os.ReadFile("../config.yaml")
	if parentErr == nil {
		return parentData, nil
	}
	return nil, err
}
