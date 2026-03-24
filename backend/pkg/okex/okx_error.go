package okex

import "fmt"

// OKXError represents a standardized OKX API error.
// It wraps the error code, message, and endpoint for consistent error handling.
type OKXError struct {
	Code     int
	Msg      string
	Endpoint string
}

// Error implements the error interface.
func (e *OKXError) Error() string {
	return fmt.Sprintf("OKX %s error (code=%d): %s", e.Endpoint, e.Code, e.Msg)
}

// Unwrap implements the errors.Unwrap interface for error chaining.
// Returns nil as OKXError is a leaf error type.
func (e *OKXError) Unwrap() error {
	return nil
}
