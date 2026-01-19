package multiversx

import (
	"context"
)

// ClientMultiversXSigner defines the interface for signing MultiversX transactions
type ClientMultiversXSigner interface {
	// Address returns the bech32 address of the signer
	Address() string

	// SignTransaction signs the transaction object/bytes and returns the signature hex
	// In strict MultiversX terms, we sign the canonical JSON of the transaction fields.
	// For this interface, we pass the bytes to be signed.
	Sign(ctx context.Context, message []byte) ([]byte, error)
}
