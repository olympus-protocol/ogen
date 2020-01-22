# Olympus Tx Package

This package contains an implementation for serialize, deserialize, sign and verify different Olympus transactions.

Unfortunately the security of this package relies on the indexed state of the blockchain, it is not possible to use outside the content of the Olympus code.

The library is separated on tree main packages: 

1. `txverifier:` Used to verify the transactions with the payloads.
2. `txpayloads:` Used to encode, decode payloads.
3. `txbuilder:` Used to create and sign transactions.

For more information about the Olympus transactions system please refer to [OLYP-0002](https://github.com/grupokindynos/olyps/blob/master/olyps/olyp-0002.md) and [OLYP-0004](https://github.com/grupokindynos/olyps/blob/master/olyps/olyp-0004.md)