# Scheduler 设计文档

## 概述

Scheduler 是一个基于 `robfig/cron/v3` 的定时任务调度器，支持一次性任务和重复执行任务。

## 架构设计

### 目录结构

```
backend/internal/service/scheduler/
├── scheduler.go          # 主调度器实现
├── executor.go           # 执行器（已内联到 scheduler.go）
├── errors.go             # 错误类型定义
├── registry.go           # Handler 注册辅助
├── scheduler_test.go     # 单元测试
└── handlers/             # 具体 Handler 实现
    └── example_handler.go # 示例 Handler 模板
```

### 核心组件

1. **Scheduler** - 主调度器
   - 生命周期管理（Start/Stop）
   - 任务加载和验证
   - 调度循环
   - 并发控制

2. **Executor** - 执行器注册表
   - Handler 注册和查找
   - 线程安全

3. **TaskHandler** - 任务执行器接口
   ```go
   type TaskHandler interface {
       Name() string
       Execute(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error
   }
   ```

## 配置

在 `config.yaml` 中添加：

```yaml
scheduler:
  enabled: true         # 是否启用调度器
  max_concurrency: 5    # 最大并发执行数
  check_interval: 10    # 检查周期（秒）
  default_timeout: 300  # 默认超时时间（秒）
```

## 数据库修改

### CronTask 表新增字段

```sql
ALTER TABLE cron_task ADD COLUMN timeout_seconds INTEGER NOT NULL DEFAULT 300;
```

## 使用示例

### 1. 实现 Handler

复制 `handlers/example_handler.go` 模板，实现你的业务逻辑：

```go
type MyHandler struct {
    svcCtx *svc.ServiceContext
}

func NewMyHandler(svcCtx *svc.ServiceContext) *MyHandler {
    return &MyHandler{svcCtx: svcCtx}
}

func (h *MyHandler) Name() string {
    return "my_task"  // 与 CronTask.ExecType 匹配
}

func (h *MyHandler) Execute(ctx context.Context, task *model.CronTask, execution *model.CronExecution) error {
    // 解析参数
    var params MyParams
    if err := json.Unmarshal([]byte(task.Raw), &params); err != nil {
        return scheduler.NewNonRetryableError(fmt.Errorf("failed to parse params: %w", err))
    }

    // 执行业务逻辑
    // ...

    return nil
}
```

### 2. 注册 Handler

在 `registry.go` 中注册：

```go
func (s *Scheduler) RegisterDefaultHandlers() {
    s.RegisterHandler(handlers.NewMyHandler(s.svcCtx))
}
```

### 3. 创建任务

```go
task := &model.CronTask{
    Name:           "我的任务",
    Spec:           "0 30 2 * * *",  // 每天 2:30 执行（6 位 cron 表达式）
    Type:           model.TaskTypeRecurring,
    ExecType:       "my_task",       // 必须与 Handler.Name() 匹配
    Raw:            `{"key": "value"}`,
    Enabled:        true,
    MaxRetries:     3,
    TimeoutSeconds: 600,
}
```

### 4. 启动调度器

```go
svcCtx := svc.NewServiceContext(cfg)
scheduler := NewScheduler(svcCtx)

// 注册 Handler
scheduler.RegisterDefaultHandlers()

// 启动
if err := scheduler.Start(); err != nil {
    log.Fatal(err)
}

// 停止
defer scheduler.Stop()
```

## 特性

### 1. 任务类型

- **Once** - 一次性任务，执行后标记为 `completed`
- **Recurring** - 重复执行任务，根据 cron 表达式调度

### 2. 并发控制

- 通过信号量限制最大并发执行数
- 配置项：`scheduler.max_concurrency`

### 3. 跳过策略

如果同一任务的上次执行尚未完成，新触发会被跳过。

### 4. 重试机制

- 支持配置最大重试次数 (`MaxRetries`)
- 只有返回 `RetryableError` 才会重试
- 重试会创建新的执行记录

### 5. 超时控制

- 任务级别：`CronTask.TimeoutSeconds`
- 默认值：`scheduler.default_timeout`

### 6. 有效期

- `ValidFrom` - 有效期开始
- `ValidUntil` - 有效期结束

## 错误处理

### 可重试错误

```go
return scheduler.NewRetryableError(errors.New("临时错误"))
```

### 不可重试错误

```go
return scheduler.NewNonRetryableError(errors.New("永久错误"))
```

## 注意事项

1. **Cron 表达式格式**：支持 6 位格式（秒 分 时 日 月 周）
2. **Handler 名称**：必须唯一，与 `CronTask.ExecType` 完全匹配
3. **服务重启**：会自动加载所有启用的任务
4. **日志记录**：执行日志记录在 `cron_execution_log` 表中

## 待办事项

- [ ] 添加更多实际业务 Handler
- [ ] 添加执行统计和监控
- [ ] 添加管理 API（可选）
- [ ] 添加更多单元测试（覆盖重试逻辑等）
