Go BLS HD Wallet tools
------------------
> This library is using the BLS12-381 curve to generate HD Wallets.
> This should not be used for Bitcoin or any other cryptocurrency using the secp256k1 curve.
>
 - BIP32 - https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki
 - BLS Implementation - https://github.com/phoreproject/bls
 
### Get this library

        go get github.com/grupokindynos/ogen/utils/hdwallets

### Example

        // Generate a random 256 bit seed
        seed, err := hdwallets.GenerateSeed(256)

        // Create a master private key:
        // To create extended keys, you must add the prefix definitions.
        // To use Bitcoin defaults use nil as prefix param.
        masterprv := hdwallet.NewMaster(seed, nil)

        // Convert a private key to public key
        // To convert an extended key into the public form, you need to pass
        // de prefix defeinitions. To use the Bitcoin defaults, pass nil.
        masterpub := masterprv.Neuter(nil)
        
        // Generate hardened child key based on private or public key
        childprv, err := masterprv.Child(HardenedKeyStart + 0)
        childpub, err := masterpub.Child(HardenedKeyStart + 0)

        // Generate new child key based on private or public key
        childprv, err := masterprv.Child(0)
        childpub, err := masterpub.Child(0)
