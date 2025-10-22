# Mocktail - AI Agent Instructions

## Project Overview

Mocktail is a Go-based API mocking tool for indie developers and small teams. It parses OpenAPI 3.x schemas and spins up mock HTTP servers with realistic, schema-aware responses.

**Status**: Early stage with parser, mock server, data generator, and `generate` CLI command implemented. Focus on simplicity over enterprise patterns—no external frameworks beyond Cobra and kin-openapi.

## Architecture

### Core Data Flow

```text
Schema File → Parser → Normalized Schema → Mock Server → Generator → HTTP Responses
                            ↓
                    CLI Generate Command → Test Payloads
```

1. **Parser** (`internal/parser/`): Reads OpenAPI YAML/JSON, validates with `doc.Validate(ctx)`, normalizes to internal `Schema` struct with `Endpoint` and `Parameter` models
2. **Mock Server** (`internal/mock/`): Routes HTTP requests based on path, dispatches by method to matched endpoint, uses Generator for response data, wraps with logging middleware
3. **Generator** (`internal/generator/`): Creates realistic mock data from `*openapi3.Schema` using seeded `math/rand` (not crypto/rand) for reproducibility

### Key Architectural Decisions

**Interface-Based Extension**: `Parser` interface (`Parse(filepath string) (*Schema, error)`) enables future GraphQL support. See `NewOpenAPIParser()` for implementation pattern—constructor returns concrete type, not interface.

**Schema Normalization**: Raw `*openapi3.T` converted to simplified `Schema`/`Endpoint`/`Parameter` models in parser layer. Original doc stored in `Schema.Raw` for Generator access. This decouples server routing logic from OpenAPI-specific types while preserving schema metadata for generation.

**Stateless Mock Server**: No database or state. Responses generated deterministically from schema + seed. Path patterns (presence of `{id}` path param) determine response structure in `generateMockResponse()`:
- `/pets` (no params) → `{"data": [...], "total": N}` (list with 2 items)
- `/pets/{id}` (has param) → `{"id": "...", "name": "..."}` (single object)

**Generator Seed Pattern**: `NewGenerator(seed int64)` ensures reproducible mock data across runs. CLI uses `time.Now().UnixNano()` for server, custom seed via `--seed` flag for `generate` command. All randomization via `gen.rng` field, not global `rand`.

## Development Workflow

### Build & Test Essentials

```bash
# Critical Makefile targets (run `make help` for all)
make build           # → bin/mocktail
make test            # Fast tests (no -v by default)
make test-coverage   # Coverage report
make test-verbose    # Verbose test output
make fmt             # go fmt ./...
make clean           # Remove bin/

# Running mock server
./bin/mocktail mock examples/petstore.yaml        # Default port 8080
./bin/mocktail mock examples/petstore.yaml -p 3000

# Testing live server
curl http://localhost:8080/health                 # {"status":"ok","server":"mocktail"}
curl http://localhost:8080/pets                   # List endpoint (2-item array)
curl http://localhost:8080/pets/abc123            # Single resource

# Generate command (CLI for test fixtures)
./bin/mocktail generate examples/petstore.yaml --path /pets --method GET --seed 42
./bin/mocktail generate examples/petstore.yaml --path /pets --method POST --count 3
```

### Testing Strategy (Critical)

**No test frameworks**—plain `testing` package only. No testify, ginkgo, etc. Patterns:

**Server Tests** (`internal/mock/server_test.go`): Port isolation + goroutines + graceful shutdown

```go
server := NewServer(schema, 8081)  // UNIQUE port per test! (8081, 8082, 8083...)
errChan := make(chan error, 1)
go func() { errChan <- server.Start() }()
time.Sleep(100 * time.Millisecond)  // Wait for server startup (critical!)

resp, err := http.Get("http://localhost:8081/test")
if err != nil { t.Fatalf(...) }
defer resp.Body.Close()  // Always defer Close()

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
server.Stop(ctx)  // Test cleanup—always stop servers
```

**Parser Tests** (`internal/parser/parser_test.go`): Use `t.TempDir()` for file isolation, inline test schemas as YAML strings written via `os.WriteFile()`.

**Table-Driven with Callbacks** (`internal/generator/generator_test.go`): Flexible assertion pattern via `check` function:

```go
tests := []struct {
    name   string
    schema *openapi3.Schema
    check  func(t *testing.T, result string)  // Custom validation logic
}{
    {
        name: "email format",
        schema: &openapi3.Schema{
            Type:   &openapi3.Types{"string"},
            Format: "email",
        },
        check: func(t *testing.T, result string) {
            if !contains(result, "@example.com") {
                t.Errorf("Expected email format, got: %s", result)
            }
        },
    },
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        result := gen.generateString(tt.schema)
        tt.check(t, result)
    })
}
```

