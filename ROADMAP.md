# Kode Roadmap

Kode is a contrarian AI coding agent — a functional architecture hijack of
opencode that replaces generate-and-pray with a structured
**Plan → Critique → Generate → Verify → Apply → Test** workflow.

## Fork Strategy

```
opencode upstream (origin/dev @ 47f3332)
  │
  └─ github.com/EmmyCodes234/kode (our fork)
       │
       ├── packages/opencode/     ← TS monorepo (upstream code, 3 files patched)
       ├── cmd/kode/              ← Go CLI (our engine)
       ├── internal/              ← Go packages (graph, verify, execution)
       └── bin/kode.exe           ← compiled binary
```

**Git strategy:** Keep opencode's TS code pristine where possible. Our Go engine
is additive (new files). Only 3 TS files are modified — easily rebased.

---

## Phase 0: Fork Health (immediate)

Deepen the shallow clone to enable clean upstream merges.

- [ ] Fetch full history from opencode upstream
- [ ] Add opencode as a second remote
- [ ] Set up a rebase workflow for the 3 patched TS files
- [ ] Verify `git merge opencode/dev` would work cleanly

---

## Phase 1: Standalone CLI (current — v1.0.0)

The Go binary works independently of the TS monorepo. It provides deterministic
safety without needing node_modules or opencode's runtime.

```
kode plan     — Build surgical 8K context graph from entry files
kode verify   — Verify file content or hunks through 4-gate check
kode stats    — Analyze gatekeeper audit log for failure patterns
```

- [x] Plan command (Phase 0)
- [x] Verify command (Phase 1)
- [x] Verify command — telemetry logger (v1.0.0)
- [x] Stats command (v1.0.0)

---

## Phase 2: LLM Integration (next)

Kode currently cannot generate patches — it can only verify them. To close the
loop, we need an LLM provider interface in Go.

**Options:**
- **A:** Shell out to opencode's Bun/TS runtime for LLM calls (reuse existing providers)
- **B:** Implement a lightweight Go LLM client (OpenAI/Anthropic only, ~200 lines)
- **C:** Use opencode's SDK as a subprocess (heavier but more complete)

**Implementation: Native Go OpenAI-compatible client** — works with any
OpenAI-compatible endpoint (OpenAI, LiteLLM, Ollama, local inference).
Configured via env vars: KODE_LLM_API_KEY, KODE_LLM_ENDPOINT, KODE_LLM_MODEL.

Bun-based opencode provider federation is available as a future extension
for non-OpenAI providers (Anthropic, Google, Groq, etc.).

- [x] `kode generate <prompt>` — call LLM, return structured hunks
- [x] `kode run <prompt>` — full generate → verify → apply pipeline (--apply flag on generate)
- [x] Wire `--model` flag through to the LLM provider
- [ ] `kode apply <hunks>` — verify + apply directly from a JSON file (can use verify --input)

---

## Phase 3: Full Loop

The Plan → Generate → Verify → Apply → Test cycle, fully automated.

- [x] `kode loop <task>` — entry point for the full cycle
- [x] Auto-retry on verify failure (3 rounds, already built in executor)
- [x] Test step (run `go test`, `npm test`, or custom command after apply)
- [x] Rollback on test failure (restore from snapshot)

---

## Phase 4: Upstream Sync

Regularly merge opencode updates while keeping our patches.

- [ ] Document the 3 patched TS files and their changes
- [ ] Set up CI that tests both Go (`go test ./...`) and TS (if node_modules available)
- [ ] Investigate upstreaming gatekeeper.ts as an optional plugin

---

## Phase 5: Polish

- [ ] `kode explain <error-id>` — deep Markdown explanation of gate failures
- [ ] `kode init` — scaffold `.kode.yaml` with architecture rules
- [ ] Dynamic graph expansion — fetch missing symbols on demand
- [ ] Better CLI output (colors, spinners, progress bars)
- [ ] `go install github.com/EmmyCodes234/kode/cmd/kode@latest`

---

## Design Principles

1. **Zero-baggage verification** — gate runs in <50ms, user never notices
2. **Fail-closed by default** — if the binary is missing, no patch gets through
3. **Tiered escalation** — 2 silent auto-retries, then human-in-the-loop
4. **Telemetry-driven** — every gate call is logged; stats inform architecture
5. **Upstream-friendly** — minimize TS modifications, prefer subprocess IPC
