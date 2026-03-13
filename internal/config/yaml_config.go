package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type MihomoConfig struct {
	MixedPort          int                    `yaml:"mixed-port,omitempty"`
	RedirPort          int                    `yaml:"redir-port,omitempty"`
	TproxyPort         int                    `yaml:"tproxy-port,omitempty"`
	Authentication     []string               `yaml:"authentication,omitempty"`
	AllowLan           bool                   `yaml:"allow-lan,omitempty"`
	Mode               string                 `yaml:"mode,omitempty"`
	LogLevel           string                 `yaml:"log-level,omitempty"`
	IPv6               bool                   `yaml:"ipv6,omitempty"`
	ExternalController string                 `yaml:"external-controller,omitempty"`
	ExternalUI         string                 `yaml:"external-ui,omitempty"`
	Secret             string                 `yaml:"secret,omitempty"`
	Tun                map[string]interface{} `yaml:"tun,omitempty"`
	Experimental       map[string]interface{} `yaml:"experimental,omitempty"`
	DNS                DNSConfig              `yaml:"dns,omitempty"`
	StoreSelected      bool                   `yaml:"store-selected,omitempty"`
	FindProcessMode    string                 `yaml:"find-process-mode,omitempty"`
	Proxies            []ProxyConfig          `yaml:"proxies,omitempty"`
	ProxyGroups        []ProxyGroupConfig     `yaml:"proxy-groups,omitempty"`
	Rules              []string               `yaml:"rules,omitempty"`
}

type ProxyConfig struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Server   string `yaml:"server"`
	Port     int    `yaml:"port"`
	Cipher   string `yaml:"cipher,omitempty"`
	Password string `yaml:"password,omitempty"`
	UDP      bool   `yaml:"udp,omitempty"`
	Username string `yaml:"username,omitempty"`
	UUID     string `yaml:"uuid,omitempty"`
	AlterId  string `yaml:"alterId,omitempty"`
	Network  string `yaml:"network,omitempty"`
	TLS      bool   `yaml:"tls,omitempty"`
	SkipCert bool   `yaml:"skip-cert-verify,omitempty"`
	Path     string `yaml:"path,omitempty"`
	Host     string `yaml:"host,omitempty"`
	ALPN     string `yaml:"alpn,omitempty"`
	Address  string `yaml:"address,omitempty"`
	Delay    string `yaml:"delay,omitempty"`
	URL      string `yaml:"url,omitempty"`
	Rename   string `yaml:"rename,omitempty"`
}

type ProxyGroupConfig struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Proxies  []string `yaml:"proxies,omitempty"`
	URL      string   `yaml:"url,omitempty"`
	Interval int      `yaml:"interval,omitempty"`
}

type DNSConfig struct {
	Enable            bool     `yaml:"enable"`
	Listen            string   `yaml:"listen"`
	EnhancedMode      string   `yaml:"enhanced-mode"`
	Nameserver        []string `yaml:"nameserver"`
	Fallback          []string `yaml:"fallback"`
	DefaultNameserver []string `yaml:"default-nameserver"`
	FakeIPFilter      []string `yaml:"fake-ip-filter"`
}

func GetConfigPath() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "data", "config.yaml")
}

func LoadConfig() (*MihomoConfig, error) {
	configPath := GetConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config MihomoConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

func SaveConfig(config *MihomoConfig) error {
	configPath := GetConfigPath()

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	if err := CopyConfigToMihomo(); err != nil {
		return fmt.Errorf("failed to copy config to Mihomo: %v", err)
	}

	return nil
}
