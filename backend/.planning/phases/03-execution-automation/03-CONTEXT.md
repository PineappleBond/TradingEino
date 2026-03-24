# Phase 3: Execution Automation - Context

**Gathered:** 2026-03-24
**Status:** Ready for planning

<domain>
## Phase Boundary

自主交易执行，Level 1 自主性（仅限明确指令）。实现 OKX 交易工具链和 Executor Agent，使系统能够根据 OKXWatcher 的明确指令执行交易操作。

**成功标准：**
1. 用户可通过 `okx-place-order-tool` 下达限价单和市价单
2. 用户可通过 `okx-cancel-order-tool` 取消待处理订单
3. 用户可通过 `okx-get-order-tool` 查询订单状态
4. Executor Agent 仅在 OKXWatcher 明确指令下执行交易（Level 1 自主性）
5. 止损/止盈订单使用 OKX 原生 `sl_tp` 算法订单类型
6. 订单响应验证 OKX `sCode`/`sMsg` 字段（检测静默失败）

</domain>

<decisions>
## Implementation Decisions

### 订单工具清单（共 9 个工具）
| 工具名称 | 功能 | 优先级 |
|----------|------|--------|
| `okx-place-order` | 下单（限价/市价/POST_ONLY/FOK/IOC） | P0 |
| `okx-cancel-order` | 取消单个订单 | P0 |
| `okx-get-order` | 查询单个订单状态 | P0 |
| `okx-attach-sl-tp` | 为已有订单附加止盈止损 | P0 |
| `okx-place-order-with-sl-tp` | 下单时同时设置止盈止损 | P0 |
| `okx-batch-place-order` | 批量下单（最多同时下 20 单） | P1 |
| `okx-batch-cancel-order` | 批量撤单（最多同时撤 20 单） | P1 |
| `okx-get-order-history` | 查询历史订单（支持时间范围筛选） | P1 |
| `okx-close-position` | 平仓（支持全部平仓和部分平仓，按百分比） | P1 |

### 参数设计

**place-order 工具参数：**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| instID | string | 是 | 交易对，如 `ETH-USDT-SWAP` |
| side | string | 是 | 订单方向：`buy` 或 `sell` |
| posSide | string | 否 | 仓位模式：`long`/`short`/`net`，默认 `net` |
| ordType | string | 是 | 订单类型：`market`/`limit`/`post_only`/`fok`/`ioc` |
| size | string | 是 | 订单数量（合约张数） |
| price | string | 条件必填 | 订单价格，`limit` 和 `post_only` 订单必填，`market` 订单留空 |

**close-position 工具参数：**
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| instID | string | 是 | 交易对，如 `ETH-USDT-SWAP` |
| posSide | string | 否 | 仓位方向：`long`/`short`/`net`，默认查询全部 |
| percentage | number | 否 | 平仓百分比（0-100），默认 100（全部平仓） |

**精度规则（在工具 Info 描述中说明）：**
- `size`（数量）：合约张数，整数，如 `1`、`10`、`100`
- `price`（价格）：根据交易对确定小数位数，参考 OKX API `tick_sz` 字段
- 精度错误由 OKX API 直接拒绝，工具不自动修正

**止盈止损工具参数：**

**okx-attach-sl-tp**（附加到已有订单）：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| instID | string | 是 | 交易对 |
| ordId | string | 是 | 已有订单 ID |
| slTriggerPx | string | 条件 | 止损触发价格，和 tpTriggerPx 至少填一个 |
| slOrderPx | string | 否 | 止损委托价格，留空表示市价单 |
| tpTriggerPx | string | 条件 | 止盈触发价格，和 slTriggerPx 至少填一个 |
| tpOrderPx | string | 否 | 止盈委托价格，留空表示市价单 |

**okx-place-order-with-sl-tp**（下单时带止盈止损）：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| instID | string | 是 | 交易对 |
| side | string | 是 | 订单方向 |
| posSide | string | 否 | 仓位模式，默认 net |
| ordType | string | 是 | 主订单类型 |
| size | string | 是 | 主订单数量 |
| price | string | 条件 | 主订单价格（limit/post_only 必填） |
| slTriggerPx | string | 条件 | 止损触发价格，和 tpTriggerPx 至少填一个 |
| slOrderPx | string | 否 | 止损委托价格，留空表示市价单 |
| tpTriggerPx | string | 条件 | 止盈触发价格，和 slTriggerPx 至少填一个 |
| tpOrderPx | string | 否 | 止盈委托价格，留空表示市价单 |

### 速率限制与并发

**速率限制配置：**
| 工具类型 | 限流值 | 说明 |
|----------|--------|------|
| 所有订单工具 | 5 req/s | OKX 交易接口保守限制 |

**并发安全：**
- 限流器保证并发安全（`rate.Limiter`）
- 防重复下单机制：基于唯一请求 ID 去重（相同请求 ID 拒绝）

**批量操作限制：**
- OKX 批量下单/撤单 API 最多支持 20 个订单/批次
- 工具需要验证输入数量，超过 20 个时返回错误

### Agent 架构

**Executor Agent：**
- **类型**：ChatModelAgent（和 RiskOfficer、SentimentAnalyst 一致）
- **名称**：`ExecutorAgent` 或 `TradingExecutor`
- **自主性级别**：Level 1 — 仅执行 OKXWatcher 的明确指令，不主动分析市场或发起交易
- **工具调用权限**：只有 Executor Agent 可以调用订单工具（通过工具描述约束，非技术强制）
- **风控检查**：Executor 不需要内置风控检查，由 OKXWatcher 负责协调 RiskOfficer

