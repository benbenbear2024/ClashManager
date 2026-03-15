package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"clash-manager/internal/service"
)

func main() {
	// 订阅源 URL
	url := "https://clashgithub.com/wp-content/uploads/rss/20260315.txt"

	// 发起 HTTP 请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Failed to fetch subscription: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read body: %v\n", err)
		return
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Encoding: %s\n", resp.Header.Get("Content-Encoding"))
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("Body length: %d\n", len(body))

	// 转换为字符串
	content := string(body)
	fmt.Printf("First 500 chars: %s\n", content[:min(500, len(content))])
	fmt.Printf("Last 500 chars: %s\n", content[max(0, len(content)-500):])

	// 尝试 base64 解码
	decoded, err := service.TryBase64Decode(content)
	if err != nil {
		fmt.Printf("Base64 decode failed: %v\n", err)
		// 尝试直接处理原始内容
		decoded = content
	} else {
		fmt.Printf("Base64 decoded successfully, length: %d\n", len(decoded))
		fmt.Printf("Decoded first 500 chars: %s\n", decoded[:min(500, len(decoded))])
	}

	// 尝试按换行符分割
	lines := strings.Split(decoded, "\n")
	fmt.Printf("Total lines: %d\n", len(lines))

	// 打印前10行
	for i, line := range lines {
		if i >= 10 {
			break
		}
		line = strings.TrimSpace(line)
		if line != "" {
			fmt.Printf("Line %d: %s\n", i, line[:min(100, len(line))])
		}
	}

	// 尝试解析前5个链接
	count := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		node, err := service.ParseLink(line)
		if err != nil {
			fmt.Printf("Failed to parse line: %v\n", err)
		} else {
			fmt.Printf("Parsed node: %s (Type: %s)\n", node.Name, node.Type)
			count++
		}

		if count >= 5 {
			break
		}
	}

	fmt.Printf("Parsed %d nodes\n", count)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
