#!/usr/bin/env node

const fs = require("fs");
const path = require("path");

function usage() {
  console.error(
    [
      "Usage:",
      "  node .codex/skills/mf-backfill/scripts/mf_fetch_to_scrape_json.js <input.js> [more inputs.js]",
      "",
      "Behavior:",
      "  - Converts saved MoneyForward /cf/fetch responses into mf_scrape_YYYY-MM.json payloads.",
      "  - If the input file name matches _mf_fetch_YYYY-MM.js, output is inferred as mf_scrape_YYYY-MM.json",
      "    in the same directory.",
    ].join("\n"),
  );
}

function parseArgs(argv) {
  if (argv.length === 0 || argv.includes("-h") || argv.includes("--help")) {
    usage();
    process.exit(argv.length === 0 ? 1 : 0);
  }

  return argv;
}

function inferOutputPath(inputPath) {
  const dirname = path.dirname(inputPath);
  const basename = path.basename(inputPath);
  const match = basename.match(/^_mf_fetch_(\d{4}-\d{2})\.js$/);
  if (!match) {
    throw new Error(
      `cannot infer output name from ${basename}; expected _mf_fetch_YYYY-MM.js`,
    );
  }

  return path.join(dirname, `mf_scrape_${match[1]}.json`);
}

function extractRange(source) {
  const startDate = source.match(/startDate:\s*'([^']+)'/);
  const endDate = source.match(/endDate:\s*'([^']+)'/);
  if (!startDate || !endDate) {
    throw new Error("failed to parse startDate/endDate");
  }

  return `${startDate[1]} - ${endDate[1]}`;
}

function extractAppendPayload(source) {
  const marker = '$(".list_body").append(\'';
  const start = source.indexOf(marker);
  if (start === -1) {
    throw new Error('failed to locate $(".list_body").append(...) payload');
  }

  let i = start + marker.length;
  let escaped = false;
  for (; i < source.length; i += 1) {
    const ch = source[i];
    if (escaped) {
      escaped = false;
      continue;
    }
    if (ch === "\\") {
      escaped = true;
      continue;
    }
    if (ch === "'") {
      break;
    }
  }

  if (i >= source.length) {
    throw new Error("failed to find end of appended HTML string");
  }

  return source.slice(start + marker.length, i);
}

function decodeJsString(raw) {
  return raw
    .replace(/\\n/g, "\n")
    .replace(/\\\//g, "/")
    .replace(/\\"/g, '"')
    .replace(/\\'/g, "'")
    .replace(/\\\\/g, "\\");
}

function normalizeText(value) {
  return value.replace(/\s+/g, " ").trim();
}

function firstCapture(row, patterns) {
  for (const pattern of patterns) {
    const match = row.match(pattern);
    if (!match) {
      continue;
    }

    for (const value of match.slice(1)) {
      if (value != null) {
        return normalizeText(value);
      }
    }
  }

  return "";
}

function parseRows(html) {
  const rowMatches = [...html.matchAll(/<tr class="transaction_list[\s\S]*?<\/tr>/g)];
  return rowMatches
    .map((match) => {
      const row = match[0];
      const memo = firstCapture(row, [
        /<td class="memo form-switch-td">[\s\S]*?<div class="noform">[\s\S]*?<span>([^<]*)<\/span>/,
      ]);

      return {
        transaction_id: firstCapture(row, [
          /name="user_asset_act\[id\]"[^>]*value="([^"]+)"/,
          /<input value="([^"]+)" type="hidden" name="user_asset_act\[id\]"/,
        ]),
        date: firstCapture(row, [/<td class="date"[\s\S]*?<span>([^<]+)<\/span>/]),
        merchant: firstCapture(row, [/<td class="content">[\s\S]*?<span>([^<]+)<\/span>/]),
        amount: firstCapture(row, [/<span class="offset">([^<]+)<\/span>/]),
        payment_method: firstCapture(row, [/<td class="note calc"[^>]*>([\s\S]*?)<\/td>/]),
        large_category: firstCapture(row, [
          /<a class="btn btn-small dropdown-toggle v_l_ctg"[^>]*>\s*([^<]+?)\s*<\/a>/,
        ]),
        middle_category: firstCapture(row, [
          /<a class="btn btn-small dropdown-toggle v_m_ctg"[^>]*>\s*([^<]+?)\s*<\/a>/,
        ]),
        memo: memo === "" ? null : memo,
        is_transfer_ui: row.includes("（振替）") ? true : null,
        is_deleted_ui: row.includes("削除済") ? true : null,
      };
    })
    .filter((row) => row.transaction_id !== "" && row.date !== "" && row.merchant !== "");
}

function convertFile(inputPath) {
  const source = fs.readFileSync(inputPath, "utf8");
  const outputPath = inferOutputPath(inputPath);
  const payload = {
    scraped_at: new Date().toISOString(),
    page_url: "https://moneyforward.com/cf",
    range: extractRange(source),
    row_count: 0,
    rows: parseRows(decodeJsString(extractAppendPayload(source))),
  };
  payload.row_count = payload.rows.length;

  fs.writeFileSync(outputPath, `${JSON.stringify(payload, null, 2)}\n`);
  console.log(
    `${path.basename(inputPath)} -> ${path.basename(outputPath)} rows=${payload.row_count} range=${payload.range}`,
  );
}

function main() {
  const inputs = parseArgs(process.argv.slice(2));
  for (const inputPath of inputs) {
    convertFile(inputPath);
  }
}

main();
