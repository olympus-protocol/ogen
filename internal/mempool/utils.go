package mempool

import (
	"encoding/binary"
)

type PoolType uint64

const (
	PoolTypeDeposit PoolType = iota
	PoolTypeExit
	PoolTypePartialExit
	PoolTypeLatestNonce
	PoolTypeGovernanceVote
	PoolTypeCoinProof
	PoolTypeVote
	PoolTypeTx
)

func appendKey(k []byte, t PoolType) []byte {
	var key []byte
	switch t {
	case PoolTypeDeposit:
		key = append(key, []byte("deposit-")...)
		key = append(key, k...)
		return key
	case PoolTypeExit:
		key = append(key, []byte("exit-")...)
		key = append(key, k...)
		return key
	case PoolTypePartialExit:
		key = append(key, []byte("partial_exit-")...)
		key = append(key, k...)
		return key
	case PoolTypeLatestNonce:
		key = append(key, []byte("latest_nonce-")...)
		key = append(key, k...)
		return key
	case PoolTypeGovernanceVote:
		key = append(key, []byte("governance_vote-")...)
		key = append(key, k...)
		return key
	case PoolTypeCoinProof:
		key = append(key, []byte("coin_proof-")...)
		key = append(key, k...)
		return key
	case PoolTypeVote:
		key = append(key, []byte("vote-")...)
		key = append(key, k...)
		return key
	case PoolTypeTx:
		key = append(key, []byte("-tx-")...)
		key = append(key, k...)
		return key
	default:
		return k
	}
}

func appendKeyWithNonce(k [20]byte, nonce uint64) [28]byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, nonce)
	var key [28]byte
	copy(key[0:20], k[:])
	copy(key[20:27], buf)
	return key
}

func nonceToBytes(nonce uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, nonce)
	return b
}
