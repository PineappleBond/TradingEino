# TradingEino

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.26.1-blue)](https://go.dev)
[![Eino](https://img.shields.io/badge/Eino-v0.8.4-purple)](https://github.com/cloudwego/eino)

[English Version](./README.md)

---

## 项目概述

TradingEino 是一个基于 Cloudwego Eino 框架的 AI 多 Agent 加密货币交易系统，监控 OKX 市场、分析技术指标和情绪，并自主执行交易。

## 核心功能

- **多 Agent 架构** - DeepAgent 协调器（OKXWatcher）调度专业子 Agent 进行技术分析、情绪分析、持仓管理和交易执行
- **OKX 交易所集成** - 完整的 REST API 支持，包括市场数据、账户管理和交易操作，带速率限制
- **技术分析** - 内置 20+ 指标（MACD、RSI、布林带、KDJ、ATR），通过 K 线数据分析
- **情绪分析** - 资金费率监控和市场情绪评估
- **风险管理** - 独立风险监控层，包括持仓追踪、保证金率预警和强平价格监控
- **RAG 决策记忆** - Redis Stack + m3e-base 嵌入，用于历史决策存储和检索（计划中）
- **Web 界面** - 嵌入式 SPA 前端，访问地址 `http://localhost:10098`

## 架构设计

```
┌─────────────────────────────────────────────────────────┐
│         OKXWatcher (DeepAgent - 总协调器)                │
│  触发：定时调度 / 价格异动                               │
│  职责：市场分析、策略生成、信号路由                       │
└─────────────┬───────────────────────────────────────────┘
              │
    ┌─────────┼──────────────────────┬────────────────────────┐
    │         │                      │                        │
┌───▼───┐ ┌──▼──────┐ ┌────────────▼────────┐ ┌─────▼──────┐
│Techno │ │Sentiment│ │   PositionManager   │ │  Executor  │
│       │ │Analyst  │ │                     │ │  (Level 1) │
└───────┘ └─────────┘ └─────────────────────┘ └────────────┘
```

### Agent 职责

| Agent                | 类型              | 工具                                                                                                        | 输出                |
|----------------------|-----------------|-----------------------------------------------------------------------------------------------------------|-------------------|
| **OKXWatcher**       | DeepAgent (协调器) | 无（仅调度子 Agent）                                                                                             | 市场分析、策略、执行指令      |
| **TechnoAgent**      | ChatModelAgent  | `okx-candlesticks-tool`                                                                                   | 趋势分析、支撑阻力位、指标信号   |
| **SentimentAnalyst** | ChatModelAgent  | `okx-get-funding-rate-tool`                                                                               | 资金费率分析、市场情绪       |
| **PositionManager**  | ChatModelAgent  | `okx-get-positions-tool`, `okx-get-orders-tool`, `okx-account-balance-tool`, `okx-liquidation-price-tool` | 持仓风险、盈亏、保证金预警     |
| **Executor**         | ChatModelAgent  | `okx-place-order-tool`, `okx-cancel-order-tool`, `okx-get-order-tool`, `okx-close-position-tool`          | 交易执行（Level 1 自主权） |

## 技术栈

| 组件            | 技术选型                                         |
|---------------|----------------------------------------------|
| **语言**        | Go 1.26.1                                    |
| **AI 框架**     | Cloudwego Eino v0.8.4                        |
| **Web 框架**    | Gin v1.12.0                                  |
| **数据库**       | SQLite3 (纯 Go，无 CGO)                         |
| **ORM**       | GORM v1.31.1                                 |
| **配置管理**      | Viper v1.21.0                                |
| **定时器**       | robfig/cron/v3                               |
| **技术分析**      | go-talib (后续迁移到 github.com/cinar/indicator) |
| **限流**        | golang.org/x/time/rate                       |
| **向量存储** (计划) | Redis Stack + RediSearch                     |
| **嵌入模型** (计划) | Ollama + m3e-base                            |

## 项目结构

```
backend/
├── cmd/
│   └── server/              # 应用入口
├── internal/
│   ├── agent/               # 多 Agent 系统
│   │   ├── okx_watcher/     # DeepAgent 协调器
│   │   ├── risk_officer/    # 风控分析（待重构）
│   │   ├── sentiment_analyst/ # 情绪分析
│   │   └── tools/           # Agent 工具（OKX 封装）
│   ├── api/                 # HTTP API 层
│   │   ├── handler/         # 请求处理器
│   │   ├── middleware/      # Gin 中间件
│   │   ├── request/         # 请求 DTO
│   │   └── response/        # 响应工具
│   ├── config/              # 配置加载
│   ├── logger/              # 结构化日志（slog）
│   ├── model/               # GORM 模型
│   ├── repository/          # 数据访问层
│   ├── scheduler/           # 任务调度
│   ├── service/             # 业务逻辑
│   └── svc/                 # 服务上下文（依赖注入）
├── pkg/
│   └── okex/                # OKX API 客户端
│       ├── api/             # REST/WebSocket 客户端
│       ├── models/          # OKX 响应模型
│       └── requests/        # OKX 请求模型
├── web/                     # 嵌入式前端（SPA）
├── data/                    # SQLite 数据库
├── logs/                    # 日志文件
├── etc/
│   └── config.yaml          # 应用配置
└── .planning/               # 项目文档
    ├── PROJECT.md           # 项目愿景
    ├── REQUIREMENTS.md      # 需求追踪
    ├── ROADMAP.md           # 开发路线图
    └── codebase/            # 架构文档
```

## 快速开始

### 前置条件

- Go 1.26.1 或更高版本
- OKX API 凭证（API Key、Secret Key、Passphrase）
- LLM API 访问（DashScope/阿里云或 OpenAI 兼容接口）

### 安装

```bash
# 克隆仓库
git clone https://github.com/PineappleBond/TradingEino.git
cd TradingEino/backend

# 安装依赖
go mod download

# 配置应用
cp etc/config.example.yaml etc/config.yaml
# 编辑 etc/config.yaml，填入 OKX 凭证和 API 密钥

# 构建
go build -o server ./cmd/server

# 运行
./server
```

Web 界面访问地址：`http://localhost:10098`

### 配置项

| 配置项                  | 说明             | 默认值                 |
|----------------------|----------------|---------------------|
| `server.port`        | HTTP 服务器端口     | 10098               |
| `okx.api_key`        | OKX API key    | -                   |
| `okx.secret_key`     | OKX secret key | -                   |
| `okx.passphrase`     | OKX passphrase | -                   |
| `okx.sandbox`        | 使用 OKX 沙盒      | true                |
| `chatmodel.endpoint` | LLM API 端点     | 阿里云 DashScope       |
| `chatmodel.api_key`  | LLM API key    | -                   |
| `db.type`            | 数据库类型          | sqlite              |
| `db.path`            | SQLite 数据库路径   | data/TradingEino.db |

## 开发路线图

| 阶段       | 目标                                 | 状态     |
|----------|------------------------------------|--------|
| **阶段 1** | 基础安全（错误处理、限流、单例、优雅关闭）              | ✅ 完成   |
| **阶段 2** | 分析层（重构子 Agent 为 ChatModelAgent）    | 🔄 进行中 |
| **阶段 3** | 执行自动化（交易工具、Executor Agent Level 1） | ⏳ 计划中  |
| **阶段 4** | RAG 记忆（Redis Stack、决策保存/搜索）        | ⏳ 计划中  |

## 关键决策 (ADR)

| 决策                            | 理由                               | 状态    |
|-------------------------------|----------------------------------|-------|
| DeepAgent 仅用于 OKXWatcher      | 避免层级冗余，子 Agent 使用 ChatModelAgent | ✅ 已批准 |
| 分析/执行分离                       | 清晰的审计追踪，独立测试                     | ✅ 已批准 |
| Tool 原子化                      | 每个 Tool 只做一件事                    | ✅ 已批准 |
| RAG 使用 Redis Stack + m3e-base | 本地 Embedding，无需外部 API            | ✅ 已批准 |
| 独立风控层                         | 实时监控，可覆盖决策                       | ✅ 已批准 |
| Executor 从 Level 1 开始         | 仅执行明确指令，逐步建立信任                   | ⏳ 待定  |

## 安全特性

- **速率限制** - 所有 API 工具都有保守的限流（交易端点 5 req/s）
- **错误处理** - 正确的错误传播（返回 `"", err` 模式）
- **上下文传播** - 取消操作端到端有效
- **单例模式** - 使用 `sync.Once` 实现线程安全的 Agent 初始化
- **优雅关闭** - 退出时正确的资源清理（Server → Scheduler → Agents → DB）
- **Executor Level 1** - 仅执行协调器的明确指令

## API 端点

| 端点                   | 方法                  | 说明          |
|----------------------|---------------------|-------------|
| `/api/v1/health`     | GET                 | 健康检查        |
| `/api/v1/cron/tasks` | GET/POST/PUT/DELETE | Cron 任务管理   |
| `/`                  | GET                 | Web 前端（SPA） |

## 可用工具

| 工具                     | 用途             | 限流       |
|------------------------|----------------|----------|
| `okx-candlesticks`     | K 线数据 + 20+ 指标 | 10 req/s |
| `okx-get-positions`    | 查询当前持仓         | 5 req/s  |
| `okx-get-funding-rate` | 资金费率数据         | 10 req/s |
| `okx-place-order`      | 下市价/限价单        | 5 req/s  |
| `okx-cancel-order`     | 撤销挂单           | 5 req/s  |
| `okx-get-order`        | 查询订单状态         | 10 req/s |
| `okx-close-position`   | 平仓             | 5 req/s  |

## 许可证

TradingEino 采用 [Apache License 2.0](LICENSE) 开源许可证。

---

## 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 免责声明

本软件仅供教育用途。加密货币交易存在重大亏损风险。开发者不对因使用本软件造成的任何财务损失负责。在实盘交易前，务必在沙盒环境中充分测试。