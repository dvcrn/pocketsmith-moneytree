- Repo: dvcrn/pocketsmith-moneytree
# Repository Guidelines

## Project Structure & Module Organization
- `main.go` is the entrypoint that wires the sync workflow.
- `moneytree/` contains Moneytree API and sync logic.
- `sanitizer/` contains text normalization helpers and tests (e.g., `sanitizer/sanitizer_test.go`).
- `Dockerfile` defines the container build; `go.mod`/`go.sum` track Go dependencies.

## Build, Test, and Development Commands
Use `mise` tasks (see `mise.toml`) rather than `make`.

```bash
mise run build   # Compile the binary in the repo root
mise run run     # Run the sync locally
mise run test    # Run all Go tests
mise run format  # Apply gofmt to the codebase
```

For container publishing (pushes to registry):
```bash
mise run docker-build
```

## Coding Style & Naming Conventions
- Format with `gofmt`; Go uses tabs for indentation.
- Package names are short, lowercase (e.g., `sanitizer`).
- Exported identifiers use `CamelCase`; files use lowercase with underscores (e.g., `sanitizer_test.go`).

## Testing Guidelines
- Tests use Go’s standard `testing` package and live alongside code.
- Run all tests with `mise run test` (`go test ./...`).
- Name files `*_test.go` and functions `TestXxx` (e.g., `TestSanitize`).

## Configuration & Secrets
- Required environment variables: `MONEYTREE_USERNAME`, `MONEYTREE_PASSWORD`, `MONEYTREE_API_KEY`, `POCKETSMITH_TOKEN`.
- Alternatively pass flags: `-username`, `-password`, `-apikey`, `-pocketsmith-token`.
- Never commit credentials; use local environment or CI secrets.

## Commit & Pull Request Guidelines
- Follow recent history: short, present-tense, capitalized subjects with no period (e.g., “Add Docker publish GitHub Action”).
- PRs should describe the change and why, list test commands run, and link issues when relevant.
