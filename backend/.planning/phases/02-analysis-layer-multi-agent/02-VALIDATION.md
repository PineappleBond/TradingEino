---
phase: 2
slug: analysis-layer-multi-agent
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (`testing` package) |
| **Config file** | None (standard Go testing) |
| **Quick run command** | `go test ./internal/agent/<specific_agent>/... -v` |
| **Full suite command** | `go test ./internal/agent/... -count=1` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/agent/<specific_agent>/... -v`
- **After every plan wave:** Run `go test ./internal/agent/... -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 01 | 1 | ANAL-02 | unit | `go test ./internal/agent/techno_agent/... -v` | ❌ W0 | ⬜ pending |
| 02-01-02 | 01 | 1 | ANAL-02 | unit | `go test ./internal/agent/techno_agent/agent_test.go -run TestTechnoAgent_Analysis -v` | ❌ W0 | ⬜ pending |
| 02-02-01 | 02 | 2 | ANAL-03 | unit | `go test ./internal/agent/flow_analyzer/... -v` | ❌ W0 | ⬜ pending |
| 02-02-02 | 02 | 2 | ANAL-03 | unit | `go test ./internal/agent/flow_analyzer/agent_test.go -run TestFlowAnalyzer_Orderbook -v` | ❌ W0 | ⬜ pending |
| 02-03-01 | 03 | 2 | ANAL-04 | unit | `go test ./internal/agent/position_manager/... -v` | ❌ W0 (risk_officer exists) | ⬜ pending |
| 02-04-01 | 04 | 3 | ANAL-05 | integration | `go test ./internal/agent/okx_watcher/orchestration_test.go -v` | ❌ W0 | ⬜ pending |
| 02-05-01 | 05 | 3 | ANAL-06 | lint/unit | `go test ./internal/agent/... -run TestAgentFiles -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/agent/techno_agent/agent_test.go` — stubs for ANAL-02
- [ ] `internal/agent/flow_analyzer/agent_test.go` — stubs for ANAL-03
- [ ] `internal/agent/position_manager/agent_test.go` — stubs for ANAL-04 (or rename risk_officer tests)
- [ ] `internal/agent/okx_watcher/orchestration_test.go` — stubs for ANAL-05
- [ ] `internal/agent/agent_files_test.go` — DESCRIPTION.md + SOUL.md presence (ANAL-06)
- [ ] Tool tests for any new tools created during Phase 2

*Note: Existing test infrastructure from Phase 1 and Phase 3 should be reused.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| SubAgent personality quality | ANAL-06 | Subjective assessment | Review DESCRIPTION.md and SOUL.md files for clarity and completeness |
| OKXWatcher coordination effectiveness | ANAL-05 | Multi-agent behavior hard to unit test | Run end-to-end agent query and observe SubAgent delegation |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
