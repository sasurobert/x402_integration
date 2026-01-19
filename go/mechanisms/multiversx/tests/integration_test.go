package multiversx_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"x402-integration/mechanisms/multiversx"
	"x402-integration/mechanisms/multiversx/exact/facilitator"

	"github.com/coinbase/x402/go/types"
)

func TestFacilitatorVerify_EGLD(t *testing.T) {
	// Mock MultiversX API (Simulation Endpoint)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transaction/simulate" {
			t.Errorf("Expected path /transaction/simulate, got %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Return success simulation
		resp := multiversx.SimulationResponse{}
		resp.Data.Result.Status = "success"
		resp.Data.Result.Hash = "mock_hash_123"

		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Setup Facilitator
	scheme := facilitator.NewExactMultiversXScheme(server.URL)

	// Create Payload (Simulate Client)
	// Direct EGLD payment
	rp := multiversx.ExactRelayedPayload{
		Scheme: multiversx.SchemeExact,
	}
	rp.Data.Receiver = "erd1receiver"
	rp.Data.Sender = "erd1sender"
	rp.Data.Value = "100" // Atomic units
	rp.Data.Nonce = 1
	rp.Data.Signature = "aabbcc"

	payloadBytes, _ := json.Marshal(rp)
	var rpMap map[string]interface{}
	json.Unmarshal(payloadBytes, &rpMap)

	paymentPayload := types.PaymentPayload{
		Payload: rpMap,
	}

	// Create Requirements
	req := types.PaymentRequirements{
		PayTo:  "erd1receiver",
		Amount: "100",
		Asset:  "EGLD",
	}

	// Verify
	resp, err := scheme.Verify(context.Background(), paymentPayload, req)
	if err != nil {
		t.Fatalf("Verification failed: %v", err)
	}

	if !resp.IsValid {
		t.Error("Expected IsValid=true")
	}
}

func TestFacilitatorVerify_ESDT_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := multiversx.SimulationResponse{}
		resp.Data.Result.Status = "success"
		resp.Data.Result.Hash = "mock_esdt_hash"
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	scheme := facilitator.NewExactMultiversXScheme(server.URL)

	// Use Real Bech32 Address (Bob) for Strict Verification
	payTo := "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"
	_, pubBytes, err := multiversx.DecodeBech32(payTo)
	if err != nil {
		t.Fatalf("Failed to decode test address: %v", err)
	}
	payToHex := hex.EncodeToString(pubBytes)

	// Token: USDC-123 -> hex ("555344432d313233")
	tokenHex := hex.EncodeToString([]byte("USDC-123"))
	// Amount: 100 -> hex ("64")
	amountHex := "64"

	// Data: "MultiESDTNFTTransfer@<receiver_hex>@01@<token_hex>@00@<amount_hex>"
	// The facilitator expects this exact format.
	dataString := fmt.Sprintf("MultiESDTNFTTransfer@%s@01@%s@00@%s", payToHex, tokenHex, amountHex)

	rp := multiversx.ExactRelayedPayload{}
	rp.Data.Data = dataString
	rp.Data.Value = "0"
	rp.Data.Receiver = "erd1sender" // Self-transfer
	rp.Data.Sender = "erd1sender"
	rp.Data.Signature = "dummy_sig"

	payloadBytes, _ := json.Marshal(rp)
	var rpMap map[string]interface{}
	json.Unmarshal(payloadBytes, &rpMap)

	paymentPayload := types.PaymentPayload{
		Payload: rpMap,
	}

	req := types.PaymentRequirements{
		PayTo:  payTo, // Bech32
		Amount: "100",
		Asset:  "USDC-123",
	}

	resp, err := scheme.Verify(context.Background(), paymentPayload, req)
	if err != nil {
		t.Fatalf("Verification failed: %v", err)
	}
	if !resp.IsValid {
		t.Error("IsValid should be true")
	}
}

func TestFacilitatorVerify_EGLD_Alias_MultiESDT(t *testing.T) {
	// Verify that EGLD-000000 via MultiESDT payload is accepted
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := multiversx.SimulationResponse{}
		resp.Data.Result.Status = "success"
		resp.Data.Result.Hash = "mock_egld_alias_hash"
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	scheme := facilitator.NewExactMultiversXScheme(server.URL)

	// PayTo: Bob
	payTo := "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"
	_, pubBytes, _ := multiversx.DecodeBech32(payTo)
	payToHex := hex.EncodeToString(pubBytes)

	// Token: EGLD-000000
	// hex("EGLD-000000") = 45474c442d303030303030
	tokenHex := hex.EncodeToString([]byte("EGLD-000000"))
	amountHex := "64" // 100

	dataString := fmt.Sprintf("MultiESDTNFTTransfer@%s@01@%s@00@%s", payToHex, tokenHex, amountHex)

	rp := multiversx.ExactRelayedPayload{}
	rp.Data.Data = dataString
	rp.Data.Value = "0"
	rp.Data.Receiver = "erd1sender"
	rp.Data.Sender = "erd1sender"
	rp.Data.Signature = "dummy_sig"

	payloadBytes, _ := json.Marshal(rp)
	var rpMap map[string]interface{}
	json.Unmarshal(payloadBytes, &rpMap)

	paymentPayload := types.PaymentPayload{
		Payload: rpMap,
	}

	req := types.PaymentRequirements{
		PayTo:  payTo,
		Amount: "100",
		Asset:  "EGLD-000000",
	}

	resp, err := scheme.Verify(context.Background(), paymentPayload, req)
	if err != nil {
		t.Fatalf("Verification failed: %v", err)
	}
	if !resp.IsValid {
		t.Error("IsValid should be true for EGLD-000000 via MultiESDT")
	}
}
