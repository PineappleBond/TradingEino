# TradingEino - Claude Code 配置

## 项目概述

TradingEino 是一个基于 Cloudwego Eino 框架的 AI 多 Agent 加密货币交易系统，监控 OKX 市场、分析技术指标和情绪，并自主执行交易。

**核心功能：**
- 多 Agent 协作（OKXWatcher 协调器 + 专业 SubAgents）
- OKX API 集成（K 线数据、持仓查询、交易执行）
- 技术指标分析（MACD、RSI、布林带等 20+ 指标）
- RAG 决策记忆（Redis Stack + m3e-base）
- 独立风控层（实时监控、熔断机制）

---

## 沟通规范

- **使用中文沟通** - 和用户沟通的内容使用中文
- **代码注释用英文** - 函数注释、行内注释使用英文
- **提交信息用英文** - git commit message 使用英文

---

## OKX API 使用规范

**必须使用 `.claude/skills/okex-skill` 技能** 处理所有 OKX API 相关任务：

### 何时调用 okex-skill

- 调用 OKX REST API（账户管理、交易执行、行情数据、资金管理）
- 连接 WebSocket API（实时订阅、推送处理）
- 实现交易相关的 Tool（下单、撤单、持仓查询、风控检查）
- 调试 API 调用问题（签名、速率限制、错误处理）

### API 模块列表

| 模块 | 功能 | 文档 |
|------|------|------|
| **Account** | 账户余额、持仓、杠杆、账单 | [references/account.md](.claude/skills/okex-skill/references/account.md) |
| **Trade** | 下单、撤单、订单查询、算法交易 | [references/trade.md](.claude/skills/okex-skill/references/trade.md) |
| **Market** | 行情 Ticker、深度、K 线、成交 | [references/market.md](.claude/skills/okex-skill/references/market.md) |
| **Funding** | 充值、提现、划转、账单 | [references/funding.md](.claude/skills/okex-skill/references/funding.md) |
| **PublicData** | 合约信息、资金费率、持仓总量 | [references/public_data.md](.claude/skills/okex-skill/references/public_data.md) |
| **SubAccount** | 子账户管理、APIKey 管理 | [references/sub_account.md](.claude/skills/okex-skill/references/sub_account.md) |
| **TradeData** | 交易大数据、多空比、成交量 | [references/trade_data.md](.claude/skills/okex-skill/references/trade_data.md) |
| **WebSocket** | 实时推送、订阅管理 | [references/websocket.md](.claude/skills/okex-skill/references/websocket.md) |

### OKXClient 使用方式

项目中已内置 OKX 客户端，通过 `ServiceContext` 访问：

```go
import (
    "github.com/PineappleBond/TradingEino/backend/internal/svc"
    "github.com/PineappleBond/TradingEino/backend/pkg/okex"
    "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"
    "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/market"
    "golang.org/x/time/rate"
)

// ServiceContext 中包含已初始化的 OKXClient
type ServiceContext struct {
    // ...
    OKXClient *api.Client  // 已通过 mustInitOKXClient 初始化
}

// 调用示例 - 获取 K 线数据
candlesticks, err := svcCtx.OKXClient.Rest.Market.GetCandlesticks(market.GetCandlesticks{
    InstID: "BTC-USDT-SWAP",
    Bar:    string(okex.Bar1H),
    Limit:  "100",
})

// 调用示例 - 获取持仓
positions, err := svcCtx.OKXClient.Rest.Account.GetPositions(account.GetPositions{
    InstID: []string{"ETH-USDT-SWAP"},
})

// Tool 实现时需要添加速率限制
type OkxTool struct {
    svcCtx  *svc.ServiceContext
    limiter *rate.Limiter  // 必须有限流器
}

func (c *OkxTool) InvokableRun(ctx context.Context, args string) (string, error) {
    // 1. 等待限流
    if err := c.limiter.Wait(ctx); err != nil {
        return "", fmt.Errorf("rate limiter wait failed: %w", err)
    }

    // 2. 调用 API
    result, err := c.svcCtx.OKXClient.Rest.Market.GetTicker(...)
    if err != nil {
        return "", err  // 返回错误，不是错误字符串
    }

    // 3. 验证 OKX 响应码
    if result.Code != "0" {
        return "", fmt.Errorf("OKX API error: %s", result.Msg)
    }

    return json.Marshal(result.Data)
}
```

