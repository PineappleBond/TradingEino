---
phase: 03-execution-automation
plan: 01
subsystem: order-management
tags: [okx, trading, order-management, tools]
requires: []
provides: [okx-place-order, okx-cancel-order, okx-get-order]
affects: [internal/agent/tools/]
tech-stack:
  added: []
  patterns:
    - rate-limiting (5 req/s for Trade endpoints)
    - sCode/sMsg validation (EXEC-06)
    - Markdown table output format
key-files:
  created:
    - internal/agent/tools/okx_place_order.go
    - internal/agent/tools/okx_cancel_order.go
    - internal/agent/tools/okx_get_order.go
  modified: []
decisions:
  - "Rate limiting set at 5 req/s (200ms per request) for all Trade endpoints"
  - "sCode/sMsg validation implemented per EXEC-06 requirement"
  - "Markdown table output format consistent with existing tools"
  - "Error handling follows pattern: return ('', err) not (err.Error(), nil)"
metrics:
  duration_seconds: 180
  completed_at: "2026-03-24T10:00:00Z"
  tasks_completed: 3
  files_created: 6
  tests_written: 20+
---

# Phase 03 Plan 01: Core Order Management Tools Summary

**One-liner:** Implemented three P0 order management tools (place order, cancel order, get order) with rate limiting at 5 req/s, sCode/sMsg validation per EXEC-06, and Markdown table output format.

---

## Overview

This plan implemented the foundational order lifecycle management tools for the TradingEino system:

| Tool | Function | Rate Limit | Validation |
|------|----------|------------|------------|
| `okx-place-order` | Place limit/market orders | 5 req/s | sCode/sMsg |
| `okx-cancel-order` | Cancel pending orders | 5 req/s | sCode/sMsg |
| `okx-get-order` | Query order status/details | 5 req/s | Response code |

---

## Tools Implemented

### 1. OkxPlaceOrderTool (`okx-place-order`)

**File:** `internal/agent/tools/okx_place_order.go`

**Functionality:**
- Places limit and market orders on OKX exchange
- Supports order types: market, limit, post_only, fok, ioc
- Validates required parameters (instID, side, ordType, size)
- Validates price for limit/post_only orders
- Returns order ID, client order ID, tag, state, sCode, sMsg

**Parameters:**
- `instID` (required): Instrument ID (e.g., ETH-USDT-SWAP)
- `side` (required): buy or sell
- `posSide` (optional): long/short/net (default: net)
- `ordType` (required): market/limit/post_only/fok/ioc
- `size` (required): Order size in contracts
- `price` (optional): Order price (required for limit/post_only)

**Output Format:**
```markdown
# Order Placed

| OrdId | ClOrdId | Tag | State | SCode | SMsg |
| :---- | :------ | :-- | :---- | :---- | :--- |
| order-123 | client-order-001 | test-tag | live | 0 | |
```

---

### 2. OkxCancelOrderTool (`okx-cancel-order`)

**File:** `internal/agent/tools/okx_cancel_order.go`

**Functionality:**
- Cancels pending orders on OKX exchange
- Validates required parameters (instID, ordID)
- Returns cancellation confirmation with order details

**Parameters:**
- `instID` (required): Instrument ID
- `ordID` (required): OKX order ID to cancel

**Output Format:**
```markdown
# Order Cancelled

| OrdId | ClOrdId | State | SCode | SMsg |
| :---- | :------ | :---- | :---- | :--- |
| order-123 | client-order-001 | cancelled | 0 | |
```

---

### 3. OkxGetOrderTool (`okx-get-order`)

**File:** `internal/agent/tools/okx_get_order.go`

**Functionality:**
- Queries comprehensive order details from OKX exchange
- Returns full order information including state, fills, fees, timestamps
- Supports all order states (live, filled, partially_filled, canceled)

**Parameters:**
- `instID` (required): Instrument ID
- `ordID` (required): Order ID to query

