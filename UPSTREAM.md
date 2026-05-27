# Upstream Tracking

Base: anomalyco/opencode @ v1.15.10
Fork date: 2026-05-15
Last upstream sync: 2026-05-27
Repository flattened: 2026-05-27

## Cherry-Pick Protocol

1. `git fetch upstream`
2. Review changelog for: bug fixes (take), new providers (take), new tools (evaluate), architecture changes (evaluate carefully)
3. `git cherry-pick <commit>` for each wanted change
4. Resolve conflicts against our modifications
5. Run full test suite before merging
6. Update this file with sync date and cherry-picked commits
7. **Budget: ~2-4 hours per upstream release**

## File Classification

### Kode-Only (never conflicts with upstream)
- `cmd/kode/`           — Go CLI entry point
- `internal/`           — Go verification engine (17 packages)
- `web/`                — React landing page (trykode.xyz)
- `packages/kode/src/bridge/` — Go↔TS IPC bridge (gatekeeper, ghost)
- `packages/kode/src/tool/verify-gate.ts` — Shared verification-on-write utility

### Modified from Upstream (cherry-pick carefully)
- `packages/kode/src/tool/edit.ts`       — verification hooks added
- `packages/kode/src/tool/write.ts`      — verification hooks added
- `packages/kode/src/tool/apply_patch.ts` — verification hooks (refactored to shared utility)
- `packages/kode/src/provider/`        — Kode provider added
- `packages/kode/src/agent/`           — Kode identity prompts
- `packages/kode/src/session/prompt/`  — Rebranded prompts
- `packages/core/src/`                 — Effect service IDs rebranded (@opencode/ → @kode/)
- `packages/ui/src/`                   — Theme, i18n, and CSS identifiers rebranded
- 158+ total files modified for rebrand

### Upstream-Identical (safe to cherry-pick)
- Most of `packages/kode/src/` (minus rebrand changes)
- `packages/llm/`
- `packages/plugin/`
- `packages/effect-drizzle-sqlite/`
- `packages/http-recorder/`

## External npm Dependencies (unchanged)

These are real npm packages from upstream — we reference them by their published name:
- `@gitlab/opencode-gitlab-auth` — GitLab auth integration
- `opencode-gitlab-auth` — GitLab auth
- `opencode-poe-auth` — Poe auth

## Sync Log

| Date | Upstream Version | Changes Taken | Notes |
|------|-----------------|---------------|-------|
| 2026-05-27 | v1.15.10 | Initial fork + full rebrand | Repository flattened, vendored/ eliminated |
