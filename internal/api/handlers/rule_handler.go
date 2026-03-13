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

	// 获取搜索和过滤参数
	keyword := c.Query("keyword")
	filterType := c.Query("type")
	filterTarget := c.Query("target")

	// 过滤规则
	filteredRules := make([]model.Rule, 0)
	for _, rule := range rules {
		// 处理规则类型（去除可能的前缀如 "- "）
		ruleType := strings.TrimPrefix(rule.Type, "- ")
		ruleType = strings.TrimSpace(ruleType)

		// 关键词搜索 - 搜索匹配内容和目标
		if keyword != "" {
			keywordLower := strings.ToLower(keyword)
			payloadLower := strings.ToLower(rule.Payload)
			targetLower := strings.ToLower(rule.Target)
			typeLower := strings.ToLower(ruleType)
			if !strings.Contains(payloadLower, keywordLower) && 
			   !strings.Contains(targetLower, keywordLower) && 
			   !strings.Contains(typeLower, keywordLower) {
				continue
			}
		}

		// 类型过滤
		if filterType != "" && ruleType != filterType {
			continue
		}

		// 目标过滤
		if filterTarget != "" && rule.Target != filterTarget {
			continue
		}

		filteredRules = append(filteredRules, rule)
	}

	c.JSON(http.StatusOK, gin.H{
		"rules":      filteredRules,
		"total":      len(filteredRules),
		"page":       1,
		"pageSize":   len(filteredRules),
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
	updateCount := 0

	// 只在开始时调用一次 ListRules()
	existingRules, err := h.Service.ListRules()
	if err != nil {
		fmt.Printf("[ImportRules] Failed to list rules: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list rules"})
		return
	}

	// 创建规则映射，用于快速查找 - 以匹配内容(Payload)为键
	ruleMap := make(map[string]model.Rule)
	for _, rule := range existingRules {
		key := rule.Payload
		ruleMap[key] = rule
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}

		// 去除类型字段中可能的前缀（如 "- "）
		ruleType := strings.TrimSpace(parts[0])
		ruleType = strings.TrimPrefix(ruleType, "- ")
		ruleType = strings.TrimSpace(ruleType)

		// 验证规则类型是否在支持的列表中
		validTypes := map[string]bool{
			"DOMAIN-SUFFIX": true,
			"DOMAIN":        true,
			"DOMAIN-KEYWORD": true,
			"IP-CIDR":       true,
			"SRC-IP-CIDR":   true,
			"GEOIP":         true,
			"MATCH":         true,
		}
		if !validTypes[ruleType] {
			fmt.Printf("[ImportRules] Skipping invalid rule type: %s\n", ruleType)
			continue
		}

		rule := &model.Rule{
			Type:    ruleType,
			Payload: strings.TrimSpace(parts[1]),
			Target:  strings.TrimSpace(parts[2]),
		}

		if len(parts) >= 4 && strings.TrimSpace(parts[3]) == "no-resolve" {
			rule.NoResolve = true
		}

		// 以匹配内容(Payload)为键进行查找
		key := rule.Payload

		// 检查是否存在相同匹配内容的规则
		if existingRule, found := ruleMap[key]; found {
			// 更新现有规则
			if err := h.Service.UpdateRule(existingRule.ID, rule); err != nil {
				fmt.Printf("[ImportRules] Failed to update rule: %v\n", err)
				continue
			}
			updateCount++
			fmt.Printf("[ImportRules] Updated rule: Type=%s, Payload=%s, Target=%s, NoResolve=%v\n", rule.Type, rule.Payload, rule.Target, rule.NoResolve)
		} else {
			// 不存在则添加新规则
			if err := h.Service.CreateRule(rule); err != nil {
				fmt.Printf("[ImportRules] Failed to import rule: %v\n", err)
				continue
			}
			importCount++
			fmt.Printf("[ImportRules] Imported new rule: Type=%s, Payload=%s, Target=%s, NoResolve=%v\n", rule.Type, rule.Payload, rule.Target, rule.NoResolve)
		}
	}

	fmt.Printf("[ImportRules] Total: imported=%d, updated=%d\n", importCount, updateCount)
	c.JSON(http.StatusOK, gin.H{
		"message":      "Rules imported successfully",
		"import_count": importCount,
		"update_count": updateCount,
	})
}

func (h *RuleHandler) GetTags(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"tags": []string{},
	})
}
