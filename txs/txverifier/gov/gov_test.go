package gov_txverifier

import (
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/state"
	gov_txpayload "github.com/olympus-protocol/ogen/txs/txpayloads/gov"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

var chainState = &state.State{
	UtxoState: state.UtxoState{
		UTXOs: map[chainhash.Hash]state.Utxo{},
	},
	GovernanceState: state.GovernanceState{
		Proposals: map[chainhash.Hash]state.GovernanceProposal{},
	},
	UserState: state.UserState{
		Users: map[chainhash.Hash]state.User{},
	},
	WorkerState: state.WorkerState{
		Workers: map[chainhash.Hash]state.Worker{},
	},
}

var gov GovTxVerifier

func init() {
	rows := []state.Utxo{
		{
			OutPoint:          p2p.OutPoint{TxHash: chainhash.DoubleHashH([]byte("row-1")), Index: 1},
			PrevInputsPubKeys: [][48]byte{},
			Owner:             "olpub1tesqm4sstyq96lm92h5hqe4wsdp4cvz8d9pcrfzjkl7st2nz4gdsrecflm",
			Amount:            10 * 1e8,
		},
		{
			OutPoint:          p2p.OutPoint{TxHash: chainhash.DoubleHashH([]byte("row-2")), Index: 1},
			PrevInputsPubKeys: [][48]byte{},
			Owner:             "olpub1aerqc5azajv3vk2lmueckw2hy2f45jt5mxwk33k768ztf7vp80nshaghcx",
			Amount:            50 * 1e8,
		},
		{
			OutPoint:          p2p.OutPoint{TxHash: chainhash.DoubleHashH([]byte("row-3")), Index: 1},
			PrevInputsPubKeys: [][48]byte{},
			Owner:             "olpub17upfvla9r4nfedxtk73w37f87a3mppq7k0f86z5jhnethhq0zx0q9spntk",
			Amount:            1000 * 1e8,
		},
		{
			OutPoint:          p2p.OutPoint{TxHash: chainhash.DoubleHashH([]byte("row-4")), Index: 1},
			PrevInputsPubKeys: [][48]byte{},
			Owner:             "olpub1ejvtu2mwn5az2aja53w527a7wl6hyrk9mk5pqxxh9czu6c846n0qxvpx4k",
			Amount:            578 * 1e8,
		},
		{
			OutPoint:          p2p.OutPoint{TxHash: chainhash.DoubleHashH([]byte("budget-funds")), Index: 1},
			PrevInputsPubKeys: [][48]byte{},
			Owner:             "",
			Amount:            20000 * 1e8,
		},
	}
	for _, row := range rows {
		chainState.UtxoState.UTXOs[serializer.Hash(&row)] = row
	}
	gov = NewGovTxVerifier(chainState, &params.Mainnet)
}

var mockPayloadUpload1 = gov_txpayload.PayloadUpload{
	BurnedUtxo:    p2p.OutPoint{},
	PubKey:        [48]byte{},
	Sig:           [96]byte{},
	Name:          "",
	URL:           "",
	PayoutAddress: "",
	Amount:        0,
	Cycles:        0,
}
var mockPayloadUpload2 = gov_txpayload.PayloadUpload{
	BurnedUtxo:    p2p.OutPoint{},
	PubKey:        [48]byte{},
	Sig:           [96]byte{},
	Name:          "",
	URL:           "",
	PayoutAddress: "",
	Amount:        0,
	Cycles:        0,
}
var mockPayloadUpload3 = gov_txpayload.PayloadUpload{
	BurnedUtxo:    p2p.OutPoint{},
	PubKey:        [48]byte{},
	Sig:           [96]byte{},
	Name:          "",
	URL:           "",
	PayoutAddress: "",
	Amount:        0,
	Cycles:        0,
}
var mockPayloadUpload4 = gov_txpayload.PayloadUpload{
	BurnedUtxo:    p2p.OutPoint{},
	PubKey:        [48]byte{},
	Sig:           [96]byte{},
	Name:          "",
	URL:           "",
	PayoutAddress: "",
	Amount:        0,
	Cycles:        0,
}
var mockPayloadUpload5 = gov_txpayload.PayloadUpload{
	BurnedUtxo:    p2p.OutPoint{},
	PubKey:        [48]byte{},
	Sig:           [96]byte{},
	Name:          "",
	URL:           "",
	PayoutAddress: "",
	Amount:        0,
	Cycles:        0,
}

var mockPayloadUploadBatch = []gov_txpayload.PayloadUpload{mockPayloadUpload1, mockPayloadUpload2, mockPayloadUpload3, mockPayloadUpload4, mockPayloadUpload5}

var mockPayloadRevoke1 = gov_txpayload.PayloadRevoke{
	GovID:  chainhash.Hash{},
	PubKey: [48]byte{},
	Sig:    [96]byte{},
}
var mockPayloadRevoke2 = gov_txpayload.PayloadRevoke{
	GovID:  chainhash.Hash{},
	PubKey: [48]byte{},
	Sig:    [96]byte{},
}
var mockPayloadRevoke3 = gov_txpayload.PayloadRevoke{
	GovID:  chainhash.Hash{},
	PubKey: [48]byte{},
	Sig:    [96]byte{},
}
var mockPayloadRevoke4 = gov_txpayload.PayloadRevoke{
	GovID:  chainhash.Hash{},
	PubKey: [48]byte{},
	Sig:    [96]byte{},
}
var mockPayloadRevoke5 = gov_txpayload.PayloadRevoke{
	GovID:  chainhash.Hash{},
	PubKey: [48]byte{},
	Sig:    [96]byte{},
}

var mockPayloadRevokeBatch = []gov_txpayload.PayloadRevoke{mockPayloadRevoke1, mockPayloadRevoke2, mockPayloadRevoke3, mockPayloadRevoke4, mockPayloadRevoke5}
