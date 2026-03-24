package okex

import (
	"errors"
	"testing"
)

func TestOKXError_StructFields(t *testing.T) {
	err := &OKXError{
		Code:     50001,
		Msg:      "Account not found",
		Endpoint: "GetPositions",
	}

	if err.Code != 50001 {
		t.Errorf("expected Code to be 50001, got %d", err.Code)
	}
	if err.Msg != "Account not found" {
		t.Errorf("expected Msg to be 'Account not found', got %s", err.Msg)
	}
	if err.Endpoint != "GetPositions" {
		t.Errorf("expected Endpoint to be 'GetPositions', got %s", err.Endpoint)
	}
}

func TestOKXError_ErrorFormat(t *testing.T) {
	err := &OKXError{
		Code:     50001,
		Msg:      "Account not found",
		Endpoint: "GetPositions",
	}

	expected := "OKX GetPositions error (code=50001): Account not found"
	if err.Error() != expected {
		t.Errorf("expected Error() to return '%s', got '%s'", expected, err.Error())
	}
}

func TestOKXError_Unwrap(t *testing.T) {
	err := &OKXError{
		Code:     50001,
		Msg:      "Account not found",
		Endpoint: "GetPositions",
	}

	// Unwrap should return nil for OKXError
	if err.Unwrap() != nil {
		t.Errorf("expected Unwrap() to return nil, got %v", err.Unwrap())
	}
}

func TestOKXError_ErrorsAs(t *testing.T) {
	var err error = &OKXError{
		Code:     50001,
		Msg:      "Account not found",
		Endpoint: "GetPositions",
	}

	var okxErr *OKXError
	if !errors.As(err, &okxErr) {
		t.Fatal("expected errors.As to succeed")
	}

	if okxErr.Code != 50001 {
		t.Errorf("expected Code to be 50001, got %d", okxErr.Code)
	}
	if okxErr.Msg != "Account not found" {
		t.Errorf("expected Msg to be 'Account not found', got %s", okxErr.Msg)
	}
	if okxErr.Endpoint != "GetPositions" {
		t.Errorf("expected Endpoint to be 'GetPositions', got %s", okxErr.Endpoint)
	}
}
