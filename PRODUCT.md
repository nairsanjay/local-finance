# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

People in India who want useful personal-finance intelligence without giving a cloud service access to their bank and credit-card history.

## Product Purpose

LocalFinance turns exported bank and credit-card statements into a private, searchable view of spending, cash flow, budgets, subscriptions, merchants, salary, and card rewards. Success means a user can understand and act on their finances while their data remains on their own computer.

## Positioning

LocalFinance combines India-specific statement parsing and narration intelligence with a fully offline architecture: the React interface, Go backend, and SQLite database ship together as one portable executable with no cloud sync, telemetry, SMS scraping, or account aggregator.

## Operating Context

Users import PDF, CSV, Excel, and password-protected statements from Indian banks and credit-card providers, then review and organize the resulting local ledger in a browser-based interface served from the application binary.

## Capabilities and Constraints

- Imports HDFC, ICICI, Axis, SBI, and generic statement formats through an extensible parser registry.
- Understands common Indian payment narrations including UPI, POS, IMPS, NEFT, RTGS, bank charges, and card billing details.
- Provides transaction search, cash-flow analysis, budgeting, subscriptions, merchant intelligence, salary insights, card recommendations, transfer reconciliation, and annual review.
- Preserves user-edited categories, tags, and notes across repeat imports, with deterministic transaction identity and idempotent upserts.
- Runs offline and stores data in local SQLite. Application features must not add cloud services, telemetry, remote logging, SMS scraping, third-party aggregators, or external API calls.
- Distribution remains a single CGO-free binary for Windows, macOS, and Linux.

## Brand Commitments

The product name is LocalFinance. The established identity uses a shield-and-rupee mark, an understated dark neutral palette, and emerald as its primary accent.

## Evidence on Hand

- Product details and setup: `README.md`
- Implemented and planned scope: `ROADMAP.md`
- Product mark: `frontend/src/assets/logo.svg`
- Existing React application: `frontend/src/`
- No testimonials, customer logos, usage benchmarks, or third-party endorsements are available; future work must not fabricate them.

## Product Principles

- Financial data stays on the user's machine.
- Imported data remains deterministic while user edits remain durable.
- India-specific financial behavior is a first-class domain, not an afterthought.
- Setup and distribution remain portable and low-friction.
- Insights should be actionable without obscuring their underlying transactions.

## Accessibility & Inclusion

The web interface should remain keyboard-accessible, responsive, readable at common zoom levels, and usable with reduced motion.
