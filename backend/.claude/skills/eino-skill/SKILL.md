---
name: eino-skill
description: How to build LLM applications with the Eino framework in Go. Make sure to use this skill whenever the user mentions "eino", wants to build an LLM application, needs to orchestrate workflows/chains, asks about Eino components (ChatModel, Tool, Retriever, Graph, Workflow, Agent), or is debugging Eino applications. This skill covers everything from basic concepts to advanced patterns like interrupt/resume, state management, RAG pipelines, and ADK multi-agent collaboration.
type: skill
---

# Eino Framework Skill - 完整指南

Eino [`'aino`] 是一个用 Go 编写的 LLM 应用开发框架，由字节跳动开源，灵感来自 LangChain 和 Google ADK，但遵循 Go 语言的约定。

## 何时使用本技能

- 用户提到 "eino" 或 "CloudWeGo Eino"
- 需要构建 LLM 应用、Agent、Workflow 或 Chain
- 需要集成 ChatModel、Tool、Retriever、Embedding、DocumentLoader 等组件
- 需要实现 RAG（检索增强生成）管道
- 需要实现 Interrupt/Resume（人机协作）模式
- 需要构建多 Agent 协作系统（DeepAgent、Supervisor、Plan-Execute）
- 需要调试 Eino 应用问题

## 快速开始 - 10 分钟上手

### Hello World - 最简单的对话应用

```go
package main

import (
    "context"
    "fmt"
    "github.com/cloudwego/eino-ext/components/model/openai"
    "github.com/cloudwego/eino/adk"
    "github.com/cloudwego/eino/schema"
)

func main() {
    ctx := context.Background()

    // 创建 ChatModel
    chatModel, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{
        APIKey: "sk-your-key",
        Model:  "gpt-4o",
    })

    // 创建 Agent（最简单的对话 Agent）
    agent := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
        Model:       chatModel,
        Instruction: "你是有用的助手",
    })

    // 创建 Runner 并执行
    runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent})
    iter := runner.Query(ctx, "你好，介绍一下你自己")

    for {
        event, ok := iter.Next()
        if !ok {
            break
        }
        fmt.Println(event.Message.Content)
    }
}
```

### 带工具的 Agent

```go
// 定义工具
type WeatherInput struct {
    City string `json:"city" jsonschema:"required,description=城市名称"`
}

weatherTool, _ := tool.InferTool(
    "get_weather",
    "查询城市天气",
    func(ctx context.Context, input WeatherInput) (string, error) {
        return "晴天，25°C", nil
    },
)

// 创建带工具的 Agent
agent := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Model:       chatModel,
    Instruction: "你是有用的助手，使用工具帮助用户",
    ToolsConfig: adk.ToolsConfig{
        ToolsNodeConfig: compose.ToolsNodeConfig{
            Tools: []tool.BaseTool{weatherTool},
        },
    },
})
```

## 核心概念

### Runnable 模式（流式处理）

| 模式 | 输入 | 输出 | 用途 |
|------|------|------|------|
| **Invoke** | `I` | `O` | 请求响应 |
| **Stream** | `I` | `StreamReader[O]` | 流式生成 |
| **Collect** | `StreamReader[I]` | `O` | 聚合输入 |
| **Transform** | `StreamReader[I]` | `StreamReader[O]` | 双向流式处理 |

### 组件接口

所有组件实现 `Typer` 和 `Checker` 接口，支持自动回调注入：

```go
type Typer interface {
    GetType() string
}

type Checker interface {
    Needed(timing CallbackTiming) bool
}
```

## 主题索引

### 编排模式

| 模式 | 说明 | 适用场景 | 参考文档 |
|------|------|----------|----------|
| **Chain** | 顺序编排 | 线性流程 | [references/chain.md](references/chain.md) |
| **Graph** | DAG 编排，支持状态管理 | 复杂流程、条件分支 | [references/graph.md](references/graph.md) |
| **Workflow** | 字段映射编排 | 数据结构转换 | 本节内示例 |

### 组件使用

| 组件 | 说明 | 参考文档 |
|------|------|----------|
| **ChatModel** | LLM 模型调用 | 本节内代码示例 |
| **Tool** | 工具定义与调用 | [references/tool.md](references/tool.md) |
| **Retriever** | 文档检索 | [references/rag.md](references/rag.md) |
| **Embedding** | 文本向量化 | [references/rag.md](references/rag.md) |
| **Prompt** | 模板格式化 | 本节内示例 |

