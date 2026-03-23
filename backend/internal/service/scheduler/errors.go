package scheduler

// RetryableErrorImpl 可重试错误实现
type RetryableErrorImpl struct {
	Err error
}

// NewRetryableError 创建可重试错误
func NewRetryableError(err error) *RetryableErrorImpl {
	return &RetryableErrorImpl{Err: err}
}

func (e *RetryableErrorImpl) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "retryable error"
}

func (e *RetryableErrorImpl) IsRetryable() bool {
	return true
}

// NonRetryableErrorImpl 不可重试错误实现
type NonRetryableErrorImpl struct {
	Err error
}

// NewNonRetryableError 创建不可重试错误
func NewNonRetryableError(err error) *NonRetryableErrorImpl {
	return &NonRetryableErrorImpl{Err: err}
}

func (e *NonRetryableErrorImpl) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "non-retryable error"
}

func (e *NonRetryableErrorImpl) IsRetryable() bool {
	return false
}
