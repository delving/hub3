# Repository Guidelines

## Project Structure & Module Organization
Hub3 is a Go monorepo. Primary runtime code lives in `ikuzo/`, with the `ikuzo/ikuzoctl` CLI as the main entrypoint. Domain-specific logic, API bindings, and generated assets live under `hub3/` (for fragments, EAD processing, and server rendering). Configuration helpers sit in `config/`, while environment-ready templates and TOML manifests land in `configs/` (e.g., `configs/development` for local playbooks). Shared assets, docs, and scripts are in `docs/`, `doc-templates/`, `tools/`, and top-level helpers such as `build-api-docs.sh`. Test data typically accompanies each package via `*_test.go` files and `testdata/` folders.

## Build, Test, and Development Commands
- `make build`: compiles `ikuzo/ikuzoctl` into `build/`.
- `make build-static`: produces a Linux-static binary for deployments.
- `make test`: runs `go tool richgo test -cover ./ikuzo/...` and `staticcheck`.
- `make pre-commit`: tidy modules, run a heavier race-enabled suite, and static analysis.
- `go tool air` or `make run-dev`: start the live-reload dev server (requires `.air.toml`).

## Coding Style & Naming Conventions
Follow Go 1.x practices: tabs for indentation, CamelCase for exported identifiers, and lower_snake for configuration files and directories. Always run `gofmt`/`goimports` plus `staticcheck` on touched files; CI expects modules to be tidy and generated assets updated via `go generate ./...` when proto or embedded resources change.

## Testing Guidelines
Prefer table-driven unit tests beside their packages (`*_test.go`). Use `go test ./ikuzo/...` for focused checks; integration and semantic validations live in scripts like `test_semantic.sh`. Maintain coverage by extending existing suites when adding features and include fixtures under `testdata/` when parsing structured files.

## Commit & Pull Request Guidelines
Commit history follows Conventional Commits (`feat:`, `fix:`, `chore:`). Write concise messages that describe the scope. PRs should link relevant issues, summarise behaviour, note config or schema changes, and document manual steps. Include the commands you ran (`make test`, `go test ./...`) and attach screenshots or sample payloads when you touch HTTP responses or rendered fragments.

## Security & Configuration Tips
Never commit secrets; keep credentials out of `configs/` directories and prefer environment overrides. Review `hub3.toml` and `config/` defaults before deploying, and document changes to access control or indexing pipelines in `docs/` so downstream agents stay aligned.
