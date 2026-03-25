# Eino Interrupt/Resume 详解

Interrupt/Resume 机制允许在应用执行过程中暂停，等待外部输入（如用户确认）后继续执行。

## 核心概念

### Interrupt（中断）
在节点执行过程中触发暂停，保存当前状态。

### Resume（恢复）
从中断点继续执行，可选传递额外数据。

### Checkpoint（检查点）
持久化存储执行状态，用于恢复。

## 完整示例

### 1. 基础 Interrupt

```go
package main

import (
    "context"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
)

type ApprovalState struct {
    PendingApproval bool
    ApprovalResult  string
}

func main() {
    ctx := context.Background()

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
                // 恢复后继续执行
                state.PendingApproval = false
                state.ApprovalResult = data.Status
                return "已执行敏感操作：" + data.Status, nil
            }

            // 首次执行，触发中断
            state.PendingApproval = true
            compose.Interrupt(ctx, map[string]any{
                "action":  "approval",
                "message": "需要用户批准此敏感操作",
                "details": input,
            })
            return "", nil
        },
    ))

    _ = g.AddEdge(compose.START, "sensitive_op")
    _ = g.AddEdge("sensitive_op", compose.END)

    // 编译（带 CheckpointStore）
    store := &MemoryStore{data: make(map[string][]byte)}
    runnable, err := g.Compile(ctx,
        compose.WithCheckPointStore(store),
        compose.WithGraphName("ApprovalGraph"),
    )
    if err != nil {
        panic(err)
    }

    // 首次执行（会中断）
    _, err = runnable.Invoke(ctx, "删除重要数据")
    if compose.IsInterruptError(err) {
        // 获取中断信息
        interruptInfo := compose.GetInterruptInfo(ctx)
        fmt.Printf("中断：%v\n", interruptInfo.Info)
        fmt.Printf("中断点：%v\n", interruptInfo.NodePath)
    }

    // 恢复执行（带数据）
    ctx = compose.ResumeWithData(ctx, interruptInfo.InterruptID, ApprovalResult{
        Status: "approved",
    })
    result, _ := runnable.Invoke(ctx, "删除重要数据",
        compose.WithCheckPointID(interruptInfo.CheckPointID),
    )
    fmt.Println(result)
}

type ApprovalResult struct {
    Status string
}

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
```

### 2. 编译时配置中断点

```go
// 在特定节点前自动中断
runnable, err := graph.Compile(ctx,
    compose.WithCheckPointStore(store),
    compose.WithInterruptBeforeNodes([]string{"ToolsNode", "DeleteNode"}),
)

// 执行时会自动在指定节点前中断
result, err := runnable.Invoke(ctx, input)
if compose.IsInterruptError(err) {
    // 等待用户确认...
}
```

### 3. 带状态恢复

```go
type WorkflowState struct {
    Step       int
    Data       map[string]any
    Approved   bool
}

// 创建图
g := compose.NewGraph[string, string](
    compose.WithGenLocalState(func(ctx context.Context) *WorkflowState {
        return &WorkflowState{Data: make(map[string]any)}
    }),
)

// 节点 1：收集数据
_ = g.AddLambdaNode("collect", compose.InvokableLambda(
    func(ctx context.Context, input string) (string, error) {
        state := compose.GetState[*WorkflowState](ctx)
        state.Data["input"] = input
        state.Step = 1
        return input, nil
    },
))

// 节点 2：需要审批
_ = g.AddLambdaNode("approve", compose.InvokableLambda(
    func(ctx context.Context, input string) (string, error) {
        state := compose.GetState[*WorkflowState](ctx)

        // 检查是否已恢复
        isResume, hasData, data := compose.GetResumeContext[ApprovalDecision](ctx)
        if isResume && hasData {
            state.Approved = data.Approved
            state.Step = 3
            if data.Approved {
                return "已批准，继续执行", nil
            }
            return "已拒绝，停止执行", nil
        }

        // 触发中断
        compose.Interrupt(ctx, map[string]any{
            "type":    "approval",
            "message": "请批准此操作",
            "data":    state.Data,
        })
        return "", nil
    },
))

// 节点 3：执行
_ = g.AddLambdaNode("execute", compose.InvokableLambda(
    func(ctx context.Context, input string) (string, error) {
        state := compose.GetState[*WorkflowState](ctx)
        if !state.Approved {
            return "操作已取消", nil
        }
        state.Step = 4
        return "执行完成：最终数据 = " + fmt.Sprintf("%v", state.Data), nil
    },
))

// 边
_ = g.AddEdge(compose.START, "collect")
_ = g.AddEdge("collect", "approve")
_ = g.AddEdge("approve", "execute")
_ = g.AddEdge("execute", compose.END)
```

