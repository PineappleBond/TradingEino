---
phase: 03-execution-automation
plan: 02
subsystem: execution
tags: [stop-loss, take-profit, OKX, algo-order, risk-management]
dependency_graph:
  requires: []
  provides:
    - okx-attach-sl-tp-tool
    - okx-place-order-with-sl-tp-tool
  affects:
    - internal/agent/tools/
tech_stack:
  added: []
  patterns:
    - TDD (Test-Driven Development)
    - Rate limiting (5 req/s)
    - sCode/sMsg validation
key_files:
  created:
    - internal/agent/tools/okx_attach_sl_tp.go
    - internal/agent/tools/okx_attach_sl_tp_test.go
    - internal/agent/tools/okx_place_order_with_sl_tp.go
    - internal/agent/tools/okx_place_order_with_sl_tp_test.go
  modified:
    - internal/agent/tools/okx_place_order.go
    - internal/agent/tools/okx_cancel_order.go
    - internal/agent/tools/okx_cancel_order_test.go
decisions:
  - Use OKX PlaceAlgoOrder API with conditional order type for SL/TP
  - Separate tools for attaching SL/TP vs placing order with SL/TP
  - Markdown table output format for consistency
metrics:
  duration_seconds: 600
  completed_at: "2026-03-24T00:00:00Z"
---

# Phase 03 Plan 02: Stop-Loss/Take-Profit Tools Summary

**One-liner:** Implemented two P0 stop-loss/take-profit tools using OKX native algo orders with rate limiting and sCode/sMsg validation.

---

## Tools Implemented

### 1. okx-attach-sl-tp-tool

Attaches stop-loss and/or take-profit to existing orders.

**File:** `internal/agent/tools/okx_attach_sl_tp.go`

**Parameters:**
- `instID` (required): 交易对
- `ordId` (required): 已有订单 ID
- `slTriggerPx` (optional): 止损触发价格
- `slOrderPx` (optional): 止损委托价格
- `tpTriggerPx` (optional): 止盈触发价格
- `tpOrderPx` (optional): 止盈委托价格

**Validation:**
- At least one of slTriggerPx or tpTriggerPx must be provided
- Uses PlaceAlgoOrder with conditional order type (EXEC-05)
- Validates sCode/sMsg in response (EXEC-06)

**Output:** Markdown table with algoId, sCode, sMsg

---

### 2. okx-place-order-with-sl-tp-tool

Places a new order with stop-loss and/or take-profit in a single call.

**File:** `internal/agent/tools/okx_place_order_with_sl_tp.go`

**Parameters:**
- `instID` (required): 交易对
- `side` (required): buy/sell
- `posSide` (optional): long/short/net (default: net)
- `ordType` (required): market/limit/post_only/fok/ioc
- `size` (required): 订单数量
- `price` (optional): 订单价格 (limit/post_only 必填)
- `slTriggerPx` (optional): 止损触发价格
- `slOrderPx` (optional): 止损委托价格
- `tpTriggerPx` (optional): 止盈触发价格
- `tpOrderPx` (optional): 止盈委托价格

**Validation:**
- At least one of slTriggerPx or tpTriggerPx must be provided
- Uses PlaceAlgoOrder with conditional order type (EXEC-05)
- Validates sCode/sMsg in response (EXEC-06)

**Output:** Markdown table with algoId, instID, side, sCode, sMsg plus SL/TP details

---

## Test Coverage

| Tool | Tests | Status |
|------|-------|--------|
| OkxAttachSlTpTool | 6 tests | All passing |
| OkxPlaceOrderWithSlTpTool | 6 tests | All passing |

**Test scenarios covered:**
- Attach SL/TP to existing order returns algoId
- Attach SL-only (no TP) returns algoId
- Attach TP-only (no SL) returns algoId
- Neither SL nor TP provided returns error
- OKX sCode != 0 returns error with details
- Rate limiter configuration verified

---

## Technical Implementation

### Rate Limiting
Both tools use rate limiting at 5 req/s (200ms per request) for Trade endpoint:
```go
limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1)
```

### sCode/sMsg Validation (EXEC-06)
Both tools validate OKX algo order response:
```go
if algoResult.SCode != 0 {
    return "", fmt.Errorf("algo order failed: sCode=%d, sMsg=%s", int64(algoResult.SCode), algoResult.SMsg)
}
```

### Markdown Table Output
Consistent output format across both tools:
```markdown
## SL/TP Order Attached

| algoId         | sCode | sMsg |
| :------------- | :---- | :--- |
| algo-123456    | 0     |      |
```

---

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed okx_place_order.go SCode formatting**
- **Found during:** Task 1 implementation
- **Issue:** fmt.Sprintf format %s has arg result.SCode of wrong type JSONInt64
- **Fix:** Cast to int64: fmt.Sprintf("%d", int64(result.SCode))
- **Files modified:** internal/agent/tools/okx_place_order.go
- **Commit:** 71b83a8

**2. [Rule 1 - Bug] Fixed okx_cancel_order.go SCode Float64 method**
- **Found during:** Task 2 implementation
- **Issue:** result.SCode.Float64 undefined (type JSONFloat64 has no field or method Float64)
- **Fix:** Use direct conversion: fmt.Sprintf("%.0f", float64(result.SCode))
- **Files modified:** internal/agent/tools/okx_cancel_order.go

**3. [Rule 1 - Bug] Fixed duplicate test function**
- **Found during:** Task 2 testing
- **Issue:** TestCancelOrderResponseParsing redeclared in okx_place_order_test.go and okx_cancel_order_test.go
- **Fix:** Removed duplicate from okx_place_order_test.go
- **Files modified:** internal/agent/tools/okx_place_order_test.go

**4. [Rule 1 - Bug] Fixed unused imports in test files**
- **Found during:** Task 2 testing
- **Issue:** okx_cancel_order_test.go imported encoding/json and fmt but not used
- **Fix:** Removed unused imports
- **Files modified:** internal/agent/tools/okx_cancel_order_test.go

---

## Requirements Satisfied

| Requirement | Status | Description |
|-------------|--------|-------------|
| EXEC-05 | Satisfied | Uses OKX native sl_tp algo order type via PlaceAlgoOrder |
| EXEC-06 | Satisfied | sCode/sMsg validation in both tools |

---

## Commits

```
5287ef6 feat(03-02): implement okx-place-order-with-sl-tp-tool
c4a6b29 feat(03-02): implement okx-attach-sl-tp-tool
```

---

## Self-Check: PASSED

All created files exist:
- internal/agent/tools/okx_attach_sl_tp.go
- internal/agent/tools/okx_attach_sl_tp_test.go
- internal/agent/tools/okx_place_order_with_sl_tp.go
- internal/agent/tools/okx_place_order_with_sl_tp_test.go

All commits exist and tests pass.
