# TradingEino Frontend - Claude.md

## 项目概述

TradingEino 前端是一个定时任务管理系统，采用赛博朋克风格设计。

## 技术栈

- **React 18** + **TypeScript**
- **Vite 5** - 构建工具
- **Ant Design 5** - 主要 UI 组件库
- **React Router 6** - 路由管理
- **Zustand** - 状态管理
- **Axios** - HTTP 客户端
- **Day.js** - 时间处理
- **Ahooks** - React Hooks 库

## 构建与开发

```bash
# 安装依赖
npm install

# 开发模式
npm run dev

# 构建（输出到 ../backend/web/dist）
npm run build

# 预览构建结果
npm run preview
```

## 项目结构

```
frontend/
├── src/
│   ├── layouts/          # 布局组件
│   │   ├── MainLayout.tsx
│   │   └── MainLayout.css
│   ├── router/           # 路由配置
│   │   └── index.tsx
│   ├── pages/            # 页面组件
│   │   ├── task/         # 任务管理相关页面
│   │   └── log/          # 日志相关页面
│   ├── components/       # 通用组件
│   ├── store/            # Zustand 状态管理
│   ├── utils/            # 工具函数
│   ├── styles/           # 全局样式
│   │   ├── theme.ts      # Ant Design 主题配置
│   │   └── global.css    # 全局 CSS
│   ├── App.tsx
│   └── main.tsx
├── package.json
├── vite.config.ts
└── tsconfig.json
```

## 可用 Skill

### 前端设计相关

- **frontend-design** - 创建独特的前端界面和组件
- **ant-design** - Ant Design 组件使用指南（antd 6.x, Pro 5, X 2）
- **shadcn** - shadcn/ui 组件管理（如 Ant Design 缺少所需组件时使用）
- **react-performance-optimization** - React 性能优化

### TypeScript 相关

- **typescript-advanced-types** - TypeScript 高级类型系统

### API 相关

- **api-design-principles** - REST/GraphQL API 设计原则
- **api-documentation** - API 文档创建

### 代码质量

- **code-review-excellence** - 代码审查最佳实践
- **requesting-code-review** - 请求代码审查
- **receiving-code-review** - 接收和处理代码审查反馈

### 流程相关

- **brainstorming** - 创意构思和设计
- **writing-plans** - 编写实现计划
- **executing-plans** - 执行计划
- **systematic-debugging** - 系统化调试
- **test-driven-development** - 测试驱动开发

## 核心开发原则

### 1. 避免重复造轮子

**尽量避免自己造轮子，优先使用以下资源：**

1. **Ant Design 组件库** - 优先使用 Ant Design 已有组件
   - 查询官方文档：https://ant.design/components/
   - 使用 `ant-design` skill 获取组件使用指南

2. **shadcn/ui 组件库** - 当 Ant Design 缺少所需组件时
   - 使用 `shadcn` skill 搜索和管理组件
   - 通过 CLI 添加：`npx shadcn@latest add <component>`

3. **网络搜索** - 寻找第三方组件库
   - npm 搜索相关组件
   - GitHub 寻找高质量开源组件

### 2. 组件组合优于自定义实现

- 使用 Ant Design 的组合式组件（如 `Card` + `Form` + `Table`）
- shadcn 原则：组合而非重构（如 Settings 页面 = Tabs + Card + Form）
- 使用内置 variant 而非自定义样式

### 3. 样式规范

- 优先使用 Ant Design Token 系统进行主题定制
- 使用语义化颜色（`colorPrimary`、`colorBgContainer` 等）
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

## 路由配置

| 路径 | 说明 |
|------|------|
| `/` | 根路径，重定向到 `/task` |
| `/task` | 定时任务列表 |
| `/task/execution` | 执行记录 |
| `/log/execution` | 执行日志 |
| `/log/system` | 系统日志 |

## 后端 API

后端 API 前缀：`/api`

### 定时任务 (CronTask)
- `GET /api/crontask` - 任务列表
- `GET /api/crontask/:id` - 任务详情
- `POST /api/crontask` - 创建任务
- `PUT /api/crontask/:id` - 更新任务
- `DELETE /api/crontask/:id` - 删除任务
- `POST /api/crontask/:id/enable` - 启用任务
- `POST /api/crontask/:id/disable` - 禁用任务
- `POST /api/crontask/:id/start` - 启动任务
- `POST /api/crontask/:id/stop` - 停止任务

### 执行记录 (CronExecution)
- `GET /api/cronexecution` - 执行列表
- `GET /api/cronexecution/:id` - 执行详情
- `GET /api/cronexecution/task/:task_id` - 按任务 ID 查询

### 执行日志 (CronExecutionLog)
- `GET /api/cronexecutionlog` - 日志列表
- `GET /api/cronexecutionlog/:id` - 日志详情
- `GET /api/cronexecutionlog/execution/:execution_id` - 按执行 ID 查询

### 系统日志 (SystemLog)
- `GET /api/systemlog/files` - 日志文件列表
- `GET /api/systemlog/files/:filename` - 日志内容
- `GET /api/systemlog/search` - 搜索日志
- `GET /api/systemlog/stats` - 统计信息

## Vite 配置

- **开发服务器端口**: 5173
- **API 代理**: `/api` -> `http://localhost:8080`
- **构建输出**: `../backend/web/dist`
- **路径别名**: `@` -> `src/`

## 注意事项

1. **构建后验证**: 确保 `backend/web/dist` 目录包含构建产物
2. **开发模式**: 运行 `npm run dev` 后，通过 `http://localhost:5173` 访问
3. **生产模式**: 后端 Go 服务会嵌入 `dist` 目录的静态文件
