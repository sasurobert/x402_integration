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

	priceMap, ok := price.(map[string]interface{})
	if !ok {
		// Try casting to AssetAmount if it was already struct (unlikely from generic JSON but possible)
		if pStruct, ok := price.(x402.AssetAmount); ok {
			// AssetAmount is valid return type but we need to validate/enhance it
			priceMap = map[string]interface{}{
				"amount": pStruct.Amount,
				"asset":  pStruct.Asset,
			}
		} else {
			return x402.AssetAmount{}, fmt.Errorf("invalid price format")
		}
	}

	amount, _ := priceMap["amount"].(string)
	asset, _ := priceMap["asset"].(string)

	if asset == "" {
		asset = "EGLD"
	}

	// Simple decimal logic for demo
	// In real world, fetch from config/chain
	// We don't set decimals in AssetAmount because x402 AssetAmount struct doesn't HAVE Decimals field!
	// x402.AssetAmount definition: Asset string, Amount string, Extra map.
	// So we just return it.

	return x402.AssetAmount{
		Asset:  asset,
		Amount: amount,
	}, nil
}

func (s *ExactMultiversXScheme) EnhancePaymentRequirements(
	ctx context.Context,
	requirements types.PaymentRequirements,
	supportedKind types.SupportedKind,
	extensions []string,
) (types.PaymentRequirements, error) {
	// Add default fields if missing
	if requirements.Extra == nil {
		requirements.Extra = make(map[string]interface{})
	}

	// Default to EGLD if no asset
	if requirements.Asset == "" {
		requirements.Asset = "EGLD"
	}

	// Ensure PayTo is present?
	if requirements.PayTo == "" {
		return requirements, fmt.Errorf("PayTo is required for MultiversX payments")
	}

	return requirements, nil
}