### 高级特性

| 特性 | 说明 | 参考文档 |
|------|------|----------|
| **Interrupt/Resume** | 人机协作、中断恢复 | [references/interrupt.md](references/interrupt.md) |
| **Callback** | 可观测性、日志追踪 | [references/callback.md](references/callback.md) |
| **RAG** | 检索增强生成 | [references/rag.md](references/rag.md) |
| **DeepAgent** | 多 Agent 层级协作 | [references/deep_agent.md](references/deep_agent.md) |
| **ADK** | Agent 开发工具包 | 本节内示例 |
| **Checkpoint** | 状态持久化 | [references/interrupt.md](references/interrupt.md) |

---

## ADK - Agent 开发工具包

### ChatModelAgent（基础对话 Agent）

```go
import "github.com/cloudwego/eino/adk"

agent := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Name:        "助手",
    Model:       chatModel,
    Instruction: "你是有用的助手",
    ToolsConfig: adk.ToolsConfig{
        ToolsNodeConfig: compose.ToolsNodeConfig{
            Tools: []tool.BaseTool{myTool},
        },
    },
    MaxIterations: 10, // 最大迭代次数
})

runner := adk.NewRunner(ctx, adk.RunnerConfig{
    Agent:           agent,
    EnableStreaming: true,
})

// 执行查询
iter := runner.Query(ctx, "你好")
for {
    event, ok := iter.Next()
    if !ok {
        break
    }
    // 处理事件
}
```

### Workflow Agent - 模式

#### LoopAgent - 循环反思模式

```go
import "github.com/cloudwego/eino/adk/prebuilt/workflow"

loopAgent := workflow.NewLoopAgent(ctx, &workflow.LoopAgentConfig{
    Name:        "反思 Agent",
    Model:       chatModel,
    Instruction: "你不断反思和改进答案",
    MaxLoops:    5, // 最大循环次数
})
```

#### ParallelAgent - 并行执行模式

```go
parallelAgent := workflow.NewParallelAgent(ctx, &workflow.ParallelAgentConfig{
    Name:          "并行 Agent",
    Model:         chatModel,
    SubAgents:     []adk.Agent{agent1, agent2, agent3},
    MergeStrategy: workflow.MergeStrategyConcat, // 合并策略
})
```

#### SequentialAgent - 顺序执行模式

```go
sequentialAgent := workflow.NewSequentialAgent(ctx, &workflow.SequentialAgentConfig{
    Name:      "顺序 Agent",
    Model:     chatModel,
    SubAgents: []adk.Agent{agent1, agent2, agent3},
})
```

### DeepAgent - 多 Agent 协作

```go
import "github.com/cloudwego/eino/adk/prebuilt/deep"

// 创建子 Agent
codeAgent := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Name:        "CodeAgent",
    Description: "负责编写和执行代码",
    Model:       codeModel,
    ToolsConfig: /* 代码工具 */,
})

searchAgent := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Name:        "SearchAgent",
    Description: "负责网络搜索",
    Model:       searchModel,
    ToolsConfig: /* 搜索工具 */,
})

// 创建 DeepAgent（主 Agent 协调子 Agent）
deepAgent, _ := deep.New(ctx, &deep.Config{
    Name:        "ExcelAgent",
    Description: "处理 Excel 任务的多 Agent 系统",
    ChatModel:   cm,
    SubAgents:   []adk.Agent{codeAgent, searchAgent},
    ToolsConfig: adk.ToolsConfig{
        Tools: []tool.BaseTool{readFileTool, treeTool},
    },
    MaxIteration: 100,
})

runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: deepAgent})
iter := runner.Run(ctx, messages)
```

**详细说明**: [references/deep_agent.md](references/deep_agent.md)

### Session 管理 - 跨 Agent 传递状态

```go
import "github.com/cloudwego/eino/adk/session"

// 创建 Session
session := session.NewSession(ctx, session.Config{
    SessionID: "unique-session-id",
})

// 在 Agent 间传递
ctx = session.WithContext(ctx)

// 访问 Session 数据
data := session.Get(ctx, "key")
session.Set(ctx, "key", value)
```

### Agent Transfer - 任务转移

