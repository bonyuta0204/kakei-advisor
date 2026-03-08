package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bonyuta0204/kakei-advisor/internal/config"
	"github.com/bonyuta0204/kakei-advisor/internal/moneyforward"
	"github.com/bonyuta0204/kakei-advisor/internal/report"
	"github.com/bonyuta0204/kakei-advisor/internal/storage"
)

func Run(args []string) error {
	if len(args) == 0 {
		return usageError()
	}

	switch args[0] {
	case "ingest-moneyforward":
		return runIngestMoneyForward(args[1:])
	case "ingest-mf-scrape":
		return runIngestMFScrape(args[1:])
	case "report-monthly":
		return runReportMonthly(args[1:])
	case "-h", "--help", "help":
		return usageError()
	default:
		return fmt.Errorf("unknown command: %s\n\n%s", args[0], usageText())
	}
}

func runIngestMoneyForward(args []string) error {
	fs := flag.NewFlagSet("ingest-moneyforward", flag.ContinueOnError)
	input := fs.String("input", "", "path to MoneyForward CSV")
	dbPath := fs.String("db", "data/finance.duckdb", "path to DuckDB file")
	rulesPath := fs.String("rules", "config/default_rules.json", "path to rules JSON")
	fs.SetOutput(os.Stdout)

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" {
		return errors.New("ingest-moneyforward requires --input")
	}

	rules, err := config.LoadRules(*rulesPath)
	if err != nil {
		return err
	}
	transactions, err := moneyforward.ParseCSV(*input, rules)
	if err != nil {
		return err
	}

	repo, err := storage.Open(*dbPath)
	if err != nil {
		return err
	}
	defer repo.Close()

	inserted, err := repo.InsertTransactions(transactions)
	if err != nil {
		return err
	}

	fmt.Printf("ingested=%d parsed=%d db=%s\n", inserted, len(transactions), *dbPath)
	return nil
}

func runIngestMFScrape(args []string) error {
	fs := flag.NewFlagSet("ingest-mf-scrape", flag.ContinueOnError)
	input := fs.String("input", "", "path to scraped MoneyForward JSON")
	dbPath := fs.String("db", "data/finance.duckdb", "path to DuckDB file")
	rulesPath := fs.String("rules", "config/default_rules.json", "path to rules JSON")
	fs.SetOutput(os.Stdout)

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" {
		return errors.New("ingest-mf-scrape requires --input")
	}

	rules, err := config.LoadRules(*rulesPath)
	if err != nil {
		return err
	}
	transactions, err := moneyforward.ParseScrapeJSON(*input, rules)
	if err != nil {
		return err
	}

	repo, err := storage.Open(*dbPath)
	if err != nil {
		return err
	}
	defer repo.Close()

	inserted, err := repo.InsertTransactions(transactions)
	if err != nil {
		return err
	}

	fmt.Printf("ingested=%d parsed=%d db=%s\n", inserted, len(transactions), *dbPath)
	return nil
}

func runReportMonthly(args []string) error {
	fs := flag.NewFlagSet("report-monthly", flag.ContinueOnError)
	dbPath := fs.String("db", "data/finance.duckdb", "path to DuckDB file")
	month := fs.String("month", "", "target month in YYYY-MM")
	outputPath := fs.String("output", "", "output markdown file")
	fs.SetOutput(os.Stdout)

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *month == "" {
		return errors.New("report-monthly requires --month")
	}

	repo, err := storage.Open(*dbPath)
	if err != nil {
		return err
	}
	defer repo.Close()

	summary, err := repo.MonthlySummary(*month)
	if err != nil {
		return err
	}

	content := report.RenderMonthly(summary)
	if *outputPath == "" {
		fmt.Print(content)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(*outputPath, []byte(content), 0o644)
}

func usageError() error {
	return fmt.Errorf("%s", usageText())
}

func usageText() string {
	return `Usage:
  kakei-advisor ingest-moneyforward --input <csv> [--db <duckdb>] [--rules <json>]
  kakei-advisor ingest-mf-scrape --input <json> [--db <duckdb>] [--rules <json>]
  kakei-advisor report-monthly --month YYYY-MM [--db <duckdb>] [--output <file>]`
}
