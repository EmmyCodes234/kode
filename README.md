<div align="center">
  <img src="./web/public/kode-logo-dark.svg" alt="Kode Logo" width="200" />
  <h1>Kode</h1>
  <p><strong>The verification-first AI coding agent. No generation without validation.</strong></p>
  <p>
    <a href="https://github.com/sicario-labs/kode/actions"><img src="https://img.shields.io/github/actions/workflow/status/sicario-labs/kode/ci.yml?branch=main" alt="CI Status" /></a>
    <a href="https://www.npmjs.com/package/kode"><img src="https://img.shields.io/npm/v/kode" alt="NPM Version" /></a>
    <a href="https://github.com/sicario-labs/kode/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License" /></a>
  </p>
</div>

---

## The Incumbent Problem
The entire AI coding market—Cursor, Copilot, Cline—relies on **generate-and-pray**. An LLM generates code, and the tool writes it directly to your filesystem. You, the human, are forced to be the verification layer. You review diffs, run compilers, catch hallucinated imports, and roll back broken patches.

## The Kode Thesis
**Kode is the world's first verification-first AI coding agent.** 

Every generated patch passes through a deterministic, compiled Go binary that runs 6 verification gates in under 50ms before it ever touches your filesystem. If any gate fails, the patch is rejected, and the LLM self-corrects based on compiler-grade feedback.

The LLM is the generative engine. **Kode is the security layer.**

## 6 Deterministic Verification Gates

1. **Syntax Gate**: Dual-engine architecture. Full Tree-sitter AST validation when available, gracefully falling back to fast regex heuristics. Parses Go, TypeScript, JavaScript, Python, and Rust. Parse error = hard block.
2. **Imports Gate**: Validates every import path against the project dependency graph. Catches hallucinated packages before they compile.
3. **Calls Gate**: Checks that every function and method call references a symbol that actually exists. Eliminates the #1 source of LLM hallucinations.
4. **Blast Radius Gate**: Walks the dependency graph backward from every modified file. If downstream impact exceeds your configured threshold, the patch is blocked.
5. **Architecture Gate**: Enforces declared module boundaries. Prevents the LLM from crossing microservice lines or importing banned packages.
6. **Security Gate**: Automated vulnerability scanning on generated code. SQL injection, XSS, and hardcoded secrets are caught before they're committed.

---

## Beyond Verification

Verification is just the beginning. Kode introduces features no incumbent offers:

- **Ghost Branches**: Why run one prompt when you can run three? Kode spawns parallel git worktrees, testing multiple speculative strategies simultaneously. It evaluates the patches, tests them, and merges the highest-scoring survivor back into your working tree.
- **Blindfold Mode**: Enterprise-grade privacy. Kode SHA-256 obfuscates your identifiers (package names, functions, types) *before* LLM submission. Your proprietary logic is never exposed in plaintext to external models.
- **Critique Lenses**: A pre-generation review layer that rejects structurally flawed ideas before the LLM wastes tokens generating them.
- **MCP Server**: Want to keep using Cursor or Claude Desktop? Run `kode mcp serve` to expose Kode's verification engine as an MCP tool. Let other agents generate code, but let Kode verify it.

## Installation

Kode is a Bring Your Own Key (BYOK) platform. You provide the API key, we provide the engine. We natively support 25+ providers including OpenAI, Anthropic, Gemini, Groq, local Ollama models, and the Kode Gateway.

```bash
# Install globally via NPM
npm install -g kode
```

## Quick Start

Initialize Kode in any project:
```bash
kode init
```

Run a complete Plan → Critique → Generate → Verify → Apply → Test loop:
```bash
kode loop "Refactor the memory storage to use SQLite instead of JSON"
```

Or launch the interactive TUI:
```bash
kode tui
```

## Architecture

Kode is built for speed and safety.
- **Verification Engine**: 100% Go. Compiled, deterministic, zero-CGo fallback architecture.
- **TUI**: High-performance terminal interface built in TypeScript and React.

### Dual-Engine AST Parsing
Kode's syntax validation runs on a dual-engine architecture:
- `//go:build cgo` compiles with official Tree-sitter AST bindings for 100% precision.
- `//go:build !cgo` gracefully falls back to an ultra-fast, zero-dependency Regex implementation, ensuring Kode installs on any environment without breaking your build.

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details on setting up your local environment and submitting PRs.

## License

Kode is released under the [MIT License](LICENSE).
