package testdata

import (
	"github.com/olympus-protocol/ogen/primitives"
)

var TransferSinglePayload = primitives.TransferSinglePayload{
	To:            [20]byte{0x1, 0x2, 0x5, 0x10},
	FromPublicKey: pubB,
	Amount:        100,
	Nonce:         100,
	Fee:           100,
	Signature:     sigB,
}

var TransferSinglePayloadBytes, _ = TransferSinglePayload.Marshal()

var TransferMultiPayload = primitives.TransferMultiPayload{
	To:       [20]byte{0x1, 0x2, 0x5, 0x10},
	Amount:   100,
	Nonce:    100,
	Fee:      100,
	MultiSig: []byte{0x0},
}

var TransferMultiPayloadBytes, _ = TransferMultiPayload.Marshal()

var TxSingle = primitives.Tx{
	Version: 1,
	Type:    primitives.TxTransferSingle,
	Payload: TransferSinglePayloadBytes,
}

var TxMulti = primitives.Tx{
	Version: 1,
	Type:    primitives.TxTransferSingle,
	Payload: TransferMultiPayloadBytes,
}
