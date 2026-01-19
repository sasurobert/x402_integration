# MultiversX x402 Integration

This package provides the [MultiversX](https://multiversx.com) integration for the x402 payment protocol. It supports the "Exact" scheme, enabling gasless payments via Relayed Transactions and standard direct transfers.

## Features

- **Scheme**: `multiversx-exact-v1`
- **Asset Support**:
    - **EGLD**: Direct Native Currency transfers.
    - **ESDT/SFT/NFT**: Token transfers via `MultiESDTNFTTransfer`.
- **Modes**:
    - **Direct**: User pays gas (standard wallet interaction).
    - **Relayed (V3)**: User signs payload, Facilitator/Relayer submits and pays gas.
- **Verification**:
    - Simulation-based verification (`/transaction/simulate`) ensures 100% accuracy of transfer validity before acceptance.
    - Strict amount and receiver validation.

## Directory Structure

- `specs/`: Implementation specifications.
- `go/`: Go SDK (Server, Facilitator, Client interface).
- `typescript/`: TypeScript SDK (Client Signer, Mechanisms).

## Go Usage

### Facilitator (Server-Side)

The Facilitator validates payment requests and (optionally) relays them.

```go
import (
    "github.com/coinbase/x402/go/mechanisms/multiversx/exact/facilitator"
    "github.com/coinbase/x402/go/types"
)

// Initialize
scheme := facilitator.NewExactMultiversXScheme("https://devnet-api.multiversx.com")

// Verify a Payment
resp, err := scheme.Verify(ctx, payload, requirements)
if err != nil {
    // Handle error (e.g. invalid signature, insufficient funds)
}
if resp.IsValid {
    // Payment verified via Simulation!
}
```

### Server (Merchant-Side)

Defines payment requirements (Price, Asset).

```go
import "github.com/coinbase/x402/go/mechanisms/multiversx/exact/server"

srv := server.NewExactMultiversXScheme()

// Parse Price
amount, err := srv.ParsePrice(x402.Price{Amount: "10.0", Asset: "USDC-c70f1a"}, network)
```

## TypeScript Usage

### Client (Browser/Wallet)

Used to sign payment requests.

```typescript
import { MultiversXSigner, MultiversXMechanism } from '@x402/mechanisms-multiversx';
import { ExtensionProvider } from '@multiversx/sdk-extension-provider'; // Example

// 1. Setup Signer
const provider = ExtensionProvider.getInstance();
await provider.init();
const msigner = new MultiversXSigner(provider);

// 2. Sign Payment
const signature = await msigner.sign({
    to: "erd1...", // Facilitator or Merchant Address
    amount: "1000000",
    tokenIdentifier: "EGLD",
    resourceId: "invoice-123",
    chainId: "D"
});

// 3. Send to x402 Facilitator
const payload = {
    scheme: "multiversx-exact-v1",
    data: {
        // ... construct payload ...
        signature: signature
    }
};
```

## Testing

Run the Go integration tests:

```bash
cd go/mechanisms/multiversx
go test ./...
```
