# Eino Graph 模式详解

Graph 是有向无环图（DAG）编排模式，适合复杂流程，支持状态管理、分支循环和中断恢复。

## 基本概念

- **节点（Node）**：执行单元，如 ChatModel、ToolsNode、Lambda、ChatTemplate、Retriever
- **边（Edge）**：定义节点间的执行顺序
- **分支（Branch）**：条件执行路径
- **状态（State）**：在节点间共享的数据
- **START/END**：预留的起始和结束节点

---

## 完整示例

### 1. 简单 Graph - 单节点

```go
package main

import (
    "context"
    "fmt"
    "github.com/cloudwego/eino-ext/components/model/openai"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
)

func main() {
    ctx := context.Background()

    // 创建 ChatModel
    chatModel, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{
        APIKey: "sk-key",
        Model:  "gpt-4o",
    })

    // 1. 创建图
    g := compose.NewGraph[[]*schema.Message, *schema.Message]()

    // 2. 添加节点
    _ = g.AddChatModelNode("model", chatModel)

    // 3. 定义边
    _ = g.AddEdge(compose.START, "model")
    _ = g.AddEdge("model", compose.END)

    // 4. 编译
    runnable, err := g.Compile(ctx, compose.WithGraphName("SimpleGraph"))
    if err != nil {
        panic(err)
    }

    // 5. 执行
    msgs := []*schema.Message{{Role: "user", Content: "Hello"}}
    result, _ := runnable.Invoke(ctx, msgs)
    fmt.Println(result.Content)
}
```

---

### 2. 带状态管理的 Graph - 对话历史

```go
// 定义状态结构
type ConversationState struct {
    History []*schema.Message
    Count   int
}

// 创建带状态的图
g := compose.NewGraph[*UserInput, *schema.Message](
    compose.WithGenLocalState(func(ctx context.Context) *ConversationState {
        return &ConversationState{History: make([]*schema.Message, 0)}
    }),
)

// 添加节点（带状态处理器）
_ = g.AddLambdaNode("process", compose.InvokableLambda(
    func(ctx context.Context, input *UserInput) (*schema.Message, error) {
        // 获取状态
        state := compose.GetState[*ConversationState](ctx)
        state.Count++
        state.History = append(state.History, input.ToMessage())
        return process(input), nil
    }),
    compose.WithStatePreHandler(func(ctx context.Context, in *UserInput, state *ConversationState) (*UserInput, error) {
        // 前置处理：更新状态
        state.History = append(state.History, in.ToMessage())
        return in, nil
    }),
    compose.WithStatePostHandler(func(ctx context.Context, out *schema.Message, state *ConversationState) (*schema.Message, error) {
        // 后置处理：更新状态
        state.History = append(state.History, out)
        return out, nil
    }),
)
```

---

### 3. Tool Call Agent Graph - 带工具调用的 Graph

来自官方示例 `compose/graph/tool_call_agent`：

```go
package main

import (
    "context"
    "github.com/cloudwego/eino-ext/components/model/openai"
    "github.com/cloudwego/eino/callbacks"
    "github.com/cloudwego/eino/components/prompt"
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/components/tool/utils"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
)

// 定义请求/响应结构
type userInfoRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type userInfoResponse struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Company  string `json:"company"`
    Position string `json:"position"`
    Salary   string `json:"salary"`
}

func main() {
    ctx := context.Background()

    // 1. 创建 ChatTemplate
    systemTpl := `你是一名房产经纪人，结合用户的薪酬和工作，使用 user_info API，为其提供相关的房产信息。邮箱是必须的`
    chatTpl := prompt.FromMessages(schema.FString,
        schema.SystemMessage(systemTpl),
        schema.MessagesPlaceholder("message_histories", true),
        schema.UserMessage("{user_query}"),
    )

    // 2. 创建 ChatModel
    chatModel, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{
        BaseURL:     os.Getenv("OPENAI_BASE_URL"),
        APIKey:      os.Getenv("OPENAI_API_KEY"),
        Model:       os.Getenv("OPENAI_MODEL_NAME"),
        Temperature: gptr.Of(float32(0.7)),
    })

    // 3. 创建工具
    userInfoTool := utils.NewTool(
        &schema.ToolInfo{
            Name: "user_info",
            Desc: "根据用户的姓名和邮箱，查询用户的公司、职位、薪酬信息",
            ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
                "name":  {Type: "string", Desc: "用户的姓名"},
                "email": {Type: "string", Desc: "用户的邮箱"},
            }),
        },
        func(ctx context.Context, input *userInfoRequest) (*userInfoResponse, error) {
            return &userInfoResponse{
                Name:     input.Name,
                Email:    input.Email,
                Company:  "Awesome company",
                Position: "CEO",
                Salary:   "9999",
            }, nil
        })

    // 4. 绑定工具到 ChatModel
    info, _ := userInfoTool.Info(ctx)
    _ = chatModel.BindForcedTools([]*schema.ToolInfo{info})

    // 5. 创建 ToolsNode
    toolsNode, _ := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
        Tools: []tool.BaseTool{userInfoTool},
    })

    const (
        nodeKeyOfTemplate  = "template"
        nodeKeyOfChatModel = "chat_model"
        nodeKeyOfTools     = "tools"
    )

    // 6. 创建图
    g := compose.NewGraph[map[string]any, []*schema.Message]()

    // 7. 添加节点
    _ = g.AddChatTemplateNode(nodeKeyOfTemplate, chatTpl)
    _ = g.AddChatModelNode(nodeKeyOfChatModel, chatModel)
    _ = g.AddToolsNode(nodeKeyOfTools, toolsNode)

    // 8. 定义边
    _ = g.AddEdge(compose.START, nodeKeyOfTemplate)
    _ = g.AddEdge(nodeKeyOfTemplate, nodeKeyOfChatModel)
    _ = g.AddEdge(nodeKeyOfChatModel, nodeKeyOfTools)
    _ = g.AddEdge(nodeKeyOfTools, compose.END)

    // 9. 编译
    r, _ := g.Compile(ctx)

    // 10. 执行
    out, _ := r.Invoke(ctx, map[string]any{
        "message_histories": []*schema.Message{},
        "user_query":        "我叫 zhangsan, 邮箱是 zhangsan@bytedance.com, 帮我推荐一处房产",
    })

    for _, msg := range out {
        fmt.Println(msg)
    }
}
```

