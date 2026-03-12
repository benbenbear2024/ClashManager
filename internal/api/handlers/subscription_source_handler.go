package handlers

import (
	"fmt"
	"net/http"
	"time"

	"clash-manager/internal/model"
	"clash-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type SubscriptionSourceHandler struct {
	Service *service.SubscriptionService
}

func NewSubscriptionSourceHandler() *SubscriptionSourceHandler {
	return &SubscriptionSourceHandler{Service: service.NewSubscriptionService()}
}

func (h *SubscriptionSourceHandler) ListSources(c *gin.Context) {
	sources, err := h.Service.ListSources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sources)
}

func (h *SubscriptionSourceHandler) GetSource(c *gin.Context) {
	id := c.Param("id")
	var idUint uint
	if _, err := fmt.Sscanf(id, "%d", &idUint); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	sources, err := h.Service.ListSources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if idUint < 1 || int(idUint) > len(sources) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Source not found"})
		return
	}

	c.JSON(http.StatusOK, sources[idUint-1])
}

func (h *SubscriptionSourceHandler) CreateSource(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		URL      string `json:"url" binding:"required,url"`
		SyncMode string `json:"syncMode"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	source := &service.SubscriptionSource{
		Name:     req.Name,
		URL:      req.URL,
		SyncMode: req.SyncMode,
	}

	if source.SyncMode == "" {
		source.SyncMode = "append"
	}

	if err := h.Service.CreateSource(source); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create source: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, source)
}

func (h *SubscriptionSourceHandler) UpdateSource(c *gin.Context) {
	id := c.Param("id")
	var idUint uint
	if _, err := fmt.Sscanf(id, "%d", &idUint); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		Name     string `json:"name"`
		URL      string `json:"url" binding:"required,url"`
		SyncMode string `json:"syncMode"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	source := &service.SubscriptionSource{
		Name:     req.Name,
		URL:      req.URL,
		SyncMode: req.SyncMode,
	}

	if err := h.Service.UpdateSource(idUint, source); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update source: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, source)
}

func (h *SubscriptionSourceHandler) DeleteSource(c *gin.Context) {
	id := c.Param("id")
	var idUint uint
	if _, err := fmt.Sscanf(id, "%d", &idUint); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.Service.DeleteSource(idUint); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete source: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *SubscriptionSourceHandler) SyncSource(c *gin.Context) {
	id := c.Param("id")
	var idUint uint
	if _, err := fmt.Sscanf(id, "%d", &idUint); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.Service.SyncSource(idUint); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync source: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sync completed successfully",
	})
}

func (h *SubscriptionSourceHandler) TestSource(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required,url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nodes, err := service.ParseSubscription(req.URL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"nodesCount": len(nodes),
		"preview":    nodesPreview(nodes),
	})
}

func nodesPreview(nodes []model.Node) []map[string]interface{} {
	preview := make([]map[string]interface{}, 0, len(nodes))
	maxPreview := 10

	for i, node := range nodes {
		if i >= maxPreview {
			break
		}
		preview = append(preview, map[string]interface{}{
			"name":   node.Name,
			"type":   node.Type,
			"server": node.Server,
			"port":   node.Port,
		})
	}

	return preview
}

func init() {
	time.Now()
}
