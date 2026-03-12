package config

import (
	"os"
	"path/filepath"
)

const (
	ServerPort = ":8090"
	MihomoPath = "/docs/config.yaml"
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
