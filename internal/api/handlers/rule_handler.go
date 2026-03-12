package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"clash-manager/internal/model"
	"clash-manager/internal/repository"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type RuleHandler struct {
	Repo *repository.RuleRepository
}

func NewRuleHandler() *RuleHandler {
	return &RuleHandler{Repo: &repository.RuleRepository{}}
}

func (h *RuleHandler) ListRules(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	ruleType := c.Query("type")
	keyword := c.Query("keyword")
	target := c.Query("target")
	tag := c.Query("tag")

	// Debug logging
	fmt.Printf("[ListRules] Query params - page: %d, pageSize: %d, type: %s, keyword: %s, target: %s, tag: %s\n",
		page, pageSize, ruleType, keyword, target, tag)

	// Validate parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}

	params := &repository.RuleListParams{
		Page:     page,
		PageSize: pageSize,
		Type:     ruleType,
		Keyword:  keyword,
		Target:   target,
		Tag:      tag,
	}

	result, err := h.Repo.FindWithPagination(params)
	if err != nil {
		fmt.Printf("[ListRules] Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("[ListRules] Result - Total: %d, Page: %d, PageSize: %d, TotalPages: %d, RulesCount: %d\n",
		result.Total, result.Page, result.PageSize, result.TotalPages, len(result.Rules))

	c.JSON(http.StatusOK, result)
}

func (h *RuleHandler) CreateRule(c *gin.Context) {
	var rule model.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Repo.Create(&rule); err != nil {
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
	if err := h.Repo.Delete(uint(id)); err != nil {
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

	// Get existing rule
	existingRule, err := h.Repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	// Update fields (preserve ID)
	rule.ID = existingRule.ID
	if err := h.Repo.Update(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rule)
}

// ImportRequest represents the import request structure
type ImportRequest struct {
	Content string `json:"content"`
}

func (h *RuleHandler) ImportRules(c *gin.Context) {
	var req ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse YAML content
	var yamlConfig map[string]interface{}
	if err := yaml.Unmarshal([]byte(req.Content), &yamlConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid YAML format: " + err.Error()})
		return
	}

	// Extract rules section
	rulesSection, ok := yamlConfig["rules"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No rules section found in YAML"})
		return
	}

	rulesList, ok := rulesSection.([]interface{})
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rules format"})
		return
	}

	// Fetch all nodes and groups for name matching
	nodeRepo := &repository.NodeRepository{}
	nodes, err := nodeRepo.FindAll()
	if err != nil {
		fmt.Printf("[ImportRules] Failed to fetch nodes: %v\n", err)
	}
	groupRepo := &repository.GroupRepository{}
	groups, err := groupRepo.FindAll()
	if err != nil {
		fmt.Printf("[ImportRules] Failed to fetch groups: %v\n", err)
	}

	// Build name to ID maps
	nodeNameToID := make(map[string]uint)
	for _, n := range nodes {
		nodeNameToID[n.Name] = n.ID
	}
	groupNameToID := make(map[string]uint)
	for _, g := range groups {
		groupNameToID[g.Name] = g.ID
	}

	var rulesToImport []model.Rule
	priority := 0

	for _, ruleItem := range rulesList {
		ruleStr, ok := ruleItem.(string)
		if !ok {
			continue
		}

		// Parse rule string: TYPE,Payload,Target[,no-resolve]
		parts := strings.Split(ruleStr, ",")
		if len(parts) < 3 {
			continue
		}

		targetName := strings.TrimSpace(parts[2])

		rule := model.Rule{
			Type:     strings.TrimSpace(parts[0]),
			Payload:  strings.TrimSpace(parts[1]),
			Priority: priority,
		}

		// Check for no-resolve option
		if len(parts) >= 4 && strings.TrimSpace(parts[3]) == "no-resolve" {
			rule.NoResolve = true
		}

		// Try to match target to nodes or groups
		// Priority: node > group > built-in
		if nodeID, ok := nodeNameToID[targetName]; ok {
			rule.TargetType = "node"
			rule.Target = fmt.Sprintf("%d", nodeID)
		} else if groupID, ok := groupNameToID[targetName]; ok {
			rule.TargetType = "group"
			rule.Target = fmt.Sprintf("%d", groupID)
		} else {
			// Built-in target (DIRECT, PROXY, REJECT, etc.)
			rule.Target = targetName
			rule.TargetType = "builtin"
		}

		rulesToImport = append(rulesToImport, rule)
		priority++
	}

	if len(rulesToImport) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid rules found"})
		return
	}

	// Batch create rules
	if err := h.Repo.BatchCreate(&rulesToImport); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rules imported successfully",
		"count":   len(rulesToImport),
	})
}

// GetTags returns all unique tags from rules
func (h *RuleHandler) GetTags(c *gin.Context) {
	tags, err := h.Repo.GetAllTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"tags": tags,
	})
}
