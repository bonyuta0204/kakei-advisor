# kakei-advisor

MoneyForward の家計データを、`無料版でもできる範囲で` 取り込み、DuckDB に蓄積し、月次レビューまで回すためのリポジトリです。

この repo では次の分担を前提にしています。

- MoneyForward のログイン: 人間がやる
- MoneyForward の画面操作と `cf/fetch` / scrape: agent がやる
- データの保存、重複排除、集計: Go CLI がやる
- 日常運用の導線: repository scope の skill がやる

## これでできること

- 過去分の MoneyForward 明細を月ごとに取り込む
- 毎月の更新を同じ手順で回す
- DuckDB に家計データを蓄積する
- 月次レポートを Markdown で出す
- `owner`, `expense_type`, `is_transfer` をルールで補正する

## まず読むポイント

- この repo の日常運用は `skill + CLI` の組み合わせで回す
- `data/` 配下は個人データ置き場なのでコミットしない
- 無料版 MoneyForward では CSV ダウンロードが使えないことがあるため、基本は authenticated `cf/fetch` + scrape JSON 前提で考える

## 運用の全体像

### 1. 初回セットアップ

最初にやることは、過去分のデータをまとめて DuckDB に入れることです。

流れ:

1. MoneyForward にブラウザでログインする
2. `mf-backfill` skill を使う
3. agent が認証済みページから `cf/fetch` を使って visible rows を取得する
4. 必要に応じて helper script で raw response を `data/raw/mf_scrape_YYYY-MM.json` に変換する
5. Go CLI で DuckDB に取り込む

### 2. 月次更新

一度バックフィルが終わったら、通常運用は月次更新だけでよいです。

流れ:

1. MoneyForward にブラウザでログインする
2. `mf-monthly-update` skill を使う
3. agent が対象月の `cf/fetch` / visible rows を取得する
4. raw JSON を `data/raw/` に保存する
5. Go CLI で DuckDB に取り込む
6. 月次レポートを出す

## repository scope の skill

この repo には repository scope の skill を置いています。

- `.codex/skills/mf-backfill/`
  - 過去分を月ごとに scrape して取り込むための skill
- `.codex/skills/mf-monthly-update/`
  - 毎月の更新を回すための skill
- `.codex/skills/kakei-analyze/`
  - DuckDB を分析して、収入確認、支出傾向、貯蓄目標、分類ギャップを読むための skill

通常運用は `mf-backfill`, `mf-monthly-update`, `kakei-analyze` の 3 本を入口にします。  
通常は CLI を手で全部叩くより、まず skill を呼ぶ前提で使います。

## private memory

ユーザーが会話で明示した家計コンテキストは、repo の外にある private memory に置きます。

- 保存先: `/Users/yuta.nakamura/.codex/memories/kakei-advisor.md`
- 例:
  - 奥さんから共有口座への毎月振込
  - 借り上げ社宅控除後の給与解釈
  - MoneyForward に出ない持株会積立

このファイルは Git 管理しません。  
分析系の skill は、家計の意味解釈をする前にこのファイルを読む前提です。

## コマンド

Go CLI は「保存と集計の決定論的な処理」を担当します。

### scrape JSON を取り込む

```bash
go run ./cmd/kakei-advisor ingest-mf-scrape \
  --input data/raw/mf_scrape_2026-03.json \
  --db data/finance.duckdb \
  --rules config/default_rules.json
```

### CSV を取り込む

CSV が使える場合だけこちらを使います。

```bash
go run ./cmd/kakei-advisor ingest-moneyforward \
  --input path/to/moneyforward.csv \
  --db data/finance.duckdb \
  --rules config/default_rules.json
```

### 月次レポートを出す

```bash
go run ./cmd/kakei-advisor report-monthly \
  --db data/finance.duckdb \
  --month 2026-03 \
  --output reports/2026-03.md
```

## ふだんの使い方

### 過去分を取り込みたいとき

- Codex / Claude Code でこの repo を開く
- `mf-backfill` を使いたいと伝える
- ログインを求められたら MoneyForward にログインする
- 対象月の範囲を伝える

### 今月分だけ更新したいとき

- Codex / Claude Code でこの repo を開く
- `mf-monthly-update` を使いたいと伝える
- ログインを求められたら MoneyForward にログインする
- 対象月を指定する。省略時は当月扱い

## データの扱い

### `data/` 配下

- `data/raw/`
  - MoneyForward から scrape した raw JSON を置く
- `data/*.duckdb`
  - ローカルの家計データベース

この配下は個人データなので Git にコミットしません。

### `reports/` 配下

- 出力した月次レポートの置き場
- 必要ならここもローカル専用運用でよい

## ルールの調整

分類や owner 判定は `config/default_rules.json` で調整します。

たとえば次のようなものをここで吸収します。

- 共用カードは `owner=shared`
- PASMO チャージは `is_transfer=true`
- サブスクは `expense_type=fixed`

分類がおかしいときは、プロンプトで都度補正するより、まずこのファイルを直す方針です。

## いまの前提と制約

- 認証は自動化しない
- Chrome MCP を使った半自動運用を基本にする
- 無料版 MoneyForward では export 制限があるため scrape 前提になりやすい
- DuckDB は同時書き込みに弱いので、取込とレポートは直列実行する

## この repo のおすすめ運用

### 目先

- まず `mf-backfill` で過去分を埋める
- 次に `mf-monthly-update` で月次更新に入る

### その後

- ルールファイルを実データに合わせて育てる
- 月次レポートの出力先を Obsidian に寄せる
- 必要になったら Playwright 化を検討する

## 補足

- 詳細設計は `docs/implementation-plan.md`
- Obsidian 側の構想メモは別 repo のノートで管理している
