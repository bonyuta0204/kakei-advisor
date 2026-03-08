package report

import (
	"strings"
	"testing"

	"github.com/bonyuta0204/kakei-advisor/internal/domain"
)

func TestRenderMonthly(t *testing.T) {
	summary := domain.MonthlySummary{
		Month:          "2026-03",
		TransactionCnt: 2,
		TotalAmount:    4690,
		CategoryTotals: []domain.Breakdown{
			{Label: "食費", Amount: 3200},
			{Label: "サブスク", Amount: 1490},
		},
		OwnerTotals: []domain.Breakdown{
			{Label: "shared", Amount: 3200},
			{Label: "self", Amount: 1490},
		},
	}

	content := RenderMonthly(summary)
	if !strings.Contains(content, "# 家計レビュー 2026-03") {
		t.Fatal("report header not rendered")
	}
	if !strings.Contains(content, "| 食費 | 3,200 |") {
		t.Fatal("category breakdown not rendered")
	}
}