---

### 4. 带分支的 Graph - 条件执行

来自官方示例 `adk/helloworld`，展示如何根据工具调用决定执行路径：

```go
// 定义状态
type myState struct {
    history []*schema.Message
}

func composeGraph[I, O any](
    ctx context.Context,
    tpl prompt.ChatTemplate,
    cm model.ToolCallingChatModel,
    tn *compose.ToolsNode,
    store compose.CheckPointStore,
) (compose.Runnable[I, O], error) {
    g := compose.NewGraph[I, O](
        compose.WithGenLocalState(func(ctx context.Context) *myState {
            return &myState{}
        }),
    )

    // 添加节点
    _ = g.AddChatTemplateNode("ChatTemplate", tpl)

    _ = g.AddChatModelNode("ChatModel", cm,
        compose.WithStatePreHandler(func(ctx context.Context, in []*schema.Message, state *myState) ([]*schema.Message, error) {
            state.history = append(state.history, in...)
            return state.history, nil
        }),
        compose.WithStatePostHandler(func(ctx context.Context, out *schema.Message, state *myState) (*schema.Message, error) {
            state.history = append(state.history, out)
            return out, nil
        }),
    )

    _ = g.AddToolsNode("ToolsNode", tn,
        compose.WithStatePreHandler(func(ctx context.Context, in *schema.Message, state *myState) (*schema.Message, error) {
            return state.history[len(state.history)-1], nil
        }),
    )

    // 定义边
    _ = g.AddEdge(compose.START, "ChatTemplate")
    _ = g.AddEdge("ChatTemplate", "ChatModel")
    _ = g.AddEdge("ToolsNode", "ChatModel") // 工具调用后返回模型继续对话

    // 条件分支：根据是否有工具调用决定执行路径
    _ = g.AddBranch("ChatModel", compose.NewGraphBranch(
        func(ctx context.Context, in *schema.Message) (string, error) {
            if len(in.ToolCalls) > 0 {
                return "ToolsNode", nil
            }
            return compose.END, nil
        },
        map[string]bool{"ToolsNode": true, compose.END: true},
    ))

    // 编译（带 CheckpointStore 和中断点）
    return g.Compile(
        ctx,
        compose.WithCheckPointStore(store),
        compose.WithInterruptBeforeNodes([]string{"ToolsNode"}),
    )
}
```

---

### 5. 带 Interrupt/Resume 的 Graph - 人机协作

来自官方示例 `compose/graph/react_with_interrupt`：

