package service

import (
	"fmt"
	"strings"

	"clash-manager/internal/config"
	"clash-manager/internal/model"
)

type RuleService struct{}

func NewRuleService() *RuleService {
	return &RuleService{}
}

func (s *RuleService) ListRules() ([]model.Rule, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	rules := make([]model.Rule, 0, len(cfg.Rules))
	for i, ruleStr := range cfg.Rules {
		parts := strings.Split(ruleStr, ",")
		if len(parts) < 3 {
			continue
		}

		rule := model.Rule{
			ID:       uint(i + 1),
			Type:     strings.TrimSpace(parts[0]),
			Payload:  strings.TrimSpace(parts[1]),
			Target:   strings.TrimSpace(parts[2]),
			Priority: i,
		}

		if len(parts) >= 4 && strings.TrimSpace(parts[3]) == "no-resolve" {
			rule.NoResolve = true
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func (s *RuleService) CreateRule(rule *model.Rule) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	ruleStr := s.buildRuleString(rule)
	cfg.Rules = append(cfg.Rules, ruleStr)

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	return nil
}

func (s *RuleService) UpdateRule(id uint, rule *model.Rule) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	if id < 1 || int(id) > len(cfg.Rules) {
		return fmt.Errorf("invalid rule id: %d", id)
	}

	cfg.Rules[id-1] = s.buildRuleString(rule)

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	return nil
}

func (s *RuleService) DeleteRule(id uint) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	if id < 1 || int(id) > len(cfg.Rules) {
		return fmt.Errorf("invalid rule id: %d", id)
	}

	cfg.Rules = append(cfg.Rules[:id-1], cfg.Rules[id:]...)

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	return nil
}

func (s *RuleService) buildRuleString(rule *model.Rule) string {
	ruleStr := fmt.Sprintf("%s,%s,%s", rule.Type, rule.Payload, rule.Target)
	if rule.NoResolve {
		ruleStr += ",no-resolve"
	}
	return ruleStr
}
