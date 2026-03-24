# TradeData API - 交易大数据

## 概述

TradeData 模块提供交易大数据相关的 API，包括持仓分析、多空比、成交量等市场统计数据。

## 访问方式

```go
client.Rest.TradeData
```

## API 列表

### 1. GetSupportCoin - 获取支持的币种

获取交易大数据接口支持的币种列表。

**示例：**
```go
coins, err := client.Rest.TradeData.GetSupportCoin()
```

**响应字段：**
- `Coin` - 支持的币种列表

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trading-data-get-support-coin

---

### 2. GetTakerVolume - 获取主动成交量

获取买卖双方的主动成交量，显示资金流入流出情况。

**请求参数：**
```go
type GetTakerVolume struct {
    Ccy      string  // 币种
    InstType string  // 产品类型
    Period   string  // 统计周期
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/tradedata"

volume, err := client.Rest.TradeData.GetTakerVolume(requests.GetTakerVolume{
    Ccy:      "BTC",
    InstType: "SWAP",
    Period:   "1h",
})
```

**响应字段：**
- `Ccy` - 币种
- `BuyVol` - 主动买入成交量
- `SellVol` - 主动卖出成交量
- `Ts` - 时间戳

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trading-data-get-support-coin

---

### 3. GetMarginLendingRatio - 获取杠杆借贷比

获取货币对杠杆 quote currency 与基础资产的累计数据比率。

