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
- `main.go`: Defines `newRootCmd()` that returns configured cobra.Command, version = "0.1.0"
- **Subcommands** (one file per command):
  - `parse.go`: Parse and validate schemas with `-o/--output` flag (summary|verbose)
  - `mock.go`: Start HTTP mock server with `-p/--port` flag (default: 8080)
- Pattern: Each command has `newXxxCmd()` constructor returning `*cobra.Command` with `RunE` error handler
- Example: `mocktail parse examples/petstore.yaml -o verbose`

**OpenAPI Parser** (`internal/parser/`):

- `Parser` interface: Defines `Parse(filepath string) (*Schema, error)` contract
- `OpenAPIParser`: Uses `github.com/getkin/kin-openapi` library for OpenAPI 3.x validation
- `Schema` model: Normalized representation with Type, Version, Title, Paths, and Raw fields
- `Endpoint` model: Captures Method, Path, Summary, Description, Parameters per endpoint
- Validates specs with `doc.Validate(ctx)` before parsing
- `extractParameters()` helper pulls type info from OpenAPI schema references

**Mock Server** (`internal/mock/`):

- `Server` struct: Wraps `*http.Server`, parsed `*parser.Schema`, and port
- `NewServer(schema, port)`: Constructor pattern
- **Lifecycle**: `Start()` begins serving, `Stop(ctx)` graceful shutdown with 5s timeout
- **Routing**: One handler per path in schema, dispatches by HTTP method
- **Response Generation**: Path-based heuristics (list vs single resource detection via `{id}` in path)
  - GET list: Returns `{"data": [...], "total": N}` with 2 mock items
  - GET single: Returns `{"id": "uuid", "name": "...", "createdAt": "..."}`
  - POST: Returns 201 with resource + `"message": "Resource created successfully"`
  - PUT/PATCH: Returns resource + `"updatedAt"` timestamp
  - DELETE: Returns `{"message": "Resource deleted successfully"}`
- **Middleware**: `loggingMiddleware` logs `METHOD PATH STATUS DURATION` for every request
- **Health Check**: `/health` endpoint returns `{"status": "ok", "server": "mocktail"}`
- Signal handling: `os.Interrupt`/`SIGTERM` trigger graceful shutdown via context cancellation

**Build System**:

- Makefile with targets: `build`, `test`, `test-coverage`, `test-verbose`, `fmt`, `clean`, `deps`, `install`, `help`
- Outputs to `bin/mocktail`
- Go version: 1.25.0

### Components To Be Built

- `internal/generator/` - Payload and test data generation (schema-aware realistic data)
- `internal/monitor/` - Traffic watching & breaking change detection

## Development Workflow

### Essential Make Commands

```bash
make build           # Builds to bin/mocktail
make test            # Run all tests
make test-coverage   # Tests with coverage report
make test-verbose    # Tests with -v flag
make fmt             # Format code (go fmt ./...)
make clean           # Remove bin/ directory
make deps            # Download and tidy dependencies
make install         # Install to $GOPATH/bin
make help            # Show all available targets
```

### Running the Mock Server Locally

```bash
# Build and start mock server
make build
./bin/mocktail mock examples/petstore.yaml

# Custom port
./bin/mocktail mock examples/petstore.yaml --port 3000

# Test endpoints while server is running
curl http://localhost:8080/health
curl http://localhost:8080/pets        # GET list
curl http://localhost:8080/pets/123    # GET single resource
```

### Adding New Cobra Subcommands

Create a new file in `cmd/mocktail/` (e.g., `generate.go`) and register in `main.go`:

```go
// cmd/mocktail/generate.go
func newGenerateCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "generate <schema-file>",
        Short: "Generate test payloads",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
            return nil
        },
    }
    cmd.Flags().StringVarP(&format, "format", "f", "json", "Output format")
    return cmd
}

// cmd/mocktail/main.go - add to newRootCmd()
rootCmd.AddCommand(newGenerateCmd())
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

**Server Tests** (`internal/mock/server_test.go`):

- Start server on unique port per test (8081, 8082, etc.) to avoid conflicts
- Use goroutines with error channels: `go func() { errChan <- server.Start() }()`
- Wait for server startup: `time.Sleep(100 * time.Millisecond)` after launching
- Always test graceful shutdown with `context.WithTimeout()`
- Table-driven tests for endpoint methods: validate status codes AND response structure
- Use `checkResponse func(t *testing.T, body []byte)` callbacks for flexible assertions

**Conventions**:

- Use `strings.Contains()` for substring checks in descriptions
- Prefer `t.Fatalf()` when test cannot continue, `t.Errorf()` otherwise
- No test frameworks - plain `testing` package only
- Always `defer resp.Body.Close()` and `defer cancel()` for resources

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

### Mock Server Philosophy

- Stateless: responses should be deterministic from schema alone
- Use seed-based randomization for reproducible test data
- Prioritize "realistic" over "exhaustive" data generation
- Path-based heuristics: Detect list vs single resource by presence of `{id}` in path
- Status code mapping: POST→201, DELETE→200, others→200
- All responses include `X-Mocktail-Server: true` header for identification
- Logging format: `METHOD PATH STATUS_CODE DURATION` via custom `loggingResponseWriter`

## Key Files

- `cmd/mocktail/main.go` - CLI entry point with cobra setup
- `cmd/mocktail/parse.go` - Parse command implementation (RunE pattern)
- `cmd/mocktail/mock.go` - Mock server command with signal handling
- `cmd/mocktail/main_test.go` - Test pattern reference
- `internal/parser/parser.go` - Parser interface, OpenAPIParser, Schema models
- `internal/parser/parser_test.go` - Parser testing with t.TempDir() pattern
- `internal/mock/server.go` - HTTP mock server implementation
- `internal/mock/server_test.go` - Server testing with goroutines and deferred cleanup
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

**Next Priority**: Payload generator implementation (`internal/generator/`) with schema-aware realistic data
