package votes_txverifier

import (
	"github.com/grupokindynos/ogen/chain/index"
	"github.com/grupokindynos/ogen/p2p"
	"github.com/grupokindynos/ogen/params"
	votes_txpayload "github.com/grupokindynos/ogen/txs/txpayloads/votes"
	"github.com/grupokindynos/ogen/utils/chainhash"
)

var workerIndexMock = &index.WorkerIndex{
	Index: map[chainhash.Hash]*index.WorkerRow{},
}

var votes VotesTxVerifier

func init() {
	votes = NewVotesTxVerifier(workerIndexMock, &params.Mainnet)
}

var mockPayloadUpload1 = votes_txpayload.PayloadUploadAndUpdate{
	WorkerID: p2p.OutPoint{},
	PubKey:   [48]byte{},
	Sig:      [96]byte{},
	GovID:    chainhash.Hash{},
	Approval: false,
}
var mockPayloadUpload2 = votes_txpayload.PayloadUploadAndUpdate{
	WorkerID: p2p.OutPoint{},
	PubKey:   [48]byte{},
	Sig:      [96]byte{},
	GovID:    chainhash.Hash{},
	Approval: false,
}
var mockPayloadUpload3 = votes_txpayload.PayloadUploadAndUpdate{
	WorkerID: p2p.OutPoint{},
	PubKey:   [48]byte{},
	Sig:      [96]byte{},
	GovID:    chainhash.Hash{},
	Approval: false,
}
var mockPayloadUpload4 = votes_txpayload.PayloadUploadAndUpdate{
	WorkerID: p2p.OutPoint{},
	PubKey:   [48]byte{},
	Sig:      [96]byte{},
	GovID:    chainhash.Hash{},
	Approval: false,
}
var mockPayloadUpload5 = votes_txpayload.PayloadUploadAndUpdate{
	WorkerID: p2p.OutPoint{},
	PubKey:   [48]byte{},
	Sig:      [96]byte{},
	GovID:    chainhash.Hash{},
	Approval: false,
}

var mockPayloadUploadBatch = []votes_txpayload.PayloadUploadAndUpdate{mockPayloadUpload1, mockPayloadUpload2, mockPayloadUpload3, mockPayloadUpload4, mockPayloadUpload5}

var mockPayloadRevoke1 = votes_txpayload.PayloadRevoke{
	WorkerID: p2p.OutPoint{},
	PubKey:   [48]byte{},
	Sig:      [96]byte{},
	GovID:    chainhash.Hash{},
}
var mockPayloadRevoke2 = votes_txpayload.PayloadRevoke{
	WorkerID: p2p.OutPoint{},
	PubKey:   [48]byte{},
	Sig:      [96]byte{},
	GovID:    chainhash.Hash{},
}
var mockPayloadRevoke3 = votes_txpayload.PayloadRevoke{
	WorkerID: p2p.OutPoint{},
	PubKey:   [48]byte{},
	Sig:      [96]byte{},
	GovID:    chainhash.Hash{},
}
var mockPayloadRevoke4 = votes_txpayload.PayloadRevoke{
	WorkerID: p2p.OutPoint{},
	PubKey:   [48]byte{},
	Sig:      [96]byte{},
	GovID:    chainhash.Hash{},
}
var mockPayloadRevoke5 = votes_txpayload.PayloadRevoke{
	WorkerID: p2p.OutPoint{},
	PubKey:   [48]byte{},
	Sig:      [96]byte{},
	GovID:    chainhash.Hash{},
}

var mockPayloadRevokeBatch = []votes_txpayload.PayloadRevoke{mockPayloadRevoke1, mockPayloadRevoke2, mockPayloadRevoke3, mockPayloadRevoke4, mockPayloadRevoke5}
