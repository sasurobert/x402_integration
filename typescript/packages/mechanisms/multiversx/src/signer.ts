import {
    Transaction,
    Address,
    TokenPayment,
    TransactionPayload,
    ITransactionValue,
    TokenTransfer
} from '@multiversx/sdk-core';

// Interface for the Wallet Provider (e.g., Extension, Ledger)
interface ISignerProvider {
    signTransaction(transaction: Transaction): Promise<Transaction>;
    // Some providers might separate signing and sending
}

export interface PaymentRequest {
    to: string; // The Payment SC Address
    amount: string; // Amount as string (e.g. "1.5" or "1000000000000000000")
    tokenIdentifier: string; // "EGLD" or "TOKEN-123456"
    resourceId: string; // Unique Invoice ID / Resource ID
    chainId: string;
    nonce?: number;
}

export class MultiversXSigner {
    constructor(
        private provider: ISignerProvider,
        private senderAddress: string
    ) { }

    /**
     * Signs a x402 payment transaction.
     * Handles both EGLD (Direct) and ESDT (Transfer & Execute) payments.
     */
    async sign(request: PaymentRequest): Promise<string> {
        let transaction: Transaction;

        // 1. Prepare Function Call: pay@<resource_id_hex>
        // resourceId is assumed to be a string that needs hex encoding for Smart Contract arguments
        const resourceIdBuff = Buffer.from(request.resourceId, 'utf8');
        const resourceIdHex = resourceIdBuff.toString('hex');

        // Logic split: EGLD vs ESDT
        if (request.tokenIdentifier === 'EGLD') {
            // Case A: Direct EGLD Payment
            // data: pay@<resource_id_hex>
            const data = new TransactionPayload(`pay@${resourceIdHex}`);
            const value = TokenTransfer.egldFromAmount(request.amount);

            transaction = new Transaction({
                nonce: request.nonce ? BigInt(request.nonce) : undefined, // Caller usually manages nonce, or we fetch it
                value: value,
                receiver: new Address(request.to),
                sender: new Address(this.senderAddress),
                gasLimit: 10_000_000, // Standard SC call limit
                data: data,
                chainID: request.chainId
            });

        } else {
            // Case B: ESDT Payment
            // Needs to use MultiESDTNFTTransfer to send token + trigger 'pay' function
            // We use the built-in TokenTransfer to create the payload

            const payment = TokenTransfer.fungibleFromAmount(
                request.tokenIdentifier,
                request.amount,
                0 // Decimals - strictly we should know decimals. 
                // CRITICAL NOTE: sdk-core requires decimals to parse '1.5'. 
                // If 'request.amount' is atomic, we use 0. If it's nominal, we need decimals.
                // For x402, we will assume 'request.amount' is ALREADY ATOMIC units or we need decimal metadata.
                // Let's assume Atomic Units (BigInt string) for safety in protocols.
            );

            // To construct the transfer:
            // The destination is the Payment SC (request.to)
            // The method to call on destination is 'pay'
            // The argument is resourceIdHex

            // Using a factory or manual construction for MultiESDTNFTTransfer
            // Method: "MultiESDTNFTTransfer" @ <Receiver> @ <NumTokens> @ <TokenID> @ <Nonce> @ <Amount> @ <Func> @ <Args>

            // Note: sdk-core 'TokenTransfer' helper might simpler but let's be explicit for the factory
            // Actually, we can just use the Payload Builder if we had the Factory. 
            // Let's manually construct the data string for zero-dep safety if factory is complex to init.

            // Format: MultiESDTNFTTransfer@<dest_hex>@01@<token_hex>@00@<amount_hex>@pay@<resource_id_hex>

            const receiver = new Address(request.to);
            const tokenHex = Buffer.from(request.tokenIdentifier, 'utf8').toString('hex');
            // Amount must be hex, even length
            let amountBi = BigInt(request.amount); // Assuming atomic
            let amountHex = amountBi.toString(16);
            if (amountHex.length % 2 !== 0) amountHex = "0" + amountHex;

            const payHex = Buffer.from("pay", 'utf8').toString('hex');

            const dataString = `MultiESDTNFTTransfer@${receiver.hex()}@01@${tokenHex}@00@${amountHex}@${payHex}@${resourceIdHex}`;

            transaction = new Transaction({
                nonce: request.nonce ? BigInt(request.nonce) : undefined,
                value: TokenTransfer.egldFromAmount("0"), // 0 EGLD for token transfers
                receiver: new Address(this.senderAddress), // Sender sends to Self for MultiESDTNFTTransfer
                sender: new Address(this.senderAddress),
                gasLimit: 15_000_000, // Slightly higher for ESDT
                data: new TransactionPayload(dataString),
                chainID: request.chainId
            });
        }

        // 2. Sign
        const signedTx = await this.provider.signTransaction(transaction);

        // 3. Return Hash
        return signedTx.getHash().toString();
    }
}
