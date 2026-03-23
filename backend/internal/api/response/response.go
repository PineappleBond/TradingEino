package response

// Response represents the unified API response format.
// Format: {"code": 0, "message": "success", "data": {...}}
// @Description API 统一响应格式
type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// StringResponse API 字符串响应
type StringResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// SuccessResponse API 成功响应（空 data）
type SuccessResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Error codes classification
// 1xxx: Authentication errors
const (
	// CodeTokenExpired represents token expired error
	CodeTokenExpired = 1001
	// CodeTokenInvalid represents token invalid error
	CodeTokenInvalid = 1002
)

// 2xxx: Parameter errors
const (
	// CodeParameterMissing represents parameter missing error
	CodeParameterMissing = 2001
	// CodeParameterFormatError represents parameter format error
	CodeParameterFormatError = 2002
)

// 3xxx: Business errors
const (
	// CodeAccountNotFound represents account not found error
	CodeAccountNotFound = 3001
	// CodeInsufficientBalance represents insufficient balance error
	CodeInsufficientBalance = 3002
	// CodeResourceNotFound represents resource not found error
	CodeResourceNotFound = 3003
	// CodePredictionActiveCannotDelete represents error when deleting active prediction
	CodePredictionActiveCannotDelete = 3004
	// CodeOKXAccountError 账户异常
	CodeOKXAccountError = 3005
)

// 4xxx: System errors
const (
	// CodeDatabaseError represents database error
	CodeDatabaseError = 4001
	// CodeExternalAPIError represents external API error
	CodeExternalAPIError = 4002
)

// Success returns a success response with the given data.
// Code is set to 0 and message is "success".
func Success[T any](data T) *Response[T] {
	return &Response[T]{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// Error returns an error response with the given code and message.
// Data is set to zero value of type T.
func Error[T any](code int, message string) *Response[T] {
	var zero T
	return &Response[T]{
		Code:    code,
		Message: message,
		Data:    zero,
	}
}

// PageInfo represents pagination information in list responses.
type PageInfo struct {
	Page     int   `json:"page"`     // 当前页码（从 1 开始）
	PageSize int   `json:"pageSize"` // 每页数量
	Total    int64 `json:"total"`    // 总记录数
}

// PagedData represents a paged list response with pagination info.
type PagedData[T any] struct {
	Items []T      `json:"items"` // 数据列表
	Page  PageInfo `json:"page"`  // 分页信息
}
