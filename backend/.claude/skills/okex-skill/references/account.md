# Account API - 账户管理

## 概述

Account 模块提供账户相关的 API，包括账户余额、持仓、杠杆、账单等功能。

## 访问方式

```go
client.Rest.Account
```

## API 列表

### 1. GetBalance - 获取账户余额

获取账户中所有资产（非零余额）的可用余额和冻结余额。

**请求参数：**
```go
type GetBalance struct {
    Ccy []string  // 可选，币种列表，逗号分隔
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

// 获取所有币种余额
balance, err := client.Rest.Account.GetBalance(requests.GetBalance{})

// 获取特定币种余额
balance, err := client.Rest.Account.GetBalance(requests.GetBalance{
    Ccy: []string{"BTC", "USDT"},
})
```

**响应字段：**
- `Id` - 账户 ID
- `TotEq` - 账户权益（美元）
- `Details` - 各币种详情
  - `Ccy` - 币种
  - `Eq` - 余额
  - `AvailEq` - 可用余额
  - `FrozenBal` - 冻结余额

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-get-balance

---

### 2. GetPositions - 获取持仓信息

获取当前账户的持仓信息，支持净模式和多空模式。

**请求参数：**
```go
type GetPositions struct {
    InstID []string  // 可选，产品 ID 列表
    PosID  []string  // 可选，持仓 ID 列表
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

// 获取所有持仓
positions, err := client.Rest.Account.GetPositions(requests.GetPositions{})

// 获取特定产品持仓
positions, err := client.Rest.Account.GetPositions(requests.GetPositions{
    InstID: []string{"BTC-USDT-SWAP"},
})
```

**响应字段：**
- `InstId` - 产品 ID
- `PosSide` - 持仓方向（long/short/net）
- `Pos` - 持仓数量
- `AvgPx` - 平均开仓价
- `UplRatio` - 收益率
- `Lever` - 杠杆倍数
- `Margin` - 保证金

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-get-positions

---

### 3. GetConfig - 获取账户配置

获取当前账户的配置信息。

**示例：**
```go
config, err := client.Rest.Account.GetConfig()
```

**响应字段：**
- `Uid` - 用户 ID
- `Label` - 用户标签
- `RoleType` - 账户类型
- `MainUid` - 主账户 ID
- `GreeksType` - 希腊字母显示类型（PA/PB）
- `Level` - 账户等级

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-get-account-configuration

---

### 4. SetPositionMode - 设置持仓模式

设置 FUTURES 和 SWAP 的持仓模式（多空模式或净模式）。

