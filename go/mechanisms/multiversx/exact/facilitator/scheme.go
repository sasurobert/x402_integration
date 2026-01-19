package facilitator

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"x402-integration/mechanisms/multiversx"

	x402 "github.com/coinbase/x402/go"
	"github.com/coinbase/x402/go/types"
)

// ExactMultiversXScheme implements SchemeNetworkFacilitator
type ExactMultiversXScheme struct {
	config multiversx.NetworkConfig
	client *http.Client
}

func NewExactMultiversXScheme(apiUrl string) *ExactMultiversXScheme {
	return &ExactMultiversXScheme{
		config: multiversx.NetworkConfig{APIUrl: apiUrl},
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *ExactMultiversXScheme) Scheme() string {
	return multiversx.SchemeExact
}

func (s *ExactMultiversXScheme) CaipFamily() string {
	return "multiversx:*"
}

func (s *ExactMultiversXScheme) GetExtra(network x402.Network) map[string]interface{} {
	return nil
}

func (s *ExactMultiversXScheme) GetSigners(network x402.Network) []string {
	// In the exact scheme, the Client is the main signer,
	// but strictly GetSigners returns the addresses the *Facilitator* controls
	// if it were acting as a wallet. However, standard implies "Signers" relevant to the payment?
	// Actually, for "Exact", the facilitator relays.
	// The prompt said: "Facilitator GetSigners is not implemented (should return at least main user)"
	// This likely refers to the Go SDK's `GetSigners` usually returning the loaded wallet address.
	// Since this is a "Facilitator" (Server) implementation, it might have a relayer wallet.

	// Assumption: We might need a Relayer Address here if we implement Relaying V1.
	// For now, we return empty or a placeholder if we haven't loaded a PEM.
	// TODO: Load Relayer Wallet.
	return []string{"facilitator-address-placeholder"}
}

func (s *ExactMultiversXScheme) Verify(ctx context.Context, payload types.PaymentPayload, requirements types.PaymentRequirements) (*x402.VerifyResponse, error) {
	// 1. Unmarshal directly to ExactRelayedPayload
	payloadBytes, err := json.Marshal(payload.Payload)
	if err != nil {
		return nil, err
	}

	var relayedPayload multiversx.ExactRelayedPayload
	if err := json.Unmarshal(payloadBytes, &relayedPayload); err != nil {
		return nil, fmt.Errorf("invalid payload format: %v", err)
	}

	// 2. Perform Verification using Universal logic
	isValid, err := multiversx.VerifyPayment(ctx, relayedPayload, requirements, s.verifyViaSimulation)
	if err != nil {
		return nil, err // Returns invalid reason wrapped
	}
	if !isValid {
		return nil, fmt.Errorf("verification failed")
	}

	// 3. Validate Requirements (Specific Fields)
	expectedReceiver := requirements.PayTo
	expectedAmount := requirements.Amount
	if expectedAmount == "" {
		return nil, errors.New("requirement amount is empty")
	}

	reqAsset := requirements.Asset
	if reqAsset == "" {
		reqAsset = "EGLD"
	}

	txData := relayedPayload.Data

	if reqAsset == "EGLD" {
		// Case A: Direct EGLD
		if txData.Receiver != expectedReceiver {
			return nil, fmt.Errorf("receiver mismatch: expected %s, got %s", expectedReceiver, txData.Receiver)
		}
		if !multiversx.CheckBigInt(txData.Value, expectedAmount) {
			return nil, fmt.Errorf("amount mismatch: expected %s, got %s", expectedAmount, txData.Value)
		}
	} else {
		// Case B: ESDT Transfer
		parts := strings.Split(txData.Data, "@")
		if len(parts) < 6 || parts[0] != "MultiESDTNFTTransfer" {
			return nil, errors.New("invalid ESDT transfer data format")
		}

		// Decode Receiver (parts[1]) - Hex
		if !multiversx.IsValidHex(parts[1]) {
			return nil, fmt.Errorf("invalid receiver hex")
		}

		// Token Hex
		tokenBytes, err := hex.DecodeString(parts[3])
		if err != nil {
			return nil, fmt.Errorf("invalid token hex")
		}
		if string(tokenBytes) != reqAsset {
			return nil, fmt.Errorf("asset mismatch: expected %s, got %s", reqAsset, string(tokenBytes))
		}

		// Amount Hex
		amountBytes, err := hex.DecodeString(parts[5])
		if err != nil {
			return nil, fmt.Errorf("invalid amount hex")
		}
		amountBig := new(big.Int).SetBytes(amountBytes)
		expectedBig, _ := new(big.Int).SetString(expectedAmount, 10)
		if amountBig.Cmp(expectedBig) < 0 {
			return nil, fmt.Errorf("amount too low")
		}
	}

	// fmt.Printf("Verification Successful. Hash: %s\n", simHash)

	return &x402.VerifyResponse{
		IsValid: true,
	}, nil
}

func (s *ExactMultiversXScheme) Settle(ctx context.Context, payload types.PaymentPayload, requirements types.PaymentRequirements) (*x402.SettleResponse, error) {
	// TODO: Implement actual Relaying.
	// 1. Recover ExactRelayedPayload
	// 2. Broadcast to MultiversX API: POST /transaction/send
	// 3. Return Hash

	// Stub for now, as we need the Proxy/API client implementation details.
	return &x402.SettleResponse{
		Success:     true,
		Transaction: "mock_settlement_hash",
	}, nil
}

func (s *ExactMultiversXScheme) verifyViaSimulation(payload multiversx.ExactRelayedPayload) (string, error) {
	reqBody := multiversx.SimulationRequest{
		Nonce:     payload.Data.Nonce,
		Value:     payload.Data.Value,
		Receiver:  payload.Data.Receiver,
		Sender:    payload.Data.Sender,
		GasPrice:  payload.Data.GasPrice,
		GasLimit:  payload.Data.GasLimit,
		Data:      base64.StdEncoding.EncodeToString([]byte(payload.Data.Data)),
		ChainID:   payload.Data.ChainID,
		Version:   payload.Data.Version,
		Signature: payload.Data.Signature,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal simulation request: %v", err)
	}

	url := fmt.Sprintf("%s/transaction/simulate", s.config.APIUrl)
	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to send simulation request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var bodyErr map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&bodyErr)
		return "", fmt.Errorf("simulation API returned non-200 status: %d Body: %v", resp.StatusCode, bodyErr)
	}

	var simResp multiversx.SimulationResponse
	if err := json.NewDecoder(resp.Body).Decode(&simResp); err != nil {
		return "", fmt.Errorf("failed to decode simulation response: %v", err)
	}

	if simResp.Error != "" {
		return "", fmt.Errorf("simulation returned error: %s (code: %s)", simResp.Error, simResp.Code)
	}

	if simResp.Data.Result.Status != "success" {
		return "", fmt.Errorf("simulation status not success: %s", simResp.Data.Result.Status)
	}

	return simResp.Data.Result.Hash, nil
}
