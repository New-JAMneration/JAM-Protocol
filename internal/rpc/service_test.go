package rpc

import (
	"testing"
)

func TestParameters(t *testing.T) {
	service := NewRPCService()

	params, err := service.Parameters()
	if err != nil {
		t.Fatalf("Parameters() failed: %v", err)
	}

	if params == nil {
		t.Fatalf("Expected parameters, got nil")
	}

	t.Logf("Service parameters: %v", params)
}
