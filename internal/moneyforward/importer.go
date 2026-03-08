package moneyforward

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bonyuta0204/kakei-advisor/internal/config"
	"github.com/bonyuta0204/kakei-advisor/internal/domain"
)

var headerAliases = map[string][]string{
	"date":           {"日付", "利用日", "計算対象", "更新日"},
	"amount":         {"金額", "金額（円）", "利用金額", "入出金額"},
	"merchant":       {"内容", "摘要", "店舗名", "利用店名", "明細"},
	"category":       {"大項目", "カテゴリ", "項目"},
	"payment_method": {"保有金融機関", "金融機関", "決済手段"},
	"source_account": {"口座", "保有金融機関", "引き落とし口座"},
	"note":           {"メモ", "備考", "内容"},
}

func ParseCSV(path string, rules config.Rules) ([]domain.Transaction, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("csv has no data rows: %s", path)
	}

	headerIndex := indexHeaders(rows[0])
	if _, ok := headerIndex["date"]; !ok {
		return nil, fmt.Errorf("required header not found: date")
	}
	if _, ok := headerIndex["amount"]; !ok {
		return nil, fmt.Errorf("required header not found: amount")
	}

	transactions := make([]domain.Transaction, 0, len(rows)-1)
	for _, row := range rows[1:] {
		if rowIsEmpty(row) {
			continue
		}

		raw := buildRawRow(rows[0], row)
		dateValue := valueFor("date", row, headerIndex)
		amountValue := valueFor("amount", row, headerIndex)
		merchantValue := firstNonEmpty(
			valueFor("merchant", row, headerIndex),
			valueFor("note", row, headerIndex),
			valueFor("category", row, headerIndex),
			"(unknown)",
		)
		categoryValue := valueFor("category", row, headerIndex)
		paymentMethodValue := firstNonEmpty(
			valueFor("payment_method", row, headerIndex),
			valueFor("source_account", row, headerIndex),
		)
		sourceAccountValue := firstNonEmpty(
			valueFor("source_account", row, headerIndex),
			paymentMethodValue,
		)
		noteValue := valueFor("note", row, headerIndex)

		date, err := parseDate(dateValue)
		if err != nil {
			return nil, fmt.Errorf("parse date %q: %w", dateValue, err)
		}
		amount, err := parseAmount(amountValue)
		if err != nil {
			return nil, fmt.Errorf("parse amount %q: %w", amountValue, err)
		}

		fields := map[string]string{
			"merchant":       merchantValue,
			"category":       categoryValue,
			"payment_method": paymentMethodValue,
			"source_account": sourceAccountValue,
			"note":           noteValue,
		}

		rawJSON, err := json.Marshal(raw)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, domain.Transaction{
			Date:          date,
			Amount:        amount,
			Merchant:      merchantValue,
			Category:      categoryValue,
			PaymentMethod: paymentMethodValue,
			SourceAccount: sourceAccountValue,
			Owner:         config.ApplyStringRules(rules.OwnerRules, fields, "self"),
			ExpenseType:   config.ApplyStringRules(rules.ExpenseTypeRules, fields, "variable"),
			IsTransfer:    config.ApplyBoolRules(rules.TransferRules, fields, false),
			Note:          noteValue,
			RawSourceFile: path,
			RawRowHash:    hashCanonicalRow(date, amount, merchantValue, categoryValue, paymentMethodValue, sourceAccountValue, noteValue),
			RawJSON:       string(rawJSON),
		})
	}

	return transactions, nil
}

func indexHeaders(header []string) map[string]int {
	indexed := make(map[string]int)
	normalized := make(map[string]int, len(header))

	for i, column := range header {
		clean := strings.TrimSpace(strings.TrimPrefix(column, "\uFEFF"))
		normalized[clean] = i
	}

	for logical, aliases := range headerAliases {
		for _, alias := range aliases {
			if idx, ok := normalized[alias]; ok {
				indexed[logical] = idx
				break
			}
		}
	}

	return indexed
}

func valueFor(logical string, row []string, headerIndex map[string]int) string {
	idx, ok := headerIndex[logical]
	if !ok || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func buildRawRow(header, row []string) map[string]string {
	raw := make(map[string]string, len(header))
	for i, key := range header {
		if i >= len(row) {
			raw[key] = ""
			continue
		}
		raw[strings.TrimSpace(strings.TrimPrefix(key, "\uFEFF"))] = strings.TrimSpace(row[i])
	}
	return raw
}

func rowIsEmpty(row []string) bool {
	for _, col := range row {
		if strings.TrimSpace(col) != "" {
			return false
		}
	}
	return true
}

func parseDate(raw string) (time.Time, error) {
	formats := []string{
		"2006/01/02",
		"2006/1/2",
		"2006-01-02",
		"2006.01.02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, strings.TrimSpace(raw)); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported date format")
}

func parseAmount(raw string) (int64, error) {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.ReplaceAll(cleaned, "円", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")

	if cleaned == "" {
		return 0, fmt.Errorf("empty amount")
	}

	return strconv.ParseInt(cleaned, 10, 64)
}

func hashCanonicalRow(date time.Time, amount int64, merchant, category, paymentMethod, sourceAccount, note string) string {
	canonical := fmt.Sprintf(
		"%s|%d|%s|%s|%s|%s|%s",
		date.Format("2006-01-02"),
		amount,
		merchant,
		category,
		paymentMethod,
		sourceAccount,
		note,
	)
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
