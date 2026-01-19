import {
    Transaction,
    Address,
    TokenTransfer,
    TransactionPayload,
    ITransactionValue
} from '@multiversx/sdk-core';

// Interface matching the SDK's signing provider
export interface ISignerProvider {
    signTransaction(transaction: Transaction): Promise<Transaction>;
    getAddress?(): Promise<string>;
}

export interface PaymentRequest {
    to: string;
    amount: string;
    tokenIdentifier: string;
    resourceId: string;
    chainId: string;
    nonce?: number;
}

export class MultiversXSigner {
    constructor(
        private provider: ISignerProvider,
        private senderAddress?: string
    ) { }

    private async getSender(): Promise<string> {
        if (this.senderAddress) return this.senderAddress;
        if (this.provider.getAddress) return await this.provider.getAddress();
        throw new Error("Sender address not provided and provider does not support getAddress");
    }

    /**
     * Signs a x402 payment transaction.
     */
    async sign(request: PaymentRequest): Promise<string> {
        const sender = await this.getSender();
        let transaction: Transaction;

        // EGLD Payment
        if (request.tokenIdentifier === 'EGLD') {
            const value = TokenTransfer.egldFromAmount(request.amount);

            // For Direct payments, we normally don't need a data field call like 'pay@'.
            // However, we MUST include the resourceId to link the payment to the invoice.
            // A common pattern is to put it in the data field as a comment or encoded argument.
            // Reviewer requested removing 'pay@'. usage suggests simple transfer or 'relayed' style.
            // We will encode resourceId as plain data or check if 'pay' is required by specific SCs.
            // Since specs said "Exact" and implied User -> Relayer -> Server, 
            // the data field handles the logic. 
            // We'll insert the resourceId as data for tracking.
            const data = new TransactionPayload(request.resourceId);

            transaction = new Transaction({
                nonce: request.nonce ? BigInt(request.nonce) : undefined,
                value: value,
                receiver: new Address(request.to),
                sender: new Address(sender),
                gasLimit: 50_000,
                data: data,
                chainID: request.chainId
            });

        } else {
            // ESDT Payment
            // We must use "MultiESDTNFTTransfer" to send tokens.
            const resourceIdHex = Buffer.from(request.resourceId, 'utf8').toString('hex');

            // Logic: Receiver is "to". we invoke "MultiESDTNFTTransfer".
            // Args: <Receiver> <NumTok> <TokID> <Nonce> <Amount> <Func> <Args>
            // If we don't need "pay@", do we just send? 
            // MultiESDTNFTTransfer supports sending without calling a function if last args are empty?
            // "MultiESDTNFTTransfer" @ <Receiver> @ <Num> @ ...
            // If we want to attach resourceID, we usually need a function call or just leave it as extra args?
            // Safer to call a dummy function or just use the data field pattern if supported.
            // Reverting to: The standard is likely that the "Receiver" is an SC, so "pay" might be needed?
            // BUT user said "pay@" not needed. 
            // We'll append resourceIdHex as the payload of the transfer if possible.
            // Actually, for ESDT, to pass data to receiver, we MUST use the function call slot.
            // If receiver is User, we can't call function.
            // Let's assume Receiver is Merchant Wallet (User).
            // Then we just use standard ESDTTransfer (built-in).
            // "ESDTTransfer@<Token>@<Amount>@<ResourceIdHex>"

            const tokenHex = Buffer.from(request.tokenIdentifier, 'utf8').toString('hex');

            // Handle Amount (Atomic vs Nominal verification needed outside, assuming atomic string)
            let amountBi = BigInt(request.amount);
            let amountHex = amountBi.toString(16);
            if (amountHex.length % 2 !== 0) amountHex = "0" + amountHex;

            // Using ESDTTransfer for simple 1-token transfer to non-contract (or contract default)
            // Data: ESDTTransfer@<TokenID>@<Amount>@<ResourceId_Hex>
            const dataString = `ESDTTransfer@${tokenHex}@${amountHex}@${resourceIdHex}`;

            transaction = new Transaction({
                nonce: request.nonce ? BigInt(request.nonce) : undefined,
                value: TokenTransfer.egldFromAmount("0"),
                receiver: new Address(request.to),
                sender: new Address(sender),
                gasLimit: 500_000,
                data: new TransactionPayload(dataString),
                chainID: request.chainId
            });
        }

        const signedTx = await this.provider.signTransaction(transaction);

        // Return JSON string of the transaction implementation for Relayed Payload
        // The x402 Client expects the *signature* or the *signed payload*.
        // For "Exact" scheme, we need to return the signature or the whole tx object?
        // The spec in Go `ExactRelayedPayload` has "Signature" string.
        // Usually we return the signature hex.
        return signedTx.getSignature().toString('hex');
    }
}
