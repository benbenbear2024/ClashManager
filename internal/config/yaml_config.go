package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	Name              string       `yaml:"name"`
	Type              string       `yaml:"type"`
	Server            string       `yaml:"server"`
	Port              int          `yaml:"port"`
	Cipher            string       `yaml:"cipher,omitempty"`
	Password          string       `yaml:"password,omitempty"`
	UDP               bool         `yaml:"udp,omitempty"`
	Username          string       `yaml:"username,omitempty"`
	UUID              string       `yaml:"uuid,omitempty"`
	AlterId           string       `yaml:"alterId"`
	Network           string       `yaml:"network,omitempty"`
	TLS               bool         `yaml:"tls,omitempty"`
	SkipCert          bool         `yaml:"skip-cert-verify,omitempty"`
	Path              string       `yaml:"path,omitempty"`
	Host              string       `yaml:"host,omitempty"`
	SNI               string       `yaml:"sni,omitempty"`
	ALPN              string       `yaml:"alpn,omitempty"`
	Address           string       `yaml:"address"`
	Delay             string       `yaml:"delay,omitempty"`
	URL               string       `yaml:"url,omitempty"`
	Rename            string       `yaml:"rename,omitempty"`
	Up                int          `yaml:"up,omitempty"`                 // Hysteria2 上行带宽
	Down              int          `yaml:"down,omitempty"`               // Hysteria2 下行带宽
	HopInterval       int          `yaml:"hop-interval,omitempty"`       // Hysteria2 跳跃间隔
	Flow              string       `yaml:"flow,omitempty"`               // VLESS 流控
	ServerName        string       `yaml:"servername,omitempty"`         // VLESS Reality 服务器名称
	RealityOpts       *RealityOpts `yaml:"reality-opts,omitempty"`       // VLESS Reality 选项
	ClientFingerprint string       `yaml:"client-fingerprint,omitempty"` // VLESS 客户端指纹
}

