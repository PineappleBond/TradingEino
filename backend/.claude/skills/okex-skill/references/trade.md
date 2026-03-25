# Trade API - 交易执行

## 概述

Trade 模块提供交易相关的 API，包括下单、撤单、订单查询、算法交易等功能。

## 访问方式

```go
client.Rest.Trade
```

## API 列表

### 1. PlaceOrder - 下单

下达新订单，支持市价单、限价单等多种订单类型。

**请求参数：**
```go
type PlaceOrder struct {
    InstID    string  // 产品 ID
    TdMode    string  // 交易模式：cross, isolated, cash
    Side      string  // 订单方向：buy, sell
    OrdType   string  // 订单类型：market, limit, post_only, fok, ioc
    Sz        string  // 下单数量
    Px        string  // 可选，限价单价格
    PosSide   string  // 可选，持仓方向：long, short（仅适用于多空模式）
    ClOrdID   string  // 可选，客户端订单 ID
    Tag       string  // 可选，订单标签
    ReduceOnly string // 可选，只减仓：true, false
    TpTriggerPx string // 可选，止盈触发价
    TpOrdPx   string   // 可选，止盈委托价
    SlTriggerPx string // 可选，止损触发价
    SlOrdPx   string   // 可选，止损委托价
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"

// 市价单
result, err := client.Rest.Trade.PlaceOrder([]requests.PlaceOrder{{
    InstID:  "BTC-USDT-SWAP",
    TdMode:  "cross",
    Side:    "buy",
    OrdType: "market",
    Sz:      "100",  // USDT 金额
}})

// 限价单
result, err := client.Rest.Trade.PlaceOrder([]requests.PlaceOrder{{
    InstID:  "BTC-USDT-SWAP",
    TdMode:  "cross",
    Side:    "buy",
    OrdType: "limit",
    Sz:      "0.01",  // BTC 数量
    Px:      "50000", // 价格
}})
```

**响应字段：**
- `OrdId` - 订单 ID
- `ClOrdId` - 客户端订单 ID
- `SCode` - 状态码
- `SMsg` - 状态消息

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trade-place-order

---

### 2. PlaceMultipleOrders - 批量下单

批量下达订单，单次最多支持 20 个订单。

