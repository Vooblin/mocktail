# Mocktail - AI Agent Instructions

## Project Overview

Mocktail is an API mocking and contract testing tool targeting small teams and indie developers. The core value proposition:

- **Input**: OpenAPI/GraphQL schemas or staging endpoints
- **Output**: Realistic mock servers, test payloads, auto-generated contract tests, and traffic-based breaking change detection

This is an **early-stage Go project** - prioritize simplicity and indie dev workflows over enterprise patterns.

## Current Architecture

### Implemented Components

**CLI Framework** (`cmd/mocktail/`):

- Uses `cobra` (github.com/spf13/cobra v1.10.1) for command structure
- `main.go`: Defines `newRootCmd()` that returns configured cobra.Command
- `parse.go`: Implements `parse <schema-file>` subcommand with `-o/--output` flag (summary|verbose)
- Version: Hardcoded `version = "0.1.0"` constant in main.go

**OpenAPI Parser** (`internal/parser/`):

- `Parser` interface: Defines `Parse(filepath string) (*Schema, error)` contract
- `OpenAPIParser`: Uses `github.com/getkin/kin-openapi` library for OpenAPI 3.x validation
- `Schema` model: Normalized representation with Type, Version, Title, Paths, and Raw fields
- `Endpoint` model: Captures Method, Path, Summary, Description, Parameters per endpoint
- Validates specs with `doc.Validate(ctx)` before parsing
- Example usage: `mocktail parse examples/petstore.yaml -o verbose`

**Build System**:

- Makefile with targets: `build`, `test`, `test-coverage`, `lint` (fmt+vet), `clean`, `deps`, `help`
- Outputs to `bin/mocktail`
- Go version: 1.25.0

### Components To Be Built

- `internal/mock/` - HTTP mock server implementation  
- `internal/generator/` - Payload and test data generation
- `internal/monitor/` - Traffic watching & breaking change detection

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

All components follow table-driven or scenario-based testing:

**Command Tests** (`cmd/mocktail/*_test.go`):

```go
func TestParseCommand(t *testing.T) {
    cmd := newParseCmd()
    // Test command properties with plain if checks
    if cmd.Use != "parse <schema-file>" {
        t.Errorf("Expected Use 'parse <schema-file>', got '%s'", cmd.Use)
    }
    // Validate flags exist
    outputFlag := cmd.Flags().Lookup("output")
    if outputFlag == nil {
        t.Error("Expected 'output' flag to exist")
    }
}
```

**Parser Tests** (`internal/parser/parser_test.go`):

- Use `t.TempDir()` for test file isolation
- Write test OpenAPI specs inline as strings
- Test happy path validation (titles, paths, parameters)
- Test error cases (nonexistent files, invalid specs)
- Example: `TestOpenAPIParser_Parse` validates parsed Schema struct fields

**Conventions**:

- Use `strings.Contains()` for substring checks in descriptions
- Prefer `t.Fatalf()` when test cannot continue, `t.Errorf()` otherwise
- No test frameworks - plain `testing` package only

### Error Handling

- Return errors explicitly from functions with context wrapping:

  ```go
  return fmt.Errorf("failed to parse schema: %w", err)
  ```

- Cobra RunE functions return error: `RunE: func(cmd *cobra.Command, args []string) error`
- Main entry point checks `Execute()` result, prints to stderr, exits with code 1
- Parser validates OpenAPI specs with `doc.Validate(ctx)` before processing

### Project Layout (Standard Go)

- `/cmd/mocktail/` - CLI commands (each command in separate file: `parse.go`, `mock.go`, etc.)
- `/internal/parser/` - Schema parsing (OpenAPI implemented, GraphQL planned)
- `/internal/mock/` - Mock server (not yet implemented)
- `/internal/generator/` - Data generation (not yet implemented)
- `/examples/` - Sample schemas for testing (`petstore.yaml`)
- `/bin/` - Build output (gitignored)
- No `/pkg/` - all code is internal to mocktail

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

**Current Implementation** (`internal/parser/parser.go`):

- Interface-based design: `Parser` interface with `Parse(filepath string) (*Schema, error)`
- `OpenAPIParser` uses `github.com/getkin/kin-openapi` library
- Validation before parsing: `doc.Validate(ctx)` catches spec errors early
- Normalization: Converts OpenAPI-specific structures to generic `Schema`, `Endpoint`, `Parameter` types
- Parameter extraction: `extractParameters()` helper pulls type info from OpenAPI schema references

**Data Model**:

```go
type Schema struct {
    Type    string                // "openapi" or "graphql"
    Version string                // e.g., "3.0.0"
    Title   string
    Paths   map[string][]Endpoint // Path -> methods
    Raw     interface{}           // Original parsed object
}
```

**Extension Strategy**: Add GraphQL by implementing `Parser` interface (see `NewOpenAPIParser()` pattern)

## Key Files

- `cmd/mocktail/main.go` - CLI entry point with cobra setup
- `cmd/mocktail/parse.go` - Parse command implementation (RunE pattern)
- `cmd/mocktail/main_test.go` - Test pattern reference
- `internal/parser/parser.go` - Parser interface, OpenAPIParser, Schema models
- `internal/parser/parser_test.go` - Parser testing with t.TempDir() pattern
- `Makefile` - Build automation and available commands
- `go.mod` - Go 1.25.0, cobra v1.10.1, kin-openapi v0.133.0
- `examples/petstore.yaml` - Sample OpenAPI spec for testing
- `README.md` - User-facing documentation and quick start

## Next Steps for AI Agents

When implementing features:

1. Create packages in `internal/` following Standard Go Project Layout
2. Add corresponding cobra subcommands in `cmd/mocktail/` (separate file per command)
3. Follow established testing pattern: `t.TempDir()` for files, plain if checks, no frameworks
4. Implement interfaces where possible (see `Parser` for pattern)
5. Update Makefile if new build steps are needed
6. Increment version constant in main.go for releases

**Next Priority**: Mock server implementation (`internal/mock/`) with HTTP handler that uses parsed Schema
