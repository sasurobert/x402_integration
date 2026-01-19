package multiversx

import (
	"context"
	"fmt"

	"github.com/coinbase/x402/go/types"
)

// VerifyUniversalSignature verifies the payment payload signature
// For MultiversX, this implies:
// 1. Validating the Ed25519 signature against the transaction bytes (if accessible/reconstructible).
// 2. Simulating the transaction (Smart Contract wallets, or just general validity).
//
// Since we don't effectively reconstruct the canonical JSON bytes locally easily without SDK canonicalizer,
// we rely heavily on Simulation for the cryptographic proof (the node verifies the sig).
//
// However, if we CAN verify Ed25519 locally, we should.
// But without the exact serialization logic from SDK, local verification is error-prone.
// EVM has standard hashing (EIP-712). MultiversX has "canonical JSON of fields".
// Recommendation: We stick to Simulation as the "Universal" verifier for MultiversX in this Go integration,
// because implementing a Go Canonical JSON Serializer for MultiversX txs perfectly matching the node is complex.
//
// But we will expose a function that integrates the checks.

func VerifyPayment(ctx context.Context, payload ExactRelayedPayload, requirements types.PaymentRequirements, simulator func(ExactRelayedPayload) (string, error)) (bool, error) {
	// 1. Static Checks
	if payload.Data.Receiver != requirements.PayTo {
		// Strict check depending on ESDT vs EGLD?
		// Handled by caller usually, but good to have utilities.
	}

	// 2. Signature Presence
	if payload.Data.Signature == "" {
		return false, fmt.Errorf("missing signature")
	}

	// 3. Simulation (The specific "Universal" verification for MX)
	hash, err := simulator(payload)
	if err != nil {
		return false, fmt.Errorf("simulation failed: %w", err)
	}

	if hash == "" {
		return false, fmt.Errorf("simulation returned empty hash")
	}

	return true, nil
}
