package multiversx

import (
	"testing"
)

func TestGetMultiversXChainId(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"mainnet", "1", false},
		{"multiversx-devnet", "D", false},
		{"multiversx:T", "T", false},
		{"multiversx:1", "1", false},
		{"multiversx:Custom", "Custom", false},
		{"invalid", "", true},
		{"multiversx-invalid", "", true},
	}

	for _, tc := range tests {
		res, err := GetMultiversXChainId(tc.input)
		if tc.hasError {
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.input, err)
			}
			if res != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, res)
			}
		}
	}
}

func TestIsValidAddress(t *testing.T) {
	tests := []struct {
		addr  string
		valid bool
	}{
		// Valid Addresses (examples from Docs or generated)
		// Bob: erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx
		{"erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx", true},

		// Invalid Length
		{"erd1short", false},
		{"erd1toolonggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg", false},

		// Invalid HRP
		{"btc1qyu5wthldzr8wx5c9ucg83cq4jgy80zy85ryfx475fsz99m4h39s292042", false},

		// Invalid Checksum (last char changed 2 -> 3)
		{"erd1qyu5wthldzr8wx5c9ucg83cq4jgy80zy85ryfx475fsz99m4h39s292043", false},

		// Mixed Case (Bech32 allows mixed case if all same, but DecodeBech32 might be strict or assume lower. Standard says MUST be mixed or lower. )
		// Our implementation uses `strings.IndexByte(charset, data[i])`. Charset is lowercase.
		// If input is uppercase, it fails unless we lower it. Standard libraries verify case.
		// Let's check our impl logic: it doesn't ToLower. So it expects lowercase.
		// Standard bech32 usually enforces one case.
		{"ERD1QYU5WTHLDZR8WX5C9UCG83CQ4JGY80ZY85RYFX475FSZ99M4H39S292042", false}, // Currently implementation fails uppercase without Lower()
	}

	for _, tc := range tests {
		if res := IsValidAddress(tc.addr); res != tc.valid {
			// Debug failure
			_, _, err := DecodeBech32(tc.addr)
			t.Errorf("IsValidAddress(%s) = %v; expected %v. Error: %v", tc.addr, res, tc.valid, err)
		}
	}
}

func TestCheckAmount(t *testing.T) {
	_, err := CheckAmount("1000")
	if err != nil {
		t.Errorf("Basic integer failed")
	}
	_, err = CheckAmount("abc")
	if err == nil {
		t.Errorf("Invalid string passed")
	}
}