**Resource Cleanup**: Always `defer resp.Body.Close()`, `defer cancel()`, `defer server.Stop(ctx)`.

### Adding CLI Commands

Pattern from `cmd/mocktail/generate.go`:

1. Create `cmd/mocktail/newcmd.go` with `newXxxCmd() *cobra.Command` function
2. Register in `cmd/mocktail/main.go` → `rootCmd.AddCommand(newXxxCmd())`
3. Use `RunE` for error returns, `Args: cobra.ExactArgs(1)` for positional args
4. Flags via local vars + `StringVarP`/`IntVarP`/`Int64VarP` in command definition

```go
func newGenerateCmd() *cobra.Command {
    var path, method string
    var seed int64
    
    cmd := &cobra.Command{
        Use:   "generate <schema-file>",
        Short: "Generate test payloads",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            schemaFile := args[0]
            p := parser.NewOpenAPIParser()
            schema, err := p.Parse(schemaFile)
            if err != nil {
                return fmt.Errorf("failed to parse schema: %w", err)
            }
            // ... implementation
            return nil
        },
    }
    
    cmd.Flags().StringVarP(&path, "path", "p", "", "API path (required)")
    cmd.Flags().StringVarP(&method, "method", "m", "", "HTTP method (required)")
    cmd.Flags().Int64VarP(&seed, "seed", "s", time.Now().UnixNano(), "Random seed")
    
    return cmd
}
```

## Code Conventions

### Error Handling

- **Always wrap errors with context**: `return fmt.Errorf("failed to parse schema: %w", err)`
- Cobra commands use `RunE` for error returns (not `Run`)
- Main function: `newRootCmd().Execute()` returns error → print to stderr → `os.Exit(1)`
- No panics except truly exceptional cases (not implemented in codebase currently)

### Generator Pattern (Schema-Aware Data Generation)

Core implementation in `internal/generator/generator.go`:

```go
// Constructor with seed for reproducibility
gen := NewGenerator(seed int64)  // Creates gen.rng from rand.NewSource(seed)

// Main entry point: type-switch on schema.Type.Slice()[0]
func (g *Generator) GenerateFromSchema(schema *openapi3.Schema) (interface{}, error)

// Type handlers (private methods):
- generateString(schema)   → format-aware (email, uuid, date-time, uri)
- generateInteger(schema)  → respects Min/Max constraints
- generateNumber(schema)   → respects Min/Max, returns float64
- generateBoolean()        → random bool via rng.Intn(2)
- generateArray(schema)    → uses MinItems/MaxItems, recursively generates items
- generateObject(schema)   → iterates schema.Properties, builds map[string]interface{}
```

**Critical details**:
- OpenAPI types accessed via `schema.Type.Slice()[0]` (not `schema.Type` directly—it's `*openapi3.Types`)
- Enum handling: check `len(schema.Enum) > 0`, pick random index, type-assert result
- Format strings: `date-time` → RFC3339, `date` → `2006-01-02`, `email` → `user<N>@example.com`
- UUID format: custom 8-4-4-4-12 hex string (not stdlib `uuid` package)

### Mock Server Internals

Key flow in `internal/mock/server.go`:

1. **Route registration** (`Start()`): One `mux.HandleFunc(path, handler)` per schema path, closure captures `pathEndpoints` slice
2. **Method dispatch** (`handlePath()`): Iterates endpoints, matches `r.Method` (case-insensitive), returns 405 if no match
3. **Response generation** (`generateMockResponse()`):
   - First tries schema-based generation via `generator.GenerateResponse(operation, statusCode)`
   - Path pattern detection: `!strings.Contains(endpoint.Path, "{")` → list endpoint → wraps in `{"data": [item, item], "total": 2}`
   - Fallback: basic mock structure based on method (GET list vs GET single vs POST/PUT/DELETE)
4. **Status codes** (`getStatusCode()`): POST → 201, DELETE → 200 (not 204 for body), default → 200
5. **Middleware** (`loggingMiddleware`):
   ```go
   lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
   next.ServeHTTP(lrw, r)
   log.Printf("%s %s %d %v", r.Method, r.URL.Path, lrw.statusCode, duration)
   ```
   - Custom `loggingResponseWriter` captures status code via `WriteHeader()` override
   - Logs: `GET /pets 200 1.234ms` format
6. **Graceful shutdown** (`Stop()`): Uses `server.Shutdown(ctx)` with 5s timeout on SIGTERM/SIGINT

**No goroutine pooling, no caching**—every request generates fresh data from schema.

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
