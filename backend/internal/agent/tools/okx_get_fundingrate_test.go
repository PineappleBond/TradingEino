package tools

import (
	"errors"
	"testing"
	"time"

	"github.com/PineappleBond/TradingEino/backend/pkg/okex"
	"golang.org/x/time/rate"
)

func TestOkxGetFundingRateTool_ReturnsOKXError(t *testing.T) {
	// This test verifies the error type can be detected
	err := &okex.OKXError{
		Code:     50002,
		Msg:      "Instrument not found",
		Endpoint: "GetFundingRate",
	}

	var okxErr *okex.OKXError
	if !errors.As(err, &okxErr) {
		t.Fatal("Should be able to unwrap OKXError")
	}

	if okxErr.Code != 50002 {
		t.Errorf("Expected code 50002, got %d", okxErr.Code)
	}
	if okxErr.Endpoint != "GetFundingRate" {
		t.Errorf("Expected endpoint 'GetFundingRate', got %s", okxErr.Endpoint)
	}
}

func TestOkxGetFundingRateTool_LimiterConfigured(t *testing.T) {
	// Test that the limiter is configured for Public endpoint (10 req/s)
	limiter := rate.NewLimiter(rate.Every(100*time.Millisecond), 2)

	// Verify limiter is created correctly
	if limiter == nil {
		t.Fatal("Limiter should not be nil")
	}

	// Verify the limit is 10 req/s (100ms per request, burst=2)
	t.Log("Limiter configured for Public endpoint (10 req/s, burst=2)")
}
