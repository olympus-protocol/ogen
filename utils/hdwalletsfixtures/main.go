package main

import (
	"encoding/hex"
	"encoding/json"
	"os"

	"github.com/olympus-protocol/ogen/utils/hdwallets"
)

// "invalid": {
// 	"fromBase58": [
// 		{
// 			"exception": "Invalid checksum",
// 			"string": "xprvQQQQQQQQQQQQQQQQCviVfJSKyQ1mDYahRjijr5idH2WwLsEd4Hsb2Tyh8RfQMuPh7f7RtyzTtdrbdqqsunu5Mm3wDvUAKRHSC34sJ7in334"
// 		},
// 		{
// 			"exception": "Invalid buffer length",
// 			"string": "HAsbc6CgKmTYEQg2CTz7m5STEPAB"
// 		},
// 		{
// 			"exception": "Invalid parent fingerprint",
// 			"string": "xprv9tnJFvAXAXPfPnMTKfwpwnkty7MzJwELVgp4NTBquaKXy4RndyfJJCJJf7zNaVpBpzrwVRutZNLRCVLEcZHcvuCNG3zGbGBcZn57FbNnmSP"
// 		},
// 		{
// 			"exception": "Invalid private key",
// 			"string": "xprv9s21ZrQH143K3yLysFvsu3n1dMwhNusmNHr7xArzAeCc7MQYqDBBStmqnZq6WLi668siBBNs3SjiyaexduHu9sXT9ixTsqptL67ADqcaBdm"
// 		},
// 		{
// 			"exception": "Invalid index",
// 			"string": "xprv9s21ZrQYdgnodnKW4Drm1Qg7poU6Gf2WUDsjPxvYiK7iLBMrsjbnF1wsZZQgmXNeMSG3s7jmHk1b3JrzhG5w8mwXGxqFxfrweico7k8DtxR"
// 		},
// 		{
// 			"exception": "Invalid network version",
// 			"string": "1111111111111adADjFaSNPxwXqLjHLj4mBfYxuewDPbw9hEj1uaXCzMxRPXDFF3cUoezTFYom4sEmEVSQmENPPR315cFk9YUFVek73wE9"
// 		},
// 		{
// 			"exception": "Invalid network version",
// 			"string": "8FH81Rao5EgGmdScoN66TJAHsQP7phEMeyMTku9NBJd7hXgaj3HTvSNjqJjoqBpxdbuushwPEM5otvxXt2p9dcw33AqNKzZEPMqGHmz7Dpayi6Vb"
// 		},
// 		{
// 			"exception": "Invalid network version",
// 			"string": "Ltpv73XYpw28ZyVe2zEVyiFnxUZxoKLGQNdZ8NxUi1WcqjNmMBgtLbh3KimGSnPHCoLv1RmvxHs4dnKmo1oXQ8dXuDu8uroxrbVxZPA1gXboYvx"
// 		},
// 		{
// 			"exception": "Invalid buffer length",
// 			"string": "9XpNiB4DberdMn4jZiMhNGtuZUd7xUrCEGw4MG967zsVNvUKBEC9XLrmVmFasanWGp15zXfTNw4vW4KdvUAynEwyKjdho9QdLMPA2H5uyt"
// 		},
// 		{
// 			"exception": "Invalid buffer length",
// 			"string": "7JJikZQ2NUXjSAnAF2SjFYE3KXbnnVxzRBNddFE1DjbDEHVGEJzYC7zqSgPoauBJS3cWmZwsER94oYSFrW9vZ4Ch5FtGeifdzmtS3FGYDB1vxFZsYKgMc"
// 		},
// 		{
// 			"exception": "Invalid parent fingerprint",
// 			"string": "xpub67tVq9SuNQCfm2PXBqjGRAtNZ935kx2uHJaURePth4JBpMfEy6jum7Euj7FTpbs7fnjhfZcNEktCucWHcJf74dbKLKNSTZCQozdDVwvkJhs"
// 		},
// 		{
// 			"exception": "Invalid index",
// 			"string": "xpub661MyMwTWkfYZq6BEh3ywGVXFvNj5hhzmWMhFBHSqmub31B1LZ9wbJ3DEYXZ8bHXGqnHKfepTud5a2XxGdnnePzZa2m2DyzTnFGBUXtaf9M"
// 		},
// 		{
// 			"exception": "Point is not on the curve",
// 			"string": "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gYymDsxxRe3WWeZQ7TadaLSdKUffezzczTCpB8j3JP96UwE2n6w1"
// 		}
// 	],
// 	"fromSeed": [
// 		{
// 			"exception": "Seed should be at least 128 bits",
// 			"seed": "ffff"
// 		},
// 		{
// 			"exception": "Seed should be at most 512 bits",
// 			"seed": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
// 		}
// 	],
// 	"deriveHardened": [
// 		2147483648,
// 		null,
// 		"foo",
// 		-1
// 	],
// 	"derive": [
// 		4294967296,
// 		null,
// 		"foo",
// 		-1
// 	],
// 	"derivePath": [
// 		2,
// 		[
// 			2,
// 			3,
// 			4
// 		],
// 		"/",
// 		"m/m/123",
// 		"a/0/1/2",
// 		"m/0/  1  /2",
// 		"m/0/1.5/2"
// 	]
// }

