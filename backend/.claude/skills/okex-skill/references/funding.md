# Funding API - 资金管理

## 概述

Funding 模块提供资金管理相关的 API，包括充值、提现、划转、账单等功能。

## 访问方式

```go
client.Rest.Funding
```

## API 列表

### 1. GetCurrencies - 获取币种列表

获取所有支持的币种信息。

**示例：**
```go
currencies, err := client.Rest.Funding.GetCurrencies()
```

**响应字段：**
- `Ccy` - 币种
- `Name` - 币种名称
- `Chain` - 链名称
- `MinWd` - 最小提现数量
- `MinDep` - 最小充值数量
- `MaxWd` - 最大提现数量
- `DepQuota` - 充值限额
- `DepUsed` - 已用充值额度

**文档：** https://www.okex.com/docs-v5/en/#rest-api-funding-get-currencies

---

### 2. GetBalance - 获取资金账户余额

获取资金账户的余额信息。

**请求参数：**
```go
type GetBalance struct {
    Ccy []string  // 可选，币种列表
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/funding"

// 获取所有币种余额
balance, err := client.Rest.Funding.GetBalance(requests.GetBalance{})

// 获取特定币种余额
balance, err := client.Rest.Funding.GetBalance(requests.GetBalance{
    Ccy: []string{"BTC", "USDT"},
})
```

**响应字段：**
- `Ccy` - 币种
- `Bal` - 余额
- `AvailBal` - 可用余额
- `FrozenBal` - 冻结余额

**文档：** https://www.okex.com/docs-v5/en/#rest-api-funding-get-balance

---

### 3. FundsTransfer - 资金划转

在资金账户和交易账户之间划转，或主账户与子账户之间划转。

**请求参数：**
```go
type FundsTransfer struct {
    Ccy       string  // 币种
    Amt       string  // 数量
    From      string  // 转出账户类型：6=资金账户，18=统一账户
    To        string  // 转入账户类型：18=统一账户，6=资金账户
    Type      string  // 划转类型：0=境内，1=跨境
    SubAcct   string  // 可选，子账户名称
    ToSubAcct string  // 可选，转入子账户名称
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/funding"

// 从资金账户划转到交易账户
result, err := client.Rest.Funding.FundsTransfer(requests.FundsTransfer{
    Ccy:  "USDT",
    Amt:  "1000",
    From: "6",
    To:   "18",
})

// 从主账户划转到子账户
result, err := client.Rest.Funding.FundsTransfer(requests.FundsTransfer{
    Ccy:     "USDT",
    Amt:     "1000",
    From:    "6",
    To:      "6",
    Type:    "1",
    SubAcct: "",
    ToSubAcct: "subaccount1",
})
```

**响应字段：**
- `TransId` - 划转 ID
- `Ccy` - 币种
- `Amt` - 数量
- `From` - 转出账户
- `To` - 转入账户
- `State` - 状态

**文档：** https://www.okex.com/docs-v5/en/#rest-api-funding-funds-transfer

---

### 4. AssetBillsDetails - 获取资金账单明细

获取资金账户的账单明细。

**请求参数：**
```go
type AssetBillsDetails struct {
    Ccy   string  // 可选，币种
    Type  string  // 可选，账单类型
    Begin string  // 开始时间
    End   string  // 结束时间
    Limit string  // 可选，分页数量
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/funding"

bills, err := client.Rest.Funding.AssetBillsDetails(requests.AssetBillsDetails{
    Ccy:   "USDT",
    Limit: "100",
})
```

**账单类型：**
- `1` - 转账
- `2` - 交易
- `3` - 交割
- `8` - 资金费率
- `11` - 自动兑换

**文档：** https://www.okex.com/docs-v5/en/#rest-api-funding-asset-bills-details

---

### 5. GetDepositAddress - 获取充值地址

获取币种的充值地址。

**请求参数：**
```go
type GetDepositAddress struct {
    Ccy string  // 币种
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/funding"

address, err := client.Rest.Funding.GetDepositAddress(requests.GetDepositAddress{
    Ccy: "USDT",
})
```

**响应字段：**
- `Ccy` - 币种
- `Chain` - 链名称
- `Addr` - 充值地址
- `To` - 到账方式

**文档：** https://www.okex.com/docs-v5/en/#rest-api-funding-get-deposit-address

---

### 6. GetDepositHistory - 获取充值历史

获取充值历史记录。

**请求参数：**
```go
type GetDepositHistory struct {
    Ccy   string  // 可选，币种
    State string  // 可选，状态
    Begin string  // 开始时间
    End   string  // 结束时间
    Limit string  // 可选，分页数量
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/funding"

deposits, err := client.Rest.Funding.GetDepositHistory(requests.GetDepositHistory{
    Ccy:   "USDT",
    Limit: "100",
})
```

**充值状态：**
- `0` - 等待确认
- `1` - 已到账
- `2` - 充值成功
- `8` - 暂停

**文档：** https://www.okex.com/docs-v5/en/#rest-api-funding-get-deposit-history

