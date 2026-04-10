package handlers

import (
	"clash-manager/internal/config"
	"clash-manager/internal/service"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type MihomoHandler struct {
	Service *service.MihomoService
}

func NewMihomoHandler() *MihomoHandler {
	return &MihomoHandler{
		Service: service.NewMihomoService(),
	}
}

// StartMihomo 启动 Mihomo 服务
func (h *MihomoHandler) StartMihomo(c *gin.Context) {
	if err := h.Service.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Mihomo started successfully"})
}

// StopMihomo 停止 Mihomo 服务
func (h *MihomoHandler) StopMihomo(c *gin.Context) {
	if err := h.Service.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Mihomo stopped successfully"})
}

// GetMihomoStatus 获取 Mihomo 服务状态
func (h *MihomoHandler) GetMihomoStatus(c *gin.Context) {
	status, err := h.Service.Status()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": status})
}

// GetConfigContent 获取 config.yaml 的原始内容
func (h *MihomoHandler) GetConfigContent(c *gin.Context) {
	content, err := config.ReadFile(config.GetConfigPath())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"content": string(content)})
}

// SaveConfigContent 保存 config.yaml 的原始内容
func (h *MihomoHandler) SaveConfigContent(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.WriteFile(config.GetConfigPath(), []byte(req.Content)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config saved successfully"})
}

// CheckL2TPConnection 检查L2TP连接状态
func (h *MihomoHandler) CheckL2TPConnection(c *gin.Context) {
	var req struct {
		NodeName string `json:"nodeName" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	connected, err := h.Service.CheckL2TPConnection(req.NodeName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connected": connected,
		"nodeName":  req.NodeName,
	})
}

// ReloadMihomo 重载 Mihomo 配置
func (h *MihomoHandler) ReloadMihomo(c *gin.Context) {
	if err := h.Service.Reload(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Mihomo config reloaded successfully"})
}

// ProxyMihomoAPI 代理 Mihomo API 请求
func (h *MihomoHandler) ProxyMihomoAPI(c *gin.Context) {
	// 记录请求信息
	fmt.Printf("Received request: %s %s\n", c.Request.Method, c.Request.URL.Path)
	fmt.Printf("Path parameter: %s\n", c.Param("path"))

	// 获取 mihomo API 端口
	mihomoPort := config.GetMihomoAPIPort()
	fmt.Printf("Mihomo port: %d\n", mihomoPort)

	// 获取完整路径，从 /proxies 开始
	fullPath := c.Request.URL.Path
	// 移除 /api 前缀
	proxyPath := strings.TrimPrefix(fullPath, "/api")
	fmt.Printf("Proxy path: %s\n", proxyPath)

	// 构建目标 URL
	targetURL := &url.URL{
		Scheme:   "http",
		Host:     "127.0.0.1:" + strconv.Itoa(mihomoPort),
		Path:     proxyPath,
		RawQuery: c.Request.URL.RawQuery,
	}
	fmt.Printf("Target URL: %s\n", targetURL.String())

	// 创建新的请求
	req, err := http.NewRequest(c.Request.Method, targetURL.String(), c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request: " + err.Error()})
		return
	}

	// 复制请求头
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to proxy request: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 设置响应状态码
	c.Status(resp.StatusCode)

	// 复制响应体
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy response: " + err.Error()})
		return
	}
}
