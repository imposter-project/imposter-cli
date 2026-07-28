# Agent guide

## Build Commands
- Build: `make build`
- Run: `make run`
- Format code: `make fmt`
- Run all tests: `make test`
- Run specific test: `go test ./path/to/package -run TestName`
- Code coverage: `make coverage`
- HTML coverage report: `make coverage-html`

## Code Style Guidelines
- Follow Go standard formatting (gofmt)
- Import order: stdlib, then external, then internal packages
- Use logrus for logging via the logger instance
- Error handling: return errors to caller, use cobra.CheckErr()
- Prefer descriptive variable/function names over abbreviations
- Use snake_case for file names, camelCase for Go identifiers
- TypeScript/JavaScript: use 2 spaces for indentation

## Directory Structure
- cmd/: CLI commands (cobra)
- internal/: private implementation packages
- docs/: documentation files

## Documentation

User-facing product documentation is maintained in the Imposter docs repo
(`imposter-project/imposter`) and published to
[docs.imposter.sh](https://docs.imposter.sh/). That repo is the source of truth
for user docs.

When you add or change user-facing documentation, put it in the docs repo and
link to docs.imposter.sh from here — don't duplicate the content in this repo.
Reserve this repo's `docs/` for maintainer- and developer-facing material (for
example the SDK/embedding guide). Existing user-facing pages here are being
migrated to the docs repo over time.
