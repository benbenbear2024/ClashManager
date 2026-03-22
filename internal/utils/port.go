package utils

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// 常用端口列表，不会被强制结束
var commonPorts = map[int]bool{
	20:   true,  // FTP Data
	21:   true,  // FTP Control
	22:   true,  // SSH
	23:   true,  // Telnet
	25:   true,  // SMTP
	53:   true,  // DNS
	80:   true,  // HTTP
	110:  true,  // POP3
	143:  true,  // IMAP
	443:  true,  // HTTPS
	465:  true,  // SMTPS
	587:  true,  // SMTP Submission
	993:  true,  // IMAPS
	995:  true,  // POP3S
	3306: true,  // MySQL
	3389: true,  // RDP
	5432: true,  // PostgreSQL
	6379: true,  // Redis
	8080: true,  // HTTP Alternate
	8443: true,  // HTTPS Alternate
}

// CheckAndKillPortProcess 检查端口是否被占用，如果是则强制结束进程
// 返回是否成功处理了端口占用
func CheckAndKillPortProcess(port int) (bool, error) {
	// 检查是否是常用端口
	if commonPorts[port] {
		return false, fmt.Errorf("port %d is a common port and will not be killed", port)
	}

	// 检查端口是否被占用
	pid, err := findProcessByPort(port)
	if err != nil {
		// 端口未被占用，返回false
		return false, nil
	}

	// 找到占用端口的进程，尝试结束它
	if err := killProcess(pid); err != nil {
		return false, fmt.Errorf("failed to kill process %d on port %d: %v", pid, port, err)
	}

	return true, nil
}

// findProcessByPort 查找占用指定端口的进程ID
func findProcessByPort(port int) (int, error) {
	switch runtime.GOOS {
	case "windows":
		return findProcessByPortWindows(port)
	case "linux", "darwin":
		return findProcessByPortUnix(port)
	default:
		return 0, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// findProcessByPortWindows 在Windows上查找占用端口的进程
func findProcessByPortWindows(port int) (int, error) {
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to execute netstat: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	portStr := fmt.Sprintf(":%d", port)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// 检查本地地址是否包含目标端口
		if strings.Contains(fields[1], portStr) {
			// 检查状态是否为LISTENING
			if strings.ToUpper(fields[3]) == "LISTENING" {
				pidStr := fields[4]
				pid, err := strconv.Atoi(pidStr)
				if err != nil {
					continue
				}
				return pid, nil
			}
		}
	}

	return 0, fmt.Errorf("no process found listening on port %d", port)
}

// findProcessByPortUnix 在Unix系统上查找占用端口的进程
func findProcessByPortUnix(port int) (int, error) {
	// 尝试使用lsof
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-P", "-n")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 2 {
				pid, err := strconv.Atoi(fields[1])
				if err == nil {
					return pid, nil
				}
			}
		}
	}

	// 如果lsof失败，尝试使用netstat
	cmd = exec.Command("netstat", "-tulpn")
	output, err = cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to execute netstat: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	portStr := fmt.Sprintf(":%d", port)

	for _, line := range lines {
		if strings.Contains(line, portStr) && strings.Contains(line, "LISTEN") {
			fields := strings.Fields(line)
			if len(fields) >= 7 {
				// 在Linux netstat输出中，PID/Program name通常是最后一个字段
				pidProgram := fields[6]
				parts := strings.Split(pidProgram, "/")
				if len(parts) > 0 {
					pid, err := strconv.Atoi(parts[0])
					if err == nil {
						return pid, nil
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("no process found listening on port %d", port)
}

// killProcess 结束指定进程
func killProcess(pid int) error {
	switch runtime.GOOS {
	case "windows":
		return killProcessWindows(pid)
	case "linux", "darwin":
		return killProcessUnix(pid)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// killProcessWindows 在Windows上结束进程
func killProcessWindows(pid int) error {
	cmd := exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("taskkill failed: %v, output: %s", err, string(output))
	}
	return nil
}

// killProcessUnix 在Unix系统上结束进程
func killProcessUnix(pid int) error {
	cmd := exec.Command("kill", "-9", strconv.Itoa(pid))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kill failed: %v, output: %s", err, string(output))
	}
	return nil
}

// IsPortAvailable 检查端口是否可用（未被占用）
func IsPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}
