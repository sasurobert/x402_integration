# x402 Integration Standard for MultiversX (Official V2 PR Strategy)

## 1. Executive Summary
To achieve official support for MultiversX in the **x402 V2** standard, we must submit a Pull Request to the [coinbase/x402](https://github.com/coinbase/x402) repository. Unlike V1, V2 is a modular, code-first protocol. We need to implement a **MultiversX Mechanism** in both TypeScript and Go.

## 2. Repository Structure & Target Changes
The x402 V2 repo is a monorepo. We will target two main directories:
1.  `typescript/packages/mechanisms/multiversx` (Client-side SDK & Facilitator support)
2.  `go/mechanisms/multiversx` (Server-side verification)

## 3. TypeScript Implementation Plan
**Location**: `typescript/packages/mechanisms/multiversx`

### 3.1. Package Structure
We will create a new NPM workspace package.
*   `package.json`: dependency on `@multiversx/sdk-core`, `@multiversx/sdk-wallet`.
*   `src/index.ts`: Exports.
*   `src/constants.ts`: Chain IDs (1, D, T), Gas Limits.
*   `src/exact/`: Implementation of the "Exact Payment" scheme.

### 3.2. Core Components
**`MultiversXSigner`**
Implements the `Signer` interface (Abstract Class in Core).
*   **Input**: `unsignedTx` (JSON or Object).
*   **Action**: Signs using the provided MultiversX Provider (Extension, Ledger, or PEM).
*   **Output**: `(signature, txHash)`.

**`MultiversXMechanism`**
*   **`createPaymentRequest()`**: Generates the `X-402` header payload.
*   **`constructTransaction()`**: Converts a 402 payment request into a MultiversX Transaction Object (Value Transfer or SC Call).

#### Code Draft (Signer)
```typescript
import { Signer } from '@x402/core';
import { Transaction } from '@multiversx/sdk-core';

export class MultiversXSigner implements Signer {
  async sign(tx: Transaction): Promise<string> {
    // ... logic to sign using sdk-dapp or wallet provider
  }
}
```

## 4. Go Implementation Plan
**Location**: `go/mechanisms/multiversx`

### 4.1. Core Components
**`Verifier`**
Used by servers/gateways to validate payment.
*   **Input**: `Authorization: x402 <token>`, `X-402-Signature`.
*   **Logic**:
    *   Parse the token (containing TxHash or Signed Message).
    *   Verify signature against Sender Address (Ed25519).
    *   (Optional) Query Gateway/Node to confirm Tx inclusion/success.

## 5. Integration Steps (The PR)
1.  **Scaffold**: Fork `coinbase/x402`.
2.  **TypeScript**:
    *   Copy `mechanisms/evm` as a template.
    *   Replace `ethers.js` logic with `@multiversx/sdk-core`.
    *   Implement `ExactPayment` scheme for strictly Native Token (EGLD) transfers first.
3.  **Go**:
    *   Copy `mechanisms/evm` template.
    *   Implement header parsing and signature verification.
4.  **Register**:
    *   Add `multiversx` to the root `MechanismRegistry` in both TS and Go.
5.  **Test**:
    *   Add `e2e/multiversx_test.ts`: Runs a local test flow (mocked node or devnet).

## 6. Verification
*   **Unit Tests**: Jest tests for Header generation/parsing.
*   **E2E**: Simulation of a Client paying a Server.
