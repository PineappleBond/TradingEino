---
status: testing
phase: 03-execution-automation
source: [03-01-SUMMARY.md, 03-02-SUMMARY.md, 03-03-SUMMARY.md, 03-04-SUMMARY.md]
started: 2026-03-24T18:00:00Z
updated: 2026-03-24T18:00:00Z
---

## Current Test

number: 1
name: 冷启动测试 - 服务启动与 Agent 初始化
expected: |
  1. 杀死任何运行中的服务
  2. 清除临时状态（临时数据库、缓存、锁文件）
  3. 从头启动应用：go run cmd/server/main.go
  4. 服务无错误启动，Agents 初始化成功
  5. 基本健康检查或日志输出显示服务正常运行
awaiting: user response

## Tests

### 1. 冷启动测试 - 服务启动与 Agent 初始化
expected: 杀死现有服务，清除临时状态，从头启动应用。服务启动无错误，Agents 初始化成功，日志显示 OKXWatcher、RiskOfficer、SentimentAnalyst、Executor 均正常加载。
result: [pass]

### 2. 下单工具 - okx-place-order 基本功能
expected: 调用 okx-place-order 工具下达限价单或市价单。返回订单 ID、状态、sCode、sMsg。订单成功提交时返回 ordId 和 state=live。
result: [pending]

### 3. 撤单工具 - okx-cancel-order 基本功能
expected: 调用 okx-cancel-order 取消待处理订单。返回取消确认，订单状态变为 cancelled。
result: [pending]

### 4. 订单查询工具 - okx-get-order 基本功能
expected: 调用 okx-get-order 查询订单状态。返回完整订单详情，包括 ordId、instId、side、state、fillSize、avgPx 等字段。
result: [pending]

### 5. 止盈止损工具 - okx-attach-sl-tp 附加 SL/TP
expected: 调用 okx-attach-sl-tp 为已有订单附加止损止盈。返回 algoId 表示条件单创建成功。至少需要 slTriggerPx 或 tpTriggerPx 之一。
result: [pending]

### 6. 带 SL/TP 下单工具 - okx-place-order-with-sl-tp
expected: 调用 okx-place-order-with-sl-tp 下单同时设置止损止盈。返回订单 ID 和 algoId。至少需要 slTriggerPx 或 tpTriggerPx 之一。
result: [pending]

### 7. Executor Agent Level 1 自主性 - 拒绝直接命令
expected: 直接发送交易命令给 Executor（如"买入 10 张 ETH-USDT-SWAP"）。Executor 应拒绝执行，声明需要 OKXWatcher 的明确指令。
result: [pending]

### 8. Executor Agent Level 1 自主性 - 执行 OKXWatcher 命令
expected: 通过 OKXWatcher 发送交易指令。OKXWatcher 分析后命令 Executor 执行，Executor 执行订单并返回结果。
result: [pending]

### 9. 批量下单工具 - okx-batch-place-order
expected: 调用 okx-batch-place-order 批量下达最多 20 个订单。返回成功订单列表和失败订单列表（如有）。超过 20 个订单返回错误。
result: [pending]

### 10. 批量撤单工具 - okx-batch-cancel-order
expected: 调用 okx-batch-cancel-order 批量取消最多 20 个订单。返回成功取消列表和失败取消列表（如有）。
result: [pending]

### 11. 订单历史工具 - okx-get-order-history
expected: 调用 okx-get-order-history 查询历史订单。支持 instID 筛选、时间范围筛选（startTime/endTime）。返回订单列表 Markdown 表格。
result: [pending]

### 12. 平仓工具 - okx-close-position 全额平仓
expected: 调用 okx-close-position percentage=100 平掉全部仓位。返回平仓确认，显示 instId、posSide、closed percentage=100%。
result: [pending]

### 13. 平仓工具 - okx-close-position 部分平仓
expected: 调用 okx-close-position percentage=50 平掉 50% 仓位。系统查询当前仓位，计算 50% 数量，下达反向市价单。返回平仓确认和订单详情。
result: [pending]

## Summary

total: 13
passed: 1
issues: 0
pending: 12
skipped: 0

## Gaps

[none yet]
