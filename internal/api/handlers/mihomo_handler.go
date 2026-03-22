package handlers

import (
	"clash-manager/internal/config"
	"clash-manager/internal/service"
	"net/http"

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
