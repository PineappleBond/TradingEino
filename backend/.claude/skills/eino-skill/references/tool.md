# Eino Tool 工具系统详解

Tool 是 Eino 中让 LLM 能够调用外部函数/API 的机制。

## Tool 接口

### 基础接口

```go
// BaseTool - 仅提供元数据（用于模型工具选择）
type BaseTool interface {
    Info(ctx context.Context) (*schema.ToolInfo, error)
}

// InvokableTool - 可执行工具（JSON 字符串输入输出）
type InvokableTool interface {
    BaseTool
    InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error)
}

// StreamableTool - 流式工具
type StreamableTool interface {
    BaseTool
    StreamableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (*schema.StreamReader[string], error)
}
```

---

## 创建 Tool 的方式

### 方式 1：使用 InferTool（推荐）

来自官方示例 `components/tool/utils`：

```go
package main

import (
    "context"
    "github.com/cloudwego/eino/utils/tool"
)

// 定义输入结构
type WeatherInput struct {
    City string `json:"city" jsonschema:"required,description=城市名称"`
    Date string `json:"date,omitempty" jsonschema:"description=日期，格式 YYYY-MM-DD"`
}

// 定义输出结构
type WeatherOutput struct {
    Temperature int    `json:"temperature"`
    Condition   string `json:"condition"`
}

func main() {
    ctx := context.Background()

    // 快速创建工具
    weatherTool, err := tool.InferTool(
        "query_weather",                          // 工具名称
        "查询指定城市的天气信息",                   // 工具描述
        func(ctx context.Context, input WeatherInput) (*WeatherOutput, error) {
            // 实现逻辑
            return &WeatherOutput{
                Temperature: 25,
                Condition:   "晴",
            }, nil
        },
    )
    if err != nil {
        panic(err)
    }

    // 获取工具信息（用于模型绑定）
    info, _ := weatherTool.Info(ctx)
    fmt.Printf("Tool: %s, Desc: %s\n", info.Name, info.Desc)
}
```

### 方式 2：使用 NewTool 手动创建

来自官方示例 `compose/graph/tool_call_agent`：

```go
import "github.com/cloudwego/eino/components/tool/utils"

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
```

### 方式 3：手动实现 InvokableTool

```go
type WeatherTool struct {
    apiKey string
}

// 实现 Info 方法
func (t *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "query_weather",
        Desc: "查询指定城市的天气信息",
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            "city": {
                Type:     "string",
                Desc:     "城市名称",
                Required: true,
            },
            "date": {
                Type:     "string",
                Desc:     "日期，格式 YYYY-MM-DD",
                Required: false,
            },
        }),
    }, nil
}

// 实现 InvokableRun 方法
func (t *WeatherTool) InvokableRun(ctx context.Context, argsJSON string, opts ...tool.Option) (string, error) {
    // 1. 解析 JSON 输入
    var input struct {
        City string `json:"city"`
        Date string `json:"date"`
    }
    if err := json.Unmarshal([]byte(argsJSON), &input); err != nil {
        return "", err
    }

    // 2. 执行逻辑
    weather, err := t.queryWeather(input.City, input.Date)
    if err != nil {
        return "", err
    }

    // 3. 返回 JSON 结果
    resultJSON, err := json.Marshal(weather)
    if err != nil {
        return "", err
    }
    return string(resultJSON), nil
}

func (t *WeatherTool) queryWeather(city, date string) (*WeatherOutput, error) {
    // 实际查询逻辑
    return &WeatherOutput{Temperature: 25, Condition: "晴"}, nil
}
```

---

## 在 Chain/Graph 中使用 Tool

### 1. 作为 ToolsNode 添加到 Graph

来自官方示例 `compose/graph/tool_call_agent`：

```go
// 创建工具
orderTool, _ := tool.InferTool("query_order", "查询订单状态", queryOrderFunc)

// 创建 ToolsNode
toolsNode, _ := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
    Tools: []tool.BaseTool{orderTool},
})

// 添加到 Graph
g := compose.NewGraph[[]*schema.Message, []*schema.Message]()
_ = g.AddToolsNode("tools", toolsNode)
```

