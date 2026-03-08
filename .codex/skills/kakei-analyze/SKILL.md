---
name: kakei-analyze
description: Use this skill when the user wants analysis or advice from this repository's household-finance DuckDB, including income checks, savings targets, monthly review, classification gaps, and interpretation of ambiguous household flows using the private memory file.
---

# Household Finance Analysis

Use this skill when the user asks questions such as:

- whether income is being captured correctly
- what spending is large recently
- what savings target is reasonable
- what is likely transfer, shared funding, or investment
- what to improve in the data model or monthly KPI

## Preconditions

- Work only in this repository.
- Read `/Users/yuta.nakamura/.codex/memories/kakei-advisor.md` before drawing conclusions from the data.
- Treat the private memory file as user-confirmed context.
- Do not copy private-memory contents into tracked repository files unless the user explicitly asks.
- When using the DuckDB file, avoid lock conflicts by copying it to `/tmp/` first if another process is holding a write lock.

## Default workflow

1. Read the private memory file.
2. Inspect the current DuckDB contents.
3. Separate `income`, `spending`, `investment/savings-like`, and `transfer-like` flows as much as the current rules allow.
4. Call out known ambiguity instead of pretending the classification is cleaner than it is.
5. Answer the user's question with:
   - the concrete numbers found in the DB
   - the confidence level and main caveats
   - a practical interpretation for household decisions
6. If the user asks for targets or advice that depends on current external benchmarks, verify them with current primary or official sources first.

## Analysis rules

- Prefer merchant-level inspection before making category-level claims when categories are noisy.
- Treat user-confirmed shared-account funding, off-ledger savings, and salary interpretation from the private memory as stronger evidence than heuristics.
- Keep a hard boundary between:
  - observed DB facts
  - inferences from those facts
  - recommendations
- If current reporting logic mixes income and expense incorrectly, say so explicitly.
- If a recommendation depends on a rule change, suggest the rule change instead of burying it in prose.

## Typical outputs

- income capture check
- recent large spending summary
- monthly savings estimate and savings-target proposal
- classification cleanup suggestions
- candidate KPI definitions for monthly review
