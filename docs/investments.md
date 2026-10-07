# Investment statements

Open **Import Statements**, then choose **Investments** alongside **Bank & credit cards**. Attach or drag in one or multiple files, inspect each detected portfolio preview, and import ready statements individually or together. The upload page discovers supported providers and extensions from the parser registry; Zerodha holdings `.xlsx` is the first supported adapter (up to 10 MB per file). Previews never save data. **Investments** in the sidebar is the results page, showing invested amount, current value at the statement's closing prices, unrealized gain/loss, return on cost, and individual holdings. Search by symbol or ISIN, filter asset classes, and open **Statement details** to inspect every worksheet field.

These are dated holdings snapshots, not live market valuations. Returns do not include realized trades, dividends, charges, or cash flows absent from the workbook; annualized return and XIRR cannot be inferred. Different snapshots are selectable and are never added together. Re-uploading the identical file, even with a different filename, returns the existing snapshot. Delete a snapshot to remove an accidental import.

Investment records use a separate SQLite table. They never create bank accounts or ledger transactions and do not change income, expense, cash-flow, or budget calculations. Database backups, JSON export, and database reset include investments.

## Adding providers

Implement `investment.Parser` (`ID`, `Info`, `CanParse`, `Parse`) and register it in `investment.DefaultRegistry`. `Info` supplies the provider, statement name, and supported file extensions through `/api/investments/formats`. The Investments UI discovers supported providers, file picker extensions, and the upload limit from this endpoint; adding an adapter requires no provider-specific page changes. Adapters normalize provider-specific files into `models.InvestmentSnapshot` and `models.InvestmentHolding`; the service, persistence, HTTP endpoints, and UI use only these shared types. Keep additional provider fields in `Fields` and worksheet data in `Sheets`. Detection must use workbook structure rather than a user's filename or account number; ambiguous matches are rejected.

The first adapter reads Zerodha's Combined sheet when present to avoid counting repeated Equity and Mutual Funds sheets twice. It retains those original sheets for inspection. Quantities include available and pledged holdings, without adding the long-term subset. Current value is quantity times previous closing price. Cost basis is current value minus the statement's reported unrealized P&L, which preserves precision lost in rounded average prices. Combined summary totals must reconcile within ₹1.

Use fictional, generated workbooks for tests. Never commit personal holdings, account identifiers, statement files, validation databases, or screenshots containing personal data.

Zerodha documents dated holdings exports in its [holdings report guide](https://support.zerodha.com/category/console/portfolio/console-holdings/articles/holding-report), and distinguishes pledged from remaining available quantities in its [pledged holdings guide](https://support.zerodha.com/category/trading-and-markets/general-kite/kite-holdings/articles/p-symbol-kite-holdings).