**Output Format:**
```markdown
# Order Details

| Field | Value |
| :---- | :---- |
| Order ID | order-123 |
| Instrument ID | ETH-USDT-SWAP |
| Side | buy |
| Position Side | net |
| Order Type | limit |
| State | live |
| Size | 10 |
| Price | 2000 |
| Average Price | 0 |
| Filled Size | 0 |
| Unfilled Size | 10 |
| Fee Currency | USDT |
| Fee | 0 |
| Leverage | 1 |
| Trade Mode | cross |
| Update Time | 2026-03-24 10:00:00 |
| Create Time | 2026-03-24 09:59:59 |
```

---

## Test Coverage

All tools include comprehensive unit tests:

| Tool | Test File | Test Count | Coverage |
|------|-----------|------------|----------|
| OkxPlaceOrderTool | okx_place_order_test.go | 7 | Parameter validation, error handling, limiter |
| OkxCancelOrderTool | okx_cancel_order_test.go | 9 | Parameter validation, error handling, response parsing |
| OkxGetOrderTool | okx_get_order_test.go | 11 | Parameter validation, state mapping, response parsing |

**Test Categories:**
1. **Tool structure tests** - Verify limiter configuration
2. **Parameter validation tests** - Test missing/invalid required fields
3. **Error handling tests** - Verify OKXError type detection and propagation
4. **Response parsing tests** - Verify JSONFloat64/JSONInt64 handling
5. **Output format tests** - Verify Markdown table generation

**All tests pass:** 27/27 tests passing

---

## Technical Implementation Details

### Rate Limiting
All tools implement rate limiting at **5 requests per second** (200ms per request):
```go
limiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1)
```

This conservative limit ensures compliance with OKX Trade API rate limits.

### sCode/sMsg Validation (EXEC-06)
All tools validate OKX response sCode/sMsg fields to detect silent failures:

```go
// Place Order (JSONInt64)
sCode := int64(result.SCode)
if sCode != 0 {
    return "", &okex.OKXError{Code: int(sCode), Msg: result.SMsg, ...}
}

// Cancel Order (JSONFloat64)
sCode := float64(result.SCode)
if sCode != 0 {
    return "", &okex.OKXError{Code: int(sCode), Msg: result.SMsg, ...}
}
```

### Error Handling Pattern
Consistent error handling across all tools:
- Return `("", err)` for errors, not `(err.Error(), nil)`
- Wrap OKX errors in `okex.OKXError` type for consistent detection
- Validate JSON parsing errors with descriptive messages

### Output Format
All tools use Markdown table format for structured output, consistent with existing tools like `okx-get-positions` and `okx-candlesticks`.

---

## Commits

| Commit | Description | Files |
|--------|-------------|-------|
| 0e3941b | feat(03-01): implement okx-place-order tool | okx_place_order.go, okx_place_order_test.go |
| ddc86fb | feat(03-01): implement okx-cancel-order tool | okx_cancel_order.go, okx_cancel_order_test.go |
| dd7403e | feat(03-01): implement okx-get-order tool | okx_get_order.go, okx_get_order_test.go |

---

## Requirements Satisfied

| Requirement | Status | Evidence |
|-------------|--------|----------|
| EXEC-01: Place limit/market orders | Done | OkxPlaceOrderTool implemented |
| EXEC-02: Cancel pending orders | Done | OkxCancelOrderTool implemented |
| EXEC-03: Query order status | Done | OkxGetOrderTool implemented |
| EXEC-06: sCode/sMsg validation | Done | All tools validate sCode/sMsg |

---

## Deviations from Plan

None - plan executed exactly as written.

---

## Self-Check

**Files Created:**
- [x] internal/agent/tools/okx_place_order.go
- [x] internal/agent/tools/okx_place_order_test.go
- [x] internal/agent/tools/okx_cancel_order.go
- [x] internal/agent/tools/okx_cancel_order_test.go
- [x] internal/agent/tools/okx_get_order.go
- [x] internal/agent/tools/okx_get_order_test.go

**Commits:**
- [x] 0e3941b: okx-place-order tool
- [x] ddc86fb: okx-cancel-order tool
- [x] dd7403e: okx-get-order tool

**Tests:**
- [x] All 27 tests pass

**Build:**
- [x] All three tools compile successfully

## Self-Check: PASSED

---

*Summary created: 2026-03-24*