```go
import "github.com/cloudwego/eino/adk"

// 创建可转移的 Agent
agent := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Name:        "路由 Agent",
    Model:       chatModel,
    Instruction: "根据任务类型转移到专业 Agent",
    TransferConfig: adk.TransferConfig{
        TransferrableAgents: []adk.AgentTransferInfo{
            {Name: "codeAgent", Description: "处理代码任务"},
            {Name: "searchAgent", Description: "处理搜索任务"},
        },
    },
})
```

---

## Graph 编排模式

### 简单 Graph

```go
package main

import (
    "context"
    "github.com/cloudwego/eino-ext/components/model/openai"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
)

func main() {
    ctx := context.Background()

    chatModel, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{
        APIKey: "sk-key",
        Model:  "gpt-4o",
    })

    // 创建图
    g := compose.NewGraph[[]*schema.Message, *schema.Message]()

    // 添加节点
    _ = g.AddChatModelNode("model", chatModel)

    // 定义边
    _ = g.AddEdge(compose.START, "model")
    _ = g.AddEdge("model", compose.END)

    // 编译
    runnable, _ := g.Compile(ctx, compose.WithGraphName("SimpleGraph"))

    // 执行
    msgs := []*schema.Message{{Role: "user", Content: "Hello"}}
    result, _ := runnable.Invoke(ctx, msgs)
    fmt.Println(result.Content)
}
```

### 带状态的 Graph

```go
// 定义状态结构
type AppState struct {
    History []*schema.Message
    Count   int
}

// 创建带状态的图
g := compose.NewGraph[*UserInput, *schema.Message](
    compose.WithGenLocalState(func(ctx context.Context) *AppState {
        return &AppState{}
    }),
)

// 添加节点（带状态处理器）
_ = g.AddLambdaNode("process", compose.InvokableLambda(
    func(ctx context.Context, input *UserInput) (*schema.Message, error) {
        // 获取状态
        state := compose.GetState[*AppState](ctx)
        state.Count++
        state.History = append(state.History, input.ToMessage())
        return process(input), nil
    }),
    compose.WithStatePreHandler(func(ctx context.Context, in *UserInput, state *AppState) (*UserInput, error) {
        // 前置处理：更新状态
        state.History = append(state.History, in.ToMessage())
        return in, nil
    }),
    compose.WithStatePostHandler(func(ctx context.Context, out *schema.Message, state *AppState) (*schema.Message, error) {
        // 后置处理：更新状态
        state.History = append(state.History, out)
        return out, nil
    }),
)
```

### 带分支的 Graph

```go
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

### 在 ToolsNode 前自动中断（人机协作）

```go
// 编译时配置中断点
runnable, err := g.Compile(ctx,
    compose.WithCheckPointStore(store),
    compose.WithInterruptBeforeNodes([]string{"ToolsNode"}),
)

// 执行时会中断等待用户确认
_, err = runnable.Invoke(ctx, input)
if compose.IsInterruptError(err) {
    info := compose.GetInterruptInfo(ctx)
    fmt.Printf("中断点：%v\n", info.NodePath)
    fmt.Printf("中断信息：%v\n", info.Info)

    // 用户确认后恢复
    ctx = compose.Resume(ctx, info.InterruptID)
    result, _ := runnable.Invoke(ctx, input,
        compose.WithCheckPointID(info.CheckPointID),
    )
}
```

---

## Chain 编排模式

### 线性 Chain

```go
chain := compose.NewChain[*CustomerQuery, *Response]()

chain.
    AppendChatTemplate(template).
    AppendChatModel(chatModel).
    AppendLambda(compose.InvokableLambda(
        func(ctx context.Context, msg *schema.Message) (*Response, error) {
            return &Response{Answer: msg.Content}, nil
        },
    ))

runnable, _ := chain.Compile(ctx)
```

### 带工具的 Chain

```go
chain := compose.NewChain[map[string]any, string]()

chain.
    AppendChatTemplate(chatTpl).
    AppendChatModel(model).
    AppendToolsNode(toolsNode).
    AppendLambda(compose.InvokableLambda(
        func(ctx context.Context, msg *schema.Message) (string, error) {
            return msg.Content, nil
        },
    ))

runnable, _ := chain.Compile(ctx)
```

---

## Workflow 编排模式

Workflow 是基于 Graph 的更高阶抽象，通过字段映射实现数据流转。

### 简单 Workflow

```go
import "github.com/cloudwego/eino/compose"

type Input struct {
    Query string
}

