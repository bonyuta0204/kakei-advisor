package moneyforward

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bonyuta0204/kakei-advisor/internal/config"
	"github.com/bonyuta0204/kakei-advisor/internal/domain"
)

type scrapePayload struct {
	Range string      `json:"range"`
	Rows  []scrapeRow `json:"rows"`
}

type scrapeRow struct {
	TransactionID  string  `json:"transaction_id"`
	Date           string  `json:"date"`
	Merchant       string  `json:"merchant"`
	Amount         string  `json:"amount"`
	PaymentMethod  string  `json:"payment_method"`
	LargeCategory  string  `json:"large_category"`
	MiddleCategory string  `json:"middle_category"`
	Memo           *string `json:"memo"`
}

var monthDayPattern = regexp.MustCompile(`^(\d{2})/(\d{2})`)

func ParseScrapeJSON(path string, rules config.Rules) ([]domain.Transaction, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var payload scrapePayload
	if err := json.Unmarshal(content, &payload); err != nil {
		return nil, err
	}

	year, err := scrapeYearFromRange(payload.Range)
	if err != nil {
		return nil, err
	}

	transactions := make([]domain.Transaction, 0, len(payload.Rows))
	for _, row := range payload.Rows {
		date, err := parseScrapeDate(year, row.Date)
		if err != nil {
			return nil, fmt.Errorf("parse scrape date %q: %w", row.Date, err)
		}

		amount, err := parseAmount(row.Amount)
		if err != nil {
			return nil, fmt.Errorf("parse amount %q: %w", row.Amount, err)
		}

		note := ""
		if row.Memo != nil {
			note = strings.TrimSpace(*row.Memo)
		}

		category := row.LargeCategory
		if strings.TrimSpace(row.MiddleCategory) != "" {
			category = fmt.Sprintf("%s>%s", row.LargeCategory, row.MiddleCategory)
		}

		fields := map[string]string{
			"merchant":       row.Merchant,
			"category":       category,
			"payment_method": row.PaymentMethod,
			"source_account": row.PaymentMethod,
			"note":           note,
		}

		rawJSON, err := json.Marshal(row)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, domain.Transaction{
			Date:          date,
			Amount:        amount,
			Merchant:      strings.TrimSpace(row.Merchant),
			Category:      category,
			PaymentMethod: strings.TrimSpace(row.PaymentMethod),
			SourceAccount: strings.TrimSpace(row.PaymentMethod),
			Owner:         config.ApplyStringRules(rules.OwnerRules, fields, "self"),
			ExpenseType:   config.ApplyStringRules(rules.ExpenseTypeRules, fields, "variable"),
			IsTransfer:    config.ApplyBoolRules(rules.TransferRules, fields, false),
			Note:          note,
			RawSourceFile: path,
			RawRowHash:    hashCanonicalRow(date, amount, row.Merchant, category, row.PaymentMethod, row.PaymentMethod, row.TransactionID),
			RawJSON:       string(rawJSON),
		})
	}

	return transactions, nil
}

func scrapeYearFromRange(raw string) (int, error) {
	if len(raw) < 4 {
		return 0, fmt.Errorf("invalid range: %s", raw)
	}
	return strconv.Atoi(raw[:4])
}

func parseScrapeDate(year int, raw string) (time.Time, error) {
	matches := monthDayPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if len(matches) != 3 {
		return time.Time{}, fmt.Errorf("unsupported scrape date")
	}

	month, err := strconv.Atoi(matches[1])
	if err != nil {
		return time.Time{}, err
	}
	day, err := strconv.Atoi(matches[2])
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}