type RealityOpts struct {
	PublicKey   string `yaml:"public-key,omitempty"`
	ShortID     string `yaml:"short-id,omitempty"`
	ServerName  string `yaml:"server-name,omitempty"`
	Fingerprint string `yaml:"fingerprint,omitempty"`
	Show        bool   `yaml:"show,omitempty"`
	Debug       bool   `yaml:"debug,omitempty"`
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
	mihomoPath := MihomoPath

	if _, err := os.Stat(mihomoPath); err == nil {
		return mihomoPath
	}

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

	// 根据proxies自动生成rules
	config.Rules = generateRules(config.Proxies)

	// 创建自定义 YAML 编码器，使用内联格式
	data, err := marshalWithInlineProxies(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	// 如果当前配置路径不是 MihomoPath，则复制到 MihomoPath
	if configPath != MihomoPath {
		if err := CopyConfigToMihomo(); err != nil {
			return fmt.Errorf("failed to copy config to Mihomo: %v", err)
		}
	}

	return nil
}

// generateRules 根据proxies生成对应的rules
func generateRules(proxies []ProxyConfig) []string {
	rules := make([]string, 0, len(proxies))
	for i, proxy := range proxies {
		// 生成SRC-IP-CIDR规则，IP从10.0.10.2开始递增
		rule := fmt.Sprintf("SRC-IP-CIDR,10.0.10.%d/32,%s", i+2, proxy.Name)
		rules = append(rules, rule)
	}
	return rules
}

// marshalWithInlineProxies 使用内联格式序列化配置
func marshalWithInlineProxies(config *MihomoConfig) ([]byte, error) {
	// 创建一个临时结构，不包含 proxies 和 rules
	tempConfig := *config
	tempProxies := tempConfig.Proxies
	tempRules := tempConfig.Rules
	tempConfig.Proxies = nil
	tempConfig.Rules = nil

	// 序列化除 proxies 和 rules 外的内容
	baseData, err := yaml.Marshal(&tempConfig)
	if err != nil {
		return nil, err
	}

	// 序列化 rules 为内联格式
	rulesData := []byte("rules:\n")
	for _, rule := range tempRules {
		rulesData = append(rulesData, []byte("    - "+rule+"\n")...)
	}

	// 序列化 proxies 为内联格式
	proxiesData := []byte("proxies:\n")
	for _, proxy := range tempProxies {
		proxiesData = append(proxiesData, []byte("    - {")...)
		first := true

		// 保持原有的 name 和 rename 逻辑
		proxiesData, first = appendField(proxiesData, first, "name", "'"+proxy.Name+"'")
		if proxy.Rename != "" {
			proxiesData, first = appendField(proxiesData, first, "rename", "'"+proxy.Rename+"'")
		}

		// 处理基本字段（按照正确的顺序）
		proxiesData, first = appendField(proxiesData, first, "type", proxy.Type)
		proxiesData, first = appendField(proxiesData, first, "server", proxy.Server)
		proxiesData, first = appendIntField(proxiesData, first, "port", proxy.Port)

		// 对密码进行适当的引号处理
		if proxy.Password != "" {
			if !first {
				proxiesData = append(proxiesData, []byte(", ")...)
			}
			if strings.Contains(proxy.Password, " ") || strings.Contains(proxy.Password, ":") {
				proxiesData = append(proxiesData, []byte("password: '"+proxy.Password+"'")...)
			} else {
				proxiesData = append(proxiesData, []byte("password: "+proxy.Password)...)
			}
			first = false
		}

		// 处理其他字段（按照正确的顺序，移除 tls 字段）
		proxiesData, first = appendBoolField(proxiesData, first, "udp", proxy.UDP)
		proxiesData, first = appendBoolField(proxiesData, first, "skip-cert-verify", proxy.SkipCert)
		proxiesData, first = appendField(proxiesData, first, "sni", proxy.SNI)
		proxiesData, first = appendField(proxiesData, first, "network", proxy.Network)
		proxiesData, first = appendField(proxiesData, first, "cipher", proxy.Cipher)
		proxiesData, first = appendField(proxiesData, first, "uuid", proxy.UUID)
		proxiesData, first = appendField(proxiesData, first, "alterId", proxy.AlterId)
		proxiesData, first = appendField(proxiesData, first, "flow", proxy.Flow)
		proxiesData, first = appendField(proxiesData, first, "servername", proxy.ServerName)
		proxiesData, first = appendIntField(proxiesData, first, "up", proxy.Up)
		proxiesData, first = appendIntField(proxiesData, first, "down", proxy.Down)
		proxiesData, first = appendIntField(proxiesData, first, "hop-interval", proxy.HopInterval)
		proxiesData, first = appendField(proxiesData, first, "client-fingerprint", proxy.ClientFingerprint)

		// 处理 Reality 选项
		if proxy.RealityOpts != nil {
			// 如果有 reality-opts，添加 tls: true
			if !first {
				proxiesData = append(proxiesData, []byte(", ")...)
			}
			proxiesData = append(proxiesData, []byte("tls: true")...)
			first = false

			if !first {
				proxiesData = append(proxiesData, []byte(", ")...)
			}
			proxiesData = append(proxiesData, []byte("reality-opts: {")...)
			realityFirst := true

			proxiesData, realityFirst = appendField(proxiesData, realityFirst, "public-key", proxy.RealityOpts.PublicKey)
			proxiesData, realityFirst = appendField(proxiesData, realityFirst, "short-id", proxy.RealityOpts.ShortID)
			proxiesData, realityFirst = appendField(proxiesData, realityFirst, "server-name", proxy.RealityOpts.ServerName)
			proxiesData, realityFirst = appendField(proxiesData, realityFirst, "fingerprint", proxy.RealityOpts.Fingerprint)
			proxiesData, realityFirst = appendBoolField(proxiesData, realityFirst, "show", proxy.RealityOpts.Show)
			proxiesData, realityFirst = appendBoolField(proxiesData, realityFirst, "debug", proxy.RealityOpts.Debug)

			proxiesData = append(proxiesData, []byte("}")...)
			first = false
		}

		proxiesData = append(proxiesData, []byte("}\n")...)
	}

	// 拼接三部分: baseData + rulesData + proxiesData
	result := append(baseData, rulesData...)
	result = append(result, proxiesData...)
	return result, nil
}

// appendField 添加字符串字段到内联格式
func appendField(data []byte, first bool, key, value string) ([]byte, bool) {
	if value != "" {
		if !first {
			data = append(data, []byte(", ")...)
		}
		data = append(data, []byte(key+": "+value)...)
		return data, false
	}
	return data, first
}

// appendIntField 添加整数字段到内联格式
func appendIntField(data []byte, first bool, key string, value int) ([]byte, bool) {
	if value != 0 {
		if !first {
			data = append(data, []byte(", ")...)
		}
		data = append(data, []byte(key+": "+fmt.Sprintf("%d", value))...)
		return data, false
	}
	return data, first
}

// appendBoolField 添加布尔字段到内联格式
func appendBoolField(data []byte, first bool, key string, value bool) ([]byte, bool) {
	if value {
		if !first {
			data = append(data, []byte(", ")...)
		}
		data = append(data, []byte(key+": true")...)
		return data, false
	}
	return data, first
}
