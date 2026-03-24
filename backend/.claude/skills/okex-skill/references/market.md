# Market API - 市场行情数据

## 概述

Market 模块提供市场行情相关的 API，包括 Ticker、深度、K 线、成交等数据。

## 访问方式

```go
client.Rest.Market
```

## API 列表

### 1. GetTickers - 获取所有产品 Ticker

获取所有产品的最新行情快照，包括 24 小时涨跌幅、成交量等。

**请求参数：**
```go
type GetTickers struct {
    InstType string  // 产品类型：SPOT, MARGIN, SWAP, FUTURES, OPTION
    Uly      string  // 可选，标的指数（仅适用于合约/期权）
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/market"

// 获取所有_swap 产品 Ticker
tickers, err := client.Rest.Market.GetTickers(requests.GetTickers{
    InstType: "SWAP",
})

// 获取所有现货 Ticker
tickers, err := client.Rest.Market.GetTickers(requests.GetTickers{
    InstType: "SPOT",
})
```

**响应字段：**
- `InstId` - 产品 ID
- `Last` - 最新价格
- `BidPx` / `BidSz` - 买一价/量
- `AskPx` / `AskSz` - 卖一价/量
- `Vol24h` - 24 小时成交量
- `VolCcy24h` - 24 小时成交额
- `High24h` / `Low24h` - 24 小时最高/最低价
- `Open24h` - 24 小时前价格
- `SodUtc24h` - UTC 时间 0 点价格

**文档：** https://www.okex.com/docs-v5/en/#rest-api-market-data-get-tickers

---

### 2. GetTicker - 获取单个产品 Ticker

获取单个产品的最新行情快照。