### 4. 批量 Resume

```go
// 多个中断点
ctx = compose.BatchResumeWithData(ctx, map[string]any{
    "interrupt-id-1": ApprovalDecision{Approved: true},
    "interrupt-id-2": ApprovalDecision{Approved: false},
    "interrupt-id-3": CustomData{Value: "abc"},
})

result, err := runnable.Invoke(ctx, input,
    compose.WithCheckPointID("checkpoint-id"),
)
```

## 中断 API 总结

### 触发中断

```go
// 标准中断
compose.Interrupt(ctx, info any) error

// 带状态中断
compose.StatefulInterrupt(ctx, info any, state any) error

// 复合中断（多个子进程）
compose.CompositeInterrupt(ctx, info any, state any, errs ...error) error
```

### 检查中断状态

```go
// 检查是否被中断
wasInterrupted, hasState, state := compose.GetInterruptState[MyState](ctx)

// 获取中断信息
interruptInfo := compose.GetInterruptInfo(ctx)
// interruptInfo.InterruptID
// interruptInfo.CheckPointID
// interruptInfo.NodePath
// interruptInfo.Info

// 获取恢复上下文
isResumeFlow, hasData, data := compose.GetResumeContext[ResumeData](ctx)

// 获取当前节点地址
addr := compose.GetCurrentAddress(ctx)
```

### 恢复执行

```go
// 无数据恢复
ctx = compose.Resume(ctx, "interrupt-id-1", "interrupt-id-2")

// 带数据恢复
ctx = compose.ResumeWithData(ctx, "interrupt-id", myData)

// 批量恢复
ctx = compose.BatchResumeWithData(ctx, map[string]any{
    "id1": data1,
    "id2": data2,
})
```

## CheckpointStore 实现

```go
type CheckPointStore interface {
    Get(ctx context.Context, id string) (data []byte, existed bool, err error)
    Set(ctx context.Context, id string, data []byte) error
}

// 内存实现（测试用）
type MemoryStore struct {
    mu   sync.Mutex
    data map[string][]byte
}

func (m *MemoryStore) Get(ctx context.Context, id string) ([]byte, bool, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    d, ok := m.data[id]
    return d, ok, nil
}

func (m *MemoryStore) Set(ctx context.Context, id string, data []byte) error {
    m.mu.Lock()
    defer m.mu.Unlock()
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

## 常见场景

### 1. 敏感操作审批

```go
// 在删除、转账等操作前中断等待用户确认
compose.Interrupt(ctx, map[string]any{
    "action":  "delete",
    "target":  "user-123",
    "message": "确定要删除此用户吗？",
})
```

### 2. 人机协作写作

```go
// 每生成一段内容后等待用户反馈
compose.Interrupt(ctx, map[string]any{
    "action":  "review",
    "content": generatedContent,
    "step":    currentStep,
})
```

### 3. 多步骤表单

```go
// 每步完成后等待用户确认
compose.Interrupt(ctx, map[string]any{
    "action":     "step_complete",
    "step":       stepNumber,
    "stepData":   stepData,
    "totalSteps": 5,
})
```