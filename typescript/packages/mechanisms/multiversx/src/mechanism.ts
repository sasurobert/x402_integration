import { PaymentRequest } from './signer';

export class MultiversXMechanism {
    /**
     * Helper to generate the details required for the x402 header.
     * In the standard, the server usually sends a challenge. 
     * This helper can be used by the server (node) or client to format the payment details.
     */
    static createPaymentDetails(
        recipient: string,
        amount: string,
        tokenIdentifier: string,
        resourceId: string,
        chainId: string
    ): PaymentRequest {
        return {
            to: recipient,
            amount,
            tokenIdentifier,
            resourceId,
            chainId
        };
    }

    /**
     * Returns the 'mechanism' string identifier for registration.
     */
    static get name(): string {
        return 'multiversx';
    }
}
