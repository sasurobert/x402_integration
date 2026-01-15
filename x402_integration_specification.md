MultiversX x402 Integration Specification (V2)
1. Executive Summary
This document outlines the technical specification for integrating MultiversX support into the official coinbase/x402 V2 repository. The goal is to enable strictly-typed, code-first payment mechanisms for MultiversX within the x402 protocol, supporting both client-side generation (TypeScript) and server-side verification (Go) of payments.

2. Architecture Overview
To achieve full protocol compatibility, we must introduce a new mechanism multiversx into the x402 monorepo. This involves updates to two primary directories in the upstream repository:

Client-Side (TypeScript): typescript/packages/mechanisms/multiversx
Responsible for generating X-402 headers and signing transactions.
Used by Wallets (Facilitators) and Clients (Browsers/Apps).
Server-Side (Go): go/mechanisms/multiversx
Responsible for verifying X-402 headers and payment confirmation.
Used by Gateways and Servers.
3. Component Specification: TypeScript
Location: typescript/packages/mechanisms/multiversx

3.1. Package Structure
This will be a new NPM workspace package.

Dependencies:
@x402/protocol (Core types)
@multiversx/sdk-core
@multiversx/sdk-wallet
Exports: MultiversXMechanism, MultiversXSigner.
3.2. Class: MultiversXSigner
Implements the Signer abstract class from @x402/protocol.

Responsibility: Converts a generic x402 payment request into a signed MultiversX transaction.
Inputs:
unsignedTx: A specialized object or JSON describing the payment intent (Standard Payment SC call).
provider: An interface for the wallet (Extension, Ledger, PEM).
Logic:
Parse the unsignedTx to identify the Recipient (Payment SC), Function (pay), and Arguments (resource_id).
Construct a Transaction object using @multiversx/sdk-core.
Sign the transaction using the provider.
Broadcast via a Proxy/API (optional, or return signed blob).
Return the txHash as the payment token.
3.3. Class: MultiversXMechanism
createPaymentRequest(): Generates the initial WWW-Authenticate: x402 ... challenge or helper headers.
constructTransaction(): Helper to format the low-level transaction payload.
3.4. Code Draft (Signer)
import { Signer } from '@x402/protocol';
import { Transaction, Address, TokenPayment } from '@multiversx/sdk-core';
export class MultiversXSigner implements Signer {
  constructor(private provider: any, private networkConfig: any) {}
  async sign(request: any): Promise<string> {
    // 1. Construct Transaction
    const tx = new Transaction({
        receiver: new Address(request.to),
        value: TokenPayment.egldFromAmount(request.amount),
        data: request.data, // e.g. "pay@<resource_id>"
        gasLimit: 10_000_000,
        // ... nonce, chainID from networkConfig
    });
    // 2. Sign
    await this.provider.signTransaction(tx);
    // 3. Broadcast (if not handled by caller) & Return Hash
    // const hash = await this.provider.sendTransaction(tx);
    return tx.getHash().toString();
  }
}
4. Component Specification: Go (Server)
Location: go/mechanisms/multiversx

4.1. Package Structure
Dependencies:
Standard Go libs.
Could use mx-chain-go or simple HTTP calls for v1.
Structs: Verifier, PaymentDetails.
4.2. Struct: Verifier
Responsibility: Validates that an incoming request contains a legitimate payment for the requested resource.
Input: HTTP Header Authorization: x402 <token>.
Logic:
Parse: Extract <token> (Transaction Hash).
Verify: Query the MultiversX API (e.g., https://api.multiversx.com/transactions/<hash>).
Validate Fields:
receiver: Must match the expected Payment Smart Contract address.
function: Must be pay (or equivalent).
args: Must contain the correct resource_id.
status: Must be success.
value: Must meet the price requirement.
Security Check: Ensure the Transaction Hash is unique (prevent replay attacks) and recent.
5. Smart Contract Requirements
To support the "Exact Payment" scheme, a canonical Smart Contract should exist on MultiversX.

Endpoints:
pay(resource_id: bytes): Payable endpoint. Accepts EGLD/ESDT. Emits event Payment { sender, resource_id, amount }.
withdraw(): Admin only.
Events: Critical for the Go Verifier to confirm payment without deep transaction parsing if possible (using event logs is more robust).
6. Integration & PR Strategy
Phase 1: Local Development
Fork coinbase/x402.
Scaffold typescript/packages/mechanisms/multiversx based on the defined package structure.
Scaffold go/mechanisms/multiversx with basic stub functions.
Phase 2: Implementation
TS: Implement MultiversXSigner using @multiversx/sdk-core. Add unit tests mocking the provider.
Go: Implement Verifier with HTTP client for MultiversX API. Add unit tests with mocked API responses.
Phase 3: Registration & End-to-End
Register multiversx in the root MechanismRegistry of the monorepo.
Create e2e/multiversx_test.ts:
Simulate a client creating a payment request.
Simulate a "server" verifying it (by hitting the Go logic or a mock thereof).
Use a local Devnet or mock transaction objects for speed.
Phase 4: Verification Checklist
 TS SDK builds and passes npm test.
 Go package builds and passes go test.
 E2E flow demonstrates a successful "payment" cycle (even if mocked on chain).
 Code follows monorepo linting/formatting rules.
7. Verification Plan
Automated Tests
TypeScript Unit Tests: jest tests within the package (Run: cd typescript/packages/mechanisms/multiversx && npm test).
Go Unit Tests: Standard Go tests (Run: cd go/mechanisms/multiversx && go test ./...).
E2E Tests: Integration test in the root e2e folder.
Manual Verification
Signer Check: Use a simple script to invoke MultiversXSigner with a MultiversX DeFi Wallet extension mock or actual connection, verifying the output is a valid signed transaction hash.
Verifier Check: Feed a real mainnet transaction hash (that called the pay function) to the Go Verifier and assert it returns true.