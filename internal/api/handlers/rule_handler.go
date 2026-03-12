package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"clash-manager/internal/model"
	"clash-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type RuleHandler struct {
	Service *service.RuleService
}

func NewRuleHandler() *RuleHandler {
	return &RuleHandler{Service: service.NewRuleService()}
}

func (h *RuleHandler) ListRules(c *gin.Context) {
	rules, err := h.Service.ListRules()
	if err != nil {
		fmt.Printf("[ListRules] Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules":      rules,
		"total":      len(rules),
		"page":       1,
		"pageSize":   len(rules),
		"totalPages": 1,
	})
}

func (h *RuleHandler) CreateRule(c *gin.Context) {
	var rule model.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Service.CreateRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rule)
}

func (h *RuleHandler) DeleteRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := h.Service.DeleteRule(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *RuleHandler) UpdateRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var rule model.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Service.UpdateRule(uint(id), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *RuleHandler) ImportRules(c *gin.Context) {
	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lines := strings.Split(req.Content, "\n")
	importCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}

		rule := &model.Rule{
			Type:    strings.TrimSpace(parts[0]),
			Payload: strings.TrimSpace(parts[1]),
			Target:  strings.TrimSpace(parts[2]),
		}

		if len(parts) >= 4 && strings.TrimSpace(parts[3]) == "no-resolve" {
			rule.NoResolve = true
		}

		if err := h.Service.CreateRule(rule); err != nil {
			fmt.Printf("[ImportRules] Failed to import rule: %v\n", err)
			continue
		}
		importCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rules imported successfully",
		"count":   importCount,
	})
}

func (h *RuleHandler) GetTags(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"tags": []string{},
	})
}
