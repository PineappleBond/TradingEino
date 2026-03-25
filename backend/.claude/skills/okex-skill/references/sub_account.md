# SubAccount API - 子账户管理

## 概述

SubAccount 模块提供子账户管理相关的 API，包括子账户列表、APIKey 管理、余额查询、划转等功能。

## 访问方式

```go
client.Rest.SubAccount
```

**注意：** 子账户 API 仅主账户可用

## API 列表

### 1. ViewList - 查看子账户列表

获取主账户下的所有子账户列表。

**请求参数：**
```go
type ViewList struct {
    Enable     string  // 可选，是否启用：true/false
    SubAcct    string  // 可选，子账户名称
    Page       string  // 可选，页码
    Limit      string  // 可选，每页数量
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/subaccount"

// 获取所有子账户
accounts, err := client.Rest.SubAccount.ViewList(requests.ViewList{})

// 获取已启用的子账户
accounts, err := client.Rest.SubAccount.ViewList(requests.ViewList{
    Enable: "true",
    Limit:  "10",
})
```

**响应字段：**
- `SubAcct` - 子账户名称
- `Label` - 子账户标签
- `Enable` - 是否启用
- `CreateTime` - 创建时间
- `Mobile` - 手机号

**文档：** https://www.okex.com/docs-v5/en/#rest-api-subaccount-view-sub-account-list

---

### 2. CreateAPIKey - 创建子账户 APIKey

为主账户下的子账户创建 APIKey。

**请求参数：**
```go
type CreateAPIKey struct {
    SubAcct    string   // 子账户名称
    Label      string   // APIKey 标签
    Passphrase string   // APIKey 密码
    IP         []string // 可选，IP 白名单
    Access     string   // 可选，权限：read_only/trade
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/subaccount"

// 创建只读 APIKey
result, err := client.Rest.SubAccount.CreateAPIKey(requests.CreateAPIKey{
    SubAcct:    "subaccount1",
    Label:      "read-only-key",
    Passphrase: "MyPassphrase123",
    Access:     "read_only",
})

// 创建带 IP 白名单的交易 APIKey
result, err := client.Rest.SubAccount.CreateAPIKey(requests.CreateAPIKey{
    SubAcct:    "subaccount1",
    Label:      "trading-key",
    Passphrase: "MyPassphrase123",
    IP:         []string{"192.168.1.1", "192.168.1.2"},
    Access:     "trade",
})
```

**响应字段：**
- `SubAcct` - 子账户名称
- `Label` - APIKey 标签
- `ApiKey` - APIKey
- `SecretKey` - 密钥
- `Passphrase` - 密码
- `Access` - 权限

**文档：** https://www.okex.com/docs-v5/en/#rest-api-subaccount-create-an-apikey-for-a-sub-account

---

### 3. QueryAPIKey - 查询子账户 APIKey

查询子账户的 APIKey 信息。

**请求参数：**
```go
type QueryAPIKey struct {
    SubAcct string  // 子账户名称
    ApiKey  string  // APIKey
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/subaccount"

keyInfo, err := client.Rest.SubAccount.QueryAPIKey(requests.QueryAPIKey{
    SubAcct: "subaccount1",
    ApiKey:  "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
})
```

**响应字段：**
- `SubAcct` - 子账户名称
- `Label` - APIKey 标签
- `ApiKey` - APIKey
- `Access` - 权限
- `IP` - IP 白名单
- `CreateTime` - 创建时间

**文档：** https://www.okex.com/docs-v5/en/#rest-api-subaccount-query-the-apikey-of-a-sub-account

---

### 4. ResetAPIKey - 重置子账户 APIKey

重置子账户的 APIKey 密钥。