**请求参数：**
```go
type GetTickers struct {
    InstID string  // 产品 ID
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/market"

ticker, err := client.Rest.Market.GetTicker(requests.GetTickers{
    InstID: "BTC-USDT-SWAP",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-market-data-get-ticker

---

### 3. GetOrderBook - 获取深度数据

获取产品的订单簿深度数据。

**请求参数：**
```go
type GetOrderBook struct {
    InstID string  // 产品 ID
    Sz     string  // 可选，深度档位：1/5/10/20/40/50/100/200/400
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/market"

// 获取 40 档深度
orderBook, err := client.Rest.Market.GetOrderBook(requests.GetOrderBook{
    InstID: "BTC-USDT-SWAP",
    Sz:     "40",
})
```

**响应字段：**
- `Bids` - 买单深度 [[价格，数量], ...]
- `Asks` - 卖单深度 [[价格，数量], ...]
- `Ts` - 时间戳

**文档：** https://www.okex.com/docs-v5/en/#rest-api-market-data-get-order-book

---

### 4. GetOrderBookFull - 获取完整深度

获取产品的完整订单簿深度（所有档位）。

**请求参数：**
```go
type GetOrderBook struct {
    InstID string  // 产品 ID
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/market"

orderBook, err := client.Rest.Market.GetOrderBookFull(requests.GetOrderBook{
    InstID: "BTC-USDT-SWAP",
})
```

**文档：** https://www.okx.com/docs-v5/en/#order-book-trading-market-data-get-full-order-book

---

### 5. GetCandlesticks - 获取 K 线数据

获取产品的 K 线数据，最多返回 1440 条。

**请求参数：**
```go
type GetCandlesticks struct {
    InstID string  // 产品 ID
    Bar    string  // K 线类型：1m/3m/5m/15m/30m/1H/2H/4H/6H/12H/1D/1W/1M/3M/6M/1Y
    After  string  // 可选，返回早于指定时间的数据（毫秒时间戳）
    Before string  // 可选，返回晚于指定时间的数据（毫秒时间戳）
    Limit  string  // 可选，返回数量，最大 1440
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/market"

// 获取 BTC 1 小时 K 线，最多 100 条
candles, err := client.Rest.Market.GetCandlesticks(requests.GetCandlesticks{
    InstID: "BTC-USDT-SWAP",
    Bar:    "1H",
    Limit:  "100",
})
```

**响应字段：**
- `[0]` - 时间戳（毫秒）
- `[1]` - 开盘价
- `[2]` - 最高价
- `[3]` - 最低价
- `[4]` - 收盘价
- `[5]` - 成交量
- `[6]` - 成交额

**文档：** https://www.okex.com/docs-v5/en/#rest-api-market-data-get-candlesticks

---

### 6. GetCandlesticksHistory - 获取历史 K 线

获取历史年份的 K 线数据。

**请求参数：**
```go
type GetCandlesticks struct {
    InstID string  // 产品 ID
    Bar    string  // K 线类型
    After  string  // 可选
    Before string  // 可选
    Limit  string  // 可选
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/market"

candles, err := client.Rest.Market.GetCandlesticksHistory(requests.GetCandlesticks{
    InstID: "BTC-USDT-SWAP",
    Bar:    "1D",
    Limit:  "365",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-market-data-get-candlesticks

---

### 7. GetIndexCandlesticks - 获取指数 K 线

获取指数 K 线数据。

**请求参数：**
```go
type GetCandlesticks struct {
    InstID string  // 指数 ID
    Bar    string  // K 线类型
    After  string  // 可选
    Before string  // 可选
    Limit  string  // 可选
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/market"

candles, err := client.Rest.Market.GetIndexCandlesticks(requests.GetCandlesticks{
    InstID: "BTC-USDT",
    Bar:    "1H",
    Limit:  "100",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-market-data-get-index-candlesticks

---

### 8. GetMarkPriceCandlesticks - 获取标记价格 K 线

获取标记价格 K 线数据。

**请求参数：**
```go
type GetCandlesticks struct {
    InstID string  // 产品 ID
    Bar    string  // K 线类型
    After  string  // 可选
    Before string  // 可选
    Limit  string  // 可选
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/market"

candles, err := client.Rest.Market.GetMarkPriceCandlesticks(requests.GetCandlesticks{
    InstID: "BTC-USDT-SWAP",
    Bar:    "1H",
    Limit:  "100",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-market-data-get-mark-price-candlesticks

---

### 9. GetTrades - 获取最近成交

获取产品最近的成交记录。

**请求参数：**
```go
type GetTrades struct {
    InstID string  // 产品 ID
    Limit  string  // 可选，返回数量，最大 100
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/market"

trades, err := client.Rest.Market.GetTrades(requests.GetTrades{
    InstID: "BTC-USDT-SWAP",
    Limit:  "50",
})
```

**响应字段：**
- `TradeId` - 成交 ID
- `Px` - 成交价
- `Sz` - 成交量
- `Side` - 成交方向：buy/sell
- `Ts` - 时间戳

**文档：** https://www.okex.com/docs-v5/en/#rest-api-market-data-get-trades

---

### 10. GetIndexComponents - 获取指数成分股

获取指数的成分股信息。

**请求参数：**
```go
type GetIndexComponents struct {
    InstID string  // 指数 ID
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/market"

components, err := client.Rest.Market.GetIndexComponents(requests.GetIndexComponents{
    InstID: "BTC-USDT",
})
```

**响应字段：**
- `InstId` - 成分股 ID
- `Exch` - 交易所
- `Wgt` - 权重

**文档：** https://www.okex.com/docs-v5/en/#rest-api-market-data-get-index-components

---

### 11. Get24HTotalVolume - 获取 24 小时总成交量

获取平台 24 小时总成交量（以 USD 计价）。

**示例：**
```go
volume, err := client.Rest.Market.Get24HTotalVolume()
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-market-data-get-24h-total-volume

---

## K 线类型

| Bar | 周期 |
|-----|------|
| `1m` | 1 分钟 |
| `3m` | 3 分钟 |
| `5m` | 5 分钟 |
| `15m` | 15 分钟 |
| `30m` | 30 分钟 |
| `1H` | 1 小时 |
| `2H` | 2 小时 |
| `4H` | 4 小时 |
| `6H` | 6 小时 |
| `12H` | 12 小时 |
| `1D` | 1 天 |
| `1W` | 1 周 |
| `1M` | 1 月 |
| `3M` | 3 月 |
| `6M` | 6 月 |
| `1Y` | 1 年 |

---

## 响应结构示例

### GetTicker 响应

```go
type Ticker struct {
    Code string `json:"code"`
    Msg  string `json:"msg"`
    Data []struct {
        InstId      string `json:"instId"`
        Last        string `json:"last"`
        BidPx       string `json:"bidPx"`
        BidSz       string `json:"bidSz"`
        AskPx       string `json:"askPx"`
        AskSz       string `json:"askSz"`
        Vol24h      string `json:"vol24h"`
        VolCcy24h   string `json:"volCcy24h"`
        High24h     string `json:"high24h"`
        Low24h      string `json:"low24h"`
        Open24h     string `json:"open24h"`
        SodUtc24h   string `json:"sodUtc24h"`
    } `json:"data"`
}
```

### GetOrderBook 响应

```go
type OrderBook struct {
    Code string `json:"code"`
    Msg  string `json:"msg"`
    Data []struct {
        Bids [][]string `json:"bids"`  // [[price, size], ...]
        Asks [][]string `json:"asks"`  // [[price, size], ...]
        Ts   string     `json:"ts"`
    } `json:"data"`
}
```

### GetCandlesticks 响应

```go
type Candle struct {
    Code string `json:"code"`
    Msg  string `json:"msg"`
    Data [][]string `json:"data"`
    // 每个数组：[ts, o, h, l, c, vol, volCcy]
}
```

---

## 注意事项

1. **无需鉴权**：Market API 都是公开数据，无需签名
2. **速率限制**：注意 API 的调用频率限制
3. **数据精度**：价格和数量的精度因产品而异
4. **时间戳**：所有时间戳均为毫秒
5. **K 线限制**：单次最多返回 1440 条 K 线数据
