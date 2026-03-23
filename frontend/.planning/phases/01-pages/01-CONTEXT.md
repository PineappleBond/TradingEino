# TradingEino 前端页面开发需求文档

## 项目概述

TradingEino 前端是一个定时任务管理系统，采用赛博朋克风格设计（经典霓虹紫蓝配色）。

## 技术栈

- **React 18** + **TypeScript**
- **Vite 5** - 构建工具
- **Ant Design 5** - 主要 UI 组件库
- **React Router 6** - 路由管理
- **Zustand** - 状态管理
- **Axios** - HTTP 客户端
- **Day.js** - 时间处理
- **Ahooks** - React Hooks 库

## 后端 API 模块

### 1. 定时任务 (CronTask)

**API 端点**: `/api/crontask`

| 操作 | 方法 | 端点 | 说明 |
|------|------|------|------|
| 列表 | GET | `/api/crontask` | 支持分页和筛选 |
| 详情 | GET | `/api/crontask/:id` | 获取任务详情 |
| 创建 | POST | `/api/crontask` | 创建新任务 |
| 更新 | PUT | `/api/crontask/:id` | 更新任务 |
| 删除 | DELETE | `/api/crontask/:id` | 删除任务 |
| 启用 | POST | `/api/crontask/:id/enable` | 启用任务 |
| 禁用 | POST | `/api/crontask/:id/disable` | 禁用任务 |
| 启动 | POST | `/api/crontask/:id/start` | 启动任务（设置下次执行时间） |
| 停止 | POST | `/api/crontask/:id/stop` | 停止任务 |

**筛选参数**:
- `page`: 页码（默认 1）
- `pageSize`: 每页数量（默认 10，提供 10/20/50/100 选项）
- `status`: 任务状态（pending/running/completed/stopped/failed）
- `type`: 任务类型（once/recurring）
- `enabled`: 是否启用（true/false）

**任务字段**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | - | 任务 ID |
| name | string | 是 | 任务名称（max 100） |
| spec | string | 是 | Cron 表达式（max 50） |
| type | enum | 是 | 任务类型：once/recurring |
| status | enum | - | 任务状态 |
| exec_type | string | 是 | 执行类型，当前仅支持 "OKXWatcher" |
| raw | JSON | 是 | 原始配置，格式：`{"symbol": "ETH-USDT-SWAP"}` |
| valid_from | datetime | 是 | 有效期开始 |
| valid_until | datetime | 是 | 有效期结束 |
| enabled | bool | - | 是否启用 |
| max_retries | int | - | 最大重试次数 |
| timeout_seconds | int | - | 超时时间（秒） |
| next_execution_at | datetime | - | 下次执行时间 |
| last_executed_at | datetime | - | 上次执行时间 |
| total_executions | uint | - | 总执行次数 |
| created_at | datetime | - | 创建时间 |
| updated_at | datetime | - | 更新时间 |

**操作逻辑**:
- 启用/禁用：无状态限制
- 启动：需要传入 `next_execution_time`（格式：`YYYY-MM-DD HH:mm:ss`）
- 停止：无状态限制

### 2. 执行记录 (CronExecution)

**API 端点**: `/api/cronexecution`

| 操作 | 方法 | 端点 | 说明 |
|------|------|------|------|
| 列表 | GET | `/api/cronexecution` | 支持分页和筛选 |
| 详情 | GET | `/api/cronexecution/:id` | 获取执行详情 |
| 按任务查询 | GET | `/api/cronexecution/task/:task_id` | 分页获取指定任务的执行记录 |

**筛选参数**:
- `page`: 页码
- `pageSize`: 每页数量
- `task_id`: 任务 ID
- `status`: 执行状态（pending/running/success/failed/retried/cancelled）
- `start_time`: 开始时间
- `end_time`: 结束时间

**执行记录字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 执行记录 ID |
| task_id | uint | 关联的任务 ID |
| scheduled_at | datetime | 计划执行时间 |
| started_at | datetime | 实际开始时间 |
| completed_at | datetime | 完成时间 |
| status | enum | 执行状态 |
| retry_count | int | 重试次数 |
| error | string | 错误信息 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### 3. 执行日志 (CronExecutionLog)

**API 端点**: `/api/cronexecutionlog`

| 操作 | 方法 | 端点 | 说明 |
|------|------|------|------|
| 列表 | GET | `/api/cronexecutionlog` | 支持分页和筛选 |
| 详情 | GET | `/api/cronexecutionlog/:id` | 获取日志详情 |
| 按执行 ID 查询 | GET | `/api/cronexecutionlog/execution/:execution_id` | 分页获取指定执行的日志 |

**筛选参数**:
- `page`: 页码
- `pageSize`: 每页数量
- `execution_id`: 执行记录 ID
- `level`: 日志级别（info/warn/error/debug）
- `from`: 日志来源

**执行日志字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 日志 ID |
| execution_id | uint | 关联的执行 ID |
| from | string | 日志来源 |
| level | string | 日志级别 |
| message | string | 日志内容 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### 4. 系统日志 (SystemLog)

**API 端点**: `/api/systemlog`

| 操作 | 方法 | 端点 | 说明 |
|------|------|------|------|
| 文件列表 | GET | `/api/systemlog/files` | 获取日志文件列表 |
| 文件内容 | GET | `/api/systemlog/files/:filename` | 获取指定文件内容 |
| 搜索日志 | GET | `/api/systemlog/search` | 搜索日志内容 |
| 统计信息 | GET | `/api/systemlog/stats` | 获取日志统计 |

