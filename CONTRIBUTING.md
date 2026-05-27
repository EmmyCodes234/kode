# Contributing to Kode

First off, thank you for considering contributing to Kode! It's people like you that make Kode a world-class verification engine for AI coding.

## The Architecture
Kode is split into two primary domains:
1. **The Engine (Go)**: Located in `internal/` and `cmd/kode/`. This is the brain. It handles the 6 verification gates, the AST/regex parsing, the git tracking, and the MCP server.
2. **The TUI (TypeScript/React)**: Located in `packages/kode/`. This is the interactive terminal interface that acts as the primary user surface.

## Setting up your Development Environment

### Prerequisites
- [Go](https://golang.org/dl/) 1.21 or higher
- [Bun](https://bun.sh/) 1.1 or higher
- [Node.js](https://nodejs.org/) 20.x or higher

### Building the Engine
```bash
# Clone the repository
git clone https://github.com/sicario-labs/kode.git
cd kode

# Run the test suite
go test ./...

# Build the CLI binary
go build -o bin/kode.exe ./cmd/kode
```

### Running the TUI
```bash
cd packages/kode
bun install
bun run build

# Run the TUI
bun run dev
```

## Pull Request Process

1. Ensure any new verification gates or logic in the Go engine is covered by tests (`go test ./...`).
2. If modifying the AST parsers (`internal/verify/multilang_treesitter.go`), ensure you also update the regex fallback (`internal/verify/multilang_regex.go`) to maintain graceful degradation.
3. Update the `README.md` or documentation if you are adding new CLI flags or capabilities.
4. Submit your PR with a clear description of the problem and your solution.

## Code of Conduct

Please treat all maintainers and contributors with respect. Constructive critique is highly encouraged; personal attacks are strictly forbidden.
