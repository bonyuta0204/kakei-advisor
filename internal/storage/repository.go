package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/duckdb/duckdb-go/v2"

	"github.com/bonyuta0204/kakei-advisor/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func Open(dbPath string) (*Repository, error) {
	if dbPath != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, err
	}

	repo := &Repository{db: db}
	if err := repo.ensureSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return repo, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) InsertTransactions(transactions []domain.Transaction) (int, error) {
	stmt, err := r.db.Prepare(`
		INSERT INTO transactions (
			txn_date,
			amount,
			merchant,
			category,
			payment_method,
			source_account,
			owner,
			expense_type,
			is_transfer,
			note,
			raw_source_file,
			raw_row_hash,
			raw_json,
			created_at
		)
		SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		WHERE NOT EXISTS (
			SELECT 1 FROM transactions WHERE raw_row_hash = ?
		)
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	inserted := 0
	for _, txn := range transactions {
		result, err := stmt.Exec(
			txn.Date.Format("2006-01-02"),
			txn.Amount,
			txn.Merchant,
			txn.Category,
			txn.PaymentMethod,
			txn.SourceAccount,
			txn.Owner,
			txn.ExpenseType,
			txn.IsTransfer,
			txn.Note,
			txn.RawSourceFile,
			txn.RawRowHash,
			txn.RawJSON,
			time.Now(),
			txn.RawRowHash,
		)
		if err != nil {
			return inserted, err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return inserted, err
		}
		inserted += int(affected)
	}

	return inserted, nil
}

func (r *Repository) MonthlySummary(month string) (domain.MonthlySummary, error) {
	start, end, err := monthRange(month)
	if err != nil {
		return domain.MonthlySummary{}, err
	}

	summary := domain.MonthlySummary{Month: month}

	row := r.db.QueryRow(`
		SELECT COALESCE(COUNT(*), 0), COALESCE(SUM(ABS(amount)), 0)
		FROM transactions
		WHERE txn_date >= ? AND txn_date < ? AND is_transfer = FALSE
	`, start, end)
	if err := row.Scan(&summary.TransactionCnt, &summary.TotalAmount); err != nil {
		return domain.MonthlySummary{}, err
	}

	categoryRows, err := r.db.Query(`
		SELECT category, COALESCE(SUM(ABS(amount)), 0) AS total_amount
		FROM transactions
		WHERE txn_date >= ? AND txn_date < ? AND is_transfer = FALSE
		GROUP BY category
		ORDER BY total_amount DESC, category ASC
	`, start, end)
	if err != nil {
		return domain.MonthlySummary{}, err
	}
	defer categoryRows.Close()

	for categoryRows.Next() {
		var entry domain.Breakdown
		if err := categoryRows.Scan(&entry.Label, &entry.Amount); err != nil {
			return domain.MonthlySummary{}, err
		}
		summary.CategoryTotals = append(summary.CategoryTotals, entry)
	}

	ownerRows, err := r.db.Query(`
		SELECT owner, COALESCE(SUM(ABS(amount)), 0) AS total_amount
		FROM transactions
		WHERE txn_date >= ? AND txn_date < ? AND is_transfer = FALSE
		GROUP BY owner
		ORDER BY total_amount DESC, owner ASC
	`, start, end)
	if err != nil {
		return domain.MonthlySummary{}, err
	}
	defer ownerRows.Close()

	for ownerRows.Next() {
		var entry domain.Breakdown
		if err := ownerRows.Scan(&entry.Label, &entry.Amount); err != nil {
			return domain.MonthlySummary{}, err
		}
		summary.OwnerTotals = append(summary.OwnerTotals, entry)
	}

	return summary, nil
}

func (r *Repository) ensureSchema() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			txn_date DATE NOT NULL,
			amount BIGINT NOT NULL,
			merchant VARCHAR NOT NULL,
			category VARCHAR,
			payment_method VARCHAR,
			source_account VARCHAR,
			owner VARCHAR NOT NULL,
			expense_type VARCHAR NOT NULL,
			is_transfer BOOLEAN NOT NULL,
			note VARCHAR,
			raw_source_file VARCHAR NOT NULL,
			raw_row_hash VARCHAR NOT NULL,
			raw_json VARCHAR NOT NULL,
			created_at TIMESTAMP NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_raw_row_hash ON transactions(raw_row_hash);`)
	return err
}

func monthRange(month string) (string, string, error) {
	start, err := time.Parse("2006-01", month)
	if err != nil {
		return "", "", fmt.Errorf("month must be YYYY-MM: %w", err)
	}
	end := start.AddDate(0, 1, 0)
	return start.Format("2006-01-02"), end.Format("2006-01-02"), nil
}
