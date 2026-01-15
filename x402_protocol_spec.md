# Technical Specification: x402 Protocol Integration (Official V2)

## 1. Overview
This is the most complex integration. It requires contributing logic to the **Official Coinbase x402 Monorepo**. We must implement the `multiversx` mechanism in both **Typescript** (for Client/Facilitator) and **Go** (for Validation/Server).

## 2. Directory Structure (Target)
```
x402/
├── typescript/
│   └── packages/
│       └── mechanisms/
│           └── multiversx/  <-- NEW PACKAGE
│               ├── src/
│               │   ├── index.ts
│               │   ├── constants.ts
│               │   ├── types.ts
│               │   └── signer.ts
│               └── package.json
└── go/
    └── mechanisms/
        └── multiversx/      <-- NEW PACKAGE
            ├── payments.go
            └── verifier.go
```

## 3. TypeScript Implementation Details
**Class**: `MultiversXSigner`
**Implements**: `Signer` (from `@x402/protocol`)

**Logic**:
1.  Receive `PaymentRequest` (Header data).
2.  Resolve `recipient` (Smart Contract Address).
3.  Construct `Transaction payload`:
    *   Function: `pay` (on the Standard Payment SC).
    *   Arguments: `resource_id` (from header).
4.  Sign and Broadcast.
5.  Return `token` (The Transaction Hash).

## 4. Go Implementation Details
**Struct**: `MultiversXVerifier`

**Logic**:
1.  Parse `Authorization: x402 <tx_hash>`.
2.  Query MultiversX Node API (`/transactions/{hash}`).
3.  Verify:
    *   Receiver == Expected Payment SC.
    *   Function == `pay`.
    *   Argument == `resource_id`.
    *   Status == `success`.
4.  Return `true`.

## 5. Pre-Requisite
We need a **Standard Payment Smart Contract** on MultiversX Mainnet/Testnet to act as the recipient. This spec assumes that contract exists.

## 6. Development Strategy
1.  **Phase 1**: Local Development (Clone Repo).
2.  **Phase 2**: Implement Dummy Signer (Mocked).
3.  **Phase 3**: Connect `@multiversx/sdk-core`.
4.  **Phase 4**: Submit PR.
