package repository

import (
	"fmt"

	"clash-manager/internal/model"
)

type RuleListParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Type     string `json:"type"`
	Keyword  string `json:"keyword"`
	Target   string `json:"target"` // 过滤目标名称
	Tag      string `json:"tag"`    // 过滤标签
}

type RuleListResult struct {
	Rules      []model.Rule `json:"rules"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"pageSize"`
	TotalPages int          `json:"totalPages"`
}

type RuleRepository struct{}

func (r *RuleRepository) Create(rule *model.Rule) error {
	return DB.Create(rule).Error
}

func (r *RuleRepository) BatchCreate(rules *[]model.Rule) error {
	if len(*rules) == 0 {
		return nil
	}
	return DB.Create(&rules).Error
}

func (r *RuleRepository) FindAll() ([]model.Rule, error) {
	var rules []model.Rule
	// Order by precedence if we add a Priority field, otherwise ID
	err := DB.Order("priority asc, id asc").Find(&rules).Error
	return rules, err
}

// FindWithPagination returns rules with pagination and filtering
func (r *RuleRepository) FindWithPagination(params *RuleListParams) (*RuleListResult, error) {
	var rules []model.Rule
	var total int64

	query := DB.Model(&model.Rule{})

	fmt.Printf("[FindWithPagination] Input params - Page: %d, PageSize: %d, Type: %s, Keyword: %s, Target: %s, Tag: %s\n",
		params.Page, params.PageSize, params.Type, params.Keyword, params.Target, params.Tag)

	// Filter by type
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
		fmt.Printf("[FindWithPagination] Applied type filter: %s\n", params.Type)
	}

	// Filter by keyword (search in Payload, Target, Remark)
	if params.Keyword != "" {
		keyword := "%" + params.Keyword + "%"
		query = query.Where("payload LIKE ? OR target LIKE ? OR remark LIKE ?", keyword, keyword, keyword)
		fmt.Printf("[FindWithPagination] Applied keyword filter: %s\n", params.Keyword)
	}

	// Filter by target name
	if params.Target != "" {
		query = query.Where("target = ?", params.Target)
		fmt.Printf("[FindWithPagination] Applied target filter: %s\n", params.Target)
	}

	// Filter by tag
	if params.Tag != "" {
		query = query.Where("tag = ?", params.Tag)
		fmt.Printf("[FindWithPagination] Applied tag filter: %s\n", params.Tag)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		fmt.Printf("[FindWithPagination] Count error: %v\n", err)
		return nil, err
	}

	fmt.Printf("[FindWithPagination] Total count: %d\n", total)

	// Calculate offset
	offset := (params.Page - 1) * params.PageSize

	// Fetch data
	if err := query.Order("priority asc, id asc").Offset(offset).Limit(params.PageSize).Find(&rules).Error; err != nil {
		fmt.Printf("[FindWithPagination] Find error: %v\n", err)
		return nil, err
	}

	fmt.Printf("[FindWithPagination] Fetched %d rules\n", len(rules))

	// Calculate total pages
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	return &RuleListResult{
		Rules:      rules,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *RuleRepository) Delete(id uint) error {
	return DB.Delete(&model.Rule{}, id).Error
}

func (r *RuleRepository) Update(rule *model.Rule) error {
	return DB.Save(rule).Error
}

func (r *RuleRepository) FindByID(id uint) (*model.Rule, error) {
	var rule model.Rule
	err := DB.First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetAllTags returns all unique non-empty tags from rules
func (r *RuleRepository) GetAllTags() ([]string, error) {
	var tags []string
	err := DB.Model(&model.Rule{}).
		Where("tag != ''").
		Where("tag IS NOT NULL").
		Distinct("tag").
		Pluck("tag", &tags).
		Error
	return tags, err
}

// DeleteInvalidRules deletes rules that reference non-existent nodes or groups
func (r *RuleRepository) DeleteInvalidRules(validNodeIDs, validGroupIDs []uint) (int64, error) {
	nodeIDMap := make(map[uint]bool)
	for _, id := range validNodeIDs {
		nodeIDMap[id] = true
	}
	groupIDMap := make(map[uint]bool)
	for _, id := range validGroupIDs {
		groupIDMap[id] = true
	}

	// Find all rules with invalid target references
	var rulesToDelete []model.Rule
	err := DB.Where("target_id > 0").Find(&rulesToDelete).Error
	if err != nil {
		return 0, err
	}

	var invalidRuleIDs []uint
	for _, rule := range rulesToDelete {
		isValid := false
		if rule.TargetType == "node" && nodeIDMap[rule.TargetID] {
			isValid = true
		} else if rule.TargetType == "group" && groupIDMap[rule.TargetID] {
			isValid = true
		}

		if !isValid {
			invalidRuleIDs = append(invalidRuleIDs, rule.ID)
		}
	}

	// Delete invalid rules
	if len(invalidRuleIDs) > 0 {
		result := DB.Delete(&model.Rule{}, invalidRuleIDs)
		return result.RowsAffected, result.Error
	}

	return 0, nil
}
