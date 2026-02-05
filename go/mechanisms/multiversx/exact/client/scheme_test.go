package client

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"x402-integration/mechanisms/multiversx"

	"github.com/coinbase/x402/go/types"
)

// MockSigner matches ClientMultiversXSigner
type MockSigner struct {
	addr string
}

func (m *MockSigner) Address() string {
	return m.addr
}
func (m *MockSigner) Sign(ctx context.Context, message []byte) ([]byte, error) {
	return []byte("signature"), nil
}

const (
	// Valid Bech32 Addresses for Testing (Alice/Bob Devnet)
	testPayTo  = "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx" // Alice (using Bob's valid address)
	testSender = "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx" // Bob
	testAsset  = "TEST-123456"
	testAmount = "1000000000000000000" // 1 EGLD
)

// MockNetworkProvider
type MockNetworkProvider struct {
	nonce uint64
	err   error
}

func (m *MockNetworkProvider) GetNonce(ctx context.Context, address string) (uint64, error) {
	return m.nonce, m.err
}

func TestCreatePaymentPayload_EGLD(t *testing.T) {
	signer := &MockSigner{addr: testSender}
	mockProvider := &MockNetworkProvider{nonce: 15}
	scheme := NewExactMultiversXScheme(signer, WithNetworkProvider(mockProvider))

	req := types.PaymentRequirements{
		PayTo:   testPayTo,
		Amount:  "100",
		Asset:   "EGLD",
		Network: "multiversx:D",
	}

	payload, err := scheme.CreatePaymentPayload(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create payload: %v", err)
	}

	// Verify structure
	dataBytes, _ := json.Marshal(payload.Payload)
	var rp multiversx.ExactRelayedPayload
	json.Unmarshal(dataBytes, &rp)

	if rp.Scheme != multiversx.SchemeExact {
		t.Errorf("Wrong scheme: %s", rp.Scheme)
	}
	if rp.Data.Receiver != testPayTo {
		t.Errorf("Wrong receiver: %s", rp.Data.Receiver)
	}
	if rp.Data.Value != "100" {
		t.Errorf("Wrong value: %s", rp.Data.Value)
	}
	if rp.Data.Data != "" {
		t.Errorf("Expected empty data for EGLD, got %s", rp.Data.Data)
	}
	if rp.Data.Nonce != 15 {
		t.Errorf("Wrong nonce: %d", rp.Data.Nonce)
	}
}

func TestCreatePaymentPayload_ESDT(t *testing.T) {
	signer := &MockSigner{addr: testSender}
	mockProvider := &MockNetworkProvider{nonce: 20}
	scheme := NewExactMultiversXScheme(signer, WithNetworkProvider(mockProvider))

	req := types.PaymentRequirements{
		PayTo:   testPayTo,
		Amount:  "100",
		Asset:   testAsset,
		Network: "multiversx:D",
	}

	payload, err := scheme.CreatePaymentPayload(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create payload: %v", err)
	}

	dataBytes, _ := json.Marshal(payload.Payload)
	var rp multiversx.ExactRelayedPayload
	json.Unmarshal(dataBytes, &rp)

	// ESDT check: Receiver should be Sender (Self-transfer)
	if rp.Data.Receiver != testSender {
		t.Errorf("ESDT tx receiver should be sender, got %s", rp.Data.Receiver)
	}
	if rp.Data.Value != "0" {
		t.Errorf("ESDT tx value should be 0 EGLD, got %s", rp.Data.Value)
	}
	// Check Nonce
	if rp.Data.Nonce != 20 {
		t.Errorf("Wrong nonce: %d", rp.Data.Nonce)
	}

	// Check Data field contains "MultiESDTNFTTransfer"
	if !strings.HasPrefix(rp.Data.Data, "MultiESDTNFTTransfer") {
		t.Errorf("Data should start with MultiESDTNFTTransfer, got %s", rp.Data.Data)
	}
}

func TestCreatePaymentPayload_ESDT_WithResourceID(t *testing.T) {
	signer := &MockSigner{addr: testSender}
	mockProvider := &MockNetworkProvider{nonce: 25}
	scheme := NewExactMultiversXScheme(signer, WithNetworkProvider(mockProvider))

	req := types.PaymentRequirements{
		PayTo:   testPayTo,
		Amount:  "100",
		Asset:   testAsset,
		Network: "multiversx:D",
		Extra: map[string]interface{}{
			"resourceId": "inv_123",
		},
	}

	payload, err := scheme.CreatePaymentPayload(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create payload: %v", err)
	}

	dataBytes, _ := json.Marshal(payload.Payload)
	var rp multiversx.ExactRelayedPayload
	json.Unmarshal(dataBytes, &rp)

	// Check encoded resource ID "inv_123" -> hex "696e765f313233"
	// Should be at the end
	expectedRidHex := "696e765f313233"
	if !strings.HasSuffix(rp.Data.Data, expectedRidHex) {
		t.Errorf("Data should end with resourceId hex %s, got %s", expectedRidHex, rp.Data.Data)
	}
}

func TestCreatePaymentPayload_EGLD_Alias(t *testing.T) {
	signer := &MockSigner{addr: testSender}
	mockProvider := &MockNetworkProvider{nonce: 30}
	scheme := NewExactMultiversXScheme(signer, WithNetworkProvider(mockProvider))

	req := types.PaymentRequirements{
		PayTo:   testPayTo,
		Amount:  "100",
		Asset:   "EGLD-000000", // Should be treated as EGLD if handled or ESDT token otherwise
		Network: "multiversx:D",
	}

	payload, err := scheme.CreatePaymentPayload(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create payload: %v", err)
	}

	dataBytes, _ := json.Marshal(payload.Payload)
	var rp multiversx.ExactRelayedPayload
	json.Unmarshal(dataBytes, &rp)

	// Should be ESDT transfer (MultiESDTNFTTransfer)
	// Because we treat EGLD-000000 as a token identifier for MultiESDT.

	// Value should be 0 (Native EGLD not sent via Value field in MultiESDT usually, unless implied?)
	// Actually, if using EGLD-000000 in MultiESDT, the 'value' of tx is 0, and amount is in data.
	if rp.Data.Value != "0" {
		t.Errorf("Value should be 0 for MultiESDT, got %s", rp.Data.Value)
	}

	if !strings.HasPrefix(rp.Data.Data, "MultiESDTNFTTransfer") {
		t.Errorf("Data should start with MultiESDTNFTTransfer, got %s", rp.Data.Data)
	}

	// Check token hex for EGLD-000000
	// "EGLD-000000" -> 45474c442d303030303030
	tokenHex := "45474c442d303030303030"
	if !strings.Contains(rp.Data.Data, tokenHex) {
		t.Errorf("Data should contain EGLD-000000 hex %s, got %s", tokenHex, rp.Data.Data)
	}
}