type Output struct {
    Answer string
}

w := compose.NewWorkflow[Input, Output]()

// 添加节点
w.AddChatModelNode("model", chatModel)
w.AddLambdaNode("format", formatFn)

// 定义边和字段映射
w.AddEdge(compose.START, "model",
    compose.MapFields("Query", "Content"))
w.AddEdge("model", "format",
    compose.MapFields("Content", "input"))
w.AddEdge("format", compose.END,
    compose.MapFields("output", "Answer"))
```

---

## Tool 工具系统

### 使用 InferTool 快速创建

```go
import "github.com/cloudwego/eino/utils/tool"

type WeatherInput struct {
    City string `json:"city" jsonschema:"required,description=城市名称"`
}

weatherTool, _ := tool.InferTool(
    "get_weather",
    "查询城市天气",
    func(ctx context.Context, input WeatherInput) (string, error) {
        // 实现逻辑
        return "晴天，25°C", nil
    },
)
```

### 手动实现 InvokableTool

```go
type WeatherTool struct {
    apiKey string
}

func (t *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "get_weather",
        Desc: "查询城市天气",
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            "city": {Type: "string", Desc: "城市名称", Required: true},
        }),
    }, nil
}

func (t *WeatherTool) InvokableRun(ctx context.Context, argsJSON string, opts ...tool.Option) (string, error) {
    var input struct {
        City string `json:"city"`
    }
    json.Unmarshal([]byte(argsJSON), &input)
    // 执行逻辑
    return result, nil
}
```

### ToolsNode 配置

```go
toolsNode, _ := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
    Tools: []tool.BaseTool{tool1, tool2},
    // 并行执行工具
    Parallel: true,
})
```

---

## RAG 检索增强生成

### 基础 RAG 流程

```go
// 1. 创建 Embedding
embedder, _ := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
    APIKey: apiKey,
    Model:  "text-embedding-3-small",
})

// 2. 创建 Retriever（以 Redis 为例）
retriever, _ := redis.NewRetriever(ctx, &redis.RetrieverConfig{
    Client:    redisClient,
    Index:     "knowledge_base",
    Embedding: embedder,
    TopK:      5,
})

// 3. 创建 RAG Graph
type RAGState struct {
    Query     string
    Documents []*schema.Document
    Answer    string
}

g := compose.NewGraph[string, string](
    compose.WithGenLocalState(func(ctx context.Context) *RAGState {
        return &RAGState{}
    }),
)

// 检索节点
_ = g.AddRetrieverNode("retrieve", retriever,
    compose.WithStatePreHandler(func(ctx context.Context, query string, state *RAGState) (string, error) {
        state.Query = query
        return query, nil
    }),
)

// Prompt 节点
_ = g.AddChatTemplateNode("prompt", ragTemplate,
    compose.WithStatePreHandler(func(ctx context.Context, docs []*schema.Document, state *RAGState) (map[string]any, error) {
        state.Documents = docs
        var docText string
        for _, doc := range docs {
            docText += doc.Content + "\n\n"
        }
        return map[string]any{
            "documents": docText,
            "query":     state.Query,
        }, nil
    }),
)

// 生成节点
_ = g.AddChatModelNode("generate", chatModel)

// 边
_ = g.AddEdge(compose.START, "retrieve")
_ = g.AddEdge("retrieve", "prompt")
_ = g.AddEdge("prompt", "generate")
_ = g.AddEdge("generate", compose.END)

runnable, _ := g.Compile(ctx)
answer, _ := runnable.Invoke(ctx, "公司的年假政策是什么？")
```

### 索引文档

```go
// 索引文档到向量库
func IndexDocuments(ctx context.Context, docs []*schema.Document) error {
    indexer, _ := redis.NewIndexer(ctx, &redis.IndexerConfig{
        Client:    redisClient,
        KeyPrefix: "doc:",
        Index:     "knowledge_base",
        Embedding: embedder,
    })

    ids, err := indexer.Store(ctx, docs)
    if err != nil {
        return err
    }
    fmt.Printf("索引了 %d 个文档\n", len(ids))
    return nil
}

// 示例文档
docs := []*schema.Document{
    {
        ID:      "policy-001",
        Content: "公司年假政策：员工每年享有 15 天带薪年假...",
        MetaData: map[string]any{
            "category": "hr",
            "updated":  "2024-01-01",
        },
    },
}
_ = IndexDocuments(ctx, docs)
```

---

## Interrupt/Resume 人机协作

### 基础 Interrupt 示例

```go
type ApprovalState struct {
    PendingApproval bool
    ApprovalResult  string
}

