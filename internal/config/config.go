package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	ServerPort           = ":8080"
	MihomoPath           = "/etc/mihomo/config.yaml"
	DefaultMihomoAPIPort = 9090
)

func GetDBPath() string {
	cwd, _ := os.Getwd()
	cwdDbPath := filepath.Join(cwd, "data", "clash.db")
	if _, err := os.Stat(cwdDbPath); err == nil {
		return cwdDbPath
	}

	exePath, err := os.Executable()
	if err != nil {
		return "data/clash.db"
	}

	exeDir := filepath.Dir(exePath)
	dbPath := filepath.Join(exeDir, "data", "clash.db")

	dataDir := filepath.Join(exeDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "data/clash.db"
	}

	return dbPath
}

func GetSubscriptionSourcesPath() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "data", "subscription_sources.json")
}

func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func WriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func CopyConfigToMihomo() error {
	srcPath := GetConfigPath()
	dstPath := MihomoPath

	// 确保目标路径是绝对路径
	if !filepath.IsAbs(dstPath) {
		cwd, _ := os.Getwd()
		dstPath = filepath.Join(cwd, dstPath)
	}

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	dir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(dstPath, data, 0644)
}

func GetMihomoAPIPort() int {
	// 尝试从配置文件读取
	configPath := GetConfigPath()
	data, err := os.ReadFile(configPath)
	if err == nil {
		var config struct {
			ExternalController string `yaml:"external-controller"`
		}
		if err := yaml.Unmarshal(data, &config); err == nil && config.ExternalController != "" {
			// 解析 external-controller 字段，格式可能是 "127.0.0.1:9999" 或 ":9999"
			parts := strings.Split(config.ExternalController, ":")
			if len(parts) > 1 {
				portStr := parts[len(parts)-1]
				if port, err := strconv.Atoi(portStr); err == nil && port > 0 && port <= 65535 {
					return port
				}
			}
		}
	}

	// 尝试从 MihomoPath 读取
	data, err = os.ReadFile(MihomoPath)
	if err == nil {
		var config struct {
			ExternalController string `yaml:"external-controller"`
		}
		if err := yaml.Unmarshal(data, &config); err == nil && config.ExternalController != "" {
			// 解析 external-controller 字段，格式可能是 "127.0.0.1:9999" 或 ":9999"
			parts := strings.Split(config.ExternalController, ":")
			if len(parts) > 1 {
				portStr := parts[len(parts)-1]
				if port, err := strconv.Atoi(portStr); err == nil && port > 0 && port <= 65535 {
					return port
				}
			}
		}
	}

	// 返回默认端口
	return DefaultMihomoAPIPort
}
