# Mocktail

Upload an OpenAPI/GraphQL schema, or point it at a staging endpoint, and Mocktail spins up a realistic mock server, generates sample and edge-case payloads, and auto-writes contract tests for your CI. It then watches traffic to detect breaking changes before they reach production. Perfect for small teams and indie devs shipping APIs fast.

## Features

✅ **OpenAPI 3.x Parser** - Parse and validate OpenAPI specifications with detailed endpoint analysis  
✅ **Mock Server** - HTTP mock server with realistic responses based on schema endpoints  
🚧 **Test Generator** - Coming soon  
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
│   ├── parser/        # Schema parsing logic (OpenAPI 3.x implemented)
│   ├── mock/          # Mock server implementation
│   └── generator/     # Payload and test generation (planned)
├── examples/          # Sample API schemas for testing
└── bin/               # Compiled binaries
```

## Roadmap

- [x] OpenAPI 3.x schema parser
- [x] HTTP mock server with realistic responses
- [ ] GraphQL schema parser
- [ ] Payload generator (happy path & edge cases)
- [ ] Contract test generator
- [ ] Traffic monitoring & breaking change detection

## License

See [LICENSE](LICENSE) file for details.