g := compose.NewGraph[string, string](
    compose.WithGenLocalState(func(ctx context.Context) *ApprovalState {
        return &ApprovalState{}
    }),
)

// 敏感操作节点
_ = g.AddLambdaNode("sensitive_op", compose.InvokableLambda(
    func(ctx context.Context, input string) (string, error) {
        state := compose.GetState[*ApprovalState](ctx)

        // 检查是否是恢复执行
        isResume, hasData, data := compose.GetResumeContext[ApprovalResult](ctx)
        if isResume && hasData {
            state.PendingApproval = false
            state.ApprovalResult = data.Status
            return "已执行：" + data.Status, nil
        }

        // 首次执行，触发中断
        state.PendingApproval = true
        compose.Interrupt(ctx, map[string]any{
            "action":  "approval",
            "message": "需要用户批准",
            "details": input,
        })
        return "", nil
    },
))

_ = g.AddEdge(compose.START, "sensitive_op")
_ = g.AddEdge("sensitive_op", compose.END)

// 编译（带 CheckpointStore）
store := &MemoryStore{data: make(map[string][]byte)}
runnable, _ := g.Compile(ctx,
    compose.WithCheckPointStore(store),
    compose.WithInterruptBeforeNodes([]string{"sensitive_op"}),
)

// 首次执行（会中断）
_, err := runnable.Invoke(ctx, "删除数据")
if compose.IsInterruptError(err) {
    info := compose.GetInterruptInfo(ctx)
    fmt.Printf("等待用户确认：%v\n", info.Info)

    // 用户确认后恢复
    ctx = compose.ResumeWithData(ctx, info.InterruptID, ApprovalResult{
        Status: "approved",
    })
    result, _ := runnable.Invoke(ctx, "删除数据",
        compose.WithCheckPointID(info.CheckPointID),
    )
}
```

### CheckpointStore 实现

```go
// 内存实现（测试用）
type MemoryStore struct {
    data map[string][]byte
}

func (m *MemoryStore) Get(ctx context.Context, id string) ([]byte, bool, error) {
    d, ok := m.data[id]
    return d, ok, nil
}

func (m *MemoryStore) Set(ctx context.Context, id string, data []byte) error {
    m.data[id] = data
    return nil
}

// Redis 实现（生产用）
type RedisStore struct {
    client *redis.Client
}

func (r *RedisStore) Get(ctx context.Context, id string) ([]byte, bool, error) {
    data, err := r.client.Get(ctx, id).Bytes()
    if err == redis.Nil {
        return nil, false, nil
    }
    if err != nil {
        return nil, false, err
    }
    return data, true, nil
}

func (r *RedisStore) Set(ctx context.Context, id string, data []byte) error {
    return r.client.Set(ctx, id, data, 24*time.Hour).Err()
}
```

---

## Callback 可观测性

### 基础 Callback

```go
import "github.com/cloudwego/eino/callbacks"

// 创建 Handler
handler := callbacks.NewHandlerBuilder().
    OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
        log.Printf("开始：%s (类型：%s)", info.Name, info.Type)
        return ctx
    }).
    OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
        log.Printf("结束：%s", info.Name)
        return ctx
    }).
    OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
        log.Printf("错误：%s - %v", info.Name, err)
        return ctx
    }).
    Build()

// 全局注册
callbacks.AppendGlobalHandlers(handler)
```

### 流式 Callback

```go
handler := callbacks.NewHandlerBuilder().
    OnEndWithStreamOutputFn(func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
        defer output.Close() // 必须关闭流

        var chunks []callbacks.CallbackOutput
        for {
            chunk, err := output.Recv()
            if err == io.EOF {
                break
            }
            chunks = append(chunks, chunk)
        }
        log.Printf("流式输出完成，共 %d 块", len(chunks))
        return ctx
    }).
    Build()
```

### Langfuse 集成

```go
import (
    "github.com/cloudwego/eino-ext/callbacks/langfuse"
    "github.com/cloudwego/eino/callbacks"
)

handler, _ := langfuse.NewLangfuseHandler(&langfuse.Config{
    Host:        "https://cloud.langfuse.com",
    PublicKey:   "pk-lf-...",
    SecretKey:   "sk-lf-...",
    ProjectID:   "your-project-id",
    EnableTrace: true,
})