### 现有 Tool 参考

项目中已有的 Tool 实现可作为参考：
- `internal/agent/tools/okx_candlesticks.go` - K 线数据获取（含速率限制）
- `internal/agent/tools/okx_get_positions.go` - 持仓查询
- `internal/agent/tools/okx_get_fundingrate.go` - 资金费率查询

---

## Eino 框架使用规范

**必须使用 `.claude/skills/eino-skill` 技能** 处理所有 Eino 相关任务：

### 何时调用 eino-skill

- 构建 LLM 应用、Agent、Workflow
- 使用 Eino 组件（ChatModel、Tool、Retriever、Graph、Workflow）
- 实现 RAG 流程、Interrupt/Resume、多 Agent 协作
- 调试 Eino 应用问题

### 核心组件模式

```go
// ChatModelAgent - 基础对话 Agent
agent := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Name:        "RiskOfficer",
    Model:       chatModel,
    Instruction: "你是风控专家...",
    ToolsConfig: adk.ToolsConfig{
        ToolsNodeConfig: compose.ToolsNodeConfig{
            Tools: []tool.BaseTool{okxGetPositionsTool},
        },
    },
})

// DeepAgent - 多 Agent 协调器（仅 OKXWatcher 使用）
deepAgent, _ := deep.New(ctx, &deep.Config{
    Name:        "OKXWatcher",
    Description: "OKX 盯盘代理",
    ChatModel:   svcCtx.ChatModel,
    SubAgents:   []adk.Agent{riskOfficer, sentimentAnalyst},
    MaxIteration: 10,
})

// Runner 执行
runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent})
iter := runner.Query(ctx, "分析 ETH-USDT-SWAP")
```

### Graph 编排模式

```go
// 带状态的 Graph
type AppState struct {
    History []*schema.Message
    Count   int
}

g := compose.NewGraph[*UserInput, *schema.Message](
    compose.WithGenLocalState(func(ctx context.Context) *AppState {
        return &AppState{}
    }),
)

// 条件分支
_ = g.AddBranch("ChatModel", compose.NewGraphBranch(
    func(ctx context.Context, in *schema.Message) (string, error) {
        if len(in.ToolCalls) > 0 {
            return "ToolsNode", nil
        }
        return compose.END, nil
    },
    map[string]bool{"ToolsNode": true, compose.END: true},
))
```

### Interrupt/Resume（人机协作）

```go
// 编译时配置中断点
runnable, err := g.Compile(ctx,
    compose.WithCheckPointStore(store),
    compose.WithInterruptBeforeNodes([]string{"ToolsNode"}),
)

// 恢复执行
if compose.IsInterruptError(err) {
    info := compose.GetInterruptInfo(ctx)
    ctx = compose.Resume(ctx, info.InterruptID)
    result, _ := runnable.Invoke(ctx, input,
        compose.WithCheckPointID(info.CheckPointID),
    )
}
```

---

## 编码规范

### 错误处理

```go
// ✅ 正确：返回错误
func (c *OkxGetFundingRateTool) InvokableRun(...) (string, error) {
    if err != nil {
        return "", err  // 返回空字符串和错误
    }
    if getFundingRate.Code != 0 {
        return "", fmt.Errorf("OKX API error: %s", getFundingRate.Msg)
    }
    return result, nil
}

// ❌ 错误：将错误当成功返回
func (c *OkxGetFundingRateTool) InvokableRun(...) (string, error) {
    if err != nil {
        return err.Error(), nil  // 禁止这样做
    }
}
```

### 速率限制

