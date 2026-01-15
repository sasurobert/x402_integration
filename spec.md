# Technical Specification: MultiversX Integration for Google Agentic Commerce

## 1. Executive Summary
This specification outlines the technical approach for integrating the MultiversX blockchain as a settlement layer for the Google Agentic Commerce ecosystem. By combining Google's Agent Payments Protocol (AP2) for authorization with MultiversX's high-throughput infrastructure, we aim to enable "Gasless Agentic Commerce." Autonomous AI agents will be able to discover, negotiate, and deterministically settle transactions using native EGLD/ESDT tokens without managing gas balances, leveraging MultiversX Relayed Transactions V3.

## 2. Problem Statement
**The Principal-Agent Problem in AI:** Autonomous agents need to execute transactions on behalf of users without requiring manual approval for every step, while ensuring they remain within strict bounds (budget, intent).
**The Gas Friction:** Managing crypto wallets for thousands of autonomous agents is operationally complex. Agents cannot easily pay for gas (network fees) without security risks (hot wallets) and treasury management overhead.
**Lack of Deterministic Settlement:** Probabilistic AI models need a deterministic layer to enforce the "finality" of a purchase.

## 3. Target Audience
- **User (Principal):** The human delegating purchasing power to an AI.
- **Shopping Agent (SA):** The autonomous AI discovering and purchasing items.
- **Merchant Agent (MA):** The digital storefront selling goods/services.
- **Credentials Provider (CP):** The infrastructure provider (Relayer) that manages gas and broadcasting for the agent.

## 4. Requirements

### Functional Requirements
1.  **Mandate Support:** The system must support AP2 Mandates (Intent, Cart, Payment) referencing MultiversX identities (`did:pkh:mvx:1:erd1...`).
2.  **Payment Execution:** Implement the x402 "Payment Required" flow for MultiversX.
3.  **Gasless Transactions:** Support MultiversX Relayed Transactions V3, where the Credentials Provider pays the gas.
4.  **Native Token Support:** Support ESDT (Elrond Standard Digital Token) transfers alongside EGLD.

### Non-Functional Requirements
1.  **Architecture:** "Functional Core, Imperative Shell" – pure logic for transaction construction; separate logic for side-effects (networking).
2.  **Security:** No private key storage within the core integration library.
3.  **Compliance:** No PII on-chain; GDPR compliance via off-chain mandates.
4.  **Standards:** Strict adherence to CAIP-2 (`mvx`), CAIP-10, and W3C DID specifications (Namespace registered).

## 5. System Architecture
The integration bridges the off-chain AP2 authorization layer with the on-chain MultiversX settlement layer.

### high-Level Data Flow
1.  **User** signs **Intent Mandate** (Off-chain VC).
2.  **Shopping Agent** negotiates with **Merchant Agent** (A2A Protocol).
3.  **Merchant** sends `payment-required` (x402 invoice).
4.  **Shopping Agent** constructs **Inner Transaction** (ESDT Transfer, Value=0).
5.  **User/Agent** signs Inner Transaction (`did:pkh:mvx`).
6.  **Credentials Provider** receives signed inner tx, wraps it in **Relayed V3 Tx**, signs as Relayer, and broadcasts.
7.  **MultiversX Blockchain** settles the transfer.
8.  **Merchant** observes settlement and releases goods.

### Tech Stack
-   **Language:** Python (to match `a2a-x402` and `AP2`).
-   **SDK:** `multiversx-sdk-core` (Transaction logic), `multiversx-sdk-wallet` (Signing/Addresses).
-   **Infrastructure:** MultiversX Gateway (API), Relayer Service.

## 6. UI/UX Specifications
*Note: This integration is primarily headless (Agent-to-Agent).*
-   **User Authorization:** Users sign the initial Mandate via a wallet (e.g., xPortal, Ledger) once.
-   **Agent Operation:** Invisible to the user; agents handle the heavy lifting.
-   **Transparency:** Users can view the "Chain of Trust" (Mandate → Transaction) in a block explorer if needed.

## 7. Implementation Plan
1.  **Standardization:** Register `mvx` in CAIP namespaces.
2.  **Core Implementation (`multiversx.py`):**
    -   Implement `create_payment_request`.
    -   Implement `construct_unsigned_transaction` (Inner Tx).
    -   Implement `verify_transaction_content`.
3.  **Testing:** Unit tests mocking the SDK; Integration tests on Devnet.
4.  **Documentation:** Developer guides for setting up the MultiversX scheme.

## 8. Success Metrics
-   **Zero Gas Leaks:** Agents never need EGLD balance.
-   **Latency:** Settlement detection < 6 seconds (MultiversX block time).
-   **Adoption:** Official merge into `google-agentic-commerce` repositories.

## 9. Open Questions & Assumptions
-   **Assumption:** The `AP2` Mandate structure is flexible enough to hold `ESDT` token identifiers without core schema changes.
-   **Question:** Is there a canonical public Relayer service for MultiversX we can default to, or must every CP run their own?

## 10. Glossary
-   **AP2:** Agent Payments Protocol.
-   **ESDT:** Elrond Standard Digital Token.
-   **Relayed V3:** A transaction type where a third party pays the gas.
-   **VC:** Verifiable Credential.