callbacks.AppendGlobalHandlers(handler)
```

---

## 调试技巧

### 可视化 Graph

```go
import "github.com/cloudwego/eino-ext/devops/visualize"

runnable, err := graph.Compile(ctx,
    compose.WithGraphCompileCallbacks(visualize.NewMermaidGenerator(
        "MyGraph",
        func(mermaidStr string) {
            fmt.Println(mermaidStr)
        },
    )),
)
```

### 使用 CozeLoop 追踪

```go
import (
    clc "github.com/cloudwego/eino-ext/callbacks/cozeloop"
    "github.com/coze-dev/cozeloop-go"
)

client, _ := cozeloop.NewClient(
    cozeloop.WithAPIToken(os.Getenv("COZELOOP_API_TOKEN")),
    cozeloop.WithWorkspaceID(os.Getenv("COZELOOP_WORKSPACE_ID")),
)
defer client.Close(ctx)

loopHandler := clc.NewLoopHandler(client)
callbacks.AppendGlobalHandlers(loopHandler)
```

---

## 扩展组件

| 类型 | 提供者 | 安装包 |
|------|--------|--------|
| **ChatModel** | OpenAI, Claude, Gemini, Ark, Ollama, DeepSeek, Qwen | `github.com/cloudwego/eino-ext/components/model/{provider}` |
| **Embedding** | OpenAI, Ark, Ollama | `github.com/cloudwego/eino-ext/components/embedding/{provider}` |
| **Retriever** | Redis, Milvus, Elasticsearch, Qdrant | `github.com/cloudwego/eino-ext/components/retriever/{provider}` |
| **Indexer** | Redis, Milvus, Elasticsearch | `github.com/cloudwego/eino-ext/components/indexer/{provider}` |
| **Tool** | MCP, DuckDuckGo, Google Search, Browser Use | `github.com/cloudwego/eino-ext/components/tool/{provider}` |
| **Callback** | Langfuse, Langsmith, CozeLoop | `github.com/cloudwego/eino-ext/callbacks/{provider}` |

安装：`go get github.com/cloudwego/eino-ext/components/{type}/{provider}@latest`

---

## 最佳实践

### 1. 状态管理

```go
// 定义可序列化的状态类型
type myState struct {
    history []*schema.Message
    count   int
}

// 注册可序列化类型
compose.RegisterSerializableType[myState]("my_state")

// 在 Graph 中使用
g := compose.NewGraph[I, O](
    compose.WithGenLocalState(func(ctx context.Context) *myState {
        return &myState{}
    }),
)
```

### 2. 错误处理

```go
result, err := runnable.Invoke(ctx, input)
if err != nil {
    if compose.IsInterruptError(err) {
        // 处理中断
        info := compose.GetInterruptInfo(err)
        // ...
    } else {
        // 处理其他错误
        log.Printf("执行失败：%v", err)
    }
}
```

### 3. 并发控制

```go
// 使用 BatchNode 控制并发
batchNode := compose.NewBatchNode(ctx, &compose.BatchNodeConfig{
    MaxConcurrency: 5, // 最大并发数
    MaxBatchSize:   10, // 最大批次大小
})
```

### 4. 流式处理最佳实践

```go
// 始终关闭 StreamReader
stream, err := runnable.Stream(ctx, input)
defer stream.Close()

for {
    chunk, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Printf("流错误：%v", err)
        break
    }
    // 处理 chunk
}
```

---

## 参考资源

- **官方文档**: https://www.cloudwego.io/docs/eino/
- **Eino 源码**: https://github.com/cloudwego/eino
- **官方示例**: https://github.com/cloudwego/eino-examples
- **扩展包**: https://github.com/cloudwego/eino-ext
- **llms.txt**: https://raw.githubusercontent.com/cloudwego/eino/refs/heads/main/llms.txt - 包含完整文档索引

## 参考文档详情

- [Chain 模式详解](references/chain.md)
- [Graph 模式详解](references/graph.md)
- [Tool 工具系统](references/tool.md)
- [RAG 检索增强生成](references/rag.md)
- [Interrupt/Resume 人机协作](references/interrupt.md)
- [Callback 可观测性](references/callback.md)
- [DeepAgent 多 Agent 协作](references/deep_agent.md)
