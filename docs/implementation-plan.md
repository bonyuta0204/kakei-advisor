# Implementation Plan

## 目的

MoneyForward を収集ハブとして使い、CSV が手元にある前提で以下を安定して行えるようにする。

- DuckDB への明細保存
- 最低限のルール補正
- 月次レビュー用 Markdown 出力

## 今回の実装範囲

### コマンド

- `ingest-moneyforward`
- `report-monthly`

### データモデル

保存する最小カラム:

- `txn_date`
- `amount`
- `merchant`
- `category`
- `payment_method`
- `source_account`
- `owner`
- `expense_type`
- `is_transfer`
- `note`
- `raw_source_file`
- `raw_row_hash`
- `raw_json`

### ルール

- `owner_rules`
- `expense_type_rules`
- `transfer_rules`

## 後回しにしたもの

- MoneyForward ブラウザ操作
- Agent Skill 本体
- GSS/Notion 連携
- より厳密な複式管理

## 直近の次ステップ

1. MoneyForward 実データで列名差分を確認する
2. ルールファイルを本人用に初期化する
3. 月次レポートの項目を増やす
4. Agent からこの CLI を呼ぶ導線を作る

