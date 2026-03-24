---
phase: 3
slug: execution-automation
status: draft
nyquist_compliant: false
wave_0_complete: false
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

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 01-01 | 01 | 1 | EXEC-01 | integration | `go test ./... -run TestOkxPlaceOrder` | ❌ W0 | ⬜ pending |
| 01-02 | 01 | 1 | EXEC-02 | integration | `go test ./... -run TestOkxCancelOrder` | ❌ W0 | ⬜ pending |
| 01-03 | 01 | 1 | EXEC-03 | integration | `go test ./... -run TestOkxGetOrder` | ❌ W0 | ⬜ pending |
| 02-01 | 02 | 2 | EXEC-05 | integration | `go test ./... -run TestOkxAttachSlTp` | ❌ W0 | ⬜ pending |
| 02-02 | 02 | 2 | EXEC-05 | integration | `go test ./... -run TestOkxPlaceOrderWithSlTp` | ❌ W0 | ⬜ pending |
| 02-03 | 02 | 2 | EXEC-06 | unit | `go test ./... -run TestOrderValidation` | ❌ W0 | ⬜ pending |
| 03-01 | 03 | 3 | EXEC-04 | unit | `go test ./... -run TestExecutorAgent` | ❌ W0 | ⬜ pending |
| 03-02 | 03 | 3 | EXEC-01,EXEC-02,EXEC-03 | integration | `go test ./... -run TestOrderFlow` | ❌ W0 | ⬜ pending |
| 04-01 | 04 | 4 | EXEC-01 | integration | `go test ./... -run TestBatchPlaceOrder` | ❌ W0 | ⬜ pending |
| 04-02 | 04 | 4 | EXEC-02 | integration | `go test ./... -run TestBatchCancelOrder` | ❌ W0 | ⬜ pending |
| 04-03 | 04 | 4 | EXEC-03 | integration | `go test ./... -run TestGetOrderHistory` | ❌ W0 | ⬜ pending |
| 04-04 | 04 | 4 | EXEC-01 | integration | `go test ./... -run TestClosePosition` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Wave 0 creates test stubs for all tools before implementation begins:

- [ ] `internal/agent/tools/okx_place_order_test.go` — stubs for EXEC-01, EXEC-06
- [ ] `internal/agent/tools/okx_cancel_order_test.go` — stubs for EXEC-02
- [ ] `internal/agent/tools/okx_get_order_test.go` — stubs for EXEC-03
- [ ] `internal/agent/tools/okx_attach_sl_tp_test.go` — stubs for EXEC-05
- [ ] `internal/agent/tools/okx_place_order_with_sl_tp_test.go` — stubs for EXEC-05
- [ ] `internal/agent/tools/okx_batch_place_order_test.go` — stubs for EXEC-01 (batch)
- [ ] `internal/agent/tools/okx_batch_cancel_order_test.go` — stubs for EXEC-02 (batch)
- [ ] `internal/agent/tools/okx_get_order_history_test.go` — stubs for EXEC-03
- [ ] `internal/agent/tools/okx_close_position_test.go` — stubs for EXEC-01 (close)
- [ ] `internal/agent/executor_agent/executor_agent_test.go` — stubs for EXEC-04
- [ ] Mock OKX responses for offline testing (uses existing `pkg/okex` test helpers)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Executor Agent Level 1 autonomy | EXEC-04 | Agent behavior constraint, not testable via unit test | 1. Run OKXWatcher with Executor Agent 2. Ask Executor directly to place order (should refuse) 3. Ask OKXWatcher to analyze, then execute (should execute via OKXWatcher command) |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
