# Mocktail - AI Agent Instructions

## Project Overview

Mocktail is a Go-based API mocking tool for indie developers and small teams. It parses OpenAPI 3.x schemas and spins up mock HTTP servers with realistic, schema-aware responses.

**Status**: Early stage with parser, mock server, and data generator implemented. Focus on simplicity over enterprise patterns.

## Architecture

### Core Data Flow

```text
Schema File → Parser → Normalized Schema → Mock Server → Generator → HTTP Responses
```

1. **Parser** (`internal/parser/`): Reads OpenAPI YAML/JSON, validates with `doc.Validate(ctx)`, normalizes to internal `Schema` struct
2. **Mock Server** (`internal/mock/`): Routes HTTP requests, dispatches to Generator, returns JSON responses with logging middleware
3. **Generator** (`internal/generator/`): Creates realistic mock data from OpenAPI schemas using seeded randomization

### Key Architectural Decisions

**Interface-Based Extension**: `Parser` interface enables future GraphQL support. See `NewOpenAPIParser()` for implementation pattern.

**Schema Normalization**: Raw OpenAPI structures converted to simplified `Schema`/`Endpoint`/`Parameter` models in parser layer. This decouples server logic from OpenAPI-specific types.

**Stateless Mock Server**: No database or state. Responses generated deterministically from schema + seed. Path patterns (e.g., `{id}` presence) determine response structure:

- `/pets` (list) → `{"data": [...], "total": N}`
- `/pets/123` (single) → `{"id": "...", "name": "..."}`

**Generator Seed Pattern**: `NewGenerator(seed)` ensures reproducible mock data across runs. Uses `math/rand` with custom seed, not `crypto/rand`.

## Development Workflow

### Build & Test Essentials

```bash
# Critical Makefile targets (see Makefile for all)
make build           # → bin/mocktail
make test            # Fast tests (no -v)
make test-coverage   # Coverage report
make fmt             # go fmt ./...

# Running mock server
./bin/mocktail mock examples/petstore.yaml       # Default port 8080
./bin/mocktail mock examples/petstore.yaml -p 3000

# Testing live server
curl http://localhost:8080/health  # Health check
curl http://localhost:8080/pets    # List endpoint
```

### Testing Strategy (Critical)

**No test frameworks** - plain `testing` package only. Patterns:

**Server Tests**: Port isolation + goroutines + graceful shutdown

```go
server := NewServer(schema, 8081)  // Unique port per test!
errChan := make(chan error, 1)
go func() { errChan <- server.Start() }()
time.Sleep(100 * time.Millisecond)  // Wait for startup

// Make HTTP requests...

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
server.Stop(ctx)  // Always test cleanup
```

**Parser Tests**: Use `t.TempDir()` for file isolation, inline test schemas as strings.

**Table-Driven**: See `internal/generator/generator_test.go` for callback pattern:

```go
tests := []struct {
    name   string
    schema *openapi3.Schema
    check  func(t *testing.T, result string)  // Flexible assertion
}{ /*...*/ }
```

**Resource Cleanup**: Always `defer resp.Body.Close()`, `defer cancel()`, etc.

### Adding CLI Commands

1. Create `cmd/mocktail/newcmd.go` with `newXxxCmd() *cobra.Command`
2. Register in `cmd/mocktail/main.go`: `rootCmd.AddCommand(newXxxCmd())`
3. Pattern: `RunE` returns errors, flags via `StringVarP`/`IntVarP`, exact args with `cobra.ExactArgs(1)`

Example:

```go
func newGenerateCmd() *cobra.Command {
    var format string
    return &cobra.Command{
        Use:   "generate <schema-file>",
        Short: "Generate test payloads",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
            return fmt.Errorf("not implemented: %w", err)
        },
    }
}
```

## Code Conventions

### Error Handling

- Wrap with context: `fmt.Errorf("failed to X: %w", err)`
- Cobra commands use `RunE`, return errors directly
- Main checks `Execute()` error, prints to stderr, `os.Exit(1)`

### Generator Pattern (New Feature Implementation)

See `internal/generator/generator.go`:

- Constructor with seed: `NewGenerator(seed int64)`
- Type-switch on OpenAPI schema types (string/integer/number/boolean/array/object)
- Format-aware generation: `date-time` → RFC3339, `email` → user@example.com, `uuid` → custom format
- Respect constraints: `Min`/`Max` for numbers, `MinItems`/`MaxItems` for arrays, `Enum` for enums

### Mock Server Internals

- One `mux.HandleFunc(path, ...)` per schema path in `Start()`
- `handlePath()` dispatches by HTTP method to matched endpoint
- `generateMockResponse()` tries schema-based generation first, falls back to basic mocks
- Custom middleware: `loggingMiddleware` wraps handlers with `loggingResponseWriter`
- Graceful shutdown: `server.Shutdown(ctx)` with 5s timeout on SIGTERM/SIGINT

## Project Structure

```text
cmd/mocktail/        # One file per command (main.go, parse.go, mock.go)
internal/parser/     # Parser interface, OpenAPIParser, Schema/Endpoint/Parameter models
internal/mock/       # Server struct, HTTP handlers, middleware
internal/generator/  # Generator struct, schema type generators (string/int/array/object)
examples/            # petstore.yaml for testing
bin/                 # Build output (gitignored)
```

**No `/pkg/`** - all code is internal to mocktail binary.

## Dependencies

- `github.com/spf13/cobra` v1.10.1 - CLI framework
- `github.com/getkin/kin-openapi` v0.133.0 - OpenAPI 3.x parsing/validation
- Go 1.25.0

## Next Priorities

1. ~~Payload generator~~ ✅ Implemented
2. Contract test generator (use Generator to create test fixtures)
3. GraphQL parser (implement `Parser` interface)
4. Traffic monitoring (`internal/monitor/`)

## Common Pitfalls

1. **Port conflicts in tests**: Always use unique ports (8081, 8082, etc.)
2. **Server startup race**: Add `time.Sleep(100 * time.Millisecond)` after goroutine launch
3. **Unclosed resources**: Always defer `resp.Body.Close()` in HTTP tests
4. **Schema type slices**: OpenAPI types are `*openapi3.Types`, access via `.Slice()[0]`
5. **Version bumps**: Update `version` const in `cmd/mocktail/main.go` for releases
