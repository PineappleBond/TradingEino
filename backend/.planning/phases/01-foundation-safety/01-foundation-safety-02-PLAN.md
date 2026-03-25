---
phase: 01-foundation-safety
plan: 02
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/agent/agents.go
  - internal/agent/risk_officer/agent.go
  - internal/agent/sentiment_analyst/agent.go
  - internal/agent/okx_watcher/agent.go
autonomous: true
requirements:
  - FOUND-03
  - FOUND-04
must_haves:
  truths:
    - "InitAgents 使用 sync.Once 保护初始化"
    - "InitAgents 接收 ctx 参数并传递给子 Agent"
    - "子 Agent 初始化函数使用传入的 ctx，不使用 context.Background()"
    - "AgentsModel.Close() 方法调用 cancel()"
  artifacts:
    - path: "internal/agent/agents.go"
      provides: "sync.Once + InitAgents(ctx, svcCtx)"
      contains: "agentsOnce.Do"
    - path: "internal/agent/risk_officer/agent.go"
      provides: "NewRiskOfficerAgent(ctx, ...)"
      exports: ["NewRiskOfficerAgent"]
    - path: "internal/agent/sentiment_analyst/agent.go"
      provides: "NewSentimentAnalystAgent(ctx, ...)"
      exports: ["NewSentimentAnalystAgent"]
    - path: "internal/agent/okx_watcher/agent.go"
      provides: "NewOkxWatcherAgent(ctx, ...)"
      exports: ["NewOkxWatcherAgent"]
  key_links:
    - from: "internal/agent/agents.go"
      to: "internal/agent/risk_officer/agent.go"
      via: "ctx 传递"
      pattern: "NewRiskOfficerAgent\\(ctx"
    - from: "cmd/server/main.go"
      to: "internal/agent/agents.go"
      via: "InitAgents 调用"
      pattern: "agent\\.InitAgents\\(ctx"
---

<objective>
Agent 使用 sync.Once 单例模式，上下文正确传播

Purpose: FOUND-03 要求 Agent 使用 sync.Once 而非裸全局变量，FOUND-04 要求上下文从 main 传播到所有子 Agent 和 Tool。本计划重构 InitAgents 接收 ctx 参数，并确保所有子 Agent 使用该 ctx。

Output:
- internal/agent/agents.go 更新（sync.Once + ctx 参数）
- 子 Agent 初始化函数签名更新（接收 ctx）
</objective>

<execution_context>
@/Users/leichujun/go/src/github.com/PineappleBond/TradingEino/backend/.claude/get-shit-done/workflows/execute-plan.md
@/Users/leichujun/go/src/github.com/PineappleBond/TradingEino/backend/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/REQUIREMENTS.md
@.planning/ROADMAP.md
@.planning/phases/01-foundation-safety/01-CONTEXT.md
@.planning/phases/01-foundation-safety/01-RESEARCH.md

# 现有代码参考
@internal/agent/agents.go
@internal/agent/risk_officer/agent.go
@internal/agent/sentiment_analyst/agent.go
@internal/agent/okx_watcher/agent.go
</context>

<tasks>

<task type="auto" tdd="true">
<name>Task 1: 重构 InitAgents 使用 sync.Once 和 ctx 参数</name>
<files>internal/agent/agents.go, internal/agent/agents_test.go</files>
<behavior>
- Test 1: InitAgents 签名变为 InitAgents(ctx context.Context, svcCtx *svc.ServiceContext) error
- Test 2: 使用 agentsOnce sync.Once 保护初始化
- Test 3: 传入的 ctx 传递给所有子 Agent 初始化函数
- Test 4: 多次调用 InitAgents 只执行一次初始化
- Test 5: AgentsModel 添加 Close() 方法调用 cancel()
</behavior>
<action>
重构 internal/agent/agents.go：

1. 添加 sync.Once 变量：
```go
var (
    agentsOnce sync.Once
    _agents    *AgentsModel
)
```

