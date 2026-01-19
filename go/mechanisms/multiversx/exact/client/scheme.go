package client

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"x402-integration/mechanisms/multiversx"

	x402 "github.com/coinbase/x402/go"
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

func (s *ExactMultiversXScheme) CreatePaymentPayload(ctx context.Context, requirements types.PaymentRequirements) (types.PaymentPayload, error) {
	// 1. Validate inputs
	if requirements.PayTo == "" {
		return types.PaymentPayload{}, fmt.Errorf("PayTo is required")
	}

	// 2. Prepare Transaction Data
	// In a real SDK we would fetch Nonce/Gas from network.
	// Here we assume defaults or they must be provided in Extra (or we query if we had an RPC client).
	// EVM implementation queries chain. SVM queries chain.
	// We should probably allow an RPC client injection similar to EVM/SVM.
	// For now, we'll use defaults or mock values, but standard enforces we should try to be real.
	// Let's rely on hardcoded Gas/Nonce for V1, or basic defaults.

	// Default Gas
	gasLimit := uint64(50000)
	gasPrice := uint64(1000000000)

	version := uint32(1)
	chainID := "D" // Default Devnet
	if requirements.Network != "" {
		// "multiversx:D" -> "D"
		_, ref, err := x402.Network(requirements.Network).Parse()
		if err == nil {
			chainID = ref
		}
	}

	sender := s.signer.Address()
	receiver := requirements.PayTo
	value := requirements.Amount // Already big int string

	// ESDT Logic
	dataString := ""
	// Normalize Asset
	asset := requirements.Asset

	if asset != "" && asset != "EGLD" {
		// ESDT Transfer (MultiESDTNFTTransfer)
		// Receiver becomes Sender (Self Transfer)
		receiver = sender
		value = "0"
		gasLimit = 60000000 // Higher gas

		// Encode Data: MultiESDTNFTTransfer@<DestHex>@01@<TokenHex>@00@<AmountHex>
		// We need to convert PayTo (dest) to hex.
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

		// MultiESDTNFTTransfer format
		// MultiESDTNFTTransfer format
		// MultiESDTNFTTransfer@<DestHex>@01@<TokenHex>@00@<AmountHex>@<ResourceID>

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
		// Data might be empty for simple transfer
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
		Nonce:    15, // TODO: Must fetch nonce in real impl
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

	// 6. Build Final Payload
	exactPayload := multiversx.ExactRelayedPayload{
		Scheme: multiversx.SchemeExact,
	}
	// struct copy
	exactPayload.Data.Nonce = txData.Nonce
	exactPayload.Data.Value = txData.Value
	exactPayload.Data.Receiver = txData.Receiver
	exactPayload.Data.Sender = txData.Sender
	exactPayload.Data.GasPrice = txData.GasPrice
	exactPayload.Data.GasLimit = txData.GasLimit
	exactPayload.Data.Data = txData.Data
	exactPayload.Data.ChainID = txData.ChainID
	exactPayload.Data.Version = txData.Version
	exactPayload.Data.Signature = hex.EncodeToString(sigBytes)

	// Return Map
	payloadBytes, _ := json.Marshal(exactPayload)
	var finalMap map[string]interface{}
	json.Unmarshal(payloadBytes, &finalMap)

	return types.PaymentPayload{
		X402Version: 2,
		Payload:     finalMap,
	}, nil
}
