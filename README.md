# Mocktail

🍹 A lightweight API mocking tool for indie developers and small teams. Point Mocktail at an OpenAPI schema and get a realistic mock server with schema-aware responses—no configuration needed.

## Features

✅ **OpenAPI 3.x Parser** - Parse and validate OpenAPI specifications  
✅ **Mock Server** - HTTP mock server with realistic, schema-driven responses  
✅ **Schema-Aware Generator** - Produces realistic mock data respecting types, formats, and constraints  
✅ **Contract Test Generator** - Generate test payloads from OpenAPI schemas  
🚧 **Traffic Monitor** - Coming soon

## Quick Start

### Prerequisites

- Go 1.25 or later

### Installation

```bash
# Clone the repository
git clone https://github.com/Vooblin/mocktail.git
cd mocktail

# Build the binary
make build

# Or install to $GOPATH/bin
make install
```

### Usage

```bash
# Parse and validate an OpenAPI schema
./bin/mocktail parse examples/petstore.yaml

# Parse with verbose output (shows all endpoints)
./bin/mocktail parse examples/petstore.yaml -o verbose

# Start a mock server from an OpenAPI schema
./bin/mocktail mock examples/petstore.yaml

# Start mock server on a custom port
./bin/mocktail mock examples/petstore.yaml --port 3000

# Test the mock server
curl http://localhost:8080/health
curl http://localhost:8080/pets
curl http://localhost:8080/pets/123

# Generate test payloads from schema
./bin/mocktail generate examples/petstore.yaml --path /pets --method GET --seed 42

# Generate request body for POST endpoint
./bin/mocktail generate examples/petstore.yaml --path /pets --method POST --seed 100

# Generate multiple test fixtures
./bin/mocktail generate examples/petstore.yaml --path /pets --method GET --count 5 --seed 42

# Show version
./bin/mocktail --version

# Show help
./bin/mocktail --help
```

## Development

### Building

```bash
# Build the binary
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Clean build artifacts
make clean

# See all available commands
make help
```

### Project Structure

```text
mocktail/
├── cmd/
│   └── mocktail/       # CLI entry point and commands
├── internal/           # Private application code
│   ├── parser/        # OpenAPI 3.x schema parsing and validation
│   ├── mock/          # HTTP mock server with middleware
│   └── generator/     # Schema-aware mock data generation
├── examples/          # Sample API schemas for testing
└── bin/               # Compiled binaries (gitignored)
```

## How It Works

1. **Parse**: Validates OpenAPI spec with `doc.Validate(ctx)`, normalizes to internal schema model
2. **Route**: Creates HTTP handlers for each endpoint in the schema
3. **Generate**: Produces realistic responses using seeded randomization—respects types, formats, enums, and min/max constraints
4. **Serve**: Returns JSON with appropriate status codes (POST→201, DELETE→200, etc.)

Responses are deterministic (same seed = same data) and path-aware:

- `/pets` → `{"data": [...], "total": N}` (list)
- `/pets/123` → `{"id": "...", "name": "..."}` (single resource)

## Roadmap

- [x] OpenAPI 3.x schema parser with validation
- [x] HTTP mock server with realistic responses
- [x] Schema-aware data generator (types, formats, constraints)
- [x] Contract test generator
- [ ] GraphQL schema parser
- [ ] Traffic monitoring & breaking change detection

## License

See [LICENSE](LICENSE) file for details.
