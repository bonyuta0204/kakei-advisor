package config

import (
	"encoding/json"
	"os"
	"strings"
)

type StringRule struct {
	Field   string `json:"field"`
	Match   string `json:"match"`
	Pattern string `json:"pattern"`
	Value   string `json:"value"`
}

type BoolRule struct {
	Field   string `json:"field"`
	Match   string `json:"match"`
	Pattern string `json:"pattern"`
	Value   bool   `json:"value"`
}

type Rules struct {
	OwnerRules       []StringRule `json:"owner_rules"`
	ExpenseTypeRules []StringRule `json:"expense_type_rules"`
	TransferRules    []BoolRule   `json:"transfer_rules"`
}

func LoadRules(path string) (Rules, error) {
	var rules Rules

	content, err := os.ReadFile(path)
	if err != nil {
		return rules, err
	}

	if err := json.Unmarshal(content, &rules); err != nil {
		return rules, err
	}

	return rules, nil
}

func ApplyStringRules(rules []StringRule, fields map[string]string, fallback string) string {
	for _, rule := range rules {
		if matches(rule.Match, fields[rule.Field], rule.Pattern) {
			return rule.Value
		}
	}
	return fallback
}

func ApplyBoolRules(rules []BoolRule, fields map[string]string, fallback bool) bool {
	for _, rule := range rules {
		if matches(rule.Match, fields[rule.Field], rule.Pattern) {
			return rule.Value
		}
	}
	return fallback
}

func matches(matchType, value, pattern string) bool {
	left := strings.TrimSpace(value)
	right := strings.TrimSpace(pattern)

	switch matchType {
	case "exact":
		return left == right
	case "contains":
		return strings.Contains(left, right)
	default:
		return false
	}
}
