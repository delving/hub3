# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- Build main executable: `make build`
- Run tests: `make test` or `make test-no-cache`
- Run a single test: `go test ./path/to/package -run TestName`
- Linting: `make staticcheck`
- Run development server: `make run-dev`

## Code Style Guidelines
- Follow standard Go conventions with gofmt
- Maximum line length: 140 characters
- Imports: organized by stdlib, external, then internal packages
- Error handling: use pkg/errors for wrapping errors
- Use domain-driven design with clear separation of concerns
- Tests: prefer table-driven tests with Ginkgo/Onsi framework
- Documentation: all exported functions require docstrings
- Naming: CamelCase for exported, camelCase for internal
- API implementations should follow consistent patterns with protobuf definitions
- Prefer composition over inheritance when designing interfaces

## Project Structure
- Main packages: hub3, ikuzo (newer implementation)
- Domain logic in ikuzo/domain
- Service implementations in ikuzo/service/x