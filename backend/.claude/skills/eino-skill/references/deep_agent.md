# Eino DeepAgent 详解

DeepAgent 是 Eino ADK 中的层级式多 Agent 协作框架，通过主 Agent 协调多个子 Agent 完成复杂任务。

## 核心概念

### 架构设计

```
┌─────────────────────────────────────┐
│         DeepAgent (主 Agent)         │
│  - 任务规划与分解                     │
│  - 子 Agent 调度                      │
│  - 结果整合                          │
└─────────────┬───────────────────────┘
              │
    ┌─────────┼─────────┐
    │         │         │
┌───▼───┐ ┌──▼───┐ ┌───▼───┐
│Agent A│ │Agent B│ │Agent C│
│(代码)  │ │(搜索) │ │(写作) │
└───────┘ └───────┘ └───────┘
```

### 与 SequentialAgent 的区别

| 特性 | DeepAgent | SequentialAgent |
|------|-----------|-----------------|
| 执行方式 | 动态调度，主 Agent 决定调用哪个子 Agent | 固定顺序执行 |
| 适用场景 | 复杂、不确定任务 | 流程固定的任务 |
| 灵活性 | 高 | 低 |

## 完整示例

### 1. DeepAgent 基础用法

以下示例来自官方 `adk/multiagent/deep`，展示如何创建一个 Excel 处理 DeepAgent：

```go
package main

import (
    "context"
    "github.com/cloudwego/eino/adk"
    "github.com/cloudwego/eino/adk/prebuilt/deep"
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/compose"
)

func newExcelAgent(ctx context.Context) (adk.Agent, error) {
    // 1. 创建主模型
    cm, err := utils.NewChatModel(ctx,
        utils.WithMaxTokens(4096),
        utils.WithTemperature(float32(0)),
        utils.WithTopP(float32(0)),
    )
    if err != nil {
        return nil, err
    }

    // 2. 创建子 Agent
    ca, err := agents.NewCodeAgent(ctx, operator)  // 代码 Agent
    if err != nil {
        return nil, err
    }
    wa, err := agents.NewWebSearchAgent(ctx)  // 搜索 Agent
    if err != nil {
        return nil, err
    }

    // 3. 创建工具
    readTool := tools.NewWrapTool(tools.NewReadFileTool(operator), nil, nil)
    treeTool := tools.NewWrapTool(tools.NewTreeTool(operator), nil, nil)

    // 4. 创建 DeepAgent
    deepAgent, err := deep.New(ctx, &deep.Config{
        Name:        "ExcelAgent",
        Description: "an agent for excel task",
        ChatModel:   cm,
        SubAgents:   []adk.Agent{ca, wa},
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: []tool.BaseTool{
                    readTool,
                    treeTool,
                },
            },
        },
        MaxIteration: 100,
    })
    if err != nil {
        return nil, err
    }

    return deepAgent, nil
}
```

### 2. CodeAgent 实现（子 Agent 示例）

```go
func NewCodeAgent(ctx context.Context, operator commandline.Operator) (adk.Agent, error) {
    // 创建 ChatModel
    cm, err := utils.NewChatModel(ctx,
        utils.WithMaxTokens(14125),
        utils.WithTemperature(float32(1)),
        utils.WithTopP(float32(1)),
    )
    if err != nil {
        return nil, err
    }

    // 定义工具预处理和后处理
    preprocess := []tools.ToolRequestPreprocess{tools.ToolRequestRepairJSON}

    // 创建 ChatModelAgent 作为子 Agent
    ca, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
        Name: "CodeAgent",
        Description: `This sub-agent is a code agent specialized in handling Excel files.
It receives a clear task and accomplish the task by generating Python code and execute it.`,
        Instruction: `You are a code agent. Your workflow is as follows:
1. You will be given a clear task to handle Excel files.
2. You should analyse the task and use right tools to help coding.
3. You should write python code to finish the task.
4. You are preferred to write code execution result to another file for further usages.

You are in a react mode, and you should use the following libraries:
- pandas: for data analysis and manipulation
- matplotlib: for plotting and visualization
- openpyxl: for reading and writing Excel files
`,
        Model: cm,
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: []tool.BaseTool{
                    tools.NewWrapTool(tools.NewBashTool(operator), preprocess, nil),
                    tools.NewWrapTool(tools.NewTreeTool(operator), preprocess, nil),
                    tools.NewWrapTool(tools.NewEditFileTool(operator), preprocess, nil),
                    tools.NewWrapTool(tools.NewReadFileTool(operator), preprocess, nil),
                    tools.NewWrapTool(tools.NewPythonRunnerTool(operator), preprocess, nil),
                },
            },
        },
        // 自定义输入生成
        GenModelInput: func(ctx context.Context, instruction string, input *adk.AgentInput) ([]adk.Message, error) {
            wd, ok := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey)
            if !ok {
                return nil, fmt.Errorf("work dir not found")
            }

            tpl := prompt.FromMessages(schema.Jinja2,
                schema.SystemMessage(instruction),
                schema.UserMessage(`WorkingDirectory: {{ working_dir }}
UserQuery: {{ user_query }}
CurrentTime: {{ current_time }}
`))

            return tpl.Format(ctx, map[string]any{
                "working_dir":  wd,
                "user_query":   utils.FormatInput(input.Messages),
                "current_time": utils.GetCurrentTime(),
            })
        },
        MaxIterations: 1000,
    })
    if err != nil {
        return nil, err
    }

    return ca, nil
}
```

