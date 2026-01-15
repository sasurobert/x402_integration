const sdk = require('@multiversx/sdk-core');
console.log('Transaction related keys:', Object.keys(sdk).filter(k => k.includes('Transaction')));
console.log('Proto:', Object.getOwnPropertyNames(sdk.Transaction.prototype));
