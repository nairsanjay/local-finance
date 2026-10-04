## Summary
<!-- Provide a clear and concise description of the changes proposed in this PR. -->

## Motivation / Context
<!-- Why is this change necessary? Fixes an issue? Adds support for a new bank parser? -->
Fixes #

## Type of Change
- [ ] 🏦 New bank / statement parser plugin
- [ ] 🐛 Bug fix (non-breaking change fixing an issue)
- [ ] ✨ New feature or UI improvement
- [ ] ⚡ Performance optimization or refactor
- [ ] 📚 Documentation update

## Verification & Quality Checklist
- [ ] **Zero Network Calls**: Verified that no external HTTP requests, analytics, or telemetry have been added.
- [ ] **Pure Go / Zero CGO**: Verified that backend code builds with `CGO_ENABLED=0`.
- [ ] **Privacy & No PII**: Verified that NO real bank statements, real passwords, real account numbers, or real salary figures are committed in any test fixture or file.
- [ ] **Backend Tests**: `go test -v ./...` passes cleanly.
- [ ] **Frontend Build**: `pnpm --prefix frontend build` completes without TypeScript or lint errors.
- [ ] **Local Verification**: Verified by running the application locally on `http://127.0.0.1:8080`.
