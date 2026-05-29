# Kode Roadmap

Kode is a contrarian AI coding agent — a functional architecture hijack of
opencode that replaces generate-and-pray with a structured
**Plan → Critique → Generate → Verify → Apply → Test** workflow.

## Fork Strategy

```
C:\kode
├── cmd/kode/          ← Go CLI entry point (kode.exe)
├── internal/          ← Go engine (graph, verify, execution, llm)
├── third_party/
│   └── opencode/      ← vendored opencode monorepo (TS/Bun), rebranded as Kode TUI
├── go.mod             ← Go module: github.com/kode/kode
├── bin/kode.exe       ← compiled Go binary (~10MB)
└── logs/kode.log      ← telemetry audit log
```

**Vendor model over git fork:** Clean-slate repo (`git init`) with opencode in
`vendored/opencode/` avoids upstream cadence pressure. Updates pulled on our
schedule (zig or subtree).

---

## Phase 0: Fork Health (deferred)

- [ ] Fetch full history from opencode upstream
- [ ] Set up a rebase workflow for the 158 modified TS files
- [ ] Verify `git subtree pull` works cleanly

---

## Phase 1: Standalone CLI (v1.0.0 — done)

```
kode plan     — Build surgical 8K context graph from entry files
kode verify   — Verify file content or hunks through 4-gate check
kode stats    — Analyze gatekeeper audit log for failure patterns
```

- [x] Context engine (go/parser + go/ast, 8K token cap)
- [x] 4-gate verification (syntax → imports → calls → architecture)
- [x] Executor with cumulative state, rollback, atomic commit
- [x] Telemetry and analytics (`kode stats`, `logs/kode.log`)

---

## Phase 2: LLM Integration (v1.1.0 — done)

- [x] `kode generate <prompt>` — call LLM, return structured hunks
- [x] `kode run <prompt>` — full generate → verify → apply pipeline
- [x] Wire `--model` flag through to the LLM provider

---

## Phase 3: Full Loop (v1.2.0 — done)

- [x] `kode loop <task>` — full Plan→Generate→Verify→Apply→Test cycle
- [x] Auto-retry on verify failure (3 rounds)
- [x] Test step (auto-detect: go test, npm test, cargo test)
- [x] Rollback on test failure (snapshot + restore)

---

## Phase 4: Rebrand & TUI (v2.0.0 — done)

- [x] Rebrand all user-facing strings: "opencode" → "Kode" / "kode"
- [x] New KODE ASCII logo and wordmark
- [x] CLI script name: `opencode` → `kode`
- [x] Env vars: `OPENCODE_*` → `KODE_*`
- [x] Config files: `opencode.json` → `kode.json`, `.opencode` → `.kode`
- [x] Internal URLs: `opencode.internal` → `kode.internal`
- [x] All HTTP headers, User-Agent, provider names, MCP client name
- [x] `kode tui` — unified CLI entry point that spawns TS TUI
- [x] Gatekeeper binary resolution via `KODE_BIN` env var + fallback search
- [x] 158 TS files modified for full rebrand

---

## Phase 5: Rebrand Completion (v3.0.0 — done)

- [x] Agent identity prompts (9 session prompt files) — OpenCode → Kode
- [x] Env var flags (40 OPENCODE_* → KODE_*) across 23 consumer files
- [x] HTTP headers, GitHub URLs → sicario-labs/kode
- [x] Provider IDs (opencode → kode, opencode-go → kode-go)
- [x] SDK types (OpencodeClient → KodeClient, 50+ files)
- [x] customize-kode.md skill file rebrand (49 references)
- [x] TSX imports (@opencode-ai/ → @kode/), missed in v2.0.0

---

## Phase 6: CI & Release (v3.0.0 — done)

- [x] `kode init` — scaffold `.kode/kode.json` with sensible defaults
- [x] CI pipeline — Go test + build on Linux/Windows/macOS
- [x] Release workflow — cross-platform binaries on tag push
- [x] `go install github.com/sicario-labs/kode/cmd/kode@latest`

---

## Phase 7: Polish

- [ ] `kode explain <error-id>` — deep Markdown explanation of gate failures
- [ ] Dynamic graph expansion — fetch missing symbols on demand
- [ ] Better CLI output (colors, spinners, progress bars)
- [ ] Install bun + node_modules automatically on first `kode tui`
- [ ] Document modified TS files for upstream sync

---

## Phase 8: Enterprise Dominance (v4.0.0 — Unicorn Roadmap)

- [x] **Blindfold Mode (Zero-Knowledge Obfuscation)**
  - [x] Core two-pass AST-pattern de-identification engine (`internal/blindfold`)
  - [x] Bidirectional symbol translator integration with LLM API requests
  - [x] Comment/Docstring de-identification to generic mathematical/behavioral functional specs
- [x] **Gate 7: "Live-Fire" Sandbox Replay**
  - [x] Sub-50ms local WebAssembly (WASM) / Docker sandboxed runtime container
  - [x] Heap profiling to detect and block infinite loops & memory leaks
  - [x] Network socket monitoring to prevent malicious data exfiltration
- [x] **"Zero-Human" PR Gateway (Kode CIV)**
  - [x] Continuous Integration & Verification (CIV) GitHub Action/GitLab Runner
  - [x] Multi-gate auto-merge protocol for verified, compilation-guaranteed patches
- [ ] **Living Spec Sync**
  - [ ] Bidirectional AST-to-Specification compiler contract
  - [ ] Automated updates to Mintlify docs (`docs.trykode.xyz`) on semantic drift
- [x] **Gate 7: Dynamic QR Code Tunnel Preview Upgrade**
  - [x] Tunnel localhost to public secure endpoint via localhost.run
  - [x] Render high-contrast Unicode block scannable QR Code matrix in the terminal console
- [x] **Gate 8: Headless Browser Verification**
  - [x] Build automated dev server booter and port binder detector
  - [x] Create serverless Playwright E2E step and JS script execution runner
  - [x] Sync screenshots and walkthrough recordings for visual proof walkthroughs
- [x] **Kode Command Voice: Voice-to-Action Bridge**
  - [x] Zero-CGo native audio capture backend (Win32 winmm, ALSA arecord, Mac AppleScript CoreAudio)
  - [x] Cloud/local transcription integration and execution mapping loop


---

## Design Principles

1. **Zero-baggage verification** — gate runs in <50ms, user never notices
2. **Fail-closed by default** — if the binary is missing, no patch gets through
3. **Tiered escalation** — 2 silent auto-retries, then human-in-the-loop
4. **Telemetry-driven** — every gate call is logged; stats inform architecture
5. **Upstream-friendly** — minimize TS modifications, prefer subprocess IPC