### 2. 绑定到 ChatModel（Tool Calling）

来自官方示例 `compose/graph/tool_call_agent`：

```go
chatModel, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{
    APIKey: apiKey,
    Model:  "gpt-4o",
})

// 获取工具信息
info, _ := userInfoTool.Info(ctx)

// 绑定工具到 ChatModel（强制使用工具）
_ = chatModel.BindForcedTools([]*schema.ToolInfo{info})

// 或者使用 WithTools（可选使用工具）
// toolCallingModel, err := chatModel.(model.ToolCallingChatModel).WithTools([]*schema.ToolInfo{info})
```

---

## 完整示例：客服订单查询机器人

来自官方示例，整合了多个模式：

```go
package main

import (
    "context"
    "fmt"
    "github.com/cloudwego/eino-ext/components/model/openai"
    "github.com/cloudwego/eino/components/prompt"
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/components/tool/utils"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
)

// 订单查询输入
type OrderQueryInput struct {
    OrderID string `json:"order_id" jsonschema:"required,description=订单号"`
}

// 订单信息
type OrderInfo struct {
    OrderID string   `json:"order_id"`
    Status  string   `json:"status"`
    ETA     string   `json:"eta"`
    Items   []string `json:"items"`
}

// 查询订单函数
func queryOrder(ctx context.Context, input OrderQueryInput) (*OrderInfo, error) {
    // 模拟查询数据库
    return &OrderInfo{
        OrderID: input.OrderID,
        Status:  "已发货",
        ETA:     "2024-03-26",
        Items:   []string{"商品 A", "商品 B"},
    }, nil
}

func main() {
    ctx := context.Background()

    // 1. 创建工具
    orderTool, err := tool.InferTool(
        "query_order",
        "查询订单状态和物流信息",
        queryOrder,
    )
    if err != nil {
        panic(err)
    }

    // 2. 创建 ChatModel（带工具调用）
    chatModel, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{
        APIKey: "sk-key",
        Model:  "gpt-4o",
    })

    // 绑定工具
    info, _ := orderTool.Info(ctx)
    _ = chatModel.BindForcedTools([]*schema.ToolInfo{info})

    // 3. 创建 Prompt 模板
    template := prompt.FromMessages(
        schema.FString,
        &schema.Message{
            Role:    "system",
            Content: "你是一个客服助手，使用可用的工具帮助用户查询订单信息。",
        },
        &schema.Message{
            Role:    "user",
            Content: "{query}",
        },
    )

    // 4. 创建 ToolsNode
    toolsNode, _ := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
        Tools: []tool.BaseTool{orderTool},
    })

    // 5. 创建 Graph
    g := compose.NewGraph[map[string]string, []*schema.Message]()

    _ = g.AddChatTemplateNode("prompt", template)
    _ = g.AddChatModelNode("model", chatModel, compose.WithNodeName("客服模型"))
    _ = g.AddToolsNode("tools", toolsNode)

    // 边
    _ = g.AddEdge(compose.START, "prompt")
    _ = g.AddEdge("prompt", "model")
    _ = g.AddEdge("model", "tools")
    _ = g.AddEdge("tools", compose.END)

    // 编译
    runnable, _ := g.Compile(ctx, compose.WithGraphName("CustomerServiceBot"))

    // 执行
    result, _ := runnable.Invoke(ctx, map[string]string{
        "query": "帮我查一下订单 12345 的状态",
    })

    for _, msg := range result {
        fmt.Println(msg)
    }
}
```

---

## Tool Schema 注解

使用 `jsonschema` tag 定义参数：

```go
type Input struct {
    // 必填字段
    City string `json:"city" jsonschema:"required,description=城市名称"`

    // 选填字段
    Date string `json:"date,omitempty" jsonschema:"description=日期"`

    // 枚举字段
    Priority int `json:"priority" jsonschema:"enum=1,enum=2,enum=3"`

    // 数组字段
    Tags []string `json:"tags" jsonschema:"description=标签列表"`

    // 嵌套结构
    Location Location `json:"location"`
}

type Location struct {
    Lat float64 `json:"lat" jsonschema:"description=纬度"`
    Lng float64 `json:"lng" jsonschema:"description=经度"`
}
```

