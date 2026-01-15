package server

import (
	"context"
	"fmt"

	x402 "github.com/coinbase/x402/go"
	"github.com/coinbase/x402/go/mechanisms/multiversx"
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
	// Basic implementation - we assume EGLD or ESDT
	// For now, return error or simple string pass-through
	// Real impl would need NetworkConfig to know decimals
	return x402.AssetAmount{}, fmt.Errorf("ParsePrice not yet implemented for MultiversX")
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

	return requirements, nil
}
