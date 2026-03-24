# Phase 1: Foundation & Safety - Context

**Gathered:** 2026-03-24
**Status:** Ready for planning

<domain>
## Phase Boundary

系统安全运行所需的基础设施修复：错误处理、速率限制、单例模式、上下文传播、优雅关闭。不包括新功能添加或架构重构。

</domain>

<decisions>
## Implementation Decisions

### 错误处理模式 (FOUND-01)
- 所有 Tool 的 `InvokableRun` 统一返回 `("", err)` 格式
- OKX API 错误使用 `OKXError` 结构体表示：
  - 文件位置：`pkg/okex/okx_error.go`
  - 字段：`Code int` (OKX 错误码), `Msg string` (错误消息), `Endpoint string` (请求端点，用于调试)
  - 方法：`Error() string`, `Unwrap() error` (支持 errors.As 解包)
  - 完整结构体定义：
    ```go
    type OKXError struct {
        Code     int
        Msg      string
        Endpoint string
    }
    func (e *OKXError) Error() string { return fmt.Sprintf("OKX %s error (code=%d): %s", e.Endpoint, e.Code, e.Msg) }
    func (e *OKXError) Unwrap() error  { return nil }
    ```
- Tool 层负责检查 `result.Code != 0` 并返回 `&OKXError{...}`
- 直接返回 `&OKXError` 实例，不使用 `fmt.Errorf` 包装
- 当前需要修复的 Tool：
  - `internal/agent/tools/okx_get_positions.go`: `OkxGetPositionsTool` — 无限流器，错误处理正确
  - `internal/agent/tools/okx_get_fundingrate.go`: `OkxGetFundingRateTool` — 无限流器，错误处理正确
  - `internal/agent/tools/okx_candlesticks.go`: `OkxCandlesticksTool` — 有限流器 (10 次/秒)，需调整为按端点分类限流

### 速率限制策略 (FOUND-02)
- 每个 Tool 在 `NewXxxTool` 函数内初始化自己的 `limiter *rate.Limiter`
- 严格限流配置（按端点类型分类硬编码）：
  - Trade/Account 端点：5 次/秒 (burst=1) — `rate.NewLimiter(rate.Every(200*time.Millisecond), 1)`
  - Market/Public 端点：10 次/秒 (burst=2) — `rate.NewLimiter(rate.Every(100*time.Millisecond), 2)`
  - Funding 端点：1 次/秒 (burst=1) — `rate.NewLimiter(rate.Every(time.Second), 1)`
- API 调用前必须调用 `limiter.Wait(ctx)`
- 当前需要添加限流器的 Tool：
  - `internal/agent/tools/okx_get_positions.go`: `OkxGetPositionsTool` — Account 端点，5 次/秒
  - `internal/agent/tools/okx_get_fundingrate.go`: `OkxGetFundingRateTool` — Public 端点，10 次/秒
  - `internal/agent/tools/okx_candlesticks.go`: `OkxCandlesticksTool` — Market 端点，当前 10 次/秒，需确认是否符合分类

### 单例模式重构 (FOUND-03)
- 保持全局变量 `_agents` 模式
- 添加文档说明：必须在 main 中先调用 `InitAgents()` 后才能使用 `Agents()`
- `Agents()` 返回 nil 的问题通过调用顺序规避

### 上下文传播 (FOUND-04)
- `InitAgents(ctx context.Context, svcCtx *svc.ServiceContext)` 签名修改，接收父上下文参数
- 所有子 Agent（RiskOfficer、SentimentAnalyst、OKXWatcher）初始化使用该 ctx
- 修复 `context.Background()` 的直接使用，改为传播父上下文
- Tool 执行时传播相同的 ctx
- 需要修改的文件：
  - `internal/agent/agents.go`: `InitAgents` 函数签名添加 `ctx context.Context` 参数
  - `internal/agent/agents.go`: `InitAgents` 内部使用传入的 ctx 而不是 `context.Background()`
  - `internal/agent/risk_officer/agent.go`: `NewRiskOfficerAgent` 接收 ctx 参数
  - `internal/agent/sentiment_analyst/agent.go`: `NewSentimentAnalystAgent` 接收 ctx 参数
  - `internal/agent/okx_watcher/agent.go`: `NewOkxWatcherAgent` 接收 ctx 参数

