---
name: okex-skill
description: OKX API V5 完整集成指南 - 包含 REST 和 WebSocket API 的所有功能模块，提供账户管理、交易执行、市场数据、资金管理等完整 API 文档
---

# OKX API V5 技能

## 概述

本技能提供对 OKX API V5 的完整访问，包含 REST 和 WebSocket 两种接口方式。

## 核心组件

### 1. Client - 主客户端

```go
import "github.com/PineappleBond/TradingEino/backend/pkg/okex/api"

// 创建客户端
client, err := api.NewClient(ctx, apiKey, secretKey, passphrase, destination)
```

**Destination 选项：**
- `okex.NormalServer` - 正式环境
- `okex.DemoServer` - 模拟交易环境
- `okex.AwsServer` - AWS 环境

---

## API 模块索引

### REST API

| 模块 | 功能 | 文档 |
|------|------|------|
| **Account** | 账户管理、持仓、杠杆、账单 | [account.md](references/account.md) |
| **Trade** | 下单、撤单、订单查询、算法交易 | [trade.md](references/trade.md) |
| **Market** | 行情数据、K 线、深度、成交 | [market.md](references/market.md) |
| **Funding** | 资金管理、充值、提现、划转 | [funding.md](references/funding.md) |
| **PublicData** | 公共数据、合约信息、费率 | [public_data.md](references/public_data.md) |
| **SubAccount** | 子账户管理、APIKey 管理 | [sub_account.md](references/sub_account.md) |
| **TradeData** | 交易大数据、持仓分析 | [trade_data.md](references/trade_data.md) |

### WebSocket API

| 模块 | 功能 | 文档 |
|------|------|------|
| **WS Client** | WebSocket 连接、订阅、推送 | [websocket.md](references/websocket.md) |

---

## 快速开始

### 1. 初始化客户端

```go
import (
    "context"
    "github.com/PineappleBond/TradingEino/backend/pkg/okex"
    "github.com/PineappleBond/TradingEino/backend/pkg/okex/api"
)

ctx := context.Background()
client, err := api.NewClient(ctx, apiKey, secretKey, passphrase, okex.DemoServer)
if err != nil {
    // 处理错误
}
```

### 2. 调用 REST API

```go
// 获取账户余额
balance, err := client.Rest.Account.GetBalance(requests.GetBalance{})

// 获取行情 Ticker
ticker, err := client.Rest.Market.GetTicker(requests.GetTickers{InstID: "BTC-USDT"})

// 获取持仓
positions, err := client.Rest.Account.GetPositions(requests.GetPositions{})
```

### 3. 连接 WebSocket

```go
// 连接公共频道
err := client.Ws.Connect(false)

// 连接私有频道（需要登录）
err := client.Ws.Connect(true)

// 订阅频道
err := client.Ws.Subscribe(false, []okex.ChannelName{"ticker"}, map[string]string{
    "instId": "BTC-USDT",
})
```

---

## 数据类型

### 基础类型

```go
type InstrumentType string  // 产品类型：SPOT, MARGIN, SWAP, FUTURES, OPTION
type MarginMode string      // 保证金模式：cross, isolated
type PositionSide string    // 持仓方向：long, short, net
type OrderSide string       // 订单方向：buy, sell
type OrderType string       // 订单类型：market, limit, post_only, fok, ioc
type OrderState string      // 订单状态：canceled, live, partially_filled, filled
```

### 时间格式

```go
type JSONTime time.Time  // 自动解析毫秒时间戳
```

---

## 错误处理

所有 API 调用都返回 `(response, error)` 对：

```go
result, err := client.Rest.Account.GetBalance(requests.GetBalance{})
if err != nil {
    // 处理错误
    return
}
// 使用 result
```

---

## 速率限制

客户端内置签名和请求处理，但需要在使用时自行实现速率限制。

---

## 参考资料

- [OKX API 官方文档](https://www.okex.com/docs-v5/en/)
- [OKX API V5 中文文档](https://www.okx.com/docs-v5/zh/)
