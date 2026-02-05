import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { UserSigner, Mnemonic } from '@multiversx/sdk-wallet';
import { Address, Transaction, TransactionComputer } from '@multiversx/sdk-core';
import axios from 'axios';
import { spawn, ChildProcess } from 'child_process';
import path from 'path';
import * as fs from 'fs';

const RELAYER_PORT = 3000;
const API_URL = "http://localhost:" + RELAYER_PORT;
const PROJECT_ROOT = path.resolve(__dirname, "../../multiversx-openclaw-relayer");

describe("MultiversX Relay Service E2E", () => {
    let relayerProcess: ChildProcess;

    beforeAll(async () => {
        // Ensure dist exists
        if (!fs.existsSync(path.join(PROJECT_ROOT, "dist/index.js"))) {
            throw new Error("Relayer dist not found. Please run 'npm run build' in multiversx-openclaw-relayer first.");
        }

        console.log("Starting Relayer Service...");
        // Spawn the node process
        relayerProcess = spawn('node', ['dist/index.js'], {
            cwd: PROJECT_ROOT,
            env: {
                ...process.env,
                PORT: RELAYER_PORT.toString(),
                NETWORK_PROVIDER: "https://devnet-gateway.multiversx.com",
                QUOTA_LIMIT: "100",
                DB_PATH: ":memory:"
            }
        });

        relayerProcess.stdout?.on('data', (data) => console.log(`[Relayer]: ${data}`));
        relayerProcess.stderr?.on('data', (data) => console.error(`[Relayer Error]: ${data}`));

        // Wait for health check
        let healthy = false;
        for (let i = 0; i < 20; i++) {
            try {
                await axios.get(`${API_URL}/health`);
                healthy = true;
                break;
            } catch (e) {
                await new Promise(r => setTimeout(r, 500));
            }
        }

        if (!healthy) {
            throw new Error("Relayer failed to start");
        }
        console.log("Relayer is healthy.");
    });

    afterAll(() => {
        if (relayerProcess) {
            relayerProcess.kill();
        }
    });

    it("should accept a valid signed transaction and attempt to relay it", async () => {
        // 1. Setup User
        const mnemonic = Mnemonic.generate();
        const secretKey = mnemonic.deriveKey(0);
        const userSigner = new UserSigner(secretKey);
        const address = userSigner.getAddress().bech32();

        // 2. Create Transaction
        // Omit data to match server behavior (which omits if empty)

        const tx = new Transaction({
            nonce: 0n,
            value: 0n,
            receiver: new Address(address), // send to self
            sender: new Address(address),
            gasPrice: 1000000000n,
            gasLimit: 50000n,
            chainID: "D",
            version: 1, // Explicit version 1
            options: 0
        });

        // 3. Sign
        const computer = new TransactionComputer();
        const serialized = computer.computeBytesForSigning(tx);
        const signature = await userSigner.sign(serialized);
        tx.signature = signature;

        console.log("Tx Plain Object:", JSON.stringify(tx.toPlainObject(), null, 2));

        // 4. Send to Relayer
        try {
            const response = await axios.post(`${API_URL}/relay`, tx.toPlainObject());

            console.log("Relay Response:", response.data);
            expect(response.status).toBe(200);
            expect(response.data.txHash).toBeDefined();

        } catch (e: any) {
            const status = e.response?.status;
            const data = e.response?.data;
            console.log("Relay Request Failed (Expected if no funds):", status, data);

            if (data && data.error) {
                const err = data.error;
                // Assert that we didn't fail signature validation
                expect(err).not.toMatch(/Invalid inner transaction signature/);
                expect(err).not.toMatch(/Invalid transaction payload/);
            } else {
                throw e; // Unknown failure
            }
        }
    });

    it("should reject invalid signatures", async () => {
        const mnemonic = Mnemonic.generate();
        const userSigner = new UserSigner(mnemonic.deriveKey(0));
        const address = userSigner.getAddress().bech32();

        const tx = new Transaction({
            nonce: 0n,
            value: 0n,
            receiver: new Address(address), // send to self
            sender: new Address(address),
            gasPrice: 1000000000n,
            gasLimit: 50000n,
            chainID: "D",
            version: 1, // Explicit version 1
            options: 0
        });

        // Sign with WRONG content or don't sign
        tx.signature = Buffer.from("00".repeat(64), 'hex');

        try {
            await axios.post(`${API_URL}/relay`, tx.toPlainObject());
            expect(true).toBe(false); // Should not reach here
        } catch (e: any) {
            expect(e.response.status).toBe(400); // Bad Request
            expect(e.response.data.error).toMatch(/Invalid inner transaction/);
        }
    });
});
