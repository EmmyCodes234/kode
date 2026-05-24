# Kode

Kode is a contrarian AI coding agent — a functional architecture hijack of
opencode that replaces generate-and-pray with a structured
**Plan → Critique → Generate → Verify → Apply → Test** workflow.

## Architecture

```
C:\kode
├── cmd/kode/          ← Go CLI entry point
├── internal/          ← Go engine (graph, verify, execution)
├── third_party/
│   └── opencode/      ← vendored opencode monorepo (TS/Bun)
├── go.mod             ← Go module: github.com/kode/kode
└── ROADMAP.md         ← full roadmap and fork strategy
```

The Go engine communicates with opencode's TypeScript code via subprocess IPC
(`kode.exe verify --input <json>`). The TS side (opencode TUI) calls into the
Go binary as a verification oracle before writing patches to disk.

## Commands

- `kode plan <task>` — Build surgical 8K context graph
- `kode verify --input <file>` — Verify file content through 4-gate check
- `kode stats` — Analyze gatekeeper audit log
- `kode run <task>` — Full generate→verify→apply pipeline (planned)

## Build

```bash
go build -o bin/kode.exe ./cmd/kode
```

Tests: `go test ./...` (99 tests across 5 packages)

## Upstream

Vendored from [github.com/anomalyco/opencode](https://github.com/anomalyco/opencode) v1.15.10.
TS files modified: `third_party/opencode/packages/opencode/src/bridge/gatekeeper.ts`,
`third_party/opencode/packages/opencode/src/tool/apply_patch.ts`.
