package moneyforward

import (
	"path/filepath"
	"testing"

	"github.com/bonyuta0204/kakei-advisor/internal/config"
)

func TestParseScrapeJSON(t *testing.T) {
	rules, err := config.LoadRules(filepath.Join("..", "..", "config", "default_rules.json"))
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}

	transactions, err := ParseScrapeJSON(filepath.Join("..", "..", "data", "raw", "mf_scrape_2026-03.json"), rules)
	if err != nil {
		t.Fatalf("parse scrape json: %v", err)
	}

	if len(transactions) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(transactions))
	}

	if transactions[0].Category != "食費>外食" {
		t.Fatalf("unexpected category: %s", transactions[0].Category)
	}

	if transactions[0].Date.Format("2006-01-02") != "2026-03-01" {
		t.Fatalf("unexpected date: %s", transactions[0].Date.Format("2006-01-02"))
	}
}