**请求参数：**
```go
type GetRatio struct {
    Ccy    string  // 币种
    Period string  // 统计周期
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/tradedata"

ratio, err := client.Rest.TradeData.GetMarginLendingRatio(requests.GetRatio{
    Ccy:    "BTC",
    Period: "1h",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trading-data-get-margin-lending-ratio

---

### 4. GetLongShortRatio - 获取多空持仓人数比

获取期货和永续合约中，净多头与净空头持仓人数的比率。

**请求参数：**
```go
type GetRatio struct {
    Ccy    string  // 币种
    Period string  // 统计周期
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/tradedata"

ratio, err := client.Rest.TradeData.GetLongShortRatio(requests.GetRatio{
    Ccy:    "BTC",
    Period: "1h",
})
```

**响应字段：**
- `LongAccount` - 净多头人数
- `ShortAccount` - 净空头人数
- `LongShortRatio` - 多空比
- `Ts` - 时间戳

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trading-data-get-long-short-ratio

---

### 5. GetContractsOpenInterestAndVolume - 获取合约持仓量和成交量

获取期货和永续合约的持仓总量和成交量。

**请求参数：**
```go
type GetRatio struct {
    Ccy    string  // 币种
    Period string  // 统计周期
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/tradedata"

data, err := client.Rest.TradeData.GetContractsOpenInterestAndVolume(requests.GetRatio{
    Ccy:    "BTC",
    Period: "1h",
})
```

**响应字段：**
- `Ccy` - 币种
- `Oi` - 持仓量
- `Vol` - 成交量
- `Ts` - 时间戳

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trading-data-get-contracts-open-interest-and-volume

---

### 6. GetOptionsOpenInterestAndVolume - 获取期权持仓量和成交量

获取期权的持仓总量和成交量。

**请求参数：**
```go
type GetRatio struct {
    Ccy    string  // 币种
    Period string  // 统计周期
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/tradedata"

data, err := client.Rest.TradeData.GetOptionsOpenInterestAndVolume(requests.GetRatio{
    Ccy:    "BTC",
    Period: "1h",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trading-data-get-options-open-interest-and-volume

---

### 7. GetPutCallRatio - 获取看涨/看跌比率

获取期权看涨和看跌的相对成交量比率，显示交易者的价格预期和波动率预期。

**请求参数：**
```go
type GetRatio struct {
    Ccy    string  // 币种
    Period string  // 统计周期
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/tradedata"

ratio, err := client.Rest.TradeData.GetPutCallRatio(requests.GetRatio{
    Ccy:    "BTC",
    Period: "1h",
})
```

**响应字段：**
- `CallOi` - 看涨期权持仓量
- `PutOi` - 看跌期权持仓量
- `CallVol` - 看涨期权成交量
- `PutVol` - 看跌期权成交量
- `PcRatio` - 看涨看跌比

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trading-data-get-put-call-ratio

---

### 8. GetOpenInterestAndVolumeExpiry - 获取不同到期日的持仓量和成交量

获取每个到期日的成交量和持仓量，用于查看哪些到期日最受欢迎。

**请求参数：**
```go
type GetRatio struct {
    Ccy    string  // 币种
    Period string  // 统计周期
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/tradedata"

data, err := client.Rest.TradeData.GetOpenInterestAndVolumeExpiry(requests.GetRatio{
    Ccy:    "BTC",
    Period: "1h",
})
```

**响应字段：**
- `ExpTime` - 到期时间
- `Oi` - 持仓量
- `Vol` - 成交量

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trading-data-get-open-interest-and-volume-expiry

---

### 9. GetOpenInterestAndVolumeStrike - 获取不同行权价的持仓量和成交量

获取每个行权价的持仓量和成交量，用于查看哪些行权价最受欢迎。

**请求参数：**
```go
type GetOpenInterestAndVolumeStrike struct {
    Ccy       string  // 币种
    ExpTime   string  // 到期时间
    Period    string  // 统计周期
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/tradedata"

data, err := client.Rest.TradeData.GetOpenInterestAndVolumeStrike(requests.GetOpenInterestAndVolumeStrike{
    Ccy:     "BTC",
    ExpTime: "20240628",
    Period:  "1h",
})
```

**响应字段：**
- `Strike` - 行权价
- `CallOi` - 看涨持仓量
- `PutOi` - 看跌持仓量
- `CallVol` - 看涨成交量
- `PutVol` - 看跌成交量

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trading-data-get-open-interest-and-volume-strike

---

### 10. GetTakerFlow - 获取主动大单成交量

获取期权看涨和看跌的主动大单成交量比率。

**请求参数：**
```go
type GetRatio struct {
    Ccy    string  // 币种
    Period string  // 统计周期
    ExpTime string // 可选，到期时间
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/tradedata"

flow, err := client.Rest.TradeData.GetTakerFlow(requests.GetRatio{
    Ccy:    "BTC",
    Period: "1h",
})
```

**响应字段：**
- `CallBuyVol` - 看涨主动买入成交量
- `CallSellVol` - 看涨主动卖出成交量
- `PutBuyVol` - 看跌主动买入成交量
- `PutSellVol` - 看跌主动卖出成交量

**文档：** https://www.okex.com/docs-v5/en/#rest-api-trading-data-get-taker-flow

---

## 响应结构示例

### GetLongShortRatio 响应

```go
type GetRatio struct {
    Code string `json:"code"`
    Msg  string `json:"msg"`
    Data []struct {
        Ccy            string `json:"ccy"`
        LongAccount    string `json:"longAccount"`
        ShortAccount   string `json:"shortAccount"`
        LongShortRatio string `json:"longShortRatio"`
        Ts             string `json:"ts"`
    } `json:"data"`
}
```

### GetPutCallRatio 响应

```go
type GetPutCallRatio struct {
    Code string `json:"code"`
    Msg  string `json:"msg"`
    Data []struct {
        Ccy      string `json:"ccy"`
        CallOi   string `json:"callOi"`
        PutOi    string `json:"putOi"`
        CallVol  string `json:"callVol"`
        PutVol   string `json:"putVol"`
        PcRatio  string `json:"pcRatio"`
        Ts       string `json:"ts"`
    } `json:"data"`
}
```

---

## 注意事项

1. **无需鉴权**：TradeData API 都是公开数据，无需签名
2. **数据更新**：统计数据按周期更新，注意选择合适的周期
3. **历史数据**：部分接口仅返回有限的历史数据
4. **时间戳**：所有时间戳均为毫秒
