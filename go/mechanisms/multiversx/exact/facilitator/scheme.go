package facilitator

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	x402 "github.com/coinbase/x402/go"
	"github.com/coinbase/x402/go/mechanisms/multiversx"
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
	// Return facilitator address if we have one
	return []string{}
}

func (s *ExactMultiversXScheme) Verify(ctx context.Context, payload types.PaymentPayload, requirements types.PaymentRequirements) (*x402.VerifyResponse, error) {
	// 1. Unmarshal payload to RelayedPayload
	// The payload comes as map[string]interface{} usually from generic handler
	// But SchemeNetworkFacilitator interface defines payload as types.PaymentPayload (map alias)

	// We expect the inner payload map to contain our data.
	payloadBytes, err := json.Marshal(payload.Payload)
	if err != nil {
		return nil, err
	}

	var relayedPayload multiversx.RelayedPayload
	if err := json.Unmarshal(payloadBytes, &relayedPayload); err != nil {
		return nil, fmt.Errorf("invalid payload format: %v", err)
	}

	// 2. Perform Verification (Simulation)
	simHash, err := s.verifyViaSimulation(relayedPayload)
	if err != nil {
		return nil, err
	}

	// 3. Validate Requirements
	expectedReceiver := requirements.PayTo
	txReceiver := relayedPayload.Data.Receiver
	// Note: For ESDT, receiver might be different (Self), need parsing logic from verifier.go

	tokenIdentifier := requirements.Asset
	if tokenIdentifier == "" {
		tokenIdentifier = "EGLD"
	}

	// Logic from verifier.go
	isCorrectReceiver := false

	if tokenIdentifier == "EGLD" {
		if txReceiver == expectedReceiver {
			isCorrectReceiver = true
		}
	} else {
		// Naive check for ESDT in Data
		// Ideally we need full data parsing
		if strings.Contains(relayedPayload.Data.Data, hex.EncodeToString([]byte(expectedReceiver))) {
			isCorrectReceiver = true
		}
	}

	if !isCorrectReceiver {
		return nil, errors.New("receiver mismatch")
	}
	fmt.Printf("Simulation Successful. Hash: %s\n", simHash)

	// Return Success
	// VerifyResponse only supports IsValid, InvalidReason, Payer.
	return &x402.VerifyResponse{
		IsValid: true,
	}, nil
}

func (s *ExactMultiversXScheme) Settle(ctx context.Context, payload types.PaymentPayload, requirements types.PaymentRequirements) (*x402.SettleResponse, error) {
	// TODO: Implement broadcasting (Relaying)
	return &x402.SettleResponse{
		Success:     true,
		Transaction: "mock_simulation_hash", // reused
	}, nil
}

// Internal Simulation Logic (Ported from verifier.go)
func (s *ExactMultiversXScheme) verifyViaSimulation(payload multiversx.RelayedPayload) (string, error) {
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
		// Read body for error
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
