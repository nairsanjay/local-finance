# Integration tests

`integration/` contains statement-import regressions that exercise exported parser, service, and database APIs. Run them with `CGO_ENABLED=0 go test ./tests/integration`, or run `CGO_ENABLED=0 go test ./...` for the full suite.

Shared fixture helpers live in `integration/fixtures_test.go`. They create disposable databases and fabricated PDFs in memory; the checked-in synthetic savings PDFs remain in `samples/savings/`.

Unit tests that need private package helpers remain beside their Go source in `internal/`.
