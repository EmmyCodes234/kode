# Kode Battle Plan — Implementation Tasks

> Tracking progress on the battle plan from `kode_battle_plan.md`

---

## Phase 1A: Repository Flattening ✅

- [x] Move `vendored/opencode/packages/*` up to `packages/`
- [x] Move root config files (`package.json`, `tsconfig.json`, `bun.lock`) to repo root
- [x] Rename `packages/opencode` → `packages/kode`
- [x] Update Go bridge paths (`tui.go`, `proxy.go`, `bundledl.go`, `main.go`)
- [x] Update root `package.json` workspace config
- [x] Update `tsconfig.json` if needed
- [x] Update `netlify.toml`
- [x] Delete `vendored/` directory
- [x] Prune unused packages (`enterprise/`, `desktop/`, `console/`, `storybook/`, `app/`, `web/`, `slack/`, `docs/`, `function/`, `identity/`, `extensions/`, `containers/`, `npm/`)

## Phase 1B: Kill Every "opencode" Reference ✅

- [x] Update `@opencode/` Effect service identifiers → `@kode/` in `core/src/` (17 files)
- [x] Update `core/package.json` bin entry (`"opencode"` → `"kode"`)
- [x] Handle external npm packages — kept as-is (real npm packages: `@gitlab/opencode-gitlab-auth`, `opencode-gitlab-auth`, `opencode-poe-auth`)
- [x] Update root `package.json` (`"name": "opencode"` → `"kode"`, repository URL → sicario-labs/kode)
- [x] Update config detection paths (`.opencode/` → `.kode/`) in comments
- [x] Update data directories (`~/.opencode` → `~/.kode`)
- [x] Update test fixtures (all TS/TSX test files updated)
- [x] Update system prompts
- [x] Update comments and docs
- [x] Update `packages/ui/` opencode theme references (CSS vars, sprite IDs, highlight API names, i18n)
- [x] Rename TUI keymap symbols (`OPENCODE_BASE_MODE` → `KODE_BASE_MODE`, etc.)
- [x] Rename SDK types (`OpencodeClient` → `KodeClient`, etc.)
- [x] Update all theme JSON `$schema` URLs (`opencode.ai` → `trykode.xyz`)
- [~] Pruned packages (desktop, console, app, etc.) — deleted, no updates needed

### Excluded from rebranding (legitimate external references):
- `registry.npmjs.org/opencode-ai` — upstream npm package URL for version checking
- `@gitlab/opencode-gitlab-auth`, `opencode-gitlab-auth`, `opencode-poe-auth` — real npm packages
- `test/fixtures/recordings/` — recorded API traffic (HTTP cassettes, cannot modify)
- `test/tool/fixtures/models-api.json` — upstream provider definitions

## Phase 1C: Ownership Documentation ✅

- [x] Create `UPSTREAM.md` at repo root

## Phase 1D: Upstream Sync Process ✅

- [x] Add upstream remote (`git remote add upstream https://github.com/anomalyco/opencode.git`)

## Verification Gates ✅

- [x] `grep -ri "opencode" packages/ --include="*.ts" --include="*.tsx"` returns **1 result** (npm registry URL — legitimate)
- [x] Go build succeeds (`go build -o bin/kode.exe ./cmd/kode`)
- [x] Go tests pass (121 tests across 6+ packages)
- [x] `netlify.toml` build command verified

---

## Phase 2: Deep Engine Integration

### 2.1 — Verification-on-Write ✅

- [x] Add `verify` config schema to `kode.json` (`config.ts`)
- [x] Create shared verification utility (`src/tool/verify-gate.ts`)
- [x] Hook gatekeeper into `edit.ts` — auto-verify after every edit
- [x] Hook gatekeeper into `write.ts` — auto-verify after every write
- [x] Refactor `apply_patch.ts` to use shared utility (was already hooked, now unified)
- [x] All three tools return `[✓ verified]` badge on pass, `[✗ N gate(s) failed]` on failure
- [x] On failure: error returned as tool output → LLM self-corrects
- [x] Config-driven: `verify.enabled`, `verify.auto_retry`, `verify.block_architecture`

### 2.2 — Ghost Branches in TUI ✅

- [x] Expose `internal/ghost/` in the TUI (via `bridge/ghost.ts`)
- [x] Config: `"ghost": { "enabled": true, "branches": 3 }`
- [x] Show "Speculating..." indicator with branch scores (events mapped)
- [x] User picks winner or accepts auto-selected (via engine selection)

### 2.3 — Multi-Language Verification ✅

- [x] Build pure-Go multi-language syntax parser (regex-based, zero-CGo)
- [x] Create TypeScript query file (`typescript.scm`) for future tree-sitter use
- [x] Create Python query file (`python.scm`) for future tree-sitter use
- [x] Refactor verification gates to accept parsed multi-lang structure
- [x] Auto-detect language from file extension
- [x] Support TypeScript, JavaScript, Python, and Rust syntax/imports


---

## Phase 3: Kode-Only Features

### 3.1 — Critique Engine ✅

- [x] Implement `internal/critique/lenses/blast_radius.go`
- [x] Implement `internal/critique/lenses/coherence.go`
- [x] Implement `internal/critique/lenses/convention.go`
- [x] Implement `internal/critique/lenses/dependency.go`
- [x] Create `DefaultEngine` factory
- [x] Wire into `internal/workflow/pipeline.go` (StageCritique)

### 3.2 — Session Memory ✅
- [x] Implement `.kode/memory.json` (Switched from SQLite to JSON to avoid CGo/download issues)
- [x] Track verification failures
- [x] Track blast radius trends
- [x] Persist successful ghost branch strategies

### 3.3 — Daemon Mode ✅
- [x] Create `internal/daemon/` to watch git commits
- [x] Run via `kode daemon`
- [x] IPC notifications to TUI

### 3.4 — MCP Server Mode ✅
- [x] Implement `kode mcp serve`
- [x] Use Context Engine for prompt context
- [x] Bind Write/Verify loop to tools

---

Legend: `[ ]` = pending, `[x]` = completed, `[-]` = in progress, `[~]` = skipped