### 优雅关闭 (FOUND-05)
- 使用 `os/signal` 包监听 SIGINT/SIGTERM 信号
- 显式顺序关闭流程：Server → Scheduler → Agents → DB → Logger
- main.go 需要实现的关闭逻辑：
  ```go
  // 设置信号处理
  sigChan := make(chan os.Signal, 1)
  signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

  // 等待退出信号
  <-sigChan
  logger.Info(ctx, "shutdown signal received")

  // 1. 停止 HTTP 服务器
  if err := server.Shutdown(ctx); err != nil {
      logger.Error(ctx, "failed to shutdown server", err)
  }

  // 2. 停止调度器
  if err := scheduler.Stop(); err != nil {
      logger.Error(ctx, "failed to stop scheduler", err)
  }

  // 3. 关闭 Agents
  if err := agents.Close(); err != nil {
      logger.Error(ctx, "failed to close agents", err)
  }

  // 4. 关闭数据库连接
  db, _ := svcCtx.DB.DB()
  if err := db.Close(); err != nil {
      logger.Error(ctx, "failed to close database", err)
  }

  // 5. 关闭日志记录器
  if err := logger.Close(); err != nil {
      logger.Error(ctx, "failed to close logger", err)
  }

  logger.Info(ctx, "graceful shutdown completed")
  ```
- 需要实现 `Agents.Close()` 方法：调用 `cancel()` 取消函数，等待 goroutine 完成
- 需要实现 `Logger.Close()` 方法：Flush 缓冲区并关闭文件句柄
- `DB.Close()` 通过 `gorm.DB.DB()` 获取底层 `*sql.DB` 后调用

### 代码清理
- 移除 dead code（`if false` 调试块）
  - 位置：`internal/service/scheduler/handlers/okx_watcher_handler.go:190-193`
- 替换 `fmt.Printf`/`fmt.Fprintf` 为结构化 logger 调用
  - 位置：`internal/svc/database.go:20-40`
  - 位置：`internal/service/scheduler/handlers/okx_watcher_handler.go:192`
- 清理无用的 TODO 注释
  - 位置：`pkg/chromedp-v0.15.0/` 中的 TODO 注释（如果影响理解可保留）

### Claude's Discretion
- OKXError 的具体实现风格（保持与项目现有代码一致）
- 限流器的具体 burst 值可根据实际 API 限制微调
- 优雅关闭的日志输出格式

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `OkxCandlesticksTool` (`internal/agent/tools/okx_candlesticks.go`): 已有 rate.Limiter 实现（10 次/秒），可作为其他 Tool 的参考
  - 参考 `limiter *rate.Limiter` 字段定义
  - 参考 `NewOkxCandlesticksTool` 中限流器初始化
  - 参考 `GetCandlesticks` 中 `limiter.Wait(ctx)` 调用
- `RiskOfficerAgent` (`internal/agent/risk_officer/agent.go`): 已使用 ChatModelAgent 模式，可作为 Agent 实现的参考
- `Scheduler` (`internal/service/scheduler/scheduler.go`): 已有 `Stop()` 方法和 `RetryableError` 接口定义
  - 参考 `Stop()` 方法的实现模式
  - `RetryableError` 接口可作为参考（但 Phase 1 不要求 OKXError 实现该接口）

### Established Patterns
- Tool 结构：`type XxxTool struct { svcCtx *svc.ServiceContext; limiter *rate.Limiter }`
- Tool 初始化：`func NewXxxTool(svcCtx *svc.ServiceContext) *XxxTool`
  - 在初始化函数中创建限流器：`limiter: rate.NewLimiter(rate.Every(interval), burst)`
- OKX API 调用：`svcCtx.OKXClient.Rest.Xxx.GetYyy(...)`
- 错误检查模式：
  ```go
  // 1. 等待限流
  if err := c.limiter.Wait(ctx); err != nil {
      return "", fmt.Errorf("rate limiter wait failed: %w", err)
  }

  // 2. 调用 API
  result, err := c.svcCtx.OKXClient.Rest.Xxx.GetYyy(...)
  if err != nil {
      return "", err  // 返回原始错误
  }

  // 3. 检查 OKX 响应码
  if result.Code != 0 {
      return "", &OKXError{
          Code:     result.Code,
          Msg:      result.Msg,
          Endpoint: "GetYyy",
      }
  }

  // 4. 返回结果
  return json.Marshal(result.Data)
  ```
- Agent 关闭模式：
  ```go
  func (a *AgentsModel) Close() error {
      if a.cancel != nil {
          a.cancel()
      }
      return nil
  }
  ```

