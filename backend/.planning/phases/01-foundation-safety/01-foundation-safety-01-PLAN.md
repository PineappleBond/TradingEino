---
phase: 01-foundation-safety
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - pkg/okex/okx_error.go
  - internal/agent/tools/okx_get_positions.go
  - internal/agent/tools/okx_get_fundingrate.go
  - internal/agent/tools/okx_candlesticks.go
autonomous: true
requirements:
  - FOUND-01
  - FOUND-02
must_haves:
  truths:
    - "API 工具调用失败时返回 ("", err) 格式"
    - "OKX API 错误 (Code != 0) 返回 *OKXError 类型"
    - "所有 API 工具调用前等待限流器 (limiter.Wait)"
    - "限流器配置符合端点类型 (Account 5 次/秒，Market/Public 10 次/秒)"
  artifacts:
    - path: "pkg/okex/okx_error.go"
      provides: "OKXError 统一错误类型"
      contains: "type OKXError struct"
    - path: "internal/agent/tools/okx_get_positions.go"
      provides: "OkxGetPositionsTool 错误处理 + 限流"
      exports: ["NewOkxGetPositionsTool"]
    - path: "internal/agent/tools/okx_get_fundingrate.go"
      provides: "OkxGetFundingRateTool 错误处理 + 限流"
      exports: ["NewOkxGetFundingRateTool"]
  key_links:
    - from: "internal/agent/tools/okx_get_positions.go"
      to: "pkg/okex/okx_error.go"
      via: "错误类型引用"
      pattern: "&OKXError\\{"
    - from: "internal/agent/tools/okx_get_positions.go"
      to: "golang.org/x/time/rate"
      via: "限流器调用"
      pattern: "limiter\\.Wait\\(ctx\\)"
---

<objective>
所有 API 工具返回错误规范化，并添加速率限制

Purpose: FOUND-01 要求工具返回 `("", err)` 格式的错误，FOUND-02 要求所有 API 工具配备 rate.Limiter。本计划创建 OKXError 统一错误类型，并修复三个现有 Tool 的错误处理和速率限制。

Output:
- pkg/okex/okx_error.go（新文件，OKXError 类型定义）
- 三个 Tool 文件更新（添加 limiter，使用 OKXError）
</objective>

<execution_context>
@/Users/leichujun/go/src/github.com/PineappleBond/TradingEino/backend/.claude/get-shit-done/workflows/execute-plan.md
@/Users/leichujun/go/src/github.com/PineappleBond/TradingEino/backend/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/REQUIREMENTS.md
@.planning/ROADMAP.md
@.planning/phases/01-foundation-safety/01-CONTEXT.md
@.planning/phases/01-foundation-safety/01-RESEARCH.md

# 现有代码参考
@internal/agent/tools/okx_candlesticks.go
@internal/agent/tools/okx_get_positions.go
@internal/agent/tools/okx_get_fundingrate.go
</context>

<tasks>

<task type="auto" tdd="true">
<name>Task 1: 创建 OKXError 统一错误类型</name>
<files>pkg/okex/okx_error.go, pkg/okex/okx_error_test.go</files>
<behavior>
- Test 1: OKXError 结构体包含 Code, Msg, Endpoint 字段
- Test 2: Error() 方法返回格式化字符串 "OKX {Endpoint} error (code={Code}): {Msg}"
- Test 3: Unwrap() 方法返回 nil（支持 errors.As 解包）
- Test 4: 可使用 errors.As 从 error 接口解包出 *OKXError
</behavior>
<action>
创建 pkg/okex/okx_error.go 文件，定义 OKXError 结构体：

```go
type OKXError struct {
    Code     int
    Msg      string
    Endpoint string
}

func (e *OKXError) Error() string {
    return fmt.Sprintf("OKX %s error (code=%d): %s", e.Endpoint, e.Code, e.Msg)
}

func (e *OKXError) Unwrap() error {
    return nil
}
```

同时创建 pkg/okex/okx_error_test.go 测试文件验证上述行为。
</action>
<verify>
<automated>go test ./pkg/okex/... -v -run TestOKXError</automated>
</verify>
<done>OKXError 类型定义完成，测试通过</done>
</task>

