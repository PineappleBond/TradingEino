# PublicData API - 公共数据

## 概述

PublicData 模块提供公共数据相关的 API，包括合约信息、资金费率、持仓总量等市场公开信息。

## 访问方式

```go
client.Rest.PublicData
```

## API 列表

### 1. GetInstruments - 获取合约信息

获取所有可交易产品的详细信息。

**请求参数：**
```go
type GetInstruments struct {
    InstType string  // 产品类型：SPOT, MARGIN, SWAP, FUTURES, OPTION
    Uly      string  // 可选，标的指数（仅适用于合约/期权）
    InstID   string  // 可选，产品 ID
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/public"

// 获取所有 Swap 产品
instruments, err := client.Rest.PublicData.GetInstruments(requests.GetInstruments{
    InstType: "SWAP",
})

// 获取所有现货产品
instruments, err := client.Rest.PublicData.GetInstruments(requests.GetInstruments{
    InstType: "SPOT",
})
```

**响应字段：**
- `InstId` - 产品 ID
- `InstType` - 产品类型
- `CtVal` - 合约面值
- `CtMult` - 合约乘数
- `LotSz` - 下单精度
- `TickSz` - 价格精度
- `MinSz` - 最小下单数量
- `State` - 产品状态

**文档：** https://www.okex.com/docs-v5/en/#rest-api-public-data-get-instruments

---

### 2. GetFundingRate - 获取资金费率

获取当前资金费率。

**请求参数：**
```go
type GetFundingRate struct {
    InstID string  // 产品 ID
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/public"

rate, err := client.Rest.PublicData.GetFundingRate(requests.GetFundingRate{
    InstID: "BTC-USDT-SWAP",
})
```

**响应字段：**
- `InstId` - 产品 ID
- `FundingRate` - 当前资金费率
- `NextFundingTime` - 下次资金费时间

**文档：** https://www.okx.com/docs-v5/zh/#public-data-rest-api-get-funding-rate

---

### 3. GetOpenInterest - 获取持仓总量

获取产品的持仓总量。

**请求参数：**
```go
type GetOpenInterest struct {
    InstType string  // 产品类型
    InstID   string  // 可选，产品 ID
    Uly      string  // 可选，标的指数
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/public"

oi, err := client.Rest.PublicData.GetOpenInterest(requests.GetOpenInterest{
    InstType: "SWAP",
    InstID:   "BTC-USDT-SWAP",
})
```

**响应字段：**
- `InstId` - 产品 ID
- `OiCcy` - 持仓量（币）
- `OiUsd` - 持仓量（USD）

**文档：** https://www.okex.com/docs-v5/en/#rest-api-public-data-get-open-interest

---

### 4. GetMarkPrice - 获取标记价格

获取产品的标记价格。

**请求参数：**
```go
type GetMarkPrice struct {
    InstType string  // 产品类型
    InstID   string  // 可选，产品 ID
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/public"

markPrice, err := client.Rest.PublicData.GetMarkPrice(requests.GetMarkPrice{
    InstType: "SWAP",
    InstID:   "BTC-USDT-SWAP",
})
```

**响应字段：**
- `InstId` - 产品 ID
- `MarkPx` - 标记价格
- `Ts` - 时间戳

**文档：** https://www.okex.com/docs-v5/en/#rest-api-public-data-get-mark-price

---

### 5. GetLimitPrice - 获取限价信息

获取产品的最高买入限价和最低卖出限价。

**请求参数：**
```go
type GetLimitPrice struct {
    InstID string  // 产品 ID
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/public"

limit, err := client.Rest.PublicData.GetLimitPrice(requests.GetLimitPrice{
    InstID: "BTC-USDT-SWAP",
})
```

**响应字段：**
- `InstId` - 产品 ID
- `Buy` - 最高买入限价
- `Sell` - 最低卖出限价

**文档：** https://www.okex.com/docs-v5/en/#rest-api-public-data-get-limit-price

---

### 6. GetOptionMarketData - 获取期权行情

获取期权市场的行情数据。