**筛选参数**:
- 文件列表：`page`, `pageSize`
- 文件内容：`page`, `pageSize`, `level`, `start_time`, `end_time`
- 搜索日志：`keyword`（必填）, `filename`, `level`, `start_time`, `end_time`, `page`, `pageSize`

**日志文件信息字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| filename | string | 文件名 |
| size | int64 | 文件大小 |
| mod_time | datetime | 修改时间 |
| line_count | int | 行数 |
| first_log_time | datetime | 首条日志时间 |
| last_log_time | datetime | 末条日志时间 |

**日志条目字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| time | datetime | 日志时间 |
| level | string | 日志级别 |
| msg | string | 日志内容 |

## 路由配置

| 路径 | 说明 | 备注 |
|------|------|------|
| `/` | 根路径，重定向到 `/task` | 仪表盘选项重定向到任务列表 |
| `/task` | 定时任务列表 | |
| `/task/create` | 创建定时任务 | 独立路由，弹窗形式 |
| `/task/:id` | 定时任务详情 | 独立路由 |
| `/task/:id/edit` | 编辑定时任务 | 独立路由，弹窗形式 |
| `/task/execution` | 执行记录列表 | |
| `/task/execution/:id` | 执行记录详情 | 独立路由 |
| `/log/execution` | 执行日志列表 | |
| `/log/execution/:id` | 执行日志详情 | 独立路由 |
| `/log/system` | 系统日志列表 | |
| `/log/system/:filename` | 系统日志文件内容 | 独立路由 |

## 核心开发原则

### 1. 避免重复造轮子

- **优先使用 Ant Design 组件库**
- shadcn/ui 作为备选（当 Ant Design 缺少所需组件时）
- 网络搜索寻找第三方组件库

### 2. 组件组合优于自定义实现

- 使用 Ant Design 的组合式组件
- 使用内置 variant 而非自定义样式

### 3. 样式规范

- 优先使用 Ant Design Token 系统进行主题定制
- 使用语义化颜色
- 避免全局 `.ant-*` 选择器覆盖
- 赛博朋克主题已在 `styles/theme.ts` 中配置

### 4. TypeScript

- 所有组件必须使用 TypeScript
- 定义清晰的 Props 接口
- 使用泛型提高代码复用性

### 5. 代码组织

- 单一职责：每个组件只做一件事
- 提取可复用逻辑到 `utils/` 或自定义 Hooks
- 页面组件负责数据获取，子组件负责展示

## 设计风格

### 赛博朋克主题（经典霓虹紫蓝）

- **背景色**: `#0a0a0f`（深紫黑）
- **主色**: `#bf00ff`（霓虹紫）
- **强调色**: `#00ffff`（青色）、`#ff0080`（霓虹粉）
- **效果**: 发光效果、渐变文字、扫描线动画

参考 `styles/theme.ts` 和 `styles/global.css` 中的主题配置。

## UI/UX 规范

### 分页大小

- 默认：10 条/页
- 可选：10, 20, 50, 100

### 日期时间格式

- 输入/输出格式：`YYYY-MM-DD HH:mm:ss`（匹配后端 API）
- 列表显示格式：`YYYY-MM-DD HH:mm:ss`

### 错误提示

- 使用 Ant Design 的 `message.error()`
- 表单验证错误使用内联显示

### 删除确认

- 所有删除操作需要二次确认弹窗

### 创建成功后

- 关闭弹窗，返回原页面

### 表单特殊规则

- `valid_from` / `valid_until`: 必选的日期范围选择器
- `spec`: 用户手动输入 Cron 表达式
- `exec_type`: 固定为 "OKXWatcher"（隐藏字段或只读）
- `raw.symbol`: 自由文本输入，单个交易对（如 `ETH-USDT-SWAP`）
- `type`: 无论何种类型都需要填写 `spec`

## 需要创建的文件

### API 层

```
frontend/src/api/
├── index.ts              # Axios 实例和基础配置
├── crontask.ts           # 定时任务 API
├── cronexecution.ts      # 执行记录 API
├── cronexecutionlog.ts   # 执行日志 API
└── systemlog.ts          # 系统日志 API
```

### 类型定义

```
frontend/src/types/
├── crontask.ts
├── cronexecution.ts
├── cronexecutionlog.ts
└── systemlog.ts
```

### Store

```
frontend/src/store/
└── index.ts              # Zustand store（如需要）
```

### 页面组件

```
frontend/src/pages/
├── task/
│   ├── TaskList.tsx       # 任务列表
│   ├── TaskDetail.tsx     # 任务详情
│   ├── TaskForm.tsx       # 任务表单（创建/编辑复用）
│   └── index.ts
├── execution/
│   ├── ExecutionList.tsx  # 执行记录列表
│   ├── ExecutionDetail.tsx # 执行记录详情
│   └── index.ts
├── log/
│   ├── ExecutionLogList.tsx  # 执行日志列表
│   ├── ExecutionLogDetail.tsx # 执行日志详情
│   ├── SystemLogList.tsx     # 系统日志文件列表
│   ├── SystemLogDetail.tsx   # 系统日志文件内容
│   └── index.ts
└── Dashboard.tsx         # 仪表盘（重定向用）
```

### 工具函数

```
frontend/src/utils/
├── request.ts            # 请求封装
├── format.ts             # 格式化函数
└── constants.ts          # 常量定义
```

## 路由菜单调整

当前菜单结构保持不变，但仪表盘点击后重定向到任务列表。
