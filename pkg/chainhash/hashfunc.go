// Copyright (c) 2015 The Decred developers
// Copyright (c) 2016-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chainhash

import (
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"hash"
)

// HashB calculates hash(b) and returns the resulting bytes.
func HashB(b []byte) []byte {
	h := sha256.Sum256(b)
	return h[:]
}

// HashH calculates hash(b) and returns the resulting bytes as a Hash.
func HashH(b []byte) Hash {
	return sha256.Sum256(b)
}

func DoubleHashB(b []byte) []byte {
	first := sha256.Sum256(b)
	second := sha256.Sum256(first[:])
	return second[:]
}

func DoubleHashH(b []byte) Hash {
	first := sha256.Sum256(b)
	return sha256.Sum256(first[:])
}

func calcHash(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}

func Hash160(buf []byte) []byte {
	return calcHash(calcHash(buf, sha256.New()), ripemd160.New())
}
