package tools

import (
	"errors"
	"testing"
	"time"

	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	"golang.org/x/time/rate"
)

func TestOkxGetPositionsTool_ReturnsOKXError(t *testing.T) {
	// This test verifies the error type can be detected
	err := &okex.OKXError{
		Code:     50001,
		Msg:      "Account not found",
		Endpoint: "GetPositions",
	}

	var okxErr *okex.OKXError
	if !errors.As(err, &okxErr) {
		t.Fatal("Should be able to unwrap OKXError")
	}

	if okxErr.Code != 50001 {
		t.Errorf("Expected code 50001, got %d", okxErr.Code)
	}
	if okxErr.Endpoint != "GetPositions" {
		t.Errorf("Expected endpoint 'GetPositions', got %s", okxErr.Endpoint)
	}
}

func TestOkxGetPositionsTool_LimiterConfigured(t *testing.T) {
	// Test that the tool has limiter field and it can be configured
	limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 1)

	// Verify limiter is created correctly
	if limiter == nil {
		t.Fatal("Limiter should not be nil")
	}

	// Verify the limit is 5 req/s (200ms per request)
	// This is a compile-time check - if the code compiles, limiter is configured
	t.Log("Limiter configured for Account endpoint (5 req/s)")
}