### Integration Points
- `main.go`: 需要添加信号处理和优雅关闭逻辑
  - 导入 `os/signal` 和 `syscall` 包
  - 在 `serve.Start()` 后添加信号监听
  - 按顺序关闭：Server → Scheduler → Agents → DB → Logger
- `internal/agent/agents.go`: 需要修改 `InitAgents` 签名添加 ctx 参数
  - `func InitAgents(ctx context.Context, svcCtx *svc.ServiceContext) error`
  - 内部调用 `risk_officer.NewRiskOfficerAgent(ctx, svcCtx)` 传递 ctx
  - 添加 `func Close() error` 方法调用 `cancel()`
- `internal/agent/tools/*.go`: 所有 Tool 需要添加限流器和错误处理
  - `okx_get_positions.go`: 添加 `limiter *rate.Limiter` (5 次/秒)
  - `okx_get_fundingrate.go`: 添加 `limiter *rate.Limiter` (10 次/秒)
  - `okx_candlesticks.go`: 调整限流器为 10 次/秒 (Market 端点)
- `pkg/okex/`: 需要添加 `okx_error.go` 文件
  - 定义 `OKXError` 结构体和 `Error()`, `Unwrap()` 方法
- `internal/svc/servicecontext.go`: 可选添加 `Close()` 方法统一管理资源关闭
- `internal/svc/database.go`: 修复 `fmt.Fprintf` 为 logger
- `internal/service/scheduler/handlers/okx_watcher_handler.go`: 移除 dead code 和 printf

</code_context>

<specifics>
## Specific Ideas

- "OKX API 返回的错误码太多了，取其中的 msg 就好" — OKXError.Msg 字段存储错误消息
- "文档太多不好查，建议可硬编码配置，统一管理即可" — 限流值硬编码在 Tool 初始化代码中
- "程序启动时必须初始化" — `InitAgents` 必须在 main 中调用
- 清理 dead code 和 printf 改为 logger

</specifics>

<success_criteria>
## Phase 1 Success Criteria

对应 REQUIREMENTS.md 中的 FOUND-01 到 FOUND-05：

1. **FOUND-01** — 所有 API 工具返回错误 properly (`"", err`) 而不是 (`err.Error(), nil`)
   - [ ] `okx_get_positions.go`: `InvokableRun` 返回 `("", err)`
   - [ ] `okx_get_fundingrate.go`: `InvokableRun` 返回 `("", err)`
   - [ ] `okx_candlesticks.go`: `InvokableRun` 返回 `("", err)`
   - [ ] 所有 OKX API 错误返回 `&OKXError{...}` 实例

2. **FOUND-02** — 所有 API 工具有 `rate.Limiter` (5 次/秒用于交易端点，10 次/秒用于行情端点)
   - [ ] `OkxGetPositionsTool`: limiter = 5 次/秒 (Account 端点)
   - [ ] `OkxGetFundingRateTool`: limiter = 10 次/秒 (Public 端点)
   - [ ] `OkxCandlesticksTool`: limiter = 10 次/秒 (Market 端点)

3. **FOUND-03** — Agent 使用 `sync.Once` 单例模式而不是全局变量
   - [ ] `internal/agent/agents.go`: 添加文档说明调用顺序
   - [ ] `Agents()` 返回 nil 问题通过文档规避

4. **FOUND-04** — 上下文传播 throughout agent initialization and tool execution
   - [ ] `InitAgents(ctx, svcCtx)` 接收 ctx 参数
   - [ ] 所有子 Agent 初始化使用传入的 ctx
   - [ ] 移除 `context.Background()` 的直接使用

5. **FOUND-05** — 应用程序优雅关闭 on exit
   - [ ] `main.go` 添加 SIGINT/SIGTERM 信号处理
   - [ ] 实现关闭顺序：Server → Scheduler → Agents → DB → Logger
   - [ ] `Agents.Close()` 方法实现
   - [ ] `Logger.Close()` 方法实现
   - [ ] `DB.Close()` 调用

6. **代码清理**
   - [ ] 移除 `if false` dead code
   - [ ] 替换 `fmt.Printf`/`fmt.Fprintf` 为 logger
</success_criteria>

<deferred>
## Deferred Ideas

- 自定义错误分类（业务错误/限流错误/网络错误）— 未来扩展，当前仅需 OKXError
- 限流器配置化（从 config.yaml 读取）— 当前硬编码，未来可配置
- 依赖注入模式重构 — 保持全局变量模式，不重构
- ServiceContext 统一管理限流器 — 各 Tool 自行管理

</deferred>

---

*Phase: 01-foundation-safety*
*Context gathered: 2026-03-24*
