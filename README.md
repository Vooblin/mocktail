# Mocktail

Upload an OpenAPI/GraphQL schema, or point it at a staging endpoint, and Mocktail spins up a realistic mock server, generates sample and edge-case payloads, and auto-writes contract tests for your CI. It then watches traffic to detect breaking changes before they reach production. Perfect for small teams and indie devs shipping APIs fast.

## Quick Start

### Prerequisites

- Go 1.21 or later

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
# Run mocktail
./bin/mocktail

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
│   └── mocktail/       # CLI entry point
├── internal/           # Private application code
│   ├── parser/        # Schema parsing logic (to be added)
│   ├── mock/          # Mock server implementation (to be added)
│   └── generator/     # Payload and test generation (to be added)
├── api/               # OpenAPI/schema definitions (to be added)
└── bin/               # Compiled binaries
```

## Roadmap

- [ ] OpenAPI 3.x schema parser
- [ ] GraphQL schema parser
- [ ] Realistic mock server
- [ ] Payload generator (happy path & edge cases)
- [ ] Contract test generator
- [ ] Traffic monitoring & breaking change detection

## License

See [LICENSE](LICENSE) file for details.
