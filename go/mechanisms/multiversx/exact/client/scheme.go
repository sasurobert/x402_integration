package client

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"x402-integration/mechanisms/multiversx"

	"github.com/coinbase/x402/go/types"
)

// ExactMultiversXScheme implements SchemeNetworkClient
type ExactMultiversXScheme struct {
	signer multiversx.ClientMultiversXSigner
}

func NewExactMultiversXScheme(signer multiversx.ClientMultiversXSigner) *ExactMultiversXScheme {
	return &ExactMultiversXScheme{
		signer: signer,
	}
}

func (s *ExactMultiversXScheme) Scheme() string {
	return multiversx.SchemeExact
}

func (s *ExactMultiversXScheme) GetSigners(ctx context.Context) ([]string, error) {
	return []string{s.signer.Address()}, nil
}

func (s *ExactMultiversXScheme) CreatePaymentPayload(ctx context.Context, requirements types.PaymentRequirements) (types.PaymentPayload, error) {
	// 1. Validate inputs
	if requirements.PayTo == "" {
		return types.PaymentPayload{}, fmt.Errorf("PayTo is required")
	}

	// 2. Prepare Transaction Data
	// TODO: Fetch Nonce/Gas from network or allow RPC injection
	gasLimit := uint64(50000)
	gasPrice := uint64(1000000000)

	version := uint32(1)
	chainID := "D" // Default Devnet
	if requirements.Network != "" {
		// Clean handling of ChainID from Network string
		parts := strings.Split(string(requirements.Network), ":")
		if len(parts) > 1 {
			chainID = parts[1]
		}
	}

	sender := s.signer.Address()
	receiver := requirements.PayTo
	value := requirements.Amount

	// ESDT Logic
	dataString := ""
	asset := requirements.Asset

	if asset != "" && asset != "EGLD" {
		// ESDT Transfer (MultiESDTNFTTransfer)
		// Receiver becomes Sender (Self Transfer)
		receiver = sender
		value = "0"
		gasLimit = 60000000 // Higher gas for ESDT

		// Encode Data: MultiESDTNFTTransfer@<DestHex>@01@<TokenHex>@00@<AmountHex>
		var destHex string
		if strings.HasPrefix(requirements.PayTo, "erd1") {
			_, decodedBytes, err := multiversx.DecodeBech32(requirements.PayTo)
			if err == nil {
				destHex = hex.EncodeToString(decodedBytes)
			} else {
				destHex = hex.EncodeToString([]byte(requirements.PayTo))
			}
		} else {
			destHex = hex.EncodeToString([]byte(requirements.PayTo))
		}

		tokenHex := hex.EncodeToString([]byte(requirements.Asset))

		amtBig, _ := new(big.Int).SetString(requirements.Amount, 10)
		if amtBig == nil {
			return types.PaymentPayload{}, fmt.Errorf("invalid amount")
		}
		amtHex := amtBig.Text(16)
		if len(amtHex)%2 != 0 {
			amtHex = "0" + amtHex
		}

		// Extract ResourceID from Extra if present
		var resourceIdHex string
		if rid, ok := requirements.Extra["resourceId"].(string); ok && rid != "" {
			resourceIdHex = hex.EncodeToString([]byte(rid))
		}

		if resourceIdHex != "" {
			dataString = fmt.Sprintf("MultiESDTNFTTransfer@%s@01@%s@00@%s@%s", destHex, tokenHex, amtHex, resourceIdHex)
		} else {
			dataString = fmt.Sprintf("MultiESDTNFTTransfer@%s@01@%s@00@%s", destHex, tokenHex, amtHex)
		}

	} else {
		// EGLD
		dataString = ""
	}

	// 3. Construct Payload Object
	txData := struct {
		Nonce    uint64 `json:"nonce"`
		Value    string `json:"value"`
		Receiver string `json:"receiver"`
		Sender   string `json:"sender"`
		GasPrice uint64 `json:"gasPrice"`
		GasLimit uint64 `json:"gasLimit"`
		Data     string `json:"data,omitempty"`
		ChainID  string `json:"chainID"`
		Version  uint32 `json:"version"`
	}{
		Nonce:    15, // TODO: Fetch nonce
		Value:    value,
		Receiver: receiver,
		Sender:   sender,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Data:     dataString,
		ChainID:  chainID,
		Version:  version,
	}

	// 4. Serialize for Signing (Canonical JSON)
	txBytes, err := json.Marshal(txData)
	if err != nil {
		return types.PaymentPayload{}, err
	}

	// 5. Sign
	sigBytes, err := s.signer.Sign(ctx, txBytes)
	if err != nil {
		return types.PaymentPayload{}, err
	}

	// 6. Build Final Payload Map directly
	finalMap := map[string]interface{}{
		"scheme": multiversx.SchemeExact,
		"data": map[string]interface{}{
			"nonce":     txData.Nonce,
			"value":     txData.Value,
			"receiver":  txData.Receiver,
			"sender":    txData.Sender,
			"gasPrice":  txData.GasPrice,
			"gasLimit":  txData.GasLimit,
			"data":      txData.Data,
			"chainID":   txData.ChainID,
			"version":   txData.Version,
			"signature": hex.EncodeToString(sigBytes),
		},
	}

	return types.PaymentPayload{
		X402Version: 2,
		Payload:     finalMap,
	}, nil
}
