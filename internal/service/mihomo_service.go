package service

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type MihomoService struct {
}

func NewMihomoService() *MihomoService {
	return &MihomoService{}
}

// Start 启动 Mihomo 服务
func (s *MihomoService) Start() error {
	// 检查是否是 Linux 系统
	if runtime.GOOS != "linux" {
		return fmt.Errorf("only supported on Linux")
	}

	// 检查进程是否已在运行
	if s.isMihomoRunning() {
		return fmt.Errorf("Mihomo is already running")
	}

	// 使用 systemctl 启动 mihomo 服务
	cmd := exec.Command("systemctl", "start", "mihomo")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start mihomo service: %v, output: %s", err, string(output))
	}

	return nil
}

// Stop 停止 Mihomo 服务
func (s *MihomoService) Stop() error {
	// 检查是否是 Linux 系统
	if runtime.GOOS != "linux" {
		return fmt.Errorf("only supported on Linux")
	}

	// 检查进程是否在运行
	if !s.isMihomoRunning() {
		return fmt.Errorf("Mihomo is not running")
	}

	// 使用 systemctl 停止 mihomo 服务
	cmd := exec.Command("systemctl", "stop", "mihomo")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop mihomo service: %v, output: %s", err, string(output))
	}

	return nil
}

// Status 获取 Mihomo 服务状态
func (s *MihomoService) Status() (string, error) {
	// 检查是否是 Linux 系统
	if runtime.GOOS != "linux" {
		return "unsupported", nil
	}

	// 使用 systemctl 检查 mihomo 服务状态
	cmd := exec.Command("systemctl", "is-active", "mihomo")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "stopped", nil
	}

	status := strings.TrimSpace(string(output))
	if status == "active" {
		return "running", nil
	}

	return status, nil
}

// isMihomoRunning 检查mihomo是否在运行
func (s *MihomoService) isMihomoRunning() bool {
	// 使用 systemctl 检查 mihomo 服务状态
	cmd := exec.Command("systemctl", "is-active", "mihomo")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	status := strings.TrimSpace(string(output))
	return status == "active"
}

// CheckL2TPConnection 检查L2TP连接状态
func (s *MihomoService) CheckL2TPConnection(nodeName string) (bool, error) {
	// Windows系统不支持L2TP连接检测，直接返回false
	if runtime.GOOS == "windows" {
		return false, nil
	}

	// 从节点名称中提取数字（如Name1 -> 1）
	var nodeNum int
	_, err := fmt.Sscanf(nodeName, "Name%d", &nodeNum)
	if err != nil {
		// 如果不是Name格式，尝试其他格式
		return false, fmt.Errorf("invalid node name format")
	}

	// 计算对应的IP地址
	ipAddress := fmt.Sprintf("10.0.10.%d", nodeNum+1)

	// 使用ip addr命令检查该IP是否存在
	cmd := exec.Command("ip", "addr")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check IP address: %v", err)
	}

	// 检查IP地址是否在输出中
	connected := strings.Contains(string(output), ipAddress)

	return connected, nil
}

// Reload 重载mihomo配置（发送SIGHUP信号）
func (s *MihomoService) Reload() error {
	// Windows系统不支持SIGHUP信号，直接返回
	if runtime.GOOS == "windows" {
		return nil
	}

	// 查找mihomo进程ID
	pid, err := s.findMihomoPID()
	if err != nil {
		return fmt.Errorf("failed to find mihomo process: %v", err)
	}

	// 发送SIGHUP信号重载配置
	cmd := exec.Command("kill", "-SIGHUP", fmt.Sprintf("%d", pid))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send SIGHUP signal: %v", err)
	}

	return nil
}

// findMihomoPID 查找mihomo进程的PID
func (s *MihomoService) findMihomoPID() (int, error) {
	cmd := exec.Command("pgrep", "-f", "mihomo")
	output, err := cmd.Output()
	if err != nil {
		// 尝试使用ps命令
		cmd = exec.Command("ps", "aux")
		output, err = cmd.Output()
		if err != nil {
			return 0, fmt.Errorf("failed to get process list: %v", err)
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "mihomo") && !strings.Contains(line, "grep") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					var pid int
					_, err := fmt.Sscanf(fields[1], "%d", &pid)
					if err == nil && pid > 0 {
						return pid, nil
					}
				}
			}
		}
		return 0, fmt.Errorf("mihomo process not found")
	}

	// 解析PID
	pidStr := strings.TrimSpace(string(output))
	var pid int
	_, err = fmt.Sscanf(pidStr, "%d", &pid)
	if err != nil {
		return 0, fmt.Errorf("failed to parse PID: %v", err)
	}

	return pid, nil
}
