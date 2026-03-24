# WebSocket API - 实时推送

## 概述

WebSocket API 提供实时行情推送、订单更新、账户变动等实时数据订阅功能。

## 访问方式

```go
client.Ws
```

## WebSocket 地址

### 正式环境
- 公共频道：`wss://ws.okx.com:8443/ws/v5/business`
- 私有频道：`wss://ws.okx.com:8443/ws/v5/private`

### AWS 环境
- 公共频道：`wss://wsaws.okx.com:8443/ws/v5/public`
- 私有频道：`wss://wsaws.okx.com:8443/ws/v5/private`

### 模拟交易环境
- 公共频道：`wss://wspap.okx.com:8443/ws/v5/public?brokerId=9999`
- 私有频道：`wss://wspap.okx.com:8443/ws/v5/private?brokerId=9999`

## 连接管理

### 创建客户端

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

### 连接 WebSocket

```go
// 连接公共频道（无需登录）
err := client.Ws.Connect(false)
if err != nil {
    // 处理错误
}

// 连接私有频道（需要登录）
err := client.Ws.Connect(true)
if err != nil {
    // 处理错误
}
```

### 等待登录完成

```go
// 私有频道需要等待登录完成
err := client.Ws.WaitForAuthorization()
if err != nil {
    // 处理错误
}
```

## 订阅频道

### 订阅公共频道

```go
import "github.com/PineappleBond/TradingEino/backend/pkg/okex"

// 订阅 Ticker
err := client.Ws.Subscribe(false, []okex.ChannelName{"ticker"}, map[string]string{
    "instId": "BTC-USDT-SWAP",
})

// 订阅深度
err := client.Ws.Subscribe(false, []okex.ChannelName{"books"}, map[string]string{
    "instId": "BTC-USDT-SWAP",
})

// 订阅 K 线
err := client.Ws.Subscribe(false, []okex.ChannelName{"candle1H"}, map[string]string{
    "instId": "BTC-USDT-SWAP",
})
```

### 订阅私有频道

```go
// 订阅账户余额
err := client.Ws.Subscribe(true, []okex.ChannelName{"account"}, nil)

// 订阅持仓
err := client.Ws.Subscribe(true, []okex.ChannelName{"positions"}, nil)

// 订阅订单
err := client.Ws.Subscribe(true, []okex.ChannelName{"orders"}, map[string]string{
    "instType": "SWAP",
})
```

## 常用频道列表

### 公共频道

| 频道 | 说明 | 参数 |
|------|------|------|
| `ticker` | 24 小时行情 | instId |
| `books` | 深度数据 | instId |
| `books5` | 5 档深度 | instId |
| `books15` | 15 档深度 | instId |
| `books50` | 50 档深度 | instId |
| `books-l2-tbt` | 全量深度 | instId |
| `candle1m`~`candle1Y` | K 线数据 | instId |
| `trades` | 成交数据 | instId |
| `platform-24-volume` | 24 小时成交量 | - |

### 私有频道

| 频道 | 说明 | 参数 |
|------|------|------|
| `account` | 账户余额 | - |
| `positions` | 持仓信息 | - |
| `orders` | 订单更新 | instType |
| `orders-algo` | 算法订单 | instType |
| `balance_and_position` | 余额和持仓 | - |

## 接收消息

### 设置消息频道

```go
import "github.com/PineappleBond/TradingEino/backend/pkg/okex/events"

errChan := make(chan *events.Error)
subChan := make(chan *events.Subscribe)
unSubChan := make(chan *events.Unsubscribe)
loginChan := make(chan *events.Login)
successChan := make(chan *events.Success)

client.Ws.SetChannels(errChan, subChan, unSubChan, loginChan, successChan)
```

### 接收消息

