# MultiversX Mechanism (V2)

This package implements x402 payment support for the MultiversX network.
It follows the standard x402 V2 architecture with support for the "Relayed" (Gasless) model.

## Subpackages

- `exact/server`: Server-side logic for parsing prices and checking requirements.
- `exact/facilitator`: Facilitator logic for verifying signatures and simulating transactions on-chain.
- `exact/client`: Client-side logic (Stub).

## Usage

### Server (Merchant)

```go
import (
    "github.com/coinbase/x402/go/mechanisms/multiversx/exact/server"
    "github.com/coinbase/x402/go/mechanisms/multiversx/exact/facilitator"
)

// 1. Setup Support
scheme := server.NewExactMultiversXScheme()

// 2. Setup Facilitator (for verification)
verifier := facilitator.NewExactMultiversXScheme("https://devnet-gateway.multiversx.com")

// 3. Verify Payment
simHash, err := verifier.Verify(ctx, payload, requirements)
```