---

## 工具包装 - 预处理和后处理

来自官方示例 `components/tool/middlewares`：

### 错误移除中间件

```go
import "github.com/cloudwego/eino/components/tool/middlewares"

// 将工具错误转换为友好提示
errorRemover := middlewares.NewErrorRemover()

wrapTool := tools.NewWrapTool(
    baseTool,
    nil, // 预处理
    []tools.ToolResponsePostprocess{errorRemover.PostProcess},
)
```

### JSON 修复中间件

```go
import "github.com/cloudwego/eino/components/tool/middlewares"

// 修复 LLM 生成的格式错误的 JSON
jsonFixer := middlewares.NewJSONFixer()

wrapTool := tools.NewWrapTool(
    baseTool,
    []tools.ToolRequestPreprocess{jsonFixer.PreProcess},
    nil, // 后处理
)
```

### 自定义预处理和后处理

```go
// 预处理：修复 JSON 格式
preprocess := []tools.ToolRequestPreprocess{
    func(ctx context.Context, toolName string, args string) (string, error) {
        // 修复 JSON 格式
        return fixJSON(args), nil
    },
}

// 后处理：文件操作
postprocess := []tools.ToolResponsePostprocess{
    func(ctx context.Context, toolName string, result string) (string, error) {
        // 整理结果格式
        return formatResult(result), nil
    },
}

wrapTool := tools.NewWrapTool(baseTool, preprocess, postprocess)
```

---

## MCP 工具集成

来自官方示例 `components/tool/mcptool`：

```go
import "github.com/cloudwego/eino-ext/components/tool/mcp"

// 创建 MCP 客户端
client, _ := mcp.NewClient(ctx, &mcp.Config{
    ServerURL: "http://localhost:8080/mcp",
})

// 创建 MCP 工具
mcpTool, _ := mcp.NewTool(client, "mcp_tool_name")

// 添加到 ToolsNode
toolsNode, _ := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
    Tools: []tool.BaseTool{mcpTool},
})
```

---

## 常见错误处理

```go
func queryOrder(ctx context.Context, input OrderQueryInput) (*OrderInfo, error) {
    // 参数验证
    if input.OrderID == "" {
        return nil, fmt.Errorf("订单号不能为空")
    }
    if len(input.OrderID) != 6 {
        return nil, fmt.Errorf("订单号必须是 6 位")
    }

    // 业务逻辑
    order, err := db.GetOrder(ctx, input.OrderID)
    if err != nil {
        if errors.Is(err, db.ErrNotFound) {
            return nil, fmt.Errorf("订单不存在")
        }
        return nil, fmt.Errorf("查询失败：%w", err)
    }

    return &OrderInfo{
        OrderID: order.ID,
        Status:  order.Status,
        ETA:     order.ETA,
        Items:   order.Items,
    }, nil
}
```

---

## 最佳实践

### 1. 工具命名

- 使用有意义的名称，如 `query_order`、`get_weather`
- 名称应该清晰描述工具的功能

### 2. 工具描述

- 描述应该清晰说明工具的用途
- 包含参数说明和使用场景

### 3. 参数验证

```go
func validateInput(input *WeatherInput) error {
    if input.City == "" {
        return fmt.Errorf("城市不能为空")
    }
    if len(input.City) > 100 {
        return fmt.Errorf("城市名称过长")
    }
    return nil
}
```

### 4. 错误处理

- 返回友好的错误消息
- 使用包装错误提供上下文

### 5. 工具组合

```go
// 组合多个工具
tools := []tool.BaseTool{
    weatherTool,
    orderTool,
    userInfoTool,
}

toolsNode, _ := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
    Tools: tools,
})
```

---

## 相关文档

- [Chain 模式](chain.md) - 顺序编排
- [Graph 模式](graph.md) - DAG 编排
- [Callback](callback.md) - 可观测性