**请求参数：**
```go
type GetOptionMarketData struct {
    InstID string  // 可选，产品 ID
    Uly    string  // 标的指数
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/public"

data, err := client.Rest.PublicData.GetOptionMarketData(requests.GetOptionMarketData{
    Uly: "BTC-USD",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-public-data-get-option-market-data

---

### 7. GetEstimatedDeliveryExercisePrice - 获取预计交割/行权价格

获取期权/期货的预计交割价格或行权价格。

**请求参数：**
```go
type GetEstimatedDeliveryExercisePrice struct {
    InstType string  // 产品类型
    InstID   string  // 可选，产品 ID
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/public"

price, err := client.Rest.PublicData.GetEstimatedDeliveryExercisePrice(requests.GetEstimatedDeliveryExercisePrice{
    InstType: "FUTURES",
    InstID:   "BTC-USDT-240628",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-public-data-get-estimated-delivery-Exercise-price

---

### 8. GetDiscountRateAndInterestFreeQuota - 获取折扣率和免息额度

获取杠杆交易的折扣率和免息额度信息。

**请求参数：**
```go
type GetDiscountRateAndInterestFreeQuota struct {
    Ccy string  // 币种
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/public"

quota, err := client.Rest.PublicData.GetDiscountRateAndInterestFreeQuota(requests.GetDiscountRateAndInterestFreeQuota{
    Ccy: "BTC",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-public-data-get-discount-rate-and-interest-free-quota

---

### 9. GetSystemTime - 获取系统时间

获取 API 服务器时间。

**示例：**
```go
time, err := client.Rest.PublicData.GetSystemTime()
```

**响应字段：**
- `Ts` - 服务器时间戳

**文档：** https://www.okex.com/docs-v5/en/#rest-api-public-data-get-system-time

---

### 10. GetLiquidationOrders - 获取强平订单

获取最近 7 天的强平订单记录。

**请求参数：**
```go
type GetLiquidationOrders struct {
    InstType string  // 产品类型
    InstID   string  // 可选，产品 ID
    Alias    string  // 可选，合约类型
    State    string  // 可选，状态
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/public"

orders, err := client.Rest.PublicData.GetLiquidationOrders(requests.GetLiquidationOrders{
    InstType: "SWAP",
    InstID:   "BTC-USDT-SWAP",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-public-data-get-liquidation-orders

---

### 11. GetPositionTiers - 获取持仓梯度

获取产品的持仓梯度和保证金率信息。

**请求参数：**
```go
type GetPositionTiers struct {
    InstType string  // 产品类型
    Uly      string  // 标的指数
    InstID   string  // 可选，产品 ID
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/public"

tiers, err := client.Rest.PublicData.GetPositionTiers(requests.GetPositionTiers{
    InstType: "SWAP",
    Uly:      "BTC-USDT",
})
```

**响应字段：**
- `InstId` - 产品 ID
- `Tier` - 梯度
- `Mmr` - 维持保证金率
- `Imr` - 初始保证金率
- `MaxLoan` - 最大可借数量

**文档：** https://www.okex.com/docs-v5/en/#rest-api-public-data-get-position-tiers

---

### 12. GetInterestRateAndLoanQuota - 获取利率和借贷额度

获取杠杆交易的利率和借贷额度。

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/public"

rates, err := client.Rest.PublicData.GetInterestRateAndLoanQuota()
```

**响应字段：**
- `Ccy` - 币种
- `Rate` - 利率
- `Quota` - 借贷额度

**文档：** https://www.okex.com/docs-v5/en/#rest-api-public-data-get-position-tiers

---

### 13. GetUnderlying - 获取标的指数

获取期权产品的标的指数信息。

**请求参数：**
```go
type GetUnderlying struct {
    InstType string  // 产品类型：OPTION
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/public"

underlyings, err := client.Rest.PublicData.GetUnderlying(requests.GetUnderlying{
    InstType: "OPTION",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-public-data-get-underlying

---

### 14. GetDeliveryExerciseHistory - 获取交割/行权历史

获取期权/期货的交割/行权历史记录。

**请求参数：**
```go
type GetDeliveryExerciseHistory struct {
    InstType string  // 产品类型
    InstID   string  // 可选，产品 ID
    After    string  // 可选，时间戳
    Before   string  // 可选，时间戳
    Limit    string  // 可选，数量
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/public"

history, err := client.Rest.PublicData.GetDeliveryExerciseHistory(requests.GetDeliveryExerciseHistory{
    InstType: "OPTION",
    Limit:    "100",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-public-data-get-instruments

---

## 响应结构示例

### GetInstruments 响应

```go
type GetInstruments struct {
    Code string `json:"code"`
    Msg  string `json:"msg"`
    Data []struct {
        InstId    string `json:"instId"`
        InstType  string `json:"instType"`
        CtVal     string `json:"ctVal"`
        CtMult    string `json:"ctMult"`
        LotSz     string `json:"lotSz"`
        TickSz    string `json:"tickSz"`
        MinSz     string `json:"minSz"`
        State     string `json:"state"`
    } `json:"data"`
}
```

### GetFundingRate 响应

```go
type GetFundingRate struct {
    Code string `json:"code"`
    Msg  string `json:"msg"`
    Data []struct {
        InstId            string `json:"instId"`
        FundingRate       string `json:"fundingRate"`
        NextFundingTime   string `json:"nextFundingTime"`
    } `json:"data"`
}
```

---

## 注意事项

1. **无需鉴权**：PublicData API 都是公开数据，无需签名
2. **速率限制**：注意 API 的调用频率限制
3. **数据更新**：标记价格、资金费率等数据会实时更新
4. **时间戳**：所有时间戳均为毫秒
