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

### 订单工具清单
- **基础订单工具**：`okx-place-order`、`okx-cancel-order`、`okx-get-order`
- **止盈止损工具**：`okx-attach-sl-tp`（附加到订单）、`okx-place-order-with-sl-tp`（下单时带参）
- **批量订单工具**：`okx-batch-place-order`、`okx-batch-cancel-order`
- **历史订单查询**：`okx-get-order-history`
- **平仓工具**：`okx-close-position`（支持全部平仓和部分平仓）

### 参数设计
- **参数命名**：和 OKX API 保持一致（instID, side, posSide, ordType, size, price）
- **仓位模式**：默认 net 模式，支持显式指定 long/short
- **精度处理**：在工具文档（Info 描述）中说明精度规则，Agent 负责传入正确精度
- **市价单单位**：使用合约张数作为单位（避免不同币种精度差异）
- **订单类型支持**：market, limit, post_only, fok, ioc

### 速率限制与并发
- **速率限制**：5 req/s（所有订单工具，包括批量工具）
- **并发安全**：限流器保证并发安全 + 防重复下单机制
- **防重复下单**：基于唯一 ID 去重（相同请求 ID 拒绝）

### Agent 架构
- **Executor Agent 类型**：使用 ChatModelAgent 模式（和 RiskOfficer 一致）
- **自主性级别**：Level 1，仅执行 OKXWatcher 的明确指令，不主动分析
- **工具调用权限**：只有 Executor Agent 可以调用订单工具
- **风控检查**：Executor 不需要内置风控检查，由 OKXWatcher 负责协调 RiskOfficer

### 响应格式
- **返回格式**：Markdown 表格（和现有 positions/candlesticks 工具一致）
- **订单状态**：使用 OKX 原始状态（live, filled, cancelled, rejected 等）
- **返回字段**：完整的订单字段（订单 ID、状态、价格、数量、成交数量、未成交数量、成交金额等对 Agent 有用的字段）
- **错误处理**：返回 Markdown 错误提示（在 Markdown 开头用醒目方式显示错误信息）
- **批量部分失败**：返回成功订单列表和失败订单列表（含错误原因），让 Agent 自行处理

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