<task type="auto" tdd="true">
<name>Task 2: 修复 OkxGetPositionsTool 错误处理和速率限制</name>
<files>internal/agent/tools/okx_get_positions.go, internal/agent/tools/okx_get_positions_test.go</files>
<behavior>
- Test 1: OkxGetPositionsTool 结构体包含 limiter *rate.Limiter 字段（5 次/秒）
- Test 2: InvokableRun 在 API 调用前调用 limiter.Wait(ctx)
- Test 3: OKX API 返回 Code != 0 时返回 &OKXError{Code, Msg, "GetPositions"}
- Test 4: 网络错误时直接返回原始错误 err
</behavior>
<action>
修改 OkxGetPositionsTool：

1. 添加 limiter 字段：`limiter *rate.Limiter`
2. NewOkxGetPositionsTool 初始化限流器：`rate.NewLimiter(rate.Every(200*time.Millisecond), 1)`（5 次/秒，Account 端点）
3. InvokableRun 中：
   - API 调用前：`if err := c.limiter.Wait(ctx); err != nil { return "", err }`
   - OKX 错误检查：`if result.Code != 0 { return "", &OKXError{Code: result.Code, Msg: result.Msg, Endpoint: "GetPositions"} }`
   - 移除 `fmt.Errorf("OKX API error: %s", ...)`，直接返回 &OKXError

同时创建测试文件验证限流器和错误处理。
</action>
<verify>
<automated>go test ./internal/agent/tools/... -v -run TestOkxGetPositionsTool</automated>
</verify>
<done>OkxGetPositionsTool 通过测试，限流器 5 次/秒，错误返回 OKXError</done>
</task>

<task type="auto" tdd="true">
<name>Task 3: 修复 OkxGetFundingRateTool 错误处理和速率限制</name>
<files>internal/agent/tools/okx_get_fundingrate.go, internal/agent/tools/okx_get_fundingrate_test.go</files>
<behavior>
- Test 1: OkxGetFundingRateTool 结构体包含 limiter *rate.Limiter 字段（10 次/秒）
- Test 2: InvokableRun 在 API 调用前调用 limiter.Wait(ctx)
- Test 3: OKX API 返回 Code != 0 时返回 &OKXError{Code, Msg, "GetFundingRate"}
</behavior>
<action>
修改 OkxGetFundingRateTool：

1. 添加 limiter 字段：`limiter *rate.Limiter`
2. NewOkxGetFundingRateTool 初始化限流器：`rate.NewLimiter(rate.Every(100*time.Millisecond), 2)`（10 次/秒，Public 端点）
3. InvokableRun 中应用相同模式：limiter.Wait + OKXError 返回

同时创建测试文件。
</action>
<verify>
<automated>go test ./internal/agent/tools/... -v -run TestOkxGetFundingRateTool</automated>
</verify>
<done>OkxGetFundingRateTool 通过测试，限流器 10 次/秒，错误返回 OKXError</done>
</task>

<task type="auto">
<name>Task 4: 调整 OkxCandlesticksTool 限流器配置</name>
<files>internal/agent/tools/okx_candlesticks.go</files>
<action>
OkxCandlesticksTool 已有 rate.Limiter，当前配置为 `rate.Every(time.Second/10)`（10 次/秒）。

根据 CONTEXT.md 决策：Market 端点应为 10 次/秒 (burst=2)。

修改 NewOkxCandlesticksTool 中的限流器配置为：
`limiter: rate.NewLimiter(rate.Every(100*time.Millisecond), 2)`

同时确保 InvokableRun 中的错误处理返回 &OKXError（检查 GetCandlesticks 方法）。
</action>
<verify>
<automated>go build ./internal/agent/tools/...</automated>
</verify>
<done>OkxCandlesticksTool 限流器配置正确，编译通过</done>
</task>

</tasks>

<verification>
- [ ] go test ./pkg/okex/... -v 通过
- [ ] go test ./internal/agent/tools/... -v 通过
- [ ] go build ./... 编译成功
</verification>

<success_criteria>
- pkg/okex/okx_error.go 存在，OKXError 类型定义正确
- OkxGetPositionsTool 有限流器（5 次/秒），返回 OKXError
- OkxGetFundingRateTool 有限流器（10 次/秒），返回 OKXError
- OkxCandlesticksTool 限流器调整为 10 次/秒 (burst=2)
- 所有 Tool 的 InvokableRun 返回 `("", err)` 格式
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation-safety/01-foundation-safety-01-SUMMARY.md`
</output>