### 3. WebSearchAgent 实现（子 Agent 示例）

```go
func NewWebSearchAgent(ctx context.Context) (adk.Agent, error) {
    cm, err := utils.NewChatModel(ctx)
    if err != nil {
        return nil, err
    }

    // 使用 DuckDuckGo 搜索工具
    searchTool, err := duckduckgo.NewTextSearchTool(ctx, &duckduckgo.Config{})
    if err != nil {
        return nil, err
    }

    return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
        Name:        "WebSearchAgent",
        Description: "WebSearchAgent utilizes the ReAct model to analyze input information and accomplish tasks using web search tools.",
        Model:       cm,
        ToolsConfig: adk.ToolsConfig{
            ToolsNodeConfig: compose.ToolsNodeConfig{
                Tools: []tool.BaseTool{searchTool},
            },
        },
        MaxIterations: 10,
    })
}
```

### 4. 运行 DeepAgent

```go
func main() {
    ctx := context.Background()

    // 创建 DeepAgent
    agent, err := newExcelAgent(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // 创建 Runner
    runner := adk.NewRunner(ctx, adk.RunnerConfig{
        Agent:           agent,
        EnableStreaming: true,
    })

    // 执行查询
    query := schema.UserMessage("请帮我将 questions.csv 表格中的第一列提取到一个新的 csv 中")
    iter := runner.Run(ctx, []*schema.Message{query})

    // 处理流式输出
    for {
        event, ok := iter.Next()
        if !ok {
            break
        }
        if event.Output != nil && event.Output.MessageOutput != nil {
            if event.Output.MessageOutput.IsStreaming {
                // 处理流式消息
            } else {
                // 处理完整消息
                fmt.Println(event.Output.MessageOutput.Message.Content)
            }
        }
    }
}
```

## DeepAgent 配置项

```go
type Config struct {
    Name        string          // Agent 名称
    Description string          // Agent 描述（用于主 Agent 理解何时调用此子 Agent）
    ChatModel   ChatModel       // 主模型
    SubAgents   []adk.Agent     // 子 Agent 列表
    ToolsConfig ToolsConfig     // 工具配置
    MaxIteration int            // 最大迭代次数
}
```

## 子 Agent 类型

### 1. ChatModelAgent（ReAct 模式）

最常用的子 Agent 类型，支持：
- 工具调用
- 多轮思考
- 流式输出

```go
agent := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Name:        "MyAgent",
    Description: "Agent 的描述，说明其职责",
    Model:       chatModel,
    Instruction: "详细的指令，定义 Agent 的行为",
    ToolsConfig: adk.ToolsConfig{
        ToolsNodeConfig: compose.ToolsNodeConfig{
            Tools: []tool.BaseTool{tool1, tool2},
        },
    },
    MaxIterations: 10,
})
```

### 2. 自定义 Agent

实现 `adk.Agent` 接口：

```go
type MyAgent struct {
    // 自定义字段
}

func (a *MyAgent) Run(ctx context.Context, input []*schema.Message, opts ...adk.AgentOption) (*schema.StreamReader[adk.AgentEvent], error) {
    // 自定义实现
}
```

## 工具包装

DeepAgent 支持工具包装，添加预处理和后处理：

```go
// 预处理（修复 JSON 格式等）
preprocess := []tools.ToolRequestPreprocess{tools.ToolRequestRepairJSON}

// 后处理（文件操作、格式整理等）
postprocess := []tools.ToolResponsePostprocess{tools.FilePostProcess}

// 包装工具
wrapTool := tools.NewWrapTool(baseTool, preprocess, postprocess)
```

## 上下文参数传递

```go
// 初始化上下文参数
ctx = params.InitContextParams(ctx)

// 添加参数
params.AppendContextParams(ctx, map[string]interface{}{
    params.WorkDirSessionKey:  workdir,
    params.FilePathSessionKey: inputFileDir,
    params.TaskIDKey:          id,
})

// 获取参数
wd, ok := params.GetTypedContextParams[string](ctx, params.WorkDirSessionKey)
```

## 最佳实践

### 1. 子 Agent 职责划分

- 每个子 Agent 负责一个明确的领域
- Description 要清晰，让主 Agent 知道何时调用

### 2. 工具设计

- 工具应原子化，职责单一
- 使用 WrapTool 添加预处理和后处理
- 错误处理要完善

### 3. 迭代次数控制

- 设置合理的 MaxIterations 防止无限循环
- 主 Agent 和子 Agent 可以有不同的迭代限制

### 4. 流式处理

```go
iter := runner.Run(ctx, messages)
for {
    event, ok := iter.Next()
    if !ok {
        break
    }
    // 实时处理事件
}
```

## 常见用例

### 用例 1：数据分析 Agent

- **主 Agent**: 任务协调
- **子 Agent**:
  - CodeAgent（Python 数据分析）
  - FileAgent（文件读写）

### 用例 2：研究写作 Agent

- **主 Agent**: 研究规划
- **子 Agent**:
  - SearchAgent（资料搜索）
  - ReadAgent（文献阅读）
  - WriteAgent（文章写作）

### 用例 3：客服 Agent

- **主 Agent**: 客服路由
- **子 Agent**:
  - QueryAgent（订单查询）
  - RefundAgent（退款处理）
  - FAQAgent（常见问题）