```go
import "golang.org/x/time/rate"

type OkxGetPositionsTool struct {
    svcCtx  *svc.ServiceContext
    limiter *rate.Limiter  // 必须有限流器
}

func NewOkxGetPositionsTool(svcCtx *svc.ServiceContext) *OkxGetPositionsTool {
    return &OkxGetPositionsTool{
        svcCtx:  svcCtx,
        limiter: rate.NewLimiter(rate.Every(time.Second/5), 1), // 5 req/s
    }
}

func (c *OkxGetPositionsTool) GetPositions(...) {
    c.limiter.Wait(ctx)  // API 调用前等待限流
    // API 调用
}
```

### 单例模式（替代全局变量）

```go
// ✅ 正确：sync.Once 单例
var (
    agentsOnce sync.Once
    agents     *AgentsModel
)

func GetAgents(ctx context.Context, svcCtx *svc.ServiceContext) (*AgentsModel, error) {
    var initErr error
    agentsOnce.Do(func() {
        agents, initErr = InitAgents(ctx, svcCtx)
    })
    if initErr != nil {
        return nil, initErr
    }
    return agents, nil
}

// ❌ 错误：全局变量污染
var _agents *AgentsModel  // 禁止这样做
```

### 上下文传播

```go
// ✅ 正确：传播上下文
func InitAgents(ctx context.Context, svcCtx *svc.ServiceContext) error {
    ctx, cancel := context.WithCancel(ctx)  // 从父上下文派生
    defer func() {
        if err != nil {
            cancel()  // 失败时清理
        }
    }()
    // ... 初始化逻辑
}

// ❌ 错误：直接使用 context.Background()
ctx := context.Background()  // 禁止这样做，应传播父上下文
```

### 资源清理

```go
func InitAgents(ctx context.Context, svcCtx *svc.ServiceContext) (err error) {
    ctx, cancel := context.WithCancel(ctx)

    // 使用 defer 确保清理
    defer func() {
        if err != nil {
            cancel()
            // 清理已初始化的资源
        }
    }()

    // 初始化逻辑...
    return nil
}
```

---

## Tool 实现规范

### 标准 Tool 结构

```go
type OkxGetPositionsTool struct {
    svcCtx  *svc.ServiceContext
    limiter *rate.Limiter  // 限流器
}

func NewOkxGetPositionsTool(svcCtx *svc.ServiceContext) *OkxGetPositionsTool {
    return &OkxGetPositionsTool{
        svcCtx:  svcCtx,
        limiter: rate.NewLimiter(rate.Every(time.Second/5), 1),
    }
}

func (c *OkxGetPositionsTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "okx_get_positions",
        Desc: "查询当前持仓信息",
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            "symbol": {Type: "string", Desc: "交易对，留空表示所有", Required: false},
        }),
    }, nil
}

func (c *OkxGetPositionsTool) InvokableRun(ctx context.Context, argsJSON string, opts ...tool.Option) (string, error) {
    // 1. 等待限流
    if err := c.limiter.Wait(ctx); err != nil {
        return "", fmt.Errorf("rate limiter wait failed: %w", err)
    }

    // 2. 解析参数
    var params struct {
        Symbol string `json:"symbol"`
    }
    if err := json.Unmarshal([]byte(argsJSON), &params); err != nil {
        return "", fmt.Errorf("failed to unmarshal args: %w", err)
    }

    // 3. API 调用
    result, err := c.svcCtx.OKXClient.GetPositions(ctx, params.Symbol)
    if err != nil {
        return "", err  // 返回错误，不是错误字符串
    }

    // 4. 验证 OKX 响应码
    if result.Code != 0 {
        return "", fmt.Errorf("OKX API error: %s", result.Msg)
    }

    // 5. 返回结果
    return json.Marshal(result.Data)
}
```

---

