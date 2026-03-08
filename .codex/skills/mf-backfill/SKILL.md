---
name: mf-backfill
description: Use this skill when the user wants to backfill past MoneyForward household transactions into this repository's DuckDB. This skill is for repository-scoped operation in kakei-advisor, using Chrome MCP after the user logs in manually, scraping one month at a time from MoneyForward and ingesting each month with the local Go CLI.
---

# MoneyForward Backfill

Use this skill for one-time or catch-up imports of past months.

## Preconditions

- Work only in this repository.
- Do not commit anything under `data/`.
- Read `/Users/yuta.nakamura/.codex/memories/kakei-advisor.md` before interpreting household context or classifying ambiguous rows.
- The user logs in to MoneyForward manually.
- Use Chrome MCP for browser interaction.
- Use the local Go CLI for ingest and reporting.
- Prefer the authenticated same-origin `POST /cf/fetch` response when it is available, because it returns the same visible rows as the household ledger table without fragile month-by-month UI clicking.

## Default workflow

1. Ask the user to log in to MoneyForward if not already logged in.
2. Open `https://moneyforward.com/cf`.
3. Determine the oldest accessible month before bulk ingest.
   - Test candidate months through the authenticated page with `POST /cf/fetch`.
   - Treat a `0 rows` / missing `range` response as inaccessible.
   - Record the actual oldest visible `range`, because the oldest month can be partial when MoneyForward is enforcing the rolling one-year limit.
4. For each accessible month, prefer the `cf/fetch` fast path:
   - Trigger `POST /cf/fetch` for that month from the authenticated page.
   - Save the raw response as `data/raw/_mf_fetch_YYYY-MM.js`.
   - Convert it with the bundled script:

```bash
node .codex/skills/mf-backfill/scripts/mf_fetch_to_scrape_json.js data/raw/_mf_fetch_YYYY-MM.js
```

   - This writes `data/raw/mf_scrape_YYYY-MM.json` in the repository format.
   - Run:

```bash
go run ./cmd/kakei-advisor ingest-mf-scrape --input data/raw/mf_scrape_YYYY-MM.json --db data/finance.duckdb --rules config/default_rules.json
```

5. Fallback only if `cf/fetch` is unavailable:
   - Navigate the household ledger to that month.
   - Scrape the visible transaction table directly from the DOM.
   - Save the payload as `data/raw/mf_scrape_YYYY-MM.json`.
6. After all months are ingested, optionally run:

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
- `cf/fetch` is acceptable because it is the same visible-table data source the page uses after login. Do not switch to export endpoints.
- Use the private memory file only as interpretation context. Do not write it back into repository files.
- Confirm each month's returned `range` and `row_count` before ingest.
- If the oldest accessible month starts mid-month, tell the user it is a partial month caused by MoneyForward visibility limits.
- If the table is paginated or truncated, tell the user exactly what is missing before proceeding.
- If MoneyForward shows a premium-only export modal, ignore export and continue with scrape mode.
- Keep the user informed of current month and progress count.

## Completion checklist

- Confirm which months were inaccessible due to MoneyForward limits.
- Confirm which months were ingested.
- Confirm where raw JSON files were written.
- Mention whether temporary `_mf_fetch_YYYY-MM.js` files remain or were cleaned up.
- Confirm that `data/` remains untracked by Git.
