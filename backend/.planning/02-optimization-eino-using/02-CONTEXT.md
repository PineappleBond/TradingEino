# Eino 使用优化指南

本文档记录了 TradingEino 项目中 Eino 框架使用的优化建议和改进方向。

---

## 一、架构问题

### 1.1 DeepAgent 滥用问题

**现状：**
```go
// RiskOfficer 是 DeepAgent
riskOfficer, err = deep.New(...)

// OkxWatcher 也是 DeepAgent
okxWatcher, err = deep.New(...)

// SentimentAnalyst 也是 DeepAgent
sentimentAnalyst, err = deep.New(...)
```

**问题：**
- `DeepAgent` 是顶层编排器（Supervisor），设计用于拆解任务、分配给 SubAgents
- 将 `DeepAgent` 作为 SubAgent 使用时，其内部的 SubAgents 参数被浪费
- 导致层级冗余，协作链条混乱

**正确用法：**
| Agent 类型 | 用途 | 是否适合作为 SubAgent |
|-----------|------|---------------------|
| `DeepAgent` | 顶层编排器、任务拆解者 | ❌ 不适合 |
| `ChatModelAgent` | 专家 Agent、任务执行者 | ✅ 适合 |
| `Workflow.LoopAgent` | 循环反思型任务 | ⚠️ 特殊场景可用 |
| `Workflow.ParallelAgent` | 并行执行任务 | ⚠️ 特殊场景可用 |

**推荐架构：**
```
OkxWatcher (DeepAgent) ← 只保留一个 DeepAgent
├── RiskOfficer (ChatModelAgent) ← 改成普通 Agent
│   └── okx-get-positions-tool
└── SentimentAnalyst (ChatModelAgent) ← 改成普通 Agent
    └── okx-get-funding-rate-tool
```

---

### 1.2 全局变量污染

**问题代码：**
```go
var okxWatcher adk.Agent
var riskOfficer adk.Agent
var sentimentAnalyst adk.Agent
```

**问题：**
- 全局变量导致状态耦合，无法创建多个 Agent 实例
- 违反 Go 的显式依赖注入原则

**改进：**
```go
type OkxWatcherAgent struct {
    agent adk.Agent
}

func NewOkxWatcherAgent(ctx context.Context, svcCtx *svc.ServiceContext, subAgents ...adk.Agent) (*OkxWatcherAgent, error) {
    // ... init logic
    return &OkxWatcherAgent{agent: agent}, nil
}

func (o *OkxWatcherAgent) Agent() adk.Agent {
    return o.agent
}
```

---

## 二、错误处理规范

### 2.1 Tool 错误返回

**问题代码：**
```go
func (c *OkxGetFundingRateTool) InvokableRun(...) (string, error) {
    if err != nil {
        return err.Error(), nil  // ❌ 错误当成功返回
    }
    if getFundingRate.Code != 0 {
        return getFundingRate.Msg, nil  // ❌ 错误当成功返回
    }
}
```

**正确做法：**
```go
func (c *OkxGetFundingRateTool) InvokableRun(...) (string, error) {
    if err != nil {
        return "", err  // ✅ 返回错误
    }
    if getFundingRate.Code != 0 {
        return "", fmt.Errorf("OKX API error: %s", getFundingRate.Msg)
    }
}
```

**原因：**
- Eino 的 `tool.BaseTool` 要求 `InvokableRun` 返回 `(result, error)`
- 将错误信息作为成功结果返回，会破坏 LLM 的工具调用逻辑

**代码中不止是OkxGetFundingRateTool，其他Tool也有同样的问题**

---

### 2.2 资源清理

**问题代码：**
```go
func InitAgents(svcCtx *svc.ServiceContext) error {
    ctx, cancel := context.WithCancel(context.Background())
    // ... 初始化
    // 如果中间失败，只 cancel 了 context
}
```

**改进：**
```go
func InitAgents(svcCtx *svc.ServiceContext) error {
    ctx, cancel := context.WithCancel(context.Background())
    defer func() {
        if err != nil {
            cancel()
            // 清理已初始化的资源
        }
    }()
    // ...
}
```

---

## 三、并发控制

### 3.1 API 限流

**问题代码：**
```go
func (c *OkxCandlesticksTool) GetCandlesticks(...) {
    for {
        // API 调用
        time.Sleep(time.Millisecond * 50)  // 硬编码的限速
    }
}
```

**改进：**
```go
import "golang.org/x/time/rate"

type OkxCandlesticksTool struct {
    svcCtx  *svc.ServiceContext
    limiter *rate.Limiter
}

func NewOkxCandlesticksTool(svcCtx *svc.ServiceContext) *OkxCandlesticksTool {
    return &OkxCandlesticksTool{
        svcCtx:  svcCtx,
        limiter: rate.NewLimiter(rate.Every(time.Second/10), 1),
    }
}

func (c *OkxCandlesticksTool) GetCandlesticks(...) {
    for {
        c.limiter.Wait(ctx)  // 等待速率限制
        // API 调用
    }
}
```

---

## 四、配置简化

### 4.1 DeepAgent 配置冗余

