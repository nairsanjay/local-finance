# 🇮🇳 LocalFinance

[![CI](https://github.com/usmslm102/local-finance/actions/workflows/ci.yml/badge.svg)](https://github.com/usmslm102/local-finance/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Website](https://img.shields.io/badge/Website-usamaansari.com-emerald)](https://usamaansari.com/local-finance/)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)](https://github.com/usmslm102/local-finance/releases)
[![Architecture](https://img.shields.io/badge/Architecture-100%25%20Offline-success)](https://github.com/usmslm102/local-finance)

> **A privacy-first, 100% offline personal finance intelligence app tailored for the Indian banking & credit card ecosystem.**  
> Distributed as a **single standalone executable binary** with an embedded React frontend and an embedded pure-Go SQLite database.

> [!IMPORTANT]
> ### ⚖️ Disclaimer & Notice of Liability (Use at Your Own Risk)
> **LocalFinance is an independent open-source software project distributed under the terms of the [MIT License](LICENSE).**
>
> - **"AS IS" & Use at Your Own Risk**: This software is provided **"AS IS"**, without warranty of any kind, express or implied, including but not limited to the warranties of merchantability, fitness for a particular purpose, and non-infringement. You use this application entirely at your own discretion and risk.
> - **No Financial, Tax, or Legal Advice**: LocalFinance is an offline analytical and record-keeping tool. It does not provide financial planning, accounting, investment, tax, or legal advice.
> - **No Banking Affiliation**: LocalFinance is completely independent and is not affiliated with, endorsed by, sponsored by, or connected to HDFC Bank, ICICI Bank, Axis Bank, State Bank of India (SBI), the Reserve Bank of India (RBI), or any other financial institution. All bank names, brand trademarks, and logos are property of their respective owners and are referenced solely for identification and parser compatibility.
> - **Zero Liability**: To the maximum extent permitted by applicable law, in no event shall the authors, maintainers, copyright holders, or contributors be held liable for any claim, damages, losses, or legal liabilities—whether in contract, tort (including negligence), or otherwise—arising from, out of, or in connection with the software, its calculations, parser extractions, categorization rules, data loss, miscalculated figures, financial decisions, tax assessments, or any consequences resulting from its use.
> - **User Responsibility**: Users are solely responsible for reviewing and verifying the accuracy of all imported transactions, balances, and calculations against their original official bank and credit card statements before taking any financial action.

---

## 🌟 Key Highlights

- **100% Offline & Private**: Zero cloud sync, zero telemetry, no SMS scraping, and no third-party account aggregators. All your financial data stays strictly on your local machine in an open SQLite database file (`local_finance.db` or `~/.localfinance/local_finance.db`).
- **Single-Binary Portability**: Built in **Go (Golang)** with the React 19 single-page app bundled directly into the executable via `go:embed`. Cross-compiles across Windows (`.exe`), macOS, and Linux with zero CGO dependencies (`CGO_ENABLED=0`).
- **Zero-Config Database Migrations**: Schema migrations managed automatically via embedded Goose (`github.com/pressly/goose/v3`) on startup—no external migration CLI needed.
- **Generic & Extensible Parser Engine**: Pluggable architecture that **auto-detects** bank formats, account types (Savings, Current, Credit Cards), and file structures (PDF, CSV, Excel/XLS/XLSX).
- **Native Password-Protected PDF Support**: Upload encrypted bank statements directly through the UI; LocalFinance decrypts and parses them in-memory without altering original file formatting.
- **Multi-Account & Multi-Year Bulk Statements**: Handles 10+ year bulk historic statements without missing records, correctly distinguishing multiple accounts (e.g. separate HDFC Savings and Current accounts).
- **Indian Narration Intelligence**: Native regex cleaning engine for:
  - **UPI Transfers**: Extracts Payee names, VPAs (e.g., `merchant@icici`, `user@okaxis`), Reference/UTR numbers, and apps used (GPay, PhonePe, Paytm).
  - **Card Swipes / POS**: Cleans raw terminal dumps (e.g., `POS 40124300 SWIGGY BANGALORE IN` &rarr; `Swiggy`).
  - **IMPS / NEFT / RTGS**: Resolves beneficiary names and 12-to-16 digit UTR numbers.
  - **ATM Withdrawals & Bank Charges**: Distinguishes cash withdrawals, SMS alert fees, AMC charges, and GST.
- **Credit Card Billing Intelligence**:
  - Automatically captures billing cycles, statement dates, payment due dates, minimum due, total due, finance charges, reward points, and cashback.
- **Duplicate Detection & Smart Upsert**:
  - Deterministic `SHA-256` transaction hashing prevents duplicates when uploading overlapping statement periods.
  - Idempotent SQLite upserts (`ON CONFLICT`) safely update balances while **strictly preserving** your custom category overrides, notes, and tags.
- **Optional Local Security / Password Protection**:
  - Protect local database access with an optional PIN/password stored with PBKDF2/argon2 hashing, complete with automatic lock timeout.
- **Modern Interactive Dashboa…4349 tokens truncated…avings_csv.go     # HDFC Savings/Current Account CSV parser plugin
│   │   ├── hdfc_savings_pdf.go     # HDFC Savings/Current Account PDF parser plugin (bulk & standard)
│   │   ├── hdfc_savings_xls.go     # HDFC Savings/Current Account Excel parser plugin
│   │   ├── icici_cc_pdf.go         # ICICI Bank Credit Card PDF parser plugin
│   │   ├── axis_cc_pdf.go          # Axis Bank Credit Card PDF parser plugin
│   │   └── parser.go               # Generic StatementParser interface & auto-detection Registry
│   └── service/
│       └── transaction_service.go  # Ingestion pipeline, SHA-256 fingerprinting & smart upserts
├── Makefile                        # Build and development automation targets
├── ROADMAP.md                      # Product specifications, completed ledger & future roadmap
└── README.md                       # Comprehensive user and developer manual
```

---

## 🧩 Adding a New Bank Parser (Extensibility)

Adding support for a new bank or format (e.g. SBI, Kotak, Axis, Amex) is completely modular:

1. Create a new file in `internal/parser/<bank>_<account_type>_<format>.go` (e.g. `internal/parser/sbi_savings_csv.go`).
2. Implement the `StatementParser` interface:

```go
package parser

import (
	"io"
	"strings"
	"local-finance/internal/models"
)

type SBISavingsCSVParser struct{}

func init() {
	// Automatically registers parser with the global registry
	DefaultRegistry.Register(&SBISavingsCSVParser{})
}

func (p *SBISavingsCSVParser) ID() string {
	return "sbi_savings_csv_v1"
}

func (p *SBISavingsCSVParser) Name() string {
	return "State Bank of India Savings CSV"
}

func (p *SBISavingsCSVParser) SupportedTypes() []StatementType {
	return []StatementType{TypeSavingsCSV}
}

func (p *SBISavingsCSVParser) CanParse(filename string, sample []byte) (float64, string, models.AccountType) {
	content := strings.ToUpper(string(sample))
	confidence := 0.0
	if strings.Contains(content, "STATE BANK OF INDIA") || strings.Contains(content, "SBI") {
		confidence += 0.5
	}
	if strings.Contains(content, "TXN DATE") && strings.Contains(content, "DESCRIPTION") {
		confidence += 0.4
	}
	return confidence, "State Bank of India", models.AccountTypeSavings
}

func (p *SBISavingsCSVParser) Parse(r io.Reader, opts ParseOptions) ([]ParsedTransaction, StatementMeta, error) {
	// 1. Read rows using extractor.ExtractCSV
	// 2. Normalize dates using extractor.NormalizeIndianDate
	// 3. Clean narrations using CleanNarration(raw)
	// 4. Return []ParsedTransaction and StatementMeta
}
```

---

## Monthly Review

Overview now includes a compact review of the latest month with imported data. Choose **Understand what changed** to open the detailed review in Cash Flow, or use **Review month** there to inspect another month.

- Expand statement coverage to see how many days each tracked account's reported statement ranges cover in both periods. Overlaps count once; gaps remain visible. This is a date-coverage check, not a balance audit.
- Inspect the largest category changes and the merchants contributing to them. **View transactions** shows the exact recorded debits for either period, with pagination.
- Use **Plan next month** to open the existing budget editor for the following month and category. Nothing is saved automatically.

Historical months compare full calendar months. The current month compares elapsed days, capped at the previous month's last day when it is shorter. Transfers and excluded transactions do not count as spending; credits and refunds are not deducted. Missing statement coverage can distort comparisons, so the review flags it explicitly.

The review is read-only, runs offline, and uses the app's existing authentication and discreet mode.

## 🔌 REST API Reference

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/api/health` | `GET` | Server health check and version info |
| `/api/accounts` | `GET` | List all discovered bank accounts and credit cards with current balances |
| `/api/transactions` | `GET` | List transactions with filters (`account_id`, `category_id`, `tx_type`, `search`, `start_date`, `end_date`, `page`, `page_size`) |
| `/api/statements/upload` | `POST` | Ingest statement file (multipart `file`, optional `password`, `account_id`, `parser_id`) |
| `/api/statements/preview` | `POST` | Dry-run statement parse without saving to DB (shows detected metadata and transactions) |
| `/api/statements` | `GET` | List history of uploaded statements with checksums and date ranges |
| `/api/categories` | `GET` | List spending categories with icons and color tokens |
| `/api/rules` | `GET` | List active auto-categorization keyword & regex rules |
| `/api/analytics/overview`| `GET` | Get total income, total expense, net savings, category breakdown, and monthly cash flow |
| `/api/analytics/monthly-review` | `GET` | Spending comparison and statement coverage; optional `month=YYYY-MM`, defaulting to the latest month with imported data |
| `/api/analytics/monthly-review/transactions` | `GET` | Supporting debits for `month`, `category` (empty means uncategorized), `period=current\|previous`, and `page`; 50 rows per page |
| `/api/parsers` | `GET` | List all registered bank parser plugins |

---

## 🛡️ Privacy & Local Security

- **100% Offline**: LocalFinance does not make external network requests, send telemetry, or connect to third-party servers.
- **Local SQLite Storage**: Your data lives entirely in `local_finance.db` (or `~/.localfinance/local_finance.db`). Backups can be made simply by copying this single file.
- **Local App Lock**: An optional PIN/Password can be enabled in settings to restrict local access to the dashboard.

---

## 📄 License
MIT License. Free and open source for local personal finance intelligence.