var PolisNetPrefix = &hdwallets.NetPrefix{
	ExtPub:  []byte{0x1f, 0x74, 0x90, 0xf0},
	ExtPriv: []byte{0x11, 0x24, 0xd9, 0x70},
}

type MasterKeyInfo struct {
	Seed        string `json:"seed"`
	PubKey      string `json:"pubKey"`
	PrivKey     string `json:"privKey"`
	ChainCode   string `json:"chainCode"`
	Base58      string `json:"base58"`
	Base58Priv  string `json:"base58Priv"`
	Identifier  string `json:"identifier"`
	Fingerprint string `json:"fingerprint"`
}

type ChildKeyInfo struct {
	Path        string `json:"path"`
	M           int    `json:"m"`
	Hardened    bool   `json:"hardened"`
	Index       int    `json:"index"`
	Depth       int    `json:"depth"`
	PubKey      string `json:"pubKey"`
	PrivKey     string `json:"privKey"`
	ChainCode   string `json:"chainCode"`
	Base58      string `json:"base58"`
	Base58Priv  string `json:"base58Priv"`
	Identifier  string `json:"identifier"`
	Fingerprint string `json:"fingerprint"`
}

type ValidKey struct {
	Comment  *string        `json:"comment"`
	Master   MasterKeyInfo  `json:"master"`
	Children []ChildKeyInfo `json:"children"`
}

type TestFixture struct {
	Valid []ValidKey
}

func generateValidKey(seedStr string) ValidKey {
	seed, _ := hex.DecodeString(seedStr)

	master, err := hdwallets.NewMaster(seed, PolisNetPrefix)
	if err != nil {
		panic(err)
	}
	masterPub, err := master.Neuter(PolisNetPrefix)
	if err != nil {
		panic(err)
	}
	id, _ := master.Identifier()
	fp, _ := master.Fingerprint()

	priv, err := master.BlsPrivKey()
	if err != nil {
		panic(err)
	}
	pub, err := master.BlsPubKey()
	if err != nil {
		panic(err)
	}
	return ValidKey{
		Comment: nil,
		Master: MasterKeyInfo{
			Seed:        seedStr,
			PubKey:      hex.EncodeToString(pub.Marshal()),
			PrivKey:     hex.EncodeToString(priv.Marshal()),
			ChainCode:   hex.EncodeToString(master.ChainCode()),
			Base58:      masterPub.String(),
			Base58Priv:  master.String(),
			Identifier:  hex.EncodeToString(id),
			Fingerprint: hex.EncodeToString(fp),
		},
	}
}

func main() {
	validKey := generateValidKey("000102030405060708090a0b0c0d0e0f")
	out, err := json.MarshalIndent(validKey, "", "  ")
	if err != nil {
		panic(err)
	}

	os.Stdout.Write(out)
}
