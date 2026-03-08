package config

import "testing"

func TestApplyStringRules(t *testing.T) {
	rules := []StringRule{
		{Field: "payment_method", Match: "contains", Pattern: "共用", Value: "shared"},
	}

	got := ApplyStringRules(rules, map[string]string{"payment_method": "楽天カード（共用）"}, "self")
	if got != "shared" {
		t.Fatalf("expected shared, got %s", got)
	}
}

func TestApplyBoolRules(t *testing.T) {
	rules := []BoolRule{
		{Field: "merchant", Match: "contains", Pattern: "PASMOチャージ", Value: true},
	}

	got := ApplyBoolRules(rules, map[string]string{"merchant": "PASMOチャージ"}, false)
	if !got {
		t.Fatal("expected transfer rule to match")
	}
}
