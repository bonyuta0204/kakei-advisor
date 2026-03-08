---
name: mf-monthly-update
description: Use this skill when the user wants to run the recurring monthly MoneyForward update workflow for this repository. This skill is for repository-scoped operation in kakei-advisor, using Chrome MCP after the user logs in manually, scraping the current target month from MoneyForward, ingesting it with the local Go CLI, and generating the monthly report.
---

# MoneyForward Monthly Update

Use this skill for the normal monthly refresh after the initial backfill is done.

## Preconditions

- Work only in this repository.
- Do not commit anything under `data/`.
- The user logs in to MoneyForward manually.
- Use Chrome MCP for browser interaction.
- Use the local Go CLI for ingest and reporting.

## Default workflow

1. Ask the user to log in to MoneyForward if not already logged in.
2. Open `https://moneyforward.com/cf`.
3. Move to the requested month. Default to the current month if the user does not specify one.
4. Scrape the visible transaction table.
5. Save the payload to `data/raw/mf_scrape_YYYY-MM.json`.
6. Run:

```bash
go run ./cmd/kakei-advisor ingest-mf-scrape --input data/raw/mf_scrape_YYYY-MM.json --db data/finance.duckdb --rules config/default_rules.json
go run ./cmd/kakei-advisor report-monthly --db data/finance.duckdb --month YYYY-MM
```

7. Summarize:
   - rows scraped
   - rows ingested
   - report result
   - any obvious classification gaps

## Execution rules

- Run ingest and report sequentially.
- Treat scrape JSON as the raw source of truth for that month.
- If the visible table clearly does not contain all expected rows, stop and tell the user what is incomplete.
- If categories or owner rules look wrong, suggest updating `config/default_rules.json` instead of hardcoding exceptions in the prompt.

## Output expectations

- Mention the target month explicitly.
- Mention the path of the raw JSON file.
- Mention the exact CLI commands used.
- Do not commit or push unless the user explicitly asks.

