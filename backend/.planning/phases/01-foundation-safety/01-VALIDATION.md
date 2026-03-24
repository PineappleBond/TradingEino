---
phase: 01
slug: foundation-safety
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-24
---

# Phase 01 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) + testify v1.11.1 |
| **Config file** | none — standard Go test files |
| **Quick run command** | `go test ./internal/agent/tools/... -run TestOkxGetPositionsTool -v` |
| **Full suite command** | `go test ./... -v` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/agent/tools/... -v`
- **After every plan wave:** Run `go test ./... -v`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 1 | FOUND-01 | unit | `go test ./pkg/okex/... -v` | ❌ W0 | ⬜ pending |
| 01-01-02 | 01 | 1 | FOUND-01 | unit | `go test ./internal/agent/tools/... -v` | ❌ W0 | ⬜ pending |
| 01-02-01 | 02 | 1 | FOUND-02 | unit | `go test ./internal/agent/tools/... -v` | ❌ W0 | ⬜ pending |
| 01-03-01 | 03 | 2 | FOUND-03 | unit | `go test ./internal/agent/... -v` | ❌ W0 | ⬜ pending |
| 01-04-01 | 04 | 2 | FOUND-04 | integration | `go test ./internal/agent/... -v` | ❌ W0 | ⬜ pending |
| 01-05-01 | 05 | 3 | FOUND-05 | manual | N/A — signal handling | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `pkg/okex/okx_error_test.go` — OKXError type tests
- [ ] `internal/agent/tools/okx_get_positions_test.go` — rate limiter + error handling
- [ ] `internal/agent/tools/okx_get_fundingrate_test.go` — rate limiter + error handling
- [ ] `internal/agent/tools/okx_candlesticks_test.go` — rate limiter verification
- [ ] `internal/agent/agents_test.go` — sync.Once + context propagation tests
- [ ] Framework install: `go mod download` — dependencies already in go.mod

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Graceful shutdown on SIGINT | FOUND-05 | Requires interactive signal | 1. Start app with `go run cmd/server/main.go` 2. Send Ctrl+C 3. Verify shutdown logs appear in order: Server → Scheduler → Agents → DB → Logger |
| Context cancellation propagation | FOUND-04 | Requires runtime verification | 1. Add test that cancels context during agent operation 2. Verify all child operations terminate |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
