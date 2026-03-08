---
name: mf-backfill
description: Use this skill when the user wants to backfill past MoneyForward household transactions into this repository's DuckDB. This skill is for repository-scoped operation in kakei-advisor, using Chrome MCP after the user logs in manually, scraping one month at a time from MoneyForward and ingesting each month with the local Go CLI.
---

# MoneyForward Backfill

Use this skill for one-time or catch-up imports of past months.

## Preconditions

- Work only in this repository.
- Do not commit anything under `data/`.
- The user logs in to MoneyForward manually.
- Use Chrome MCP for browser interaction.
- Use the local Go CLI for ingest and reporting.

## Default workflow

1. Ask the user to log in to MoneyForward if not already logged in.
2. Open `https://moneyforward.com/cf`.
3. For each target month:
   - Navigate the household ledger to that month.
   - Scrape the visible transaction table.
   - Save the scraped payload to `data/raw/mf_scrape_YYYY-MM.json`.
   - Run:

```bash
go run ./cmd/kakei-advisor ingest-mf-scrape --input data/raw/mf_scrape_YYYY-MM.json --db data/finance.duckdb --rules config/default_rules.json
```

4. After all months are ingested, optionally run:

```bash
go run ./cmd/kakei-advisor report-monthly --db data/finance.duckdb --month YYYY-MM
```

## Scrape payload shape

Save JSON with this shape:

```json
{
  "scraped_at": "ISO-8601",
  "page_url": "https://moneyforward.com/cf",
  "range": "YYYY/MM/DD - YYYY/MM/DD",
  "row_count": 0,
  "rows": []
}
```

Each row should include:

- `transaction_id`
- `date`
- `merchant`
- `amount`
- `payment_method`
- `large_category`
- `middle_category`
- `memo`

## Execution rules

- Ingest months sequentially, not in parallel.
- Prefer existing visible table data over hidden export endpoints.
- If the table is paginated or truncated, tell the user exactly what is missing before proceeding.
- If MoneyForward shows a premium-only export modal, ignore export and continue with scrape mode.
- Keep the user informed of current month and progress count.

## Completion checklist

- Confirm which months were ingested.
- Confirm where raw JSON files were written.
- Confirm that `data/` remains untracked by Git.

