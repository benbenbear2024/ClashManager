package config

import (
	"os"
	"path/filepath"
)

const (
	ServerPort = ":8090"
)

// GetDBPath 获取数据库文件的绝对路径
// 优先使用当前工作目录下的数据库，如果不存在则使用可执行文件目录
func GetDBPath() string {
	// 1. 首先检查当前工作目录下是否有数据库
	cwd, _ := os.Getwd()
	cwdDbPath := filepath.Join(cwd, "data", "clash.db")
	if _, err := os.Stat(cwdDbPath); err == nil {
		return cwdDbPath
	}

	// 2. 获取可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		return "data/clash.db"
	}

	exeDir := filepath.Dir(exePath)
	dbPath := filepath.Join(exeDir, "data", "clash.db")

	// 3. 确保数据目录存在
	dataDir := filepath.Join(exeDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "data/clash.db"
	}

	return dbPath
}
