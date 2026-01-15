# PR Description: Add MultiversX Payment Scheme

## Title
feat: Add MultiversX (mvx) Payment Scheme with Relayed Transactions V3

## Description
This Pull Request integrates the **MultiversX** blockchain as a supported payment rail for the Agent-to-Agent (x402) protocol.

It enables **Gasless Agentic Commerce** by leveraging MultiversX's native **Relayed Transactions V3**. This allows an autonomous agent (Client) to sign a transaction without holding EGLD for gas, while a Credentials Provider (Relayer) covers the network fees.

### Key Features
-   **New Scheme**: `MultiversXScheme` (`src/x402_a2a/schemes/multiversx.py`)
-   **Implementation**: "Functional Core" pattern for `create_payment_request` and `construct_payment_payload`.
-   **Gasless**: Constructs Inner Transactions (Value=0) with ESDT transfer data, ready to be wrapped by a Relayer.
-   **Identity**: Supports `did:pkh:mvx:1:erd1...` DID resolution (CAIP-2 `mvx` compliant).
-   **ESDT Support**: Native token transfers via the `data` field (e.g., `ESDTTransfer@...`).

### Dependencies
-   `multiversx-sdk-core` (`>=0.8.0,<1.0.0`) - Strictly pinned to resolve conflicts with `multiversx-sdk-wallet`.
-   `multiversx-sdk-wallet` (`>=1.0.0`) - Used for signing interface.
-   **Python Compatibility**: Confirmed support for **Python 3.9+** (downgraded from 3.10+).

### Verification
-   **Unit Tests**: `tests/test_multiversx_scheme.py` (Mocked).
-   **Integration Tests**: `tests/test_integration_real.py` (Real Packages).
    -   Verified end-to-end flow from requirement creation to payload signing without mocks.
    -   Verified compatibility with installed `x402` and `a2a-sdk` packages.

## Checklist
-   [x] Signed CLA.
-   [x] Added Unit Tests.
-   [x] Updated README with scheme usage.
-   [x] Verified CAIP-2 `mvx` namespace registration.

---
**Related Issues**: N/A
**Co-authored-by**: Agentic Payments Team