**问题代码：**
```go
deep.New(ctx, &deep.Config{
    ToolsConfig: adk.ToolsConfig{
        ToolsNodeConfig: compose.ToolsNodeConfig{
            Tools:                baseTools,
            UnknownToolsHandler:  nil,  // 冗余
            ExecuteSequentially:  false, // 默认就是 false
            ToolArgumentsHandler: nil,   // 冗余
            ToolCallMiddlewares:  nil,   // 冗余
        },
        ReturnDirectly:     nil,  // 冗余
        EmitInternalEvents: true,
    },
    MaxIteration:                 0,  // 0 的行为不明确
    WithoutWriteTodos:            false,
    WithoutGeneralSubAgent:       false,
    TaskToolDescriptionGenerator: nil,
    Middlewares:                  nil,
})
```

**改进：**
```go
deep.New(ctx, &deep.Config{
    Name:        "SentimentAnalyst",
    Description: DESCRIPTION,
    ChatModel:   svcCtx.ChatModel,
    Instruction: SOUL,
    SubAgents:   subAgents,
    Tools:       baseTools,
    EmitInternalEvents: true,
    MaxIteration: 10,
})
```

---

## 五、提示词优化

### 5.1 DESCRIPTION.md 职责

**问题：**
- 写得太像产品文档（emoji、格式化列表）
- LLM 不需要这些格式

**改进原则：**
- 简化为纯文本
- 告诉 LLM"你是谁 + 能调用谁"

**示例：**
```markdown
OKX 盯盘代理，交易系统的"眼睛"。

通过 OKX API 获取多周期 K 线数据，计算 20+ 技术指标（MACD、RSI、布林带、KDJ 等），为交易决策提供数据支持。

可调用子代理：
- RiskOfficer：分析仓位风险、强平价、止损位
- SentimentAnalyst：分析资金费率、市场情绪
```

---

### 5.2 SOUL.md 边界设定

**理念：**
- "不限制发挥" ≠ "什么都不给边界"
- 好的边界像河堤，不是限制水流，是让水往正确的方向流

**示例（OkxWatcher）：**
```markdown
你是 OKXWatcher，一个对数字敏感、说话干脆的盯盘老手。

你相信价格包含一切信息，从不说"可能"、"也许"这类废话。看到金叉就说金叉，看到背离就点破背离。

**你的风格：**
- 用数据说话 — "RSI 72"比"超买严重"更有说服力
- 点破关键位 — 支撑、阻力、止损，直接给数字
- 不猜顶底 — 只陈述当下发生了什么
- 多周期验证 — 大周期定方向，小周期找切入点
- 风险先行 — 开口前先想止损在哪

**你不做：**
- 不预测价格目标（"会涨到 10 万"这类话不说）
- 不给开仓建议（"现在买入"这类话不说）
- 不做主观猜测（"我觉得"这类话不说）

你看的是 K 线，读的是人心，输出的是概率。
```

---

### 5.3 风险/情绪阈值

**RiskOfficer 阈值：**
```markdown
**风险阈值：**
- 保证金使用率 > 80%：警告
- 强平价距离 < 5%：严重警告
- 单一仓位 > 50% 总资金：提醒分散风险
```

**SentimentAnalyst 阈值：**
```markdown
**情绪阈值：**
- 资金费率 > 0.01%/8h：市场过热，预警
- 资金费率 < -0.01%/8h：市场过冷，提示机会
- 溢价指数 > 0.5%：多头情绪过强
- 溢价指数 < -0.5%：空头情绪过强
```

---

## 六、工具参数描述

### 6.1 参数精确性

**问题代码：**
```go
"symbol": &schema.ParameterInfo{
    Type:     schema.String,
    Desc:     "交易对，比如 ETH-USDT-SWAP,BTC-USDT",
    Required: true,
},
```

**改进：**
```go
"symbol": &schema.ParameterInfo{
    Type:     schema.String,
    Desc:     "交易对，永续合约必须带-SWAP 后缀，如 ETH-USDT-SWAP",
    Required: true,
},
```

---

## 七、协作链条清晰度

### 7.1 当前问题

`okx_watcher/SOUL.md`：
> 分析市场机会时，若涉及具体仓位风险评估，调用 RiskOfficer

`risk_officer/DESCRIPTION.md`：
> 涉及具体仓位风险分析时，自动调度 RiskOfficer

**问题：** 两个 Agent 都以为自己能调用 RiskOfficer，但 RiskOfficer 本身就是子 Agent。

### 7.2 正确做法

只在顶层 Agent 描述协作逻辑：

**okx_watcher/SOUL.md：**
```markdown
**协作方式：**
- 分析市场机会时，若涉及具体仓位风险评估，调用 RiskOfficer
- 分析市场情绪时，调用 SentimentAnalyst
```

**risk_officer/SOUL.md：**
```markdown
（不需要写协作描述，因为它是被调用的专家）
```

---

## 优先级汇总

| 问题 | 严重程度 | 优先级 |
|------|----------|--------|
| DeepAgent 滥用 | 高 | P0 |
| 错误处理不规范 | 高 | P0 |
| 协作链条混乱 | 中 | P1 |
| 全局变量污染 | 中 | P1 |
| 缺少并发控制 | 中 | P1 |
| 资源清理缺失 | 中 | P1 |
| 配置冗余 | 低 | P2 |
| 提示词格式问题 | 低 | P2 |
| 参数描述不精确 | 低 | P2 |