**请求参数：**
```go
type CreateAPIKey struct {
    SubAcct    string   // 子账户名称
    ApiKey     string   // 原 APIKey
    Label      string   // 新标签
    Passphrase string   // 新密码
    IP         []string // 可选，新 IP 白名单
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/subaccount"

result, err := client.Rest.SubAccount.ResetAPIKey(requests.CreateAPIKey{
    SubAcct:    "subaccount1",
    ApiKey:     "old-api-key",
    Label:      "new-label",
    Passphrase: "NewPassphrase123",
    IP:         []string{"10.0.0.1"},
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-subaccount-reset-the-apikey-of-a-sub-account

---

### 5. DeleteAPIKey - 删除子账户 APIKey

删除子账户的 APIKey。

**请求参数：**
```go
type DeleteAPIKey struct {
    SubAcct string  // 子账户名称
    ApiKey  string  // APIKey
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/subaccount"

result, err := client.Rest.SubAccount.DeleteAPIKey(requests.DeleteAPIKey{
    SubAcct: "subaccount1",
    ApiKey:  "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
})
```

**文档：** https://www.okex.com/docs-v5/en/#rest-api-subaccount-delete-the-apikey-of-sub-accounts

---

### 6. GetBalance - 查询子账户余额

查询子账户交易账户的余额。

**请求参数：**
```go
type GetBalance struct {
    SubAcct string  // 子账户名称
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/subaccount"

balance, err := client.Rest.SubAccount.GetBalance(requests.GetBalance{
    SubAcct: "subaccount1",
})
```

**响应字段：**
- `SubAcct` - 子账户名称
- `Details` - 各币种余额详情
  - `Ccy` - 币种
  - `Bal` - 余额
  - `AvailBal` - 可用余额

**文档：** https://www.okex.com/docs-v5/en/#rest-api-subaccount-get-sub-account-balance

---

### 7. HistoryTransfer - 查询子账户划转历史

查询主账户与子账户之间的划转历史。

**请求参数：**
```go
type HistoryTransfer struct {
    SubAcct string  // 子账户名称
    Ccy     string  // 可选，币种
    Type    string  // 可选，划转类型
    Begin   string  // 开始时间
    End     string  // 结束时间
    Limit   string  // 可选，分页数量
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/subaccount"

history, err := client.Rest.SubAccount.HistoryTransfer(requests.HistoryTransfer{
    SubAcct: "subaccount1",
    Ccy:     "USDT",
    Limit:   "100",
})
```

**响应字段：**
- `SubAcct` - 子账户名称
- `Ccy` - 币种
- `Amt` - 数量
- `Type` - 划转类型
- `State` - 状态
- `Ts` - 时间戳

**文档：** https://www.okex.com/docs-v5/en/#rest-api-subaccount-history-of-sub-account-transfer

---

### 8. ManageTransfers - 管理子账户划转

在主账户和子账户之间划转资金，或子账户之间划转。

**请求参数：**
```go
type ManageTransfers struct {
    Ccy       string  // 币种
    Amt       string  // 数量
    From      string  // 转出类型：6=主账户，18=子账户
    To        string  // 转入类型：6=主账户，18=子账户
    FromSubAcct string // 可选，转出的子账户名称
    ToSubAcct   string // 可选，转入的子账户名称
}
```

**示例：**
```go
import requests "github.com/PineappleBond/TradingEino/backend/pkg/okex/requests/rest/subaccount"

// 主账户划转到子账户
result, err := client.Rest.SubAccount.ManageTransfers(requests.ManageTransfers{
    Ccy:       "USDT",
    Amt:       "1000",
    From:      "6",
    To:        "18",
    ToSubAcct: "subaccount1",
})

// 子账户划转到主账户
result, err := client.Rest.SubAccount.ManageTransfers(requests.ManageTransfers{
    Ccy:         "USDT",
    Amt:         "1000",
    From:        "18",
    To:          "6",
    FromSubAcct: "subaccount1",
})

// 子账户之间划转
result, err := client.Rest.SubAccount.ManageTransfers(requests.ManageTransfers{
    Ccy:         "USDT",
    Amt:         "1000",
    From:        "18",
    To:          "18",
    FromSubAcct: "subaccount1",
    ToSubAcct:   "subaccount2",
})
```

**响应字段：**
- `TransId` - 划转 ID
- `Ccy` - 币种
- `Amt` - 数量
- `State` - 状态

**文档：** https://www.okex.com/docs-v5/en/#rest-api-subaccount-master-accounts-manage-the-transfers-between-sub-accounts

---

## 响应结构示例

### ViewList 响应

```go
type ViewList struct {
    Code string `json:"code"`
    Msg  string `json:"msg"`
    Data []struct {
        SubAcct    string `json:"subAcct"`
        Label      string `json:"label"`
        Enable     string `json:"enable"`
        CreateTime string `json:"createTime"`
        Mobile     string `json:"mobile"`
    } `json:"data"`
}
```

### CreateAPIKey 响应

```go
type APIKey struct {
    Code string `json:"code"`
    Msg  string `json:"msg"`
    Data []struct {
        SubAcct    string `json:"subAcct"`
        Label      string `json:"label"`
        ApiKey     string `json:"apiKey"`
        SecretKey  string `json:"secretKey"`
        Passphrase string `json:"passphrase"`
        Access     string `json:"access"`
    } `json:"data"`
}
```

---

## 注意事项

1. **主账户专用**：所有子账户 API 仅主账户可用
2. **鉴权要求**：所有 SubAccount API 都需要私钥签名
3. **APIKey 安全**：创建 APIKey 后，SecretKey 仅显示一次，请妥善保存
4. **划转限制**：子账户之间不能直接划转，需要通过主账户
5. **IP 白名单**：建议设置 IP 白名单以提高安全性
