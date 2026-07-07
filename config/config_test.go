package config

import "testing"

func TestMetricConfigDefaultsPrometheusPortTo9408(t *testing.T) {
	cfg := MetricConfig{}

	cfg.setDefault()

	if cfg.PrometheusPort != 9408 {
		t.Fatalf("expected default prometheus port 9408, got %d", cfg.PrometheusPort)
	}
}

func TestMetricConfigPrometheusListenAddrUsesConfiguredPort(t *testing.T) {
	cfg := MetricConfig{PrometheusPort: 19408}

	cfg.setDefault()

	if cfg.PrometheusPort != 19408 {
		t.Fatalf("expected configured prometheus port 19408, got %d", cfg.PrometheusPort)
	}
	if got := cfg.PrometheusListenAddr(); got != ":19408" {
		t.Fatalf("expected prometheus listen address :19408, got %q", got)
	}
}
