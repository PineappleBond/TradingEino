# Eino Callback 详解

Callback 系统用于在应用执行过程中注入横切关注点，如日志、追踪、监控等。

## Callback 时机

| 时机 | 说明 | 触发条件 |
|------|------|----------|
| `OnStart` | 节点执行开始 | 非流式输入 |
| `OnEnd` | 节点执行成功 | 非流式输出 |
| `OnError` | 节点执行出错 | 任何错误 |
| `OnStartWithStreamInput` | 流式输入开始 | 流式输入 |
| `OnEndWithStreamOutput` | 流式输出结束 | 流式输出 |

## 完整示例

### 1. 基础 Callback（记录执行时间）

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/cloudwego/eino/callbacks"
    "github.com/cloudwego/eino/compose"
)

// 定义状态 key
type startTimeKey struct{}

func main() {
    // 创建 Handler
    handler := callbacks.NewHandlerBuilder().
        // OnStart: 记录开始时间和输入
        OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
            fmt.Printf("[%s] 开始执行节点：%s (类型：%s)\n",
                time.Now().Format("15:04:05"), info.Name, info.Type)
            fmt.Printf("  输入：%v\n", input)

            // 保存开始时间到 context
            return context.WithValue(ctx, startTimeKey{}, time.Now())
        }).
        // OnEnd: 记录执行时间和输出
        OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
            start, ok := ctx.Value(startTimeKey{}).(time.Time)
            if ok {
                fmt.Printf("[%s] 完成执行，耗时：%v\n",
                    time.Now().Format("15:04:05"), time.Since(start))
            }
            fmt.Printf("  输出：%v\n", output)
            return ctx
        }).
        // OnError: 记录错误
        OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
            fmt.Printf("[%s] 执行出错：%v\n", time.Now().Format("15:04:05"), err)
            return ctx
        }).
        Build()

    // 注册全局回调
    callbacks.AppendGlobalHandlers(handler)

    // 使用图...
    g := compose.NewGraph[string, string]()
    // ...
}
```

### 2. 流式 Callback

```go
handler := callbacks.NewHandlerBuilder().
    OnEndWithStreamOutputFn(func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
        defer output.Close() // 必须关闭流

        fmt.Printf("[%s] 流式输出开始（节点：%s）\n", time.Now().Format("15:04:05"), info.Name)

        totalTokens := 0
        for {
            chunk, err := output.Recv()
            if err == io.EOF {
                break
            }
            if err != nil {
                fmt.Printf("流读取错误：%v\n", err)
                break
            }
            // 处理流式块
            totalTokens++
            fmt.Printf("  收到块：%v\n", chunk)
        }

        fmt.Printf("[%s] 流式输出完成，总块数：%d\n", time.Now().Format("15:04:05"), totalTokens)
        return ctx
    }).
    Build()
```

### 3. Langfuse 集成（可观测性平台）

```go
import (
    "github.com/cloudwego/eino-ext/callbacks/langfuse"
    "github.com/cloudwego/eino/callbacks"
)

func initLangfuse() callbacks.Handler {
    handler, _ := langfuse.NewLangfuseHandler(&langfuse.Config{
        Host:        "https://cloud.langfuse.com",
        PublicKey:   "pk-lf-...",
        SecretKey:   "sk-lf-...",
        ProjectID:   "your-project-id",
        EnableTrace: true,
    })
    return handler
}

func main() {
    // 注册 Langfuse
    callbacks.AppendGlobalHandlers(initLangfuse())

    // 应用代码...
}
```

### 4. 自定义 Callback（日志 + 指标）

```go
type MetricsCallback struct {
    metricsChan chan<- Metric
}

type Metric struct {
    NodeName   string
    Duration   time.Duration
    InputLen   int
    OutputLen  int
    IsError    bool
}

func (c *MetricsCallback) Handler() callbacks.Handler {
    return callbacks.NewHandlerBuilder().
        OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
            return context.WithValue(ctx, startTimeKey{}, time.Now())
        }).
        OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
            start := ctx.Value(startTimeKey{}).(time.Time)
            c.metricsChan <- Metric{
                NodeName:  info.Name,
                Duration:  time.Since(start),
                InputLen:  len(fmt.Sprintf("%v", input)),
                OutputLen: len(fmt.Sprintf("%v", output)),
                IsError:   false,
            }
            return ctx
        }).
        OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
            start := ctx.Value(startTimeKey{}).(time.Time)
            c.metricsChan <- Metric{
                NodeName:  info.Name,
                Duration:  time.Since(start),
                IsError:   true,
            }
            return ctx
        }).
        Build()
}
```

## Callback 注册方式

### 1. 全局注册（应用启动时）

```go
// 非线程安全，在应用启动时调用一次
callbacks.AppendGlobalHandlers(handler1, handler2, ...)
```

### 2. 单次调用注册

```go
// 为特定调用添加回调
result, err := runnable.Invoke(ctx, input,
    compose.WithCallbacks(myHandler),
)
```

### 3. 指定节点注册

```go
// 仅为特定节点添加回调
result, err := runnable.Invoke(ctx, input,
    compose.WithCallbacks(myHandler).DesignateNode("ToolsNode"),
)
```

## RunInfo 结构

```go
type RunInfo struct {
    Name      string              // 节点名称（用户指定或自动生成）
    Type      string              // 实现类型（如 "OpenAI", "RedisRetriever"）
    Component components.Component // 组件类别（如 "ChatModel", "Tool"）
}
```

## 最佳实践

### 1. 实现 TimingChecker 跳过不需要的时机

```go
type LoggingHandler struct{}

func (h *LoggingHandler) Needed(timing callbacks.CallbackTiming) bool {
    // 只处理 OnStart 和 OnEnd
    return timing == callbacks.TimingOnStart || timing == callbacks.TimingOnEnd
}

func (h *LoggingHandler) OnStart(...) { ... }
func (h *LoggingHandler) OnEnd(...) { ... }
// 其他方法可以省略...
```

### 2. 流式回调必须关闭流

```go
OnEndWithStreamOutputFn(func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
    defer output.Close() // 必须！
    // ...
    return ctx
})
```

### 3. 使用多个 Handler

```go
// 组合多个 Handler
callbacks.AppendGlobalHandlers(
    loggingHandler,
    metricsHandler,
    langfuseHandler,
    tracingHandler,
)
```