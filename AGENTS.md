# kakei-advisor AGENTS

## Repository Overview

- This repository stores the local tooling for household-finance ingestion, normalization, and reporting.
- Treat `data/` as private user financial data. Never commit anything under `data/`.

## Working Rules

- Use repository-scoped skills in `.codex/skills/` for normal operations before inventing ad hoc workflows.
- For MoneyForward operations, the user handles login manually. Automation starts only after the authenticated page is available.
- Prefer the local Go CLI for deterministic ingest, normalization, and reporting.
- Prefer Chrome MCP for authenticated browser interaction until a more stable scripted path is intentionally introduced.

## Private Memory

- Before household-finance analysis or interpretation, read the private memory file at `/Users/yuta.nakamura/.codex/memories/kakei-advisor.md`.
- This file is outside the repository on purpose. Do not copy its contents into tracked repository files unless the user explicitly asks.
- Use the private memory file to resolve user-confirmed context such as shared-account transfers, salary interpretation, and off-ledger savings behavior.

## Documentation

- Keep long-lived docs and skill instructions aligned to the current operating model.
- Remove stale historical instructions when the workflow changes.
