---
phase: 3
slug: execution-automation
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-24
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (built-in) |
| **Config file** | none — uses project go.mod |
| **Quick run command** | `go test ./internal/agent/tools/... -run TestOkxPlaceOrder -v` |
| **Full suite command** | `go test ./internal/agent/tools/... -v` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/agent/tools/... -v`
- **After every plan wave:** Run `go test ./internal/agent/... -v`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | TDD | Status |
|---------|------|------|-------------|-----------|-------------------|-----|--------|
| 01-01 | 01 | 1 | EXEC-01 | integration | `go test ./... -run TestOkxPlaceOrder` | ✅ | ⬜ pending |
| 01-02 | 01 | 1 | EXEC-02 | integration | `go test ./... -run TestOkxCancelOrder` | ✅ | ⬜ pending |
| 01-03 | 01 | 1 | EXEC-03 | integration | `go test ./... -run TestOkxGetOrder` | ✅ | ⬜ pending |
| 02-01 | 02 | 1 | EXEC-05 | integration | `go test ./... -run TestOkxAttachSlTp` | ✅ | ⬜ pending |
| 02-02 | 02 | 1 | EXEC-05 | integration | `go test ./... -run TestOkxPlaceOrderWithSlTp` | ✅ | ⬜ pending |
| 03-01 | 03 | 2 | EXEC-04 | build | `go build ./internal/agent/executor_agent/...` | ✅ | ⬜ pending |
| 03-02 | 03 | 2 | EXEC-04 | build | `go build ./internal/agent/...` | ✅ | ⬜ pending |
| 03-03 | 03 | 2 | EXEC-04 | manual | Human verification checkpoint | N/A | ⬜ pending |
| 04-01 | 04 | 3 | EXEC-01 | integration | `go test ./... -run TestBatchPlaceOrder` | ✅ | ⬜ pending |
| 04-02 | 04 | 3 | EXEC-02 | integration | `go test ./... -run TestBatchCancelOrder` | ✅ | ⬜ pending |
| 04-03 | 04 | 3 | EXEC-03 | integration | `go test ./... -run TestGetOrderHistory` | ✅ | ⬜ pending |
| 04-04 | 04 | 3 | EXEC-01 | integration | `go test ./... -run TestClosePosition` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

**TDD Pattern:** Tests are created alongside implementation (not as separate Wave 0). Each task with `tdd="true"` creates both `.go` and `_test.go` files.

- [x] Wave 0 not required — TDD pattern used (tests with implementation)
- [x] All implementation tasks have `<automated>` verify commands
- [x] Manual-only verification documented (EXEC-04 Level 1 autonomy)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Executor Agent Level 1 autonomy | EXEC-04 | Agent behavior constraint, not testable via unit test | 1. Run OKXWatcher with Executor Agent 2. Ask Executor directly to place order (should refuse) 3. Ask OKXWatcher to analyze, then execute (should execute via OKXWatcher command) |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify (Task 03-03 is manual checkpoint, surrounded by build-verified tasks)
- [x] Wave 0 covers all MISSING references — N/A (TDD pattern)
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
