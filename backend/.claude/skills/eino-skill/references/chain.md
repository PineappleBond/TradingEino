# Eino Chain 模式详解

Chain 是 Eino 中最简单的顺序编排模式，适合线性流程。

## 基本概念

Chain 将多个节点按顺序连接，前一个节点的输出作为后一个节点的输入。

## 完整示例

### 1. 客服机器人 Chain

```go
package main

import (
    "context"
    "fmt"
    "github.com/cloudwego/eino/components/model/openai"
    "github.com/cloudwego/eino/components/prompt"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
)

// 定义输入类型
type CustomerQuery struct {
    OrderID   string
    Question  string
}

// 定义输出类型
type Response struct {
    Answer string
}

func main() {
    ctx := context.Background()

    // 1. 创建 ChatModel
    chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
        APIKey: "sk-your-api-key",
        Model:  "gpt-4o",
    })
    if err != nil {
        panic(err)
    }

    // 2. 创建 Prompt 模板
    template := prompt.FromMessages(
        schema.FString,
        &schema.Message{
            Role:    "system",
            Content: "你是一个客服助手，专门帮助用户查询订单状态。",
        },
        &schema.Message{
            Role:    "user",
            Content: "订单号：{order_id}\n问题：{question}",
        },
    )

    // 3. 创建 Chain
    chain := compose.NewChain[*CustomerQuery, *Response]()

    chain.
        // Prompt 格式化
        AppendChatTemplate(template).
        // 调用模型
        AppendChatModel(chatModel).
        // 转换为 Response 类型
        AppendLambda(compose.InvokableLambda(
            func(ctx context.Context, msg *schema.Message) (*Response, error) {
                return &Response{Answer: msg.Content}, nil
            },
        ))

    // 4. 编译
    runnable, err := chain.Compile(ctx)
    if err != nil {
        panic(err)
    }

    // 5. 执行
    result, err := runnable.Invoke(ctx, &CustomerQuery{
        OrderID:  "12345",
        Question: "我的订单什么时候发货？",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(result.Answer)
}
```

### 2. 带工具的 Chain

```go
chain := compose.NewChain[map[string]any, string]()

chain.
    AppendChatTemplate(chatTpl).
    AppendChatModel(model).
    AppendToolsNode(toolsNode).
    AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) (string, error) {
        return msg.Content, nil
    }))

runnable, err := chain.Compile(ctx)
```

## Chain 构建方法

| 方法 | 说明 |
|------|------|
| `AppendChatTemplate` | 添加 Prompt 模板节点 |
| `AppendChatModel` | 添加 ChatModel 节点 |
| `AppendToolsNode` | 添加工具执行节点 |
| `AppendLambda` | 添加自定义 Lambda 节点 |
| `AppendBranch` | 添加分支逻辑 |
| `AppendParallel` | 添加并行执行 |
| `AppendGraph` | 添加子图 |
| `AppendPassthrough` | 添加透传节点 |

## 编译选项

```go
runnable, err := chain.Compile(ctx,
    compose.WithGraphName("MyChain"),
    compose.WithMaxRunSteps(10),
    compose.WithCallbacks(myHandler),
)
```