```go
// 从结构化频道接收
select {
case err := <-client.Ws.ErrChan:
    // 处理错误
case sub := <-client.Ws.SubscribeChan:
    // 处理订阅成功
case login := <-client.Ws.LoginChan:
    // 处理登录成功
case data := <-client.Ws.StructuredEventChan:
    // 处理其他事件
}

// 或从原始频道接收
for e := range client.Ws.RawEventChan {
    // 处理原始事件
}
```

## WebSocket 事件

### 错误事件

```go
type Error struct {
    Event string `json:"event"`
    Code  string `json:"code"`
    Msg   string `json:"msg"`
}
```

### 订阅成功事件

```go
type Subscribe struct {
    Event string `json:"event"`
    Arg   struct {
        Channel string `json:"channel"`
        InstId  string `json:"instId"`
    } `json:"arg"`
}
```

### 登录成功事件

```go
type Login struct {
    Event string `json:"event"`
    ConnId string `json:"connId"`
}
```

## 取消订阅

```go
// 取消订阅
err := client.Ws.Unsubscribe(false, []okex.ChannelName{"ticker"}, map[string]string{
    "instId": "BTC-USDT-SWAP",
})
```

## 完整示例

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/PineappleBond/TradingEino/backend/pkg/okex"
    "github.com/PineappleBond/TradingEino/backend/pkg/okex/api"
    "github.com/PineappleBond/TradingEino/backend/pkg/okex/events"
)

func main() {
    ctx := context.Background()
    client, err := api.NewClient(ctx, apiKey, secretKey, passphrase, okex.DemoServer)
    if err != nil {
        panic(err)
    }

    // 设置消息频道
    errChan := make(chan *events.Error)
    client.Ws.SetChannels(errChan, nil, nil, nil, nil)

    // 连接公共频道
    if err := client.Ws.Connect(false); err != nil {
        panic(err)
    }

    // 订阅 Ticker
    if err := client.Ws.Subscribe(false, []okex.ChannelName{"ticker"}, map[string]string{
        "instId": "BTC-USDT-SWAP",
    }); err != nil {
        panic(err)
    }

    // 接收消息
    for {
        select {
        case err := <-errChan:
            fmt.Printf("Error: %s\n", err.Msg)
        default:
            data := <-client.Ws.RawEventChan
            jsonData, _ := json.MarshalIndent(data, "", "  ")
            fmt.Printf("Received: %s\n", string(jsonData))
        }
    }
}
```

## 心跳机制

WebSocket 连接需要保持心跳：
- 服务器每 30 秒发送一次 ping
- 客户端需要在收到 ping 后回复 pong
- 如果 30 秒内没有通信，连接将被关闭

客户端库会自动处理心跳，无需手动处理。

## 断线重连

客户端库会自动尝试重连：
- 重连间隔：2 秒
- 重连次数：无限

可以通过监听 `DoneChan` 来检测断线：

```go
select {
case <-client.Ws.DoneChan:
    fmt.Println("Connection closed, reconnecting...")
    client.Ws.Connect(false)
}
```

## 注意事项

1. **鉴权要求**：私有频道需要登录后才能订阅
2. **连接限制**：每个 IP 最多 24 个连接
3. **订阅限制**：每个连接最多订阅 300 个频道
4. **消息长度**：单次发送的消息不能超过 4096 字节
5. **自动重连**：客户端会自动重连，但需要重新登录和订阅
6. **上下文管理**：使用 `context.WithCancel` 来控制连接生命周期

## 错误处理

常见的错误码：

| 错误码 | 说明 |
|--------|------|
| `60001` | 连接数超限 |
| `60002` | 频道订阅超限 |
| `60003` | 重复订阅 |
| `60004` | 订阅失败 |
| `60005` | 参数错误 |
| `60006` | 登录失败 |

## 参考资料

- [WebSocket API 官方文档](https://www.okex.com/docs-v5/en/#websocket-api)
- [连接说明](https://www.okex.com/docs-v5/en/#websocket-api-connect)
- [订阅说明](https://www.okex.com/docs-v5/en/#websocket-api-subscribe)
