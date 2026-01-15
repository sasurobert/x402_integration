# MultiversX <> Google Agentic Commerce Integration Roadmap

## Rationale for Previous Actions
The initial code changes (creation of `multiversx.py` and test files) were proposed based on the **[ap2.md](ap2.md)** document provided in the active workspace. Section **5.3 Implementation Specification** of that document explicitly instructed:
> "To achieve this, a new scheme file must be created in the a2a-x402 repository. File Path: src/x402_a2a/schemes/multiversx.py"

I have now reverted these changes as requested to allow for a structured, official integration process starting from a clean state.

---

## Official Integration Roadmap

To officially integrate MultiversX as a payment rail for the Google Agentic Commerce ecosystem, the following steps must be taken. This process adheres to the open-source contribution standards for the [google-agentic-commerce](https://github.com/google-agentic-commerce) repositories.

### 1. Standardization (Verified)
The MultiversX chain namespace `mvx` is already registered in the Chain Agnostic namespaces (CAIP-2).
-   **Link**: [ChainAgnostic/namespaces/mvx](https://github.com/ChainAgnostic/namespaces/tree/main/mvx)
-   **Status**: **Ready**. No further standardization work required.

### 2. Fork and Branch Strategy
You should not work directly on the `x402_repo` main branch if you intend to upstream the changes.
- **Step 1**: Fork the [a2a-x402](https://github.com/google-agentic-commerce/a2a-x402) repository to your organization (e.g., `multiversx/a2a-x402`).
- **Step 2**: Clone your fork locally.
- **Step 3**: Create a feature branch:
  ```bash
  git checkout -b feat/add-multiversx-scheme
  ```

### 3. Implementation (The "Functional Core")
Re-apply the implementation logic but within the correct feature branch structure.
- **File**: `python/x402/src/x402/schemes/multiversx.py`
- **Content**: Implement `MultiversXScheme` class inheriting from the protocol's base classes. 
- **Dependencies**: Add `multiversx-sdk-core` and `multiversx-sdk-wallet` to `pyproject.toml` under `[project.optional-dependencies]`.

### 4. Testing
Google repositories require high test coverage.
- **Unit Tests**: Create `python/x402/tests/test_multiversx.py` mocking the network calls.
- **Integration Tests**: Provide a script in `examples/` demonstrating a flow on Devnet.
- **Run Tests**: Ensure `hatch run dev:test` (or equivalent) passes locally.

### 5. Pull Request (PR) Process
Once the code is ready:
- **CLA**: Ensure you have signed the Google [Contributor License Agreement (CLA)](https://cla.developers.google.com/).
- **PR Target**: Open a Pull Request against `google-agentic-commerce/a2a-x402:main`.
- **Description**:
  - Title: `feat: Add MultiversX (mvx) Payment Scheme`
  - Body: Explain the integration, reference the CAIP standard, and link to the MultiversX SDKs used.
  - Context: Mention support for "Relayed Transactions V3" as a key differentiator (Gasless Agents).

### 6. Documentation
- Update `README.md` in the upstream repo (as part of your PR) to list MultiversX as a supported scheme.
- Provide a `docs/multiversx.md` guide for setting up the environment variables (e.g., `MVX_CHAIN_ID`).

### 7. Governance & Extensions (AP2)
If changes are required to the Mandate structure (e.g., specific fields for ESDT tokens in the `PaymentMandate`):
- **Repo**: [google-agentic-commerce/AP2](https://github.com/google-agentic-commerce/AP2)
- **Process**: Similar Fork -> Branch -> PR flow.
- **Note**: The current proposal (x402 scheme) likely fits within the existing `PaymentMandate` flexible payload structure, minimizing the need for AP2 core changes.