---

### 7. Withdrawal - 提现

发起提现请求。

**请求参数：**
```go
type Withdrawal struct {
    Ccy     string  // 币种
    Chain   string  // 链名称
    ToAddr  string  // 提现地址
    Amt     string  // 提现数量
    Dest    string  // 提现方式：3=内部，4=链上
    Fee     string  // 手续费
    AreaCode string // 可选，手机区号
    Info    string  // 可选，附加信息
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/funding"

// 链上提现
result, err := client.Rest.Funding.Withdrawal(requests.Withdrawal{
    Ccy:    "USDT",
    Chain:  "USDT-TRC20",
    ToAddr: "TxxxxxxxxxxxxxxxxxxxxxxxxxxxxB",
    Amt:    "100",
    Dest:   "4",
    Fee:    "1",
})

// 内部提现（转账到其他 OKX 用户）
result, err := client.Rest.Funding.Withdrawal(requests.Withdrawal{
    Ccy:    "USDT",
    ToAddr: "user@example.com",
    Amt:    "100",
    Dest:   "3",
    Fee:    "0",
})
```

**响应字段：**
- `WdId` - 提现 ID
- `Ccy` - 币种
- `Chain` - 链名称
- `Amt` - 数量
- `State` - 状态

**文档：** https://www.okex.com/docs-v5/en/#rest-api-funding-withdrawal

---

### 8. GetWithdrawalHistory - 获取提现历史

获取提现历史记录。

**请求参数：**
```go
type GetWithdrawalHistory struct {
    Ccy   string  // 可选，币种
    WdId  string  // 可选，提现 ID
    State string  // 可选，状态
    Begin string  // 开始时间
    End   string  // 结束时间
    Limit string  // 可选，分页数量
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/funding"

withdrawals, err := client.Rest.Funding.GetWithdrawalHistory(requests.GetWithdrawalHistory{
    Ccy:   "USDT",
    Limit: "100",
})
```

**提现状态：**
- `-3` - 待取消
- `-2` - 已取消
- `-1` - 失败
- `0` - 等待审核
- `1` - 审核中
- `2` - 已提现
- `3` - 等待邮件确认
- `4` - 等待人工审核
- `5` - 人工审核中

**文档：** https://www.okex.com/docs-v5/en/#rest-api-funding-get-withdrawal-history

---

### 9. GetPiggyBankBalance - 获取活期宝余额

获取活期宝（赚币宝）的余额信息。

**请求参数：**
```go
type GetPiggyBankBalance struct {
    Ccy []string  // 可选，币种列表
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/funding"

balance, err := client.Rest.Funding.GetPiggyBankBalance(requests.GetPiggyBankBalance{
    Ccy: []string{"USDT", "BTC"},
})
```

**响应字段：**
- `Ccy` - 币种
- `Bal` - 余额
- `Amt` - 在途金额
- `Income` - 累计收益

**文档：** https://www.okex.com/docs-v5/en/#rest-api-funding-get-piggybank-balance

---

### 10. PiggyBankPurchaseRedemption - 活期宝申购/赎回

申购或赎回到期活期宝。

**请求参数：**
```go
type PiggyBankPurchaseRedemption struct {
    Ccy   string  // 币种
    Amt   string  // 数量
    Side  string  // 操作类型：purchase=申购，redempt=赎回
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/funding"

// 申购
result, err := client.Rest.Funding.PiggyBankPurchaseRedemption(requests.PiggyBankPurchaseRedemption{
    Ccy:  "USDT",
    Amt:  "1000",
    Side: "purchase",
})

// 赎回
result, err := client.Rest.Funding.PiggyBankPurchaseRedemption(requests.PiggyBankPurchaseRedemption{
    Ccy:  "USDT",
    Amt:  "1000",
    Side: "redempt",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-funding-piggybank-purchase-redemption

---

## 响应结构示例

### GetBalance 响应

```go
type GetBalance struct {
    Code string `json:"code"`
    Msg  string `json:"msg"`
    Data []struct {
        Ccy       string `json:"ccy"`
        Bal       string `json:"bal"`
        AvailBal  string `json:"availBal"`
        FrozenBal string `json:"frozenBal"`
    } `json:"data"`
}
```

### FundsTransfer 响应

```go
type FundsTransfer struct {
    Code string `json:"code"`
    Msg  string `json:"msg"`
    Data []struct {
        TransId string `json:"transId"`
        Ccy     string `json:"ccy"`
        Amt     string `json:"amt"`
        State   string `json:"state"`
    } `json:"data"`
}
```

---

## 注意事项

1. **鉴权要求**：所有 Funding API 都需要私钥签名
2. **提现限制**：提现需要完成身份认证和资金密码验证
3. **划转限制**：主账户与子账户之间的划转需要正确设置账户类型
4. **费率说明**：链上提现需要支付网络手续费
5. **错误处理**：始终检查响应的 `Code` 字段，非 "0" 表示错误