```go
// 编译时配置中断点
runner, err := composeGraph(
    ctx,
    newChatTemplate(ctx),
    newChatModel(ctx),
    newToolsNode(ctx),
    newCheckPointStore(ctx),
)
if err != nil {
    log.Fatal(err)
}

var history []*schema.Message

for {
    // 执行（带状态修改）
    result, err := runner.Invoke(ctx, map[string]any{
        "name":     "Megumin",
        "location": "Beijing",
    }, compose.WithCheckPointID("1"), compose.WithStateModifier(
        func(ctx context.Context, path compose.NodePath, state any) error {
            state.(*myState).history = history
            return nil
        },
    ))

    if err == nil {
        fmt.Printf("final result: %s", result.Content)
        break
    }

    // 提取中断信息
    info, ok := compose.ExtractInterruptInfo(err)
    if !ok {
        log.Fatal(err)
    }

    // 获取中断时的状态
    history = info.State.(*myState).history

    // 人工审核工具调用参数
    for i, tc := range history[len(history)-1].ToolCalls {
        fmt.Printf("will call tool: %s, arguments: %s\n", tc.Function.Name, tc.Function.Arguments)
        fmt.Print("Are the arguments as expected? (y/n): ")
        var response string
        fmt.Scanln(&response)

        if strings.ToLower(response) == "n" {
            fmt.Print("Please enter the modified arguments: ")
            scanner := bufio.NewScanner(os.Stdin)
            var newArguments string
            if scanner.Scan() {
                newArguments = scanner.Text()
            }

            // 更新工具调用参数
            history[len(history)-1].ToolCalls[i].Function.Arguments = newArguments
            fmt.Printf("Updated arguments to: %s\n", newArguments)
        }
    }
}
```

---

### 6. 多节点 Graph - RAG 模式

```go
type RAGState struct {
    Query     string
    Documents []*schema.Document
    Answer    string
}

func BuildRAG(ctx context.Context) (compose.Runnable[string, string], error) {
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
        compose.WithNodeName("文档检索"),
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
        compose.WithNodeName("RAG Prompt"),
    )

    // 生成节点
    _ = g.AddChatModelNode("generate", chatModel, compose.WithNodeName("LLM 生成"))

    // 提取节点
    _ = g.AddLambdaNode("extract", compose.InvokableLambda(
        func(ctx context.Context, msg *schema.Message) (string, error) {
            return msg.Content, nil
        }),
        compose.WithStatePostHandler(func(ctx context.Context, answer string, state *RAGState) (string, error) {
            state.Answer = answer
            return answer, nil
        }),
        compose.WithNodeName("答案提取"),
    )

    // 边
    _ = g.AddEdge(compose.START, "retrieve")
    _ = g.AddEdge("retrieve", "prompt")
    _ = g.AddEdge("prompt", "generate")
    _ = g.AddEdge("generate", "extract")
    _ = g.AddEdge("extract", compose.END)

    return g.Compile(ctx, compose.WithGraphName("RAGPipeline"))
}
```

---

### 7. 异步节点 Graph

来自官方示例 `compose/graph/async_node`：

```go
// 异步节点适用于长时间运行的任务
_ = g.AddLambdaNode("async_report", compose.InvokableLambda(
    func(ctx context.Context, input ReportInput) (*ReportOutput, error) {
        // 模拟长时间运行的任务
        time.Sleep(5 * time.Second)
        return &ReportOutput{Content: generateReport(input)}, nil
    }),
    compose.WithNodeName("异步报告生成"),
)
```

---

## 节点类型

| 方法 | 说明 | 输入/输出类型 |
|------|------|---------------|
| `AddChatTemplateNode` | 添加 Prompt 模板节点 | `map[string]any` → `[]*schema.Message` |
| `AddChatModelNode` | 添加 ChatModel 节点 | `[]*schema.Message` → `*schema.Message` |
| `AddToolsNode` | 添加工具执行节点 | `*schema.Message` → `[]*schema.Message` |
| `AddLambdaNode` | 添加自定义 Lambda 节点 | 任意类型 |
| `AddRetrieverNode` | 添加检索器节点 | `string` → `[]*schema.Document` |
| `AddEmbeddingNode` | 添加 Embedding 节点 | `[]string` → `[][]float64` |
| `AddGraphNode` | 添加子图节点 | 任意类型 |

---

## 边和分支

### 简单边

```go
g.AddEdge("node1", "node2")
```

### 字段映射边

```go
// 单字段映射
g.AddEdge("node1", "node2", compose.MapFields("output.field", "input.field"))

// 多字段映射
g.AddEdge("node1", "node2",
    compose.MapFields("a", "x"),
    compose.MapFields("b", "y"),
)
```

### 条件分支

```go
_ = g.AddBranch("ChatModel", compose.NewGraphBranch(
    func(ctx context.Context, in *schema.Message) (string, error) {
        if len(in.ToolCalls) > 0 {
            return "ToolsNode", nil
        }
        return compose.END, nil
    },
    map[string]bool{
        "ToolsNode": true,
        compose.END: true,
    },
))
```

---

## 编译选项