2. 修改 InitAgents 签名：
```go
func InitAgents(ctx context.Context, svcCtx *svc.ServiceContext) error {
    var initErr error
    agentsOnce.Do(func() {
        // 使用传入的 ctx，不是 context.Background()
        riskOfficerAgent, err := risk_officer.NewRiskOfficerAgent(ctx, svcCtx)
        if err != nil {
            initErr = err
            return
        }
        // ... 其他子 Agent
        _agents = &AgentsModel{...}
    })
    return initErr
}
```

3. 添加 Close 方法：
```go
func (a *AgentsModel) Close() error {
    if a.cancel != nil {
        a.cancel()
    }
    return nil
}
```

同时创建测试文件验证 sync.Once 行为和 ctx 传播。
</action>
<verify>
<automated>go test ./internal/agent/... -v -run TestInitAgents</automated>
</verify>
<done>InitAgents 使用 sync.Once，ctx 正确传播，Close() 方法存在</done>
</task>

<task type="auto">
<name>Task 2: 更新子 Agent 初始化函数接收 ctx 参数</name>
<files>internal/agent/risk_officer/agent.go, internal/agent/sentiment_analyst/agent.go, internal/agent/okx_watcher/agent.go</files>
<action>
更新三个子 Agent 的初始化函数签名，接收 ctx 参数：

1. risk_officer/agent.go:
```go
func NewRiskOfficerAgent(ctx context.Context, svcCtx *svc.ServiceContext) (*RiskOfficerAgent, error)
```

2. sentiment_analyst/agent.go:
```go
func NewSentimentAnalystAgent(ctx context.Context, svcCtx *svc.ServiceContext) (*SentimentAnalystAgent, error)
```

3. okx_watcher/agent.go:
```go
func NewOkxWatcherAgent(ctx context.Context, svcCtx *svc.ServiceContext, subAgents ...adk.Agent) (*OkxWatcherAgent, error)
```

每个函数内部使用传入的 ctx 创建 ChatModelAgent/DeepAgent，不使用 context.Background()。
</action>
<verify>
<automated>go build ./internal/agent/...</automated>
</verify>
<done>所有子 Agent 初始化函数接收 ctx 参数，编译通过</done>
</task>

<task type="auto">
<name>Task 3: 清理 dead code 和 fmt.Printf</name>
<files>internal/service/scheduler/handlers/okx_watcher_handler.go, internal/svc/database.go</files>
<action>
根据 CONTEXT.md 的代码清理决策：

1. internal/service/scheduler/handlers/okx_watcher_handler.go:190-193
   删除 `if false { ... }` dead code 块

2. internal/svc/database.go:20-40
   替换 `fmt.Fprintf(os.Stderr, ...)` 为 logger.Error：
```go
// 原代码：
fmt.Fprintf(os.Stderr, "db type %s not supported\n", cfg.DB.Type)
os.Exit(1)

// 替换为（在 mustInitDB 内，先获取 logger）：
log.Error(ctx, "unsupported database type", nil, "type", cfg.DB.Type)
os.Exit(1)
```

注意：database.go 中 mustInitDB 目前没有 ctx 和 logger 参数，需要添加。
</action>
<verify>
<automated>go build ./...</automated>
</verify>
<done>dead code 已删除，fmt.Printf 替换为 logger</done>
</task>

</tasks>

<verification>
- [ ] go test ./internal/agent/... -v 通过
- [ ] go build ./... 编译成功
- [ ] 无 context.Background() 在 InitAgents 内部
- [ ] sync.Once 保护 _agents 初始化
</verification>

<success_criteria>
- InitAgents(ctx, svcCtx) 接收 ctx 参数
- 使用 sync.Once 保护初始化
- AgentsModel.Close() 方法存在
- 所有子 Agent 初始化函数接收 ctx
- dead code 和 fmt.Printf 已清理
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation-safety/01-foundation-safety-02-SUMMARY.md`
</output>