**请求参数：**
```go
[]PlaceOrder  // PlaceOrder 数组
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"

orders := []requests.PlaceOrder{
    {
        InstID:  "BTC-USDT-SWAP",
        TdMode:  "cross",
        Side:    "buy",
        OrdType: "market",
        Sz:      "100",
    },
    {
        InstID:  "ETH-USDT-SWAP",
        TdMode:  "cross",
        Side:    "buy",
        OrdType: "market",
        Sz:      "100",
    },
}

result, err := client.Rest.Trade.PlaceMultipleOrders(orders)
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trade-place-multiple-orders

---

### 3. CancelOrder - 撤单

撤销未完成的订单。

**请求参数：**
```go
type CancelOrder struct {
    InstID  string  // 产品 ID
    OrdId   string  // 订单 ID
    ClOrdId string  // 可选，客户端订单 ID
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"

// 撤销单个订单
result, err := client.Rest.Trade.CancelOrder([]requests.CancelOrder{{
    InstID: "BTC-USDT-SWAP",
    OrdId:  "123456789",
}})

// 批量撤销（多个订单）
result, err := client.Rest.Trade.CancelOrder([]requests.CancelOrder{
    {InstID: "BTC-USDT-SWAP", OrdId: "123456789"},
    {InstID: "ETH-USDT-SWAP", OrdId: "987654321"},
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trade-cancel-order

---

### 4. AmendOrder - 修改订单

修改未完成的订单。

**请求参数：**
```go
type OrderList struct {
    InstID    string  // 产品 ID
    OrdId     string  // 订单 ID
    ClOrdId   string  // 可选，客户端订单 ID
    ReqSz     string  // 可选，新数量
    ReqPx     string  // 可选，新价格
    NewSz     string  // 可选，新数量（替代 ReqSz）
    NewPx     string  // 可选，新价格（替代 ReqPx）
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"

// 修改订单价格
result, err := client.Rest.Trade.AmendOrder([]requests.OrderList{{
    InstID:  "BTC-USDT-SWAP",
    OrdId:   "123456789",
    NewPx:   "51000",
}})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trade-amend-order

---

### 5. ClosePosition - 全平仓

通过市价单全平仓某个产品的所有持仓。

**请求参数：**
```go
type ClosePosition struct {
    InstID  string  // 产品 ID
    PosSide string  // 持仓方向：long, short（多空模式必填）
    MgnMode string  // 保证金模式：cross, isolated
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"

// 平掉所有 BTC-USDT-SWAP 多头持仓
result, err := client.Rest.Trade.ClosePosition(requests.ClosePosition{
    InstID:  "BTC-USDT-SWAP",
    PosSide: "long",
    MgnMode: "cross",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trade-close-positions

---

### 6. GetOrderDetail - 获取订单详情

获取订单的详细信息。

**请求参数：**
```go
type OrderDetails struct {
    InstID  string  // 产品 ID
    OrdId   string  // 订单 ID
    ClOrdId string  // 可选，客户端订单 ID
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"

order, err := client.Rest.Trade.GetOrderDetail(requests.OrderDetails{
    InstID: "BTC-USDT-SWAP",
    OrdId:  "123456789",
})
```

**响应字段：**
- `InstId` - 产品 ID
- `Side` - 订单方向
- `OrdType` - 订单类型
- `Sz` - 下单数量
- `Px` - 委托价格
- `AvgPx` - 平均成交价
- `State` - 订单状态
- `FillPx` - 最新成交价
- `FillSz` - 累计成交量

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trade-get-order-details

---

### 7. GetOrderList - 获取未完成订单

获取当前账户所有未完成的订单。

**请求参数：**
```go
type OrderList struct {
    InstType string  // 产品类型
    InstID   string  // 可选，产品 ID
    OrdType  string  // 可选，订单类型
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"

orders, err := client.Rest.Trade.GetOrderList(requests.OrderList{
    InstType: "SWAP",
    InstID:   "BTC-USDT-SWAP",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trade-get-order-list

---

### 8. GetOrderHistory - 获取订单历史

获取已完成的订单历史数据。

**请求参数：**
```go
type OrderList struct {
    InstType string  // 产品类型
    InstID   string  // 可选，产品 ID
    OrdType  string  // 可选，订单类型
    Begin    string  // 开始时间
    End      string  // 结束时间
    Limit    string  // 分页数量
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"

// 获取最近 7 天的订单历史
orders, err := client.Rest.Trade.GetOrderHistory(requests.OrderList{
    InstType: "SWAP",
    InstID:   "BTC-USDT-SWAP",
    Limit:    "100",
}, false)  // false = 最近 7 天，true = 最近 3 个月

// 获取最近 3 个月的订单历史
orders, err := client.Rest.Trade.GetOrderHistory(requests.OrderList{
    InstType: "SWAP",
    Begin:    "1711267200000",
    End:      "1711872000000",
}, true)
```

**文档：**
- 最近 7 天：https://www.okex.com/docs-v5/en/#rest-api-trade-get-order-history-last-7-days
- 最近 3 个月：https://www.okex.com/docs-v5/en/#rest-api-trade-get-order-history-last-3-months

---

### 9. GetTransactionDetails - 获取成交明细

获取最近 3 天内的成交明细。

**请求参数：**
```go
type TransactionDetails struct {
    InstType string  // 产品类型
    InstID   string  // 可选，产品 ID
    OrdId    string  // 可选，订单 ID
    Begin    string  // 开始时间
    End      string  // 结束时间
    Limit    string  // 分页数量
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"

// 获取最近 3 天的成交明细
fills, err := client.Rest.Trade.GetTransactionDetails(requests.TransactionDetails{
    InstType: "SWAP",
    Limit:    "100",
}, false)  // false = 最近 3 天，true = 最近 3 个月

// 获取最近 3 个月的成交明细
fills, err := client.Rest.Trade.GetTransactionDetails(requests.TransactionDetails{
    InstType: "SWAP",
    Begin:    "1711267200000",
    End:      "1711872000000",
}, true)
```

**响应字段：**
- `TradeId` - 成交 ID
- `OrdId` - 订单 ID
- `InstId` - 产品 ID
- `Side` - 订单方向
- `FillPx` - 成交价
- `FillSz` - 成交量
- `Fee` - 手续费

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trade-get-transaction-details

---

### 10. PlaceAlgoOrder - 下单算法订单

下达算法订单，包括条件单、止盈止损单、冰山单、TWAP 单等。

**请求参数：**
```go
type PlaceAlgoOrder struct {
    InstID      string  // 产品 ID
    TdMode      string  // 交易模式
    Side        string  // 订单方向
    OrdType     string  // 订单类型
    Sz          string  // 委托数量
    PosSide     string  // 可选，持仓方向
    AlgoOrdType string  // 算法订单类型：conditional, oco, trigger, iceberg, twap

    // 条件单参数
    TpTriggerPx string  // 止盈触发价
    TpOrdPx     string  // 止盈委托价
    SlTriggerPx string  // 止损触发价
    SlOrdPx     string  // 止损委托价
    TriggerPx   string  // 条件单触发价

    // 冰山单参数
    PxVar       string  // 价格变动
    PxSpread    string  // 价格间距
    LimitPx     string  // 限价

    // TWAP 参数
    TimeInterval string  // 时间间隔
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"

// 条件单
result, err := client.Rest.Trade.PlaceAlgoOrder(requests.PlaceAlgoOrder{
    InstID:      "BTC-USDT-SWAP",
    TdMode:      "cross",
    Side:        "buy",
    OrdType:     "limit",
    Sz:          "0.01",
    AlgoOrdType: "conditional",
    TriggerPx:   "48000",  // 触发价
    Px:          "47900",  // 委托价
})

// 止盈止损单
result, err := client.Rest.Trade.PlaceAlgoOrder(requests.PlaceAlgoOrder{
    InstID:      "BTC-USDT-SWAP",
    TdMode:      "cross",
    Side:        "sell",
    OrdType:     "limit",
    Sz:          "0.01",
    PosSide:     "long",
    AlgoOrdType: "oco",
    TpTriggerPx: "55000",  // 止盈触发价
    TpOrdPx:     "54900",  // 止盈委托价
    SlTriggerPx: "45000",  // 止损触发价
    SlOrdPx:     "45100",  // 止损委托价
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trade-place-algo-order

---

### 11. CancelAlgoOrder - 撤销算法订单

撤销未触发的算法订单。

**请求参数：**
```go
type CancelAlgoOrder struct {
    AlgoIds []string  // 算法订单 ID 列表
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"

result, err := client.Rest.Trade.CancelAlgoOrder(requests.CancelAlgoOrder{
    AlgoIds: []string{"123456", "789012"},
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trade-cancel-algo-order

---

### 12. GetAlgoOrderList - 获取算法订单列表

获取当前账户未触发的算法订单。

**请求参数：**
```go
type AlgoOrderList struct {
    AlgoOrdType string  // 算法订单类型
    InstType    string  // 产品类型
    InstID      string  // 可选，产品 ID
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/trade"

// 获取未触发的条件单
orders, err := client.Rest.Trade.GetAlgoOrderList(requests.AlgoOrderList{
    AlgoOrdType: "conditional",
    InstType:    "SWAP",
}, false)  // false = 未触发，true = 历史
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trade-get-algo-order-list

---

## 订单类型说明

| 订单类型 | 说明 |
|---------|------|
| `market` | 市价单，以最优价格立即成交 |
| `limit` | 限价单，指定价格或更优价格成交 |
| `post_only` | 只做 maker 单，确保提供流动性 |
| `fok` | 全额成交或撤销，无法全部成交则全部撤销 |
| `ioc` | 立即成交或撤销，成交多少算多少 |

---

## 订单状态说明

| 状态 | 说明 |
|-----|------|
| `live` | 等待成交 |
| `partially_filled` | 部分成交 |
| `filled` | 完全成交 |
| `canceled` | 已撤销 |
| `mmp_canceled` | MNP 撤销 |

---

## 响应结构示例

### PlaceOrder 响应

```go
type PlaceOrder struct {
    Code string `json:"code"`
    Msg  string `json:"msg"`
    Data []struct {
        OrdId   string `json:"ordId"`
        ClOrdId string `json:"clOrdId"`
        SCode   string `json:"sCode"`
        SMsg    string `json:"sMsg"`
    } `json:"data"`
}
```

### GetOrderDetail 响应

```go
type OrderList struct {
    Code string `json:"code"`
    Msg  string `json:"msg"`
    Data []struct {
        InstId    string `json:"instId"`
        Side      string `json:"side"`
        OrdType   string `json:"ordType"`
        Sz        string `json:"sz"`
        Px        string `json:"px"`
        AvgPx     string `json:"avgPx"`
        State     string `json:"state"`
        FillPx    string `json:"fillPx"`
        FillSz    string `json:"fillSz"`
        Fee       string `json:"fee"`
        AccFillSz string `json:"accFillSz"`
    } `json:"data"`
}
```

---

## 注意事项

1. **鉴权要求**：所有 Trade API 都需要私钥签名
2. **订单数量限制**：单个账户最多 100 个未完成订单
3. **批量操作限制**：批量下单/撤单最多支持 20 个订单
4. **价格精度**：注意不同产品的价格精度和数量精度要求
5. **账户余额**：下单前确保有足够的可用余额
6. **错误处理**：始终检查响应的 `sCode` 字段，非 "0" 表示错误
