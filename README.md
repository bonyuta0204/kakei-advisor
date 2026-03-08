# kakei-advisor

MoneyForward の CSV を取り込み、ルールベースで正規化し、DuckDB に保存して月次 Markdown レポートを出力するための軽量 CLI。

## 現在のスコープ

- `ingest-moneyforward`
  - MoneyForward CSV を読み込む
  - `owner`, `expense_type`, `is_transfer` をルールベースで補正する
  - DuckDB に重複排除しながら保存する
- `report-monthly`
  - 指定月の支出サマリーを Markdown で出力する

## 方針

- UI 依存の取得は agent に任せる
- 台帳への書き込みと集計は Go の CLI で決定論的に処理する
- ルールは `config/default_rules.json` で調整する

## 使い方

```bash
go run ./cmd/kakei-advisor ingest-moneyforward \
  --input testdata/moneyforward_sample.csv \
  --db data/finance.duckdb \
  --rules config/default_rules.json

go run ./cmd/kakei-advisor report-monthly \
  --db data/finance.duckdb \
  --month 2026-03 \
  --output reports/2026-03.md
```

## 設計メモ

- 詳細は `docs/implementation-plan.md`
- Obsidian 側の構想メモは別 repo のノートに管理

