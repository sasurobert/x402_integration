package server

import (
	"context"
	"fmt"

	"x402-integration/mechanisms/multiversx"

	x402 "github.com/coinbase/x402/go"
	"github.com/coinbase/x402/go/types"
)

// ExactMultiversXScheme implements SchemeNetworkServer for MultiversX
type ExactMultiversXScheme struct {
	// Config if needed
}

func NewExactMultiversXScheme() *ExactMultiversXScheme {
	return &ExactMultiversXScheme{}
}

func (s *ExactMultiversXScheme) Scheme() string {
	return multiversx.SchemeExact
}

func (s *ExactMultiversXScheme) ParsePrice(price x402.Price, network x402.Network) (x402.AssetAmount, error) {
	// Price is interface{}, usually map[string]interface{} from JSON
	// We expect "amount" and "asset" keys.

	// Try casting to AssetAmount struct first as it's the expected type
	pStruct, ok := price.(x402.AssetAmount)
	if ok {
		// Just validate/defaults
		if pStruct.Asset == "" {
			pStruct.Asset = "EGLD"
		}
		return pStruct, nil
	}

	// If not a struct, try map (e.g. from JSON interface{})
	if pMap, okMap := price.(map[string]interface{}); okMap {
		amount, _ := pMap["amount"].(string)
		asset, _ := pMap["asset"].(string)

		if asset == "" {
			asset = "EGLD"
		}

		return x402.AssetAmount{
			Asset:  asset,
			Amount: amount,
		}, nil
	}

	return x402.AssetAmount{}, fmt.Errorf("invalid price format: expected AssetAmount struct or map")
}

func (s *ExactMultiversXScheme) EnhancePaymentRequirements(
	ctx context.Context,
	requirements types.PaymentRequirements,
	supportedKind types.SupportedKind,
	extensions []string,
) (types.PaymentRequirements, error) {
	// Create a copy to avoid side effects on the passed map
	reqCopy := requirements
	if reqCopy.Extra != nil {
		newExtra := make(map[string]interface{}, len(reqCopy.Extra))
		for k, v := range reqCopy.Extra {
			newExtra[k] = v
		}
		reqCopy.Extra = newExtra
	} else {
		reqCopy.Extra = make(map[string]interface{})
	}

	// Default to EGLD if no asset
	if reqCopy.Asset == "" {
		reqCopy.Asset = "EGLD"
	}

	// Ensure PayTo is present
	if reqCopy.PayTo == "" {
		return reqCopy, fmt.Errorf("PayTo is required for MultiversX payments")
	}

	return reqCopy, nil
}
