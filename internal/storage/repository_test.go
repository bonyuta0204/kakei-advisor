package storage

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/bonyuta0204/kakei-advisor/internal/domain"
)

func TestInsertTransactionsAndMonthlySummary(t *testing.T) {
	repo, err := Open(filepath.Join(t.TempDir(), "finance.duckdb"))
	if err != nil {
		t.Fatalf("open repo: %v", err)
	}
	defer repo.Close()

	txns := []domain.Transaction{
		{
			Date:          time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			Amount:        -3200,
			Merchant:      "スーパー",
			Category:      "食費",
			PaymentMethod: "楽天カード（共用）",
			SourceAccount: "楽天カード（共用）",
			Owner:         "shared",
			ExpenseType:   "variable",
			IsTransfer:    false,
			RawSourceFile: "sample.csv",
			RawRowHash:    "hash-1",
			RawJSON:       `{"row":1}`,
		},
		{
			Date:          time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
			Amount:        -5000,
			Merchant:      "PASMOチャージ",
			Category:      "交通費",
			PaymentMethod: "みずほ銀行（個人）",
			SourceAccount: "みずほ銀行（個人）",
			Owner:         "self",
			ExpenseType:   "variable",
			IsTransfer:    true,
			RawSourceFile: "sample.csv",
			RawRowHash:    "hash-2",
			RawJSON:       `{"row":2}`,
		},
	}

	inserted, err := repo.InsertTransactions(txns)
	if err != nil {
		t.Fatalf("insert transactions: %v", err)
	}
	if inserted != 2 {
		t.Fatalf("expected 2 inserted rows, got %d", inserted)
	}

	inserted, err = repo.InsertTransactions(txns)
	if err != nil {
		t.Fatalf("reinsert transactions: %v", err)
	}
	if inserted != 0 {
		t.Fatalf("expected dedupe to skip inserts, got %d", inserted)
	}

	summary, err := repo.MonthlySummary("2026-03")
	if err != nil {
		t.Fatalf("monthly summary: %v", err)
	}
	if summary.TransactionCnt != 1 {
		t.Fatalf("expected 1 non-transfer transaction, got %d", summary.TransactionCnt)
	}
	if summary.TotalAmount != 3200 {
		t.Fatalf("expected total amount 3200, got %d", summary.TotalAmount)
	}
}
