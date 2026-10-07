# Fictional bank history samples

These checked-in PDFs contain fabricated accounts and two transactions dated April 10–11, 2026: a withdrawal of 100 and a deposit of 250, with balances of 900 and 1,150. They are used by the account identity, idempotent import, and PDF password integration tests.

Each bank (ICICI and Union Bank) has statements for accounts `123456781234`, `123456785678`, and `987654321234`. The first and third share the last four digits so tests verify that the full account identity is retained.

The PDFs have compressed content streams and deliberately large synthetic metadata, making each file larger than 64 KB to exercise full-document detection. Tests load these files directly; no generator is required. Password tests encrypt a fixture in memory to check authentication and decoding failures.
