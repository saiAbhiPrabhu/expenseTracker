# expenseTracker

Reads a Bank of America statement PDF, extracts each transaction, anonymizes it, and has JEV (TypeSafe's System One model) categorize it and judge whether it is an essential expense. The result is meant to feed a dashboard.

Everything up to the JEV call runs locally. The name and account number at the top of each page are cropped out before any text is produced.

## Overall Pipeline

```
statement PDF
  1. render each page to an image (300 dpi)
  2. crop away everything above the first red heading   <- removes name, address, account number
  3. OCR the cropped image (Tesseract)
  4. extract transaction rows (date, description, amount, section)
       check each section's printed Total against the sum of its rows
  5. remove deposits (money in)
  6. scrub descriptions: mask names, IDs and long numbers
  7. build JEV states: {"Date", "Description", "Amount"}
  8. ask JEV, one request per transaction, three questions:
       category_20   which of 20 categories (+ "Other")
       category_12   which of 12 categories (+ "Other")
       essential_vs_non_essential   probability the expense is essential
  9. validate the answers and write results
 10. export: raw answers per transaction, a key per transaction, and expenditure totals
```

Separate from the main flow:
- **Description simplification** (`simplify.go`) turns a scrubbed title into a readable one by removing filler words. It is built and tested but not yet wired into the pipeline.
- **A/B test** (`abtest.go`) compares the 20-category and 12-category answers for the same transactions. The 20-category set is the one chosen to continue with.

If a row's amount is misread by OCR, the run stops by design (the checksum would fail). See `future_isses.md` for the planned fix (partial uploads and a prompt for the user).

## Each File's Responsibility

| File | Responsibility |
|---|---|
| `main.go` | Command-line entry point: reads flags and environment, runs the pipeline, writes A/B and export files. The only place with I/O decisions. |
| `parsing.go` | PDF to text: page count, render, crop above the first red heading, OCR, per-stage snapshots, and manual OCR fixes (`-fix-ocr`). |
| `extraction.go` | Text to transactions: rows, sections, per-section checksum, deposit removal, scrubbing of personal data, and building the JEV state JSON (date and amount normalization; `parseDate` is the single date parser). |
| `simplify.go` | Removes filler words from a title. Documents the full heuristic cycle in its header comment. |
| `questions.go` | The fixed JEV questions: the two category Choice questions (built from `transaction_categories.json`, plus an "Other" catch-all) and the essential Noul question. |
| `classify.go` | Sends one request per transaction with all questions and returns validated results. Keeps partial results if a call fails midway. |
| `client.go` | The JEV HTTP client: request, bearer auth, retries with exponential backoff on 429 and 529. |
| `decide.go` | Turns JEV's raw answers into the final values: the 20-set category, a human-review flag (confidence below 0.5) and an essential yes/no (probability of 0.5 or more). |
| `export.go` | Everything exported: a SHA-256 key per transaction (date, title, amount), the raw JEV answers (JSON and CSV), and expenditure totals per month and per category. Expenditure is any outflow (a negative amount). |
| `results.go` | Typed, validated models for JEV answers (`ChoiceAnswer`, `NoulAnswer`, `ClassifiedTransaction`). Invalid answers are rejected when parsed. |
| `abtest.go` | Compares the two category sets and formats the agreement report. |
| `serve.go` | The web server: serves the built front end (`web/dist`) and mounts the `/api` routes. |
| `demo.go` | The demo's saved data and API. Replays the cached JEV run on synthetic transactions (`demo/`); no JEV calls, no per-user server state. |
| `web/` | The front end (React and TypeScript, built with Vite): home page with a folded menu, a non-functional log-in page, and the 5-step demo (placeholders for now). |
| `titles_csv.go` | TEMPORARY: loads already-extracted titles from a CSV (`-titles-csv`) to test JEV without re-running the PDF pipeline. |
| `transaction_categories.json` | The 20-category and 12-category sets with descriptions. |
| `*_test.go` | Tests for the file with the same name. |
| `CLAUDE.md`, `PERFECT_FUNCTION.md` | Working rules for the repo: behavior, and function design. |
| `future_isses.md` | Known problems and planned fixes. |
| `demo/` | Synthetic demo data (safe to commit): `transactions.csv`, its generator `build_demo_data.py`, and the saved JEV run in `cache/`. |
| `documents/`, `snapshots/`, `abtest/` | Local data: statements and labels, per-page images and text, and A/B outputs. All gitignored. |

## How to Use

### Requirements
- Go (see `go.mod`)
- Tesseract and Poppler (`pdftoppm`, `pdfinfo`): `brew install tesseract poppler`

### Setup
Create a `.env` file (gitignored) containing your key:
```
TYPESAFE_API_KEY=your-key-here
```

### Run the tests
```
go vet ./... && go test ./...
```

### Extract states from a statement (local only, no API call)
```
go run . documents/sept_statements.pdf
```
Prints the JSON states. Add `SNAPSHOT_DIR=snapshots` to save every page's render, crop and OCR text for inspection.

### Run the JEV A/B test (makes API calls)
```
set -a; source .env; set +a
go run . -abtest-dir abtest/sept documents/sept_statements.pdf
```
Writes `results.json` (every answer) and `report.md` (side-by-side comparison) to the directory. This sends one request per transaction, so it costs a small amount per run. Output is overwritten on each run.

### Export the data (local only, no API call)
```
go run . -export abtest/combined abtest/may/results.json abtest/jun/results.json abtest/jul/results.json abtest/aug/results.json abtest/sept/results.json
```
Combines the A/B `results.json` files into `jev_states.json` and `jev_states.csv` (raw answers for every transaction, oldest first, each with its key) and `expenditure_totals.json` (expenditure per month, per category, and overall).

### Run the web app (skeleton)
```
cd web && npm install && npm run build && cd ..
go run . -serve :8080
```
Open `http://localhost:8080`. The demo API reads the saved run in `demo/cache/results.json`:
- `GET /api/demo/transactions`: every transaction with JEV's answers and a review flag (confidence below 0.5).
- `GET /api/demo/categories`: the 20 categories plus "Other".
- `GET` or `POST /api/demo/summary`: dashboard numbers (expenditure per month, per category, essential split). `POST` takes `{"edits": {"<key>": {"category": "...", "essential": true}}}` and returns the numbers with those edits applied; the browser keeps the edits.

For front-end development, run `go run . -serve :8080` and `npm run dev` (in `web/`) together; Vite proxies `/api` to the Go server.

### Useful flags
- `-abtest-dir <dir>`: run JEV and write results to `<dir>`.
- `-serve <addr>`: start the web app, for example `:8080`.
- `-export <dir>`: read the `results.json` files given as arguments and write the export files to `<dir>`.
- `-fix-ocr 'OLD=NEW'`: correct a known OCR misread before extraction (repeatable). The text must appear exactly once, and the section checksum still has to pass. Example: `-fix-ocr '-177.4)=-177.41'`.
- `-titles-csv <file>`: temporary. Use titles from a CSV with `date`, `amount` and `title` columns instead of a PDF (titles are scrubbed on load).

### Privacy
Statements, snapshots and A/B outputs contain personal financial data. They live in gitignored folders and must never be committed. Only scrubbed descriptions are sent to JEV, but scrubbing is heuristic, so check results before sharing.
