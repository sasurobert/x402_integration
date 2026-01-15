package multiversx

import "math/big"

// SchemeExact is the identifier for the exact payment scheme
const SchemeExact = "multiversx-exact-v1" // or V2 if we are bold

// NetworkConfig holds configuration for a MultiversX network
type NetworkConfig struct {
	APIUrl  string
	ChainID string
}

// RelayedPayload matches the JSON sent by the Client (V3/Relayed)
type RelayedPayload struct {
	Scheme string `json:"scheme"`
	Data   struct {
		Nonce     uint64 `json:"nonce"`
		Value     string `json:"value"`
		Receiver  string `json:"receiver"`
		Sender    string `json:"sender"`
		GasPrice  uint64 `json:"gasPrice"`
		GasLimit  uint64 `json:"gasLimit"`
		Data      string `json:"data"`
		ChainID   string `json:"chainID"`
		Version   uint32 `json:"version"`
		Options   uint32 `json:"options"`
		Signature string `json:"signature"` // Hex encoded
	} `json:"data"`
}

// SimulationRequest represents the body for /transaction/simulate
type SimulationRequest struct {
	Nonce     uint64 `json:"nonce"`
	Value     string `json:"value"`
	Receiver  string `json:"receiver"`
	Sender    string `json:"sender"`
	GasPrice  uint64 `json:"gasPrice"`
	GasLimit  uint64 `json:"gasLimit"`
	Data      string `json:"data,omitempty"`
	ChainID   string `json:"chainID"`
	Version   uint32 `json:"version"`
	Signature string `json:"signature"`
}

// SimulationResponse represents the response from /transaction/simulate
type SimulationResponse struct {
	Data struct {
		Result struct {
			Status string `json:"status"`
			Hash   string `json:"hash"`
		} `json:"result"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

// Helper to check big int logic
func CheckBigInt(valStr string, expected string) bool {
	val, ok := new(big.Int).SetString(valStr, 10)
	if !ok {
		return false
	}
	exp, ok := new(big.Int).SetString(expected, 10)
	if !ok {
		return false
	}
	return val.Cmp(exp) >= 0
}
