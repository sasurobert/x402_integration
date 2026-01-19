package multiversx

import (
	"context"
	"errors"
	"testing"

	"github.com/coinbase/x402/go/types"
)

func TestVerifyPayment(t *testing.T) {
	// Setup valid payload
	validPayload := ExactRelayedPayload{}
	validPayload.Data.Receiver = "erd1receiver"
	validPayload.Data.Signature = "sig"

	req := types.PaymentRequirements{
		PayTo: "erd1receiver",
	}

	// Mock Simulator
	successSim := func(p ExactRelayedPayload) (string, error) {
		return "hash", nil
	}
	failSim := func(p ExactRelayedPayload) (string, error) {
		return "", errors.New("sim failed")
	}

	// Case 1: Success
	valid, err := VerifyPayment(context.Background(), validPayload, req, successSim)
	if err != nil || !valid {
		t.Errorf("Expected success, got valid=%v err=%v", valid, err)
	}

	// Case 2: Missing Sig
	noSig := validPayload
	noSig.Data.Signature = ""
	valid, err = VerifyPayment(context.Background(), noSig, req, successSim)
	if err == nil {
		t.Error("Expected error for missing sig")
	}

	// Case 3: Sim Fail
	valid, err = VerifyPayment(context.Background(), validPayload, req, failSim)
	if err == nil {
		t.Error("Expected error for sim failure")
	}
}
