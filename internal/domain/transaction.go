package domain

import "time"

type Transaction struct {
	Date          time.Time
	Amount        int64
	Merchant      string
	Category      string
	PaymentMethod string
	SourceAccount string
	Owner         string
	ExpenseType   string
	IsTransfer    bool
	Note          string
	RawSourceFile string
	RawRowHash    string
	RawJSON       string
}

type MonthlySummary struct {
	Month          string
	TransactionCnt int64
	TotalAmount    int64
	CategoryTotals []Breakdown
	OwnerTotals    []Breakdown
}

type Breakdown struct {
	Label  string
	Amount int64
}
