package multiversx

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// GetMultiversXChainId returns the chain ID for a given network string
// Supports "multiversx:1", "multiversx:D", "multiversx:T", or legacy short names
func GetMultiversXChainId(network string) (string, error) {
	// Normalize
	net := network

	// Map common aliases
	switch net {
	case "mainnet", "multiversx-mainnet":
		return "1", nil
	case "devnet", "multiversx-devnet":
		return "D", nil
	case "testnet", "multiversx-testnet":
		return "T", nil
	}

	// Parse CAIP-2 or custom format "multiversx:Ref"
	if strings.HasPrefix(net, "multiversx:") {
		ref := strings.TrimPrefix(net, "multiversx:")
		// Ref must be 1, D, T usually
		if ref == "1" || ref == "D" || ref == "T" {
			return ref, nil
		}
		// Allow custom
		return ref, nil
	}

	return "", fmt.Errorf("unsupported network format: %s", network)
}

// IsValidAddress checks if addres is valid Bech32 with Checksum
func IsValidAddress(address string) bool {
	// 1. Basic length check (erd1... is 62)
	if len(address) != 62 {
		return false
	}

	// 2. Full Bech32 Decode & Checksum Verify
	hrp, _, err := DecodeBech32(address)
	if err != nil {
		return false
	}

	// 3. Check HRP is "erd"
	if hrp != "erd" {
		return false
	}

	return true
}

// IsValidHex checks if string is valid hex (length check optional?)
func IsValidHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

// BytesToHex helper
func BytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}

// CheckAmount verifies decimal amount string
func CheckAmount(amount string) (*big.Int, error) {
	i, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %s", amount)
	}
	return i, nil
}
