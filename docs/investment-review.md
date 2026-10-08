# Investment PR review audit

This audit covers the distinct findings from both supplied reviews of PR #3. Repeated findings are grouped together. Intentional decisions are distinguished from code fixes.

| Finding | Resolution and evidence |
| --- | --- |
| Zerodha equities mislabeled as funds | Fixed. EQ, BE, blank, dash, and unknown instrument codes stay equities. Explicit MF/Mutual Fund markers and the Mutual Funds worksheet identify funds. `TestZerodhaAssetClasses` covers these cases. |
| Missing database read lock | Fixed. Listing acquires a read lock; export uses a private helper under its existing lock. `TestInvestmentReadsDuringRestore` exercises listing and export alongside database replacement. |
| Version-15 migration gap | Intentionally retained. Versions 16/17 have already been used by preview databases. Renumbering changes their migration identity; an earlier placeholder adds a new migration to an existing history. This is a compatibility exception to the normal sequential-addition rule, not a claim that the rule is irrelevant. `TestINDmoneyPreviewValuationMigration` covers a version-16 preview upgrading to 17; restore tests reopen version-17 databases. |
| Stale migration reference | Fixed. `docs/investments.md` correctly identifies migration 17 and explains the reserved gap. |
| Binary fixtures without generator integration | Followed the user's request to replace generators with checked-in samples. CONTRIBUTING explicitly permits fictional investment workbooks; both samples are exercised through service import and deduplication. No personal workbooks are fixtures. |
| Separate investment registry | Retained for the holdings snapshot contract. Bank parsers return debit/credit records, which cannot faithfully represent this portfolio. The repository parser skill now documents both contracts and prohibits additional provider registries or service/UI switches. `TestInvestmentAPIUsesRegisteredProvider` exercises an independent provider through the same API. |
| Unused registry Get method | Removed. Registration, format discovery, and automatic detection are the required investment operations. |
| Duplicated currency/privacy formatting | Fixed with `formatMoney` in `frontend/src/lib/formatters.ts`, shared by all three investment views, including statement-price precision. Verified in the local UI with privacy mode enabled and disabled. |
| Fragile timestamp ordering | Kept nanosecond UTC RFC3339 ordering. `Date.parse` truncates below milliseconds. Frontend tests check same-millisecond revisions in both input orders. |
| Unclear columns helper | Renamed to `headerIndexMap`; workbook helpers are independent of provider adapters. |
| Composite grouping key | Retained the collision-safe JSON tuple. A frontend regression proves delimiter-like provider/account names remain separate portfolios. Adding a wrapper would not change this behavior. |
| Missing INDmoney fixture valuation test | Fixed. `TestCheckedInINDmoneySampleImport` asserts $62.345679 current value, 0.123456789 shares, absent costs/returns, provider fields, and renamed-file deduplication. Zerodha fixture values and asset classes are also asserted. |
| Provider fields missing from UI | Fixed. Statement details displays per-holding fields and original worksheets. Both providers were inspected locally; provider values are masked in Discreet Mode. |
| Auxiliary worksheets rejected without Combined | Fixed. Summary/disclaimer sheets are retained without becoming holdings. Tests cover auxiliary sheets, separate Equity/Mutual Funds tables, missing headers/columns, and duplicate ISINs. Malformed financial tables still fail preview and import. |
| Unrelated transfer changes and misleading PR description | The branch was rebased onto main and the redundant transfer commit removed. Its diff has no bank transfer-classification changes. The PR description now describes the full frontend/backend implementation and the intentional decisions. |

## Verification

The Go suite, `go vet`, frontend currency-summary tests, frontend lint/build, and CGO-free Windows build passed. Frontend lint reports existing warnings. Local UI checks used only the fictional samples and a disposable database: previews, batch imports, duplicates, separate currency totals, search, asset filters, provider fields, original worksheets, and privacy masking. Bank income and expense totals stayed zero. No browser console errors were reported.
