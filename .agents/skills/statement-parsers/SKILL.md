---
name: statement-parsers
description: Add or repair LocalFinance bank, credit-card, and investment statement parsers while preserving the existing plugin interface, account model, and import identity.
---

# LocalFinance statement parsers

Read the repository's `AGENTS.md`, `internal/parser/parser.go`, the closest existing adapter, and the relevant extractors before changing a parser. Inspect `internal/service/transaction_service.go` when changing detection or identity; its preview and import paths must agree.

## Fit the existing plugin

Implement `StatementParser` in a format-specific file under `internal/parser/`, register through `DefaultRegistry`, and return the existing `StatementMeta` and `ParsedTransaction` types. Keep institution-specific headings, detection signals and metadata inside its adapter. Put genuinely shared mechanics in the parser package's utilities or existing extractor package. Do not add a second registry, parallel transaction model, inheritance framework, or service-level bank switch.

Detection uses confidence scores. The registry extracts compressed PDF text once before asking adapters to identify it. Test a generic filename as well as the bank's customary filename; raw PDF bytes often hide text in compressed streams. Encrypted PDFs are unlocked by the existing service/extractor path with a user-supplied password.

Extract labeled account identifiers rather than guessing from arbitrary long numbers. Populate full account numbers when available and use the existing `XX` plus last-four convention for masks. Missing identifiers can cause same-bank accounts to merge in `GetOrCreateAccount`; cover distinct-account imports whenever the format exposes identity.

## Parse complete records

Reuse the PDF, CSV, and Excel extractors, Indian date/amount normalization, and `CleanNarration`. Preserve source narration and references because import fingerprints include both. Assemble wrapped descriptions and references before cleaning them. Preserve relevant cleaner fields such as UPI VPA and card last four.

For positional PDFs, use each active table header's coordinates and exclude page headers, footers, and summary sections from transactions. A numeric zero balance is present; distinguish missing values from zero. Keep transaction amounts positive and express direction through `TxType`.

Reconcile available running balances and printed totals. Source order may be ascending or descending, including several transactions on one date: preserve or reverse the complete sequence rather than sorting solely by date. Reject incomplete or contradictory financial rows rather than presenting a partial import as complete.

Bank debits/credits and credit-card purchases/payments have different meanings. Follow the existing account type and adapter conventions; do not reclassify transfers or create merchant/person rules as a side effect of parsing.

## Investment formats

First inspect the current account types, transaction fields, and service support. Do not force holdings snapshots, units, security identifiers, valuations, or corporate actions into bank debit/credit transactions. If the existing model cannot represent the requested investment data faithfully, identify the required model/service extension before implementing the adapter. Password rules belong to the unlock path; retain no credentials or identifying examples in source or fixtures.

## Verification

Use fabricated data in the same layout, including a sample PDF when needed. Keep personal statements, decrypted exports, PANs, passwords, customer details and financial databases out of commits. Small in-memory PDF construction inside tests can cover layout variations without a separate generator command.

Cover meaningful format behavior: detection, dates, amounts/directions, identity, narration/reference continuation, zero balances, reverse and same-day ordering, page-specific headers, and reconciliation failures as applicable. Exercise both service preview and import for detection/account changes; verify repeated imports remain idempotent and distinct accounts stay distinct. Preserve the existing transaction hash and user-edit behavior.

Run affected tests first, then `CGO_ENABLED=0 go test ./...` and `go vet ./...` for the completed parser change. Do not include unrelated application redesign or sample-generation tooling.
