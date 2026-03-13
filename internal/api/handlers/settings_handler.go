package handlers

import (
	"net/http"

	"clash-manager/internal/config"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct{}

func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

type DNSConfig struct {
	Enable            bool     `json:"enable"`
	Listen            string    `json:"listen"`
	EnhancedMode      string    `json:"enhancedMode"`
	Nameserver        []string  `json:"nameserver"`
	Fallback          []string  `json:"fallback"`
	DefaultNameserver []string  `json:"defaultNameserver"`
	FakeIPFilter      []string  `json:"fakeIPFilter"`
}

func (h *SettingsHandler) GetDNS(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dnsConfig := DNSConfig{
		Enable:            cfg.DNS.Enable,
		Listen:            cfg.DNS.Listen,
		EnhancedMode:      cfg.DNS.EnhancedMode,
		Nameserver:        cfg.DNS.Nameserver,
		Fallback:          cfg.DNS.Fallback,
		DefaultNameserver: cfg.DNS.DefaultNameserver,
		FakeIPFilter:      cfg.DNS.FakeIPFilter,
	}

	c.JSON(http.StatusOK, dnsConfig)
}

func (h *SettingsHandler) SaveDNS(c *gin.Context) {
	var req DNSConfig

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cfg.DNS.Enable = req.Enable
	cfg.DNS.Listen = req.Listen
	cfg.DNS.EnhancedMode = req.EnhancedMode
	cfg.DNS.Nameserver = req.Nameserver
	cfg.DNS.Fallback = req.Fallback
	cfg.DNS.DefaultNameserver = req.DefaultNameserver
	cfg.DNS.FakeIPFilter = req.FakeIPFilter

	if err := config.SaveConfig(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
