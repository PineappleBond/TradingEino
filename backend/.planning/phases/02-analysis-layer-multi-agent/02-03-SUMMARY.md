---
phase: 02-analysis-layer-multi-agent
plan: 03
subsystem: position-management
tags: [agent, tool, okx, balance, position, risk-management]
dependency_graph:
  requires: ["02-01"]
  provides: ["PositionManager", "OkxAccountBalanceTool"]
  affects: ["OKXWatcher", "RiskOfficer"]
tech_stack:
  added: []
  patterns:
    - ChatModelAgent pattern
    - Tool rate limiting (5 req/s for Account endpoint)
    - go:embed for agent personality files
key_files:
  created:
    - internal/agent/tools/okx_account_balance.go
    - internal/agent/tools/okx_account_balance_test.go
    - internal/agent/position_manager/agent.go
    - internal/agent/position_manager/DESCRIPTION.md
    - internal/agent/position_manager/SOUL.md
    - internal/agent/position_manager/agent_test.go
  modified: []
decisions:
  - "PositionManager created as new directory (not renaming risk_officer) for backward compatibility"
  - "Account endpoint rate limit set to 5 req/s (conservative for trading APIs)"
  - "Margin ratio calculated as (equity - liability) / equity * 100%"
metrics:
  duration_seconds: 600
  completed_date: "2026-03-25"
---

# Phase 02 Plan 03: PositionManager Implementation Summary

## One-liner

创建 OkxAccountBalanceTool 和 PositionManager Agent，实现完整的账户余额监控和持仓管理能力，包含保证金率计算和风险提示。

## Task Completion

| Task | Name                                      | Commit    | Files                                                 |
|------|-------------------------------------------|-----------|-------------------------------------------------------|
| 1    | Create okx-account-balance-tool (DATA-03) | f26a174   | okx_account_balance.go, okx_account_balance_test.go   |
| 2    | Create PositionManager Agent (ANAL-04)    | a9492b8   | position_manager/agent.go, DESCRIPTION.md, SOUL.md    |

## Implementation Details

### Task 1: OkxAccountBalanceTool (DATA-03)

**Purpose:** 获取账户余额和保证金率，评估整体风险敞口

**Key Features:**
- Rate limiter: 5 req/s (200ms refill, burst 1) for Account endpoint
- Output format: Markdown table with currency, equity, available, frozen, liability
- Margin ratio calculation: (totalEquity - totalLiability) / totalEquity * 100%
- Risk warnings: < 20% severe warning, < 50% caution

**Files Created:**
- `internal/agent/tools/okx_account_balance.go` - Tool implementation
- `internal/agent/tools/okx_account_balance_test.go` - Test suite

**Tests Passed:**
- TestOkxAccountBalanceTool_Info - Verifies tool metadata
- TestOkxAccountBalanceTool_Params - Verifies no required parameters
- TestOkxAccountBalanceTool_RateLimiter - Verifies rate limiter initialization
- TestOkxAccountBalanceTool_OutputFormat - Verifies markdown table output
- TestOkxAccountBalanceTool_MarginRatioCalculation - Verifies risk calculation
- TestOkxAccountBalanceTool_EmptyBalance - Verifies empty balance handling

### Task 2: PositionManager Agent (ANAL-04)

**Purpose:** 持仓管理专家，整合仓位监控和账户余额能力

**Key Features:**
- ChatModelAgent pattern (not DeepAgent, following project convention)
- Dual tools: okx-get-positions-tool + okx-account-balance-tool
- Personality: conservative, risk-first, data-driven
- Embedded DESCRIPTION.md and SOUL.md via go:embed

**Files Created:**
- `internal/agent/position_manager/agent.go` - Agent implementation
- `internal/agent/position_manager/DESCRIPTION.md` - 15 lines capability description
- `internal/agent/position_manager/SOUL.md` - 24 lines personality definition
- `internal/agent/position_manager/agent_test.go` - Test suite

**Tests Passed:**
- TestPositionManagerAgent_Creation - Verifies agent instantiation
- TestPositionManagerAgent_AgentInterface - Verifies adk.Agent interface
- TestPositionManagerAgent_DescriptionAndSoul - Verifies content keywords
- TestPositionManagerAgent_MinimumLines - Verifies file length requirements

## Requirements Fulfilled

| Requirement | Status | Description                                    |
|-------------|--------|------------------------------------------------|
| ANAL-04     | Done   | PositionManager created as ChatModelAgent      |
| ANAL-06     | Done   | DESCRIPTION.md and SOUL.md with personality    |
| DATA-03     | Done   | okx-account-balance-tool with margin ratio     |

## Key Decisions

1. **New directory vs rename:** Created `position_manager/` as new directory instead of renaming `risk_officer/` for backward compatibility during transition.

2. **Rate limit:** Set Account endpoint rate limit to 5 req/s (200ms per request), following conservative approach for trading APIs.

3. **Margin ratio formula:** Calculated as `(totalEquity - totalLiability) / totalEquity * 100%` for risk assessment.

## Deviations from Plan

None - Plan executed exactly as written.

## Self-Check

- [x] okx_account_balance.go compiles
- [x] okx_account_balance_test.go tests pass (6/6)
- [x] position_manager/agent.go compiles
- [x] position_manager/agent_test.go tests pass (4/4)
- [x] Rate limiter set to 5 req/s (Account endpoint)
- [x] Output includes balance table and margin ratio warning
- [x] DESCRIPTION.md has 15 lines (> 10 minimum)
- [x] SOUL.md has 24 lines (> 10 minimum)
- [x] PositionManager uses both positions and balance tools

## Self-Check: PASSED

All files created, all tests passing, all requirements fulfilled.
