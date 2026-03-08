package report

import (
	"fmt"
	"strings"

	"github.com/bonyuta0204/kakei-advisor/internal/domain"
)

func RenderMonthly(summary domain.MonthlySummary) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# 家計レビュー %s\n\n", summary.Month)
	fmt.Fprintf(&b, "- 支出合計: %s円\n", formatYen(summary.TotalAmount))
	fmt.Fprintf(&b, "- 対象明細数: %d件\n", summary.TransactionCnt)
	fmt.Fprintf(&b, "- 資金移動扱いの明細は除外済み\n\n")

	b.WriteString("## カテゴリ別\n\n")
	b.WriteString("| カテゴリ | 金額 |\n")
	b.WriteString("|---|---:|\n")
	for _, entry := range summary.CategoryTotals {
		fmt.Fprintf(&b, "| %s | %s |\n", normalizeLabel(entry.Label), formatYen(entry.Amount))
	}

	b.WriteString("\n## owner別\n\n")
	b.WriteString("| owner | 金額 |\n")
	b.WriteString("|---|---:|\n")
	for _, entry := range summary.OwnerTotals {
		fmt.Fprintf(&b, "| %s | %s |\n", normalizeLabel(entry.Label), formatYen(entry.Amount))
	}

	return b.String()
}

func formatYen(amount int64) string {
	raw := fmt.Sprintf("%d", amount)
	if len(raw) <= 3 {
		return raw
	}

	var out []byte
	prefix := len(raw) % 3
	if prefix > 0 {
		out = append(out, raw[:prefix]...)
		if len(raw) > prefix {
			out = append(out, ',')
		}
	}

	for i := prefix; i < len(raw); i += 3 {
		out = append(out, raw[i:i+3]...)
		if i+3 < len(raw) {
			out = append(out, ',')
		}
	}

	return string(out)
}

func normalizeLabel(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(未設定)"
	}
	return value
}