## 项目结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # 应用入口
├── internal/
│   ├── agent/                   # Agent 定义
│   │   ├── agents.go            # Agent 初始化（单例模式）
│   │   └── okx_watcher/         # OKXWatcher Agent
│   │       ├── DESCRIPTION.md
│   │       └── SOUL.md
│   ├── config/                  # 配置管理
│   ├── handler/                 # HTTP 处理器
│   ├── model/                   # 数据模型
│   ├── repository/              # 数据访问层
│   ├── service/                 # 业务逻辑层
│   │   ├── agent/               # Agent 服务
│   │   ├── agentchat/           # 对话服务
│   │   └── okx/                 # OKX API 服务
│   └── middleware/              # 中间件
├── pkg/
│   ├── okex/                    # OKX API 客户端
│   │   └── api/
│   │       ├── trading.go       # 交易 API
│   │       └── trade_requests.go # 交易请求结构
│   └── chromedp-v0.15.0/        # Chrome 自动化（本地覆盖）
├── web/                         # 前端静态文件
├── etc/
│   ├── config.yaml              # 配置文件
│   └── config.example.yaml      # 配置示例
├── .planning/                   # GSD 规划文档
│   ├── PROJECT.md               # 项目愿景
│   ├── REQUIREMENTS.md          # 需求列表
│   ├── ROADMAP.md               # 路线图
│   ├── config.json              # 工作流配置
│   ├── codebase/                # 代码库映射
│   └── research/                # 研究报告
└── .claude/
    └── skills/
        ├── eino-skill/          # Eino 框架技能
        └── okex-skill/          # OKX API 技能
```

---

## 关键决策 (ADR)

| 决策 | 状态 | 说明 |
|------|------|------|
| DeepAgent 仅用于 OKXWatcher | ✓ 已批准 | 避免层级冗余，SubAgents 使用 ChatModelAgent |
| 分析/执行分离 | ✓ 已批准 | 清晰的审计追踪，独立测试 |
| Tool 原子化 | ✓ 已批准 | 每个 Tool 只做一件事 |
| RAG = Redis Stack + m3e-base | ✓ 已批准 | 本地 Embedding，无需外部 API |
| 独立风控层 | ✓ 已批准 | 实时监控，可覆盖交易决策 |
| Executor 从 Level 1 开始 | — 待决 | 仅执行明确指令，逐步建立信任 |

---

## 当前路线图

| 阶段 | 目标 | 状态 |
|------|------|------|
| Phase 1 | 基础安全：错误处理、限流、单例、上下文 | 待开始 |
| Phase 2 | 分析层：重构 SubAgents 为 ChatModelAgent | 待开始 |
| Phase 3 | 执行层：交易工具 + Executor Agent (Level 1) | 待开始 |
| Phase 4 | RAG 记忆：Redis Stack + 决策保存/搜索 | 待开始 |

**下一步：** `/gsd:plan-phase 1` 规划 Phase 1 执行计划

---

## 常用命令

```bash
# 启动开发服务器
go run cmd/server/main.go

# 构建生产二进制
go build -o server ./cmd/server

# 运行测试
go test ./...

# GSD 命令
/gsd:plan-phase 1      # 规划阶段 1
/gsd:execute-phase 1   # 执行阶段 1
/gsd:progress          # 查看项目进度
```

---

## 安全注意事项

1. **API 密钥保护**
   - `etc/config.yaml` 包含真实 API 密钥
   - 已添加到 `.gitignore`
   - 不要提交包含密钥的文件

2. **交易安全**
   - Executor 从 Level 1 开始（仅执行明确指令）
   - 必须先实现风控层才能进行实盘交易
   - 使用 OKX 沙盒环境进行测试

3. **错误处理**
   - 所有 Tool 必须返回 `("", err)` 而不是 `(err.Error(), nil)`
   - 验证 OKX 响应的 `sCode`/`sMsg` 字段
   - 实现 `RetryableError` 接口支持自动重试

---

## 参考资料

- **Eino 技能**: `.claude/skills/eino-skill/SKILL.md`
- **OKX API 技能**: `.claude/skills/okex-skill/SKILL.md`
- **项目文档**: `.planning/PROJECT.md`
- **路线图**: `.planning/ROADMAP.md`
- **代码规范**: `.planning/codebase/CONVENTIONS.md`
- **架构设计**: `.planning/codebase/ARCHITECTURE.md`
- **技术栈**: `.planning/codebase/STACK.md`
- **风险清单**: `.planning/codebase/CONCERNS.md`