**OKXWatcher 与 Executor 交互模式：**
```
OKXWatcher (DeepAgent 协调器)
    ↓
    | 1. 咨询 RiskOfficer（风控检查）
    ↓
    | 2. 下达明确交易指令给 Executor
    ↓
ExecutorAgent (ChatModelAgent)
    ↓
    | 3. 调用订单工具执行交易
    ↓
    | 4. 返回执行结果给 OKXWatcher
```

**Executor Agent 提示词关键约束：**
- 只负责执行，不负责分析市场或判断交易时机
- 必须等待 OKXWatcher 的明确指令才能执行交易
- 订单失败时，返回详细错误信息，由 OKXWatcher 决定下一步操作
- 不主动重试失败订单，除非 OKXWatcher 明确要求重试

### 响应格式

**place-order 返回字段（Markdown 表格）：**
```markdown
| 字段 | 说明 |
|------|------|
| ordId | OKX 订单 ID |
| clOrdId | 客户端订单 ID（如有） |
| tag | 订单标签（如有） |
| state | 订单状态 |
| sCode | 子错误码 |
| sMsg | 子错误信息 |
```

**get-order 返回字段（Markdown 表格）：**
```markdown
| 字段 | 说明 |
|------|------|
| ordId | 订单 ID |
| instId | 交易对 |
| side | 方向（buy/sell） |
| posSide | 仓位模式（long/short/net） |
| ordType | 订单类型 |
| size | 订单数量 |
| px | 委托价格 |
| avgPx | 成交价格 |
| fillSize | 已成交数量 |
| unfillSize | 未成交数量 |
| state | 订单状态 |
| cTime | 创建时间（毫秒时间戳） |
| uTime | 更新时间（毫秒时间戳） |
```

**订单状态（OKX 原始值）：**
- `live` — 待成交
- `partially_filled` — 部分成交
- `filled` — 完全成交
- `cancelled` — 已取消
- `rejected` — 已拒绝
- `expired` — 已过期（FOK/IOC 等特殊订单）

**错误处理格式：**
```markdown
**❌ 订单操作失败**

**错误代码：** {sCode}
**错误信息：** {sMsg}
**请求参数：** {instID, side, size, ...}
```

**批量操作返回格式：**
```markdown
## 成功订单
| ordId | instId | state | ... |
|-------|--------|-------|-----|
| ... | ... | ... | ... |

## 失败订单
| 请求索引 | sCode | sMsg |
|----------|-------|------|
| 0 | 51000 | 余额不足 |
```

### 日志与测试
- **日志级别**：详细日志（记录订单 ID、金额、价格、时间等完整信息）
- **测试策略**：使用 etc/config.yaml 中的模拟盘配置进行沙盒集成测试
- **异常场景**：工具层返回错误，由 Agent 自行判断是否重试或忽略

### 架构依赖
- **API 访问方式**：通过 ServiceContext.OKXClient 访问（和现有工具一致）
- **配置方式**：和现有 Agent/工具配置方式一致，不需要新增配置项
- **订单状态更新**：Agent 主动查询（Phase 3 不涉及轮询或 WebSocket 推送）
- **前端配合**：Phase 3 不涉及前端更新

### Executor Agent 提示词场景
- 只负责执行，不负责分析
- 需覆盖的执行相关场景：订单失败处理、下单时机判断、仓位管理原则
- 提示词中说明场景处理方式

</decisions>

<code_context>
## Existing Code Insights

### 可复用资产
- **pkg/okex/api/trading.go** - 已实现 PlaceOrder, CancelOrder, GetOrderDetails 基础方法
- **internal/agent/tools/okx_*.go** - 现有 Tool 实现模式（限流器、错误处理、响应验证）
- **internal/agent/agents.go** - Agent 初始化模式（sync.Once 单例、ChatModelAgent/DeepAgent 配置）

### 既定模式
- **Tool 结构**：`{ToolName} struct { svcCtx *svc.ServiceContext; limiter *rate.Limiter }`
- **错误处理**：返回 `("", err)` 而不是 `(err.Error(), nil)`
- **OKX 响应验证**：检查 `Code` 字段，非 0 时返回 `&okex.OKXError{}`
- **速率限制**：`rate.NewLimiter(rate.Every(time.Duration), burst)`
- **Agent 模式**：OKXWatcher 使用 DeepAgent，SubAgents 使用 ChatModelAgent

### 集成点
- **Executor Agent**：需要添加到 `internal/agent/agents.go` 的 AgentsModel 结构
- **订单工具**：需要添加到 `internal/agent/tools/` 目录
- **配置更新**：`etc/config.yaml` 需要配置 Executor Agent 和订单工具

</code_context>

<specifics>
## Specific Ideas

- "必须经过 OKXWatcher 的调度，必须是 OKXWatcher 的命令，才能执行" — Executor Agent 的核心约束
- "失败的订单和成功的订单，都要返回，让 Agent 自行处理" — 批量操作的部分失败处理
- "工具不需要处理异常场景，Agent 会自行判断，Agent 可能会重试，也可能忽略" — 工具层和 Agent 层的职责分离
- "etc/config.yaml 中目前就是模拟盘，随便测试" — 测试环境配置

</specifics>

<deferred>
## Deferred Ideas

- **订单修改功能**（amend-order）— 不需要，先撤单再下单即可
- **订单超时取消机制** — 不需要，由 Agent 自行调用 cancel-order
- **一键关闭全部仓位工具** — 需要 close-position 工具（支持部分平仓），属于 Phase 3
- **WebSocket 订单状态推送** — Phase 3 不涉及，Agent 主动查询
- **前端订单管理界面** — Phase 3 不涉及前端

</deferred>

---

*Phase: 03-execution-automation*
*Context gathered: 2026-03-24*