```go
runnable, err := g.Compile(ctx,
    compose.WithGraphName("MyGraph"),           // 图名称
    compose.WithMaxRunSteps(20),                // 最大执行步数
    compose.WithNodeTriggerMode(compose.AllPredecessor), // 或 AnyPredecessor
    compose.WithCallbacks(handler),             // Callback 处理器
    compose.WithCheckPointStore(store),         // Checkpoint 存储
    compose.WithInterruptBeforeNodes([]string{"ToolsNode"}), // 中断点
    compose.WithGraphCompileCallbacks(visualize.NewMermaidGenerator("MyGraph")), // 可视化
)
```

---

## 状态处理器

### StatePreHandler - 节点执行前调用

```go
compose.WithStatePreHandler(func(ctx context.Context, in InputType, state *AppState) (InputType, error) {
    // 修改输入或更新状态
    state.Count++
    return in, nil
})
```

### StatePostHandler - 节点执行后调用

```go
compose.WithStatePostHandler(func(ctx context.Context, out OutputType, state *AppState) (OutputType, error) {
    // 修改输出或更新状态
    state.History = append(state.History, out)
    return out, nil
})
```

### StateModifier - 运行时修改状态

```go
result, err := runnable.Invoke(ctx, input,
    compose.WithStateModifier(func(ctx context.Context, path compose.NodePath, state any) error {
        state.(*myState).history = newHistory
        return nil
    }),
)
```

---

## CheckpointStore 实现

### 内存实现（测试用）

```go
type MemoryStore struct {
    buf map[string][]byte
}

func (m *MemoryStore) Get(ctx context.Context, checkPointID string) ([]byte, bool, error) {
    data, ok := m.buf[checkPointID]
    return data, ok, nil
}

func (m *MemoryStore) Set(ctx context.Context, checkPointID string, checkPoint []byte) error {
    m.buf[checkPointID] = checkPoint
    return nil
}
```

### Redis 实现（生产用）

```go
type RedisStore struct {
    client *redis.Client
}

func (r *RedisStore) Get(ctx context.Context, checkPointID string) ([]byte, bool, error) {
    data, err := r.client.Get(ctx, checkPointID).Bytes()
    if err == redis.Nil {
        return nil, false, nil
    }
    if err != nil {
        return nil, false, err
    }
    return data, true, nil
}

func (r *RedisStore) Set(ctx context.Context, checkPointID string, checkPoint []byte) error {
    return r.client.Set(ctx, checkPointID, checkPoint, 24*time.Hour).Err()
}
```

---

## 调试技巧

### 可视化 Graph

```go
import "github.com/cloudwego/eino-ext/devops/visualize"

runnable, err := g.Compile(ctx,
    compose.WithGraphCompileCallbacks(visualize.NewMermaidGenerator(
        "MyGraph",
        func(mermaidStr string) {
            fmt.Println(mermaidStr)
        },
    )),
)
```

### 使用 Callback 追踪执行

```go
handler := callbacks.NewHandlerBuilder().
    OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
        log.Printf("开始节点：%s (类型：%s)", info.Name, info.Type)
        return ctx
    }).
    OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
        log.Printf("完成节点：%s", info.Name)
        return ctx
    }).
    Build()

callbacks.AppendGlobalHandlers(handler)
```

---

## 最佳实践

### 1. 定义清晰的状态结构

```go
type AppState struct {
    History   []*schema.Message
    Query     string
    Documents []*schema.Document
    Answer    string
}
```

### 2. 使用节点名称标识

```go
_ = g.AddChatModelNode("ChatModel", chatModel, compose.WithNodeName("客服模型"))
```

### 3. 在分支中使用 Map 明确合法目标

```go
map[string]bool{
    "ToolsNode": true,
    compose.END: true,
}
```

### 4. 合理使用 StatePreHandler 和 StatePostHandler

```go
// 前置处理：保存输入到历史
compose.WithStatePreHandler(func(ctx context.Context, in []*schema.Message, state *myState) ([]*schema.Message, error) {
    state.history = append(state.history, in...)
    return state.history, nil
})

// 后置处理：保存输出到历史
compose.WithStatePostHandler(func(ctx context.Context, out *schema.Message, state *myState) (*schema.Message, error) {
    state.history = append(state.history, out)
    return out, nil
})
```

### 5. 使用 CheckpointStore 支持中断恢复

```go
store := &MemoryStore{buf: make(map[string][]byte)}
runnable, _ := g.Compile(ctx,
    compose.WithCheckPointStore(store),
    compose.WithInterruptBeforeNodes([]string{"ToolsNode"}),
)
```

---

## 相关文档

- [Chain 模式](chain.md) - 顺序编排
- [Interrupt/Resume](interrupt.md) - 中断恢复
- [Callback](callback.md) - 可观测性
- [RAG](rag.md) - 检索增强生成