**请求参数：**
```go
type SetPositionMode struct {
    PosMode string  // "long_short_mode" 或 "net_mode"
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

// 设置为多空模式
result, err := client.Rest.Account.SetPositionMode(requests.SetPositionMode{
    PosMode: "long_short_mode",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-set-position-mode

---

### 5. SetLeverage - 设置杠杆倍数

设置产品的杠杆倍数。

**请求参数：**
```go
type SetLeverage struct {
    InstId  string  // 产品 ID
    Lever   string  // 杠杆倍数
    MgnMode string  // 保证金模式：cross, isolated
    PosSide string  // 可选，持仓方向（仅适用于多空模式）
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

// 设置 10 倍杠杆
result, err := client.Rest.Account.SetLeverage(requests.SetLeverage{
    InstId:  "BTC-USDT-SWAP",
    Lever:   "10",
    MgnMode: "cross",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-set-leverage

---

### 6. GetLeverage - 获取杠杆倍数

获取产品的当前杠杆倍数。

**请求参数：**
```go
type GetLeverage struct {
    InstID  []string  // 产品 ID 列表
    MgnMode string    // 保证金模式
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

leverage, err := client.Rest.Account.GetLeverage(requests.GetLeverage{
    InstID:  []string{"BTC-USDT-SWAP"},
    MgnMode: "cross",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-get-leverage

---

### 7. GetBills - 获取账单明细

获取账户的账单明细，包括所有导致余额变动的交易记录。

**请求参数：**
```go
type GetBills struct {
    Begin   string  // 开始时间（毫秒时间戳）
    End     string  // 结束时间（毫秒时间戳）
    Limit   string  // 分页数量，最大 100
    Ccy     string  // 可选，币种
    Type    string  // 可选，账单类型
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

// 获取最近 7 天的账单
bills, err := client.Rest.Account.GetBills(requests.GetBills{
    Begin: "1711267200000",
    End:   "1711872000000",
    Limit: "100",
}, false)  // false = 最近 7 天，true = 最近 3 个月

// 获取归档账单
bills, err := client.Rest.Account.GetBills(requests.GetBills{}, true)
```

**文档：**
- 最近 7 天：https://www.okex.com/docs-v5/en/#rest-api-account-get-bills-details-last-7-days
- 最近 3 个月：https://www.okex.com/docs-v5/en/#rest-api-account-get-bills-details-last-3-months

---

### 8. IncreaseDecreaseMargin - 调整保证金

增加或减少孤立持仓的保证金。

**请求参数：**
```go
type IncreaseDecreaseMargin struct {
    InstId  string  // 产品 ID
    PosSide string  // 持仓方向
    Type    string  // "add" 或 "reduce"
    Chg     string  // 保证金变动数量
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

// 增加保证金
result, err := client.Rest.Account.IncreaseDecreaseMargin(requests.IncreaseDecreaseMargin{
    InstId:  "BTC-USDT-SWAP",
    PosSide: "long",
    Type:    "add",
    Chg:     "100",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-increase-decrease-margin

---

### 9. GetMaxBuySellAmount - 获取最大可买卖数量

获取账户的最大可买卖数量。

**请求参数：**
```go
type GetMaxBuySellAmount struct {
    InstID    string  // 产品 ID
    TdMode    string  // 交易模式：cross, isolated, cash
    Ccy       string  // 可选，币种
    Px        string  // 可选，价格
    Leverage  string  // 可选，杠杆
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

maxSize, err := client.Rest.Account.GetMaxBuySellAmount(requests.GetMaxBuySellAmount{
    InstID: "BTC-USDT-SWAP",
    TdMode: "cross",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-get-maximum-buy-sell-amount-or-open-amount

---

### 10. GetFeeRates - 获取费率

获取账户的交易费率。

**请求参数：**
```go
type GetFeeRates struct {
    InstType string  // 产品类型：SPOT, MARGIN, SWAP, FUTURES, OPTION
    InstID   string  // 可选，产品 ID
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

fees, err := client.Rest.Account.GetFeeRates(requests.GetFeeRates{
    InstType: "SWAP",
})
```

**响应字段：**
- `Category` - 用户等级
- `Level` - 费率等级
- `Maker` - Maker 费率
- `Taker` - Taker 费率

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-get-fee-rates

---

### 11. GetInterestAccrued - 获取应计利息

获取账户的应计利息信息。

**请求参数：**
```go
type GetInterestAccrued struct {
    InstID   string  // 产品 ID
    Ccy      string  // 币种
    MgnMode  string  // 保证金模式
    Begin    string  // 开始时间
    End      string  // 结束时间
    Limit    string  // 分页数量
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

interest, err := client.Rest.Account.GetInterestAccrued(requests.GetInterestAccrued{
    Ccy:     "USDT",
    MgnMode: "cross",
    Limit:   "100",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-get-interest-accrued

---

### 12. GetInterestRates - 获取利率

获取当前杠杆借贷利率。

**请求参数：**
```go
type GetBalance struct {
    Ccy []string  // 可选，币种列表
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

rates, err := client.Rest.Account.GetInterestRates(requests.GetBalance{
    Ccy: []string{"USDT", "BTC"},
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-get-interest-rate

---

### 13. GetMaxLoan - 获取最大可借数量

获取账户的最大可借数量。

**请求参数：**
```go
type GetMaxLoan struct {
    InstID  string  // 产品 ID
    MgnMode string  // 保证金模式
    Ccy     string  // 币种
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

maxLoan, err := client.Rest.Account.GetMaxLoan(requests.GetMaxLoan{
    InstID:  "BTC-USDT-SWAP",
    MgnMode: "cross",
    Ccy:     "USDT",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-get-the-maximum-loan-of-instrument

---

### 14. GetMaxWithdrawals - 获取最大可提现数量

获取账户的最大可提现数量。

**请求参数：**
```go
type GetBalance struct {
    Ccy []string  // 可选，币种列表
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

maxWithdrawal, err := client.Rest.Account.GetMaxWithdrawals(requests.GetBalance{
    Ccy: []string{"USDT"},
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-get-maximum-withdrawals

---

### 15. SetGreeks - 设置希腊字母显示类型

设置期权希腊字母的显示类型。

**请求参数：**
```go
type SetGreeks struct {
    GreekType string  // "PA" (币本位) 或 "PB" (U 本位)
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

result, err := client.Rest.Account.SetGreeks(requests.SetGreeks{
    GreekType: "PA",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-set-greeks-m-bs

---

### 16. GetAccountAndPositionRisk - 获取账户和持仓风险

获取账户和持仓的风险信息。

**请求参数：**
```go
type GetAccountAndPositionRisk struct {
    InstType string  // 产品类型
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

risk, err := client.Rest.Account.GetAccountAndPositionRisk(requests.GetAccountAndPositionRisk{
    InstType: "SWAP",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-get-account-and-position-risk

---

### 17. GetMaxAvailableTradeAmount - 获取最大可交易数量

获取账户的最大可交易数量。

**请求参数：**
```go
type GetMaxAvailableTradeAmount struct {
    InstID  string  // 产品 ID
    MgnMode string  // 保证金模式
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/account"

maxAmount, err := client.Rest.Account.GetMaxAvailableTradeAmount(requests.GetMaxAvailableTradeAmount{
    InstID:  "BTC-USDT-SWAP",
    MgnMode: "cross",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-account-get-maximum-available-tradable-amount

---

## 响应结构示例

### GetBalance 响应

```go
type GetBalance struct {
    Code string `json:"code"`
    Msg  string `json:"msg"`
    Data []struct {
        Id      string `json:"id"`
        TotEq   string `json:"totEq"`
        Details []struct {
            Ccy       string `json:"ccy"`
            Eq        string `json:"eq"`
            AvailEq   string `json:"availEq"`
            FrozenBal string `json:"frozenBal"`
        } `json:"details"`
    } `json:"data"`
}
```

### GetPositions 响应

```go
type GetPositions struct {
    Code string `json:"code"`
    Msg  string `json:"msg"`
    Data []struct {
        InstId    string `json:"instId"`
        PosSide   string `json:"posSide"`
        Pos       string `json:"pos"`
        AvgPx     string `json:"avgPx"`
        UplRatio  string `json:"uplRatio"`
        Lever     string `json:"lever"`
        Margin    string `json:"margin"`
    } `json:"data"`
}
```

---

## 注意事项

1. **鉴权要求**：所有 Account API 都需要私钥签名
2. **速率限制**：注意 API 的调用频率限制
3. **时间戳**：签名时需要使用 UTC 时间，格式为 `2006-01-02T15:04:05.999Z07:00`
4. **参数验证**：确保所有必填参数都已提供
5. **错误处理**：始终检查响应的 `Code` 字段，非 "0" 表示错误
