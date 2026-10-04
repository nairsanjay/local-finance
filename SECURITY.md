# Security Policy

## Security Model & Privacy Guarantee

LocalFinance is engineered from the ground up as a **100% offline, privacy-first desktop application**:
- **Zero Remote Network Calls**: LocalFinance makes no outbound requests to third-party analytics, cloud sync services, or external APIs.
- **Local Storage**: All financial records, parsed transactions, and database keys reside solely on your local disk in an embedded SQLite database (`local_finance.db` or `~/.localfinance/local_finance.db`).
- **In-Memory Decryption**: Password-protected PDF statements are decrypted strictly in memory and are never written to disk unencrypted.

---

## ⚠️ Critical: Do NOT Post Real Statements or PII in Public Issues

Because LocalFinance parses sensitive financial documents:
- **Never upload raw bank statements, scanned PDFs, or unredacted CSV/Excel files** to public GitHub issues or discussions.
- If you are reporting a parser bug or requesting support for a new bank format:
  1. Replace all real names with generic personas (e.g. `RAHUL SHARMA`).
  2. Replace all real 10-to-16 digit bank account and credit card numbers with dummy numbers (e.g. `XXXX1234`).
  3. Redact all salary amounts, phone numbers, and home addresses.
  4. Ensure any uploaded fixture is purely synthetic.

---

## Reporting a Security Vulnerability

If you discover a potential security vulnerability within LocalFinance (such as local privilege escalation, command injection, path traversal in statement extraction, or local database credential leakage), please report it responsibly:

1. **GitHub Private Vulnerability Advisory (Preferred)**:
   Submit a report via the **Security** tab of this repository: [Report a vulnerability](https://github.com/usmslm102/local-finance/security/advisories/new).
2. **Direct Contact**:
   Alternatively, email **`usmslm102@gmail.com`** with:
   - A description of the issue and potential impact
   - Minimal reproduction steps or proof-of-concept
   - The affected version or operating system

Please give us reasonable time to investigate and address the vulnerability before disclosing it publicly.
