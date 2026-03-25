---
phase: 01-foundation-safety
plan: 03
type: execute
wave: 2
depends_on:
  - 01-foundation-safety-02
files_modified:
  - cmd/server/main.go
  - internal/logger/logger.go
autonomous: false
requirements:
  - FOUND-05
must_haves:
  truths:
    - "应用监听 SIGINT/SIGTERM 信号"
    - "收到信号后按顺序关闭：Server → Scheduler → Agents → DB → Logger"
    - "关闭完成后进程正常退出"
  artifacts:
    - path: "internal/logger/logger.go"
      provides: "Logger.Close() 方法"
      exports: ["Close"]
    - path: "cmd/server/main.go"
      provides: "信号处理 + 优雅关闭逻辑"
      contains: "signal.Notify"
  key_links:
    - from: "cmd/server/main.go"
      to: "internal/logger/logger.go"
      via: "Logger 关闭调用"
      pattern: "logger\\.Close\\(\\)"
    - from: "cmd/server/main.go"
      to: "internal/agent/agents.go"
      via: "Agents 关闭调用"
      pattern: "agent\\.Agents\\(\\)\\.Close\\(\\)"
    - from: "cmd/server/main.go"
      to: "internal/service/scheduler/scheduler.go"
      via: "Scheduler 关闭调用"
      pattern: "sch\\.Stop\\(\\)"
---

<objective>
实现应用程序优雅关闭

Purpose: FOUND-05 要求应用程序在接收到退出信号时优雅关闭，按顺序关闭 Server → Scheduler → Agents → DB → Logger，确保资源正确清理，无 goroutine 泄漏。

Output:
- cmd/server/main.go 更新（信号处理 + 优雅关闭）
- internal/logger/logger.go 添加 Close() 方法
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
@cmd/server/main.go
@internal/logger/logger.go
@internal/service/scheduler/scheduler.go
</context>

<tasks>

<task type="auto">
<name>Task 1: 为 Logger 添加 Close() 方法</name>
<files>internal/logger/logger.go</files>
<action>
为 Logger 添加 Close() 方法，用于 flush 缓冲区并关闭文件句柄：

```go
func (l *Logger) Close() error {
    // Logger 内部包装了 slog.Logger，需要访问底层 handler 的 io.Writer
    // 如果是文件输出，需要关闭文件
    // 由于 slog.Handler 不直接暴露 writer，需要在 Logger 结构体中保存 closer
    // 修改 Logger 结构体：
    type Logger struct {
        inner       *slog.Logger
        addSource   bool
        skipCallers int
        closer      io.Closer  // 新增字段
    }

    // Close 方法：
    func (l *Logger) Close() error {
        if l.closer != nil {
            return l.closer.Close()
        }
        return nil
    }
}
```

同时修改 New() 函数，在打开文件时保存 closer。
</action>
<verify>
<automated>go build ./internal/logger/...</automated>
</verify>
<done>Logger.Close() 方法存在，编译通过</done>
</task>

<task type="auto">
<name>Task 2: 在 main.go 中添加信号处理和优雅关闭</name>
<files>cmd/server/main.go</files>
<action>
修改 cmd/server/main.go 实现优雅关闭：

1. 导入必要包：`os/signal`, `syscall`

2. 在 serve.Start() 后添加信号处理：
```go
// Setup signal handling
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

logger.Info(ctx, "server started, waiting for shutdown signal")

// Wait for shutdown signal
<-sigChan
logger.Info(ctx, "shutdown signal received")

// Ordered shutdown
// 1. Stop HTTP server
if err := serve.Shutdown(ctx); err != nil {
    logger.Error(ctx, "failed to shutdown server", err)
}

// 2. Stop scheduler
if err := sch.Stop(); err != nil {
    logger.Error(ctx, "failed to stop scheduler", err)
}

// 3. Close agents
if err := agent.Agents().Close(); err != nil {
    logger.Error(ctx, "failed to close agents", err)
}

// 4. Close database
db, _ := svcCtx.DB.DB()
if err := db.Close(); err != nil {
    logger.Error(ctx, "failed to close database", err)
}

// 5. Close logger
if err := logger.Close(); err != nil {
    logger.Error(ctx, "failed to close logger", err)
}

logger.Info(ctx, "graceful shutdown completed")
```

3. 需要使用 agent.Agents() 获取 AgentsModel（Plan 02 已实现 Close 方法）
</action>
<verify>
<automated>go build ./cmd/server/...</automated>
</verify>
<done>main.go 实现信号处理和优雅关闭，编译通过</done>
</task>

<task type="checkpoint:human-verify">
<name>Task 3: 验证优雅关闭流程</name>
<files>cmd/server/main.go</files>
<action>
验证 Claude 自动化的优雅关闭流程：
1. 启动应用：`go run cmd/server/main.go`
2. 等待日志 "server started, waiting for shutdown signal"
3. 按 Ctrl+C 发送 SIGINT
4. 验证日志输出顺序：
   - "shutdown signal received"
   - "graceful shutdown completed"（最后一条）
5. 验证进程正常退出（无 goroutine 泄漏警告）
</action>
<verify>
<automated>MISSING - 需要人工验证信号处理和关闭顺序</automated>
</verify>
<done>人工确认优雅关闭验证通过</done>
</task>

</tasks>

<verification>
- [ ] go build ./cmd/server/... 编译成功
- [ ] 应用启动后按 Ctrl+C 触发关闭
- [ ] 日志显示正确的关闭顺序
- [ ] 进程正常退出
</verification>

<success_criteria>
- Logger.Close() 方法实现
- main.go 有 os/signal 处理
- 关闭顺序：Server → Scheduler → Agents → DB → Logger
- 人工验证通过
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation-safety/01-foundation-safety-03-SUMMARY.md`
</output>
