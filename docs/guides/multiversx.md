# MultiversX Merchant Integration Guide

This guide details how to integrate **x402 Agentic Payments** on the MultiversX network.

## Overview
The solution provides a **Gasless Payment Experience** for your users.
1.  **User** approves a payment in their wallet (e.g., xPortal, DeFi Wallet).
2.  **Client Application** generates a signed transaction *without* broadcasting it.
3.  **Client** sends this "Relayed Payload" to your **Merchant Server**.
4.  **Merchant Server** simulates the transaction to verify validity.
5.  **Merchant Server** (optionally) relays the transaction to the network or treats the payment as "Pending" until a relayer processes it.

## Prerequisites
-   **MultiversX Wallet**: A destination address to receive payments.
-   **MultiversX API**: A proxy/gateway URL (e.g., `https://devnet-gateway.multiversx.com` or a private node).

---

## 1. Frontend Integration (TypeScript)

Use the `@x402/multiversx` package to handle wallet interaction and payload creation.

### Installation
```bash
npm install @x402/multiversx @multiversx/sdk-core
```

### Usage
```typescript
import { ExactMultiversXScheme, MultiversXSigner } from "@x402/multiversx";
import { ExtensionProvider } from "@multiversx/sdk-extension-provider"; // or WalletConnect

// 1. Initialize Provder & Signer
const provider = ExtensionProvider.getInstance();
await provider.init();
const address = await provider.login();

const signer = new MultiversXSigner(provider, address);
const scheme = new ExactMultiversXScheme(signer);

// 2. Create Payment Requirements
const requirements = {
    payTo: "erd1merchant...",      // Your Wallet Address
    amount: "1000000000000000000", // 1 EGLD (in atomic units)
    asset: "EGLD",                 // or Token Identifier
    network: "multiversx:1",       // Mainnet (1) or Devnet (D)
    maxTimeoutSeconds: 300,
    extra: {
        resourceId: "invoice-123"  // Unique ID to track this payment
    }
};

// 3. Generate Payload
// This triggers the wallet signature prompt but DOES NOT broadcast.
const { payload } = await scheme.createPaymentPayload(1, requirements);

// 4. Send `payload` to your backend API
await api.sendPayment(payload);
```

---

## 2. Backend Verification (Go)

Use the `multiversx` mechanism package to verify the received payload.

### Installation
Ensure your `go.mod` includes the repo:
```bash
go get github.com/coinbase/x402/go
```

### Usage
```go
import (
    "fmt"
    "github.com/coinbase/x402/go/mechanisms/multiversx"
)

func HandlePayment(w http.ResponseWriter, r *http.Request) {
    // 1. Initialize Verifier with MultiversX API (Proxy)
    // Use https://gateway.multiversx.com for Mainnet
    verifier := multiversx.NewVerifier("https://devnet-gateway.multiversx.com")

    // 2. Parse Payload from Request Body (RelayedPayload struct)
    var payload multiversx.RelayedPayload
    // ... json decode ...

    // 3. Verify
    // This performs a Simulation call to the API to ensure the signature is valid.
    expectedReceiver := "erd1merchant..."
    resourceId := "invoice-123" // The ID you expect for this transaction
    amount := "1000000000000000000"

    txHash, err := verifier.ProcessRelayedPayment(payload, expectedReceiver, resourceId, amount, "EGLD")
    if err != nil {
        // Validation Failed (Invalid Signature, Wrong Amount, etc.)
        http.Error(w, "Payment Invalid: "+err.Error(), 400)
        return
    }

    // 4. Success!
    fmt.Printf("Payment Verified! Simulated Hash: %s\n", txHash)
    // Fulfill the order...
}
```

## Security Considerations
-   **Resource ID**: Always strictly validate the `resourceId` on the server to prevent replay attacks or payload reuse for different invoices.
-   **Prices**: Calculate expected amounts on the server side, do not trust client-provided amounts blindly (the verifier checks `payload.Value == expectedAmount`).
