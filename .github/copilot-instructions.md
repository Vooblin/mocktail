# Mocktail - AI Agent Instructions

## Project Overview

Mocktail is an API mocking and contract testing tool targeting small teams and indie developers. The core value proposition:

- **Input**: OpenAPI/GraphQL schemas or staging endpoints
- **Output**: Realistic mock servers, test payloads, auto-generated contract tests, and traffic-based breaking change detection

This is an **early-stage Go project** - prioritize simplicity and indie dev workflows over enterprise patterns.

## Current Architecture

### Existing Structure

- **CLI Framework**: Uses `cobra` (github.com/spf13/cobra v1.10.1) for command structure
- **Entry Point**: `cmd/mocktail/main.go` - defines `newRootCmd()` that returns configured cobra.Command
- **Version**: Hardcoded `version = "0.1.0"` constant in main.go
- **Build System**: Makefile-based workflow (see below)
- **Go Version**: 1.25.0 (go.mod)

### Components To Be Built

The `internal/` directory is currently empty. Planned structure:

- `internal/parser/` - OpenAPI 3.x/GraphQL schema parsing
- `internal/mock/` - HTTP mock server implementation  
- `internal/generator/` - Payload and test data generation

## Development Workflow

### Essential Make Commands

```bash
make build           # Builds to bin/mocktail
make test            # Run all tests
make test-coverage   # Tests with coverage report
make lint            # Runs fmt + vet
make clean           # Remove bin/ directory
make help            # Show all available targets
```

### Adding New Cobra Subcommands

When implementing features, extend `cmd/mocktail/main.go`:

```go
func newRootCmd() *cobra.Command {
    rootCmd := &cobra.Command{ /* ... */ }
    
    // Uncomment and implement as features are built:
    // rootCmd.AddCommand(newMockCmd())
    // rootCmd.AddCommand(newGenerateCmd())
    
    return rootCmd
}
```

## Code Conventions

### Testing Pattern

See `cmd/mocktail/main_test.go` for established pattern:

- Test constructors (e.g., `TestRootCommand` validates `newRootCmd()` properties)
- Use plain `if` checks with `t.Errorf()` for simple validations
- Use `strings.Contains()` for substring checks in descriptions

### Error Handling

- Return errors explicitly from functions
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Main entry point handles errors: checks `Execute()` result, prints to stderr, exits with code 1

### Project Layout (Standard Go)

- `/cmd/mocktail/` - CLI entry point and commands
- `/internal/` - Private application code (not importable by external projects)
- `/bin/` - Build output (gitignored)
- No `/pkg/` yet - only add if code needs external reuse

## Design Principles

### Target Audience Impact

**Indie developers & small teams** means:

- Single binary distribution (no complex installation)
- Minimal external dependencies (currently only cobra)
- Makefile over complex build tooling
- Clear, self-documenting CLI help text

### Mock Server Philosophy

- Stateless: responses should be deterministic from schema alone
- Use seed-based randomization for reproducible test data
- Prioritize "realistic" over "exhaustive" data generation

### Schema Parsing Strategy

- Parsers should implement common interface for pluggability
- Start with OpenAPI 3.x (most common), add GraphQL later
- Question to resolve: Best-effort vs strict schema validation

## Key Files

- `cmd/mocktail/main.go` - CLI entry point with cobra setup
- `cmd/mocktail/main_test.go` - Test pattern reference
- `Makefile` - Build automation and available commands
- `go.mod` - Go 1.25.0, cobra v1.10.1
- `README.md` - User-facing documentation and quick start

## Next Steps for AI Agents

When implementing features:
1. Create packages in `internal/` following Standard Go Project Layout
2. Add corresponding cobra subcommands in `cmd/mocktail/main.go`
3. Follow established testing pattern from `main_test.go`
4. Update Makefile if new build steps are needed
5. Increment version constant in main.go for releases

Start with MVP: OpenAPI parser → simple mock server → basic payload generation.
