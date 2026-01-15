const { TransactionPayload } = require('@multiversx/sdk-core');
console.log('Type:', typeof TransactionPayload);
try {
    const tp = new TransactionPayload("hello");
    console.log('Instance created:', tp);
} catch (e) {
    console.log('Error creating instance:', e.message);
}
console.log('Exports:', require('@multiversx/sdk-core'));
