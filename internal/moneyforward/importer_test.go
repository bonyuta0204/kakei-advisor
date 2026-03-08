package moneyforward

import (
	"path/filepath"
	"testing"

	"github.com/bonyuta0204/kakei-advisor/internal/config"
)

func TestParseCSV(t *testing.T) {
	rules, err := config.LoadRules(filepath.Join("..", "..", "config", "default_rules.json"))
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}

	transactions, err := ParseCSV(filepath.Join("..", "..", "testdata", "moneyforward_sample.csv"), rules)
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	if len(transactions) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(transactions))
	}

	if transactions[0].Owner != "shared" {
		t.Fatalf("expected shared owner, got %s", transactions[0].Owner)
	}
	if !transactions[1].IsTransfer {
		t.Fatal("expected PASMO charge to be treated as transfer")
	}
	if transactions[2].ExpenseType != "fixed" {
		t.Fatalf("expected fixed expense type, got %s", transactions[2].ExpenseType)
	}
}
