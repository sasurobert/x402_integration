# Implementation Plan - MultiversX Integration

# Goal Description
Integrate MultiversX as a payment rail for the Google Agentic Commerce (a2a-x402) protocol. This will enable agents to perform "gasless" payments using Relayed Transactions V3 on the MultiversX network.
The current `wallet.py` implementation is tightly coupled to EVM (EIP-3009). We will refactor this to support pluggable schemes and implement the MultiversX scheme.

## User Review Required
> [!IMPORTANT]
> **Architecture Change**: `wallet.py` currently hardcodes EVM logic. I propose refactoring `process_payment` to delegate to a `SchemeHandler` interface. This is a significant change to the core flow.

## Proposed Changes

### Component: `a2a-x402` (Python)

#### [NEW] `python/x402_a2a/src/x402_a2a/schemes/`
Create a new package for payment schemes.
- `__init__.py`: Registry or factory for schemes.
- `multiversx.py`: Implements `MultiversXScheme`.
    - Handles `create_payment_request` (Functional Core).
    - Handles `construct_payment_payload` (signing logic).
- `base.py`: Abstract base class for schemes (optional, if we want strict typing).

#### [MODIFY] `python/x402_a2a/src/x402_a2a/core/wallet.py`
Refactor `process_payment` to:
1.  Check `requirements.scheme` (or `network`).
2.  If `mvx` or `multiversx`, delegate to `MultiversXScheme`.
3.  Fallback to existing EVM logic (or move EVM logic to `schemes/evm.py` later, but keep inline for now to minimize diff).

#### [MODIFY] `python/x402_a2a/pyproject.toml`
Add optional dependencies:
- `multiversx-sdk-core`
- `multiversx-sdk-wallet`
- `requests`

### Component: `AP2`
*(No immediate code changes required if standard `PaymentMandate` is sufficient. Will confirm during testing.)*

## Verification Plan

### Automated Tests
1.  **Unit Tests**:
    - `python/x402_a2a/tests/test_multiversx.py`: Test `MultiversXScheme` methods.
    - Test `wallet.py` refactor ensures existing EVM tests still pass.
    - Command: `hatch run dev:test` (or `pytest`).
2.  **Integration (Manual/Script)**:
    - Create `python/examples/multiversx_payment.py`.
    - Simulate a full flow:
        - Create `PaymentRequirements` (MVX).
        - `process_payment` (Client with Mock Wallet).
        - Verify output payload matches `RelayedTransactionV3` structure.

### Manual Verification
- Verify `did:pkh:mvx` resolution locally if possible (mocked).
