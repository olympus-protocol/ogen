package coins_txverifier

import (
	"bytes"
	"github.com/grupokindynos/ogen/chain/index"
	"github.com/grupokindynos/ogen/p2p"
	"github.com/grupokindynos/ogen/params"
	coins_txpayload "github.com/grupokindynos/ogen/txs/txpayloads/coins"
	"github.com/grupokindynos/ogen/utils/chainhash"
	"reflect"
	"testing"
)

/*
6cc8363d16f80d549cd53201ad79c1a4c93d53c8f2298a8471e8102e68b25641
11f4c12ce99362efb8c0810110fd87b502d5b7cb2572e8c188f0fd0728fd52ff
72e1118268266eae0b0961c1a230632d54e3af088a5acdf77d22f925d454fad8
0d2ef3cbc1acf1730b8cc24bb2b578e6e7668747c387af15cda550d4e04dd5db
*/

var PubKey1 = [48]byte{147, 175, 97, 107, 196, 111, 244, 210, 42, 3, 68, 226, 10, 196, 245, 114, 65, 162, 83, 20, 217, 214, 216, 109, 181, 5, 142, 234, 245, 196, 17, 86, 41, 226, 151, 105, 137, 137, 189, 157, 212, 244, 51, 82, 30, 77, 116, 47}
var PubKey2 = [48]byte{173, 123, 139, 167, 43, 11, 143, 169, 86, 146, 49, 189, 165, 251, 151, 38, 225, 138, 137, 86, 8, 37, 216, 22, 192, 137, 223, 210, 214, 71, 30, 135, 93, 170, 192, 249, 99, 27, 226, 212, 9, 123, 110, 212, 79, 60, 12, 117}
var PubKey3 = [48]byte{134, 213, 17, 19, 3, 54, 137, 215, 44, 227, 142, 140, 201, 200, 156, 183, 116, 149, 80, 234, 1, 144, 145, 34, 138, 197, 76, 164, 33, 241, 184, 252, 46, 47, 107, 41, 180, 170, 88, 7, 93, 213, 180, 154, 141, 220, 222, 65}
var PubKey4 = [48]byte{176, 50, 126, 164, 6, 9, 215, 12, 25, 223, 42, 121, 157, 12, 168, 36, 37, 248, 79, 201, 208, 202, 119, 136, 192, 163, 121, 242, 182, 179, 84, 100, 251, 176, 221, 122, 186, 222, 228, 168, 169, 33, 33, 207, 104, 34, 121, 140}

var utxosIndexMock = &index.UtxosIndex{
	Index: map[chainhash.Hash]*index.UtxoRow{},
}

var coinTxVerifier CoinsTxVerifier

func init() {
	coinTxVerifier = NewCoinsTxVerifier(utxosIndexMock, &params.Mainnet)
	rows := []*index.UtxoRow{
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
		err := utxosIndexMock.Add(row)
		if err != nil {
			panic(err)
		}
	}
}

var mockPayloadValid1 = coins_txpayload.PayloadTransfer{
	AggSig: [96]byte{152, 166, 106, 194, 194, 45, 240, 30, 78, 214, 31, 119, 135, 67, 140, 238, 240, 211, 2, 218, 11, 89, 79, 12, 206, 17, 218, 65, 145, 78, 12, 33, 85, 217, 179, 173, 202, 38, 142, 23, 126, 154, 2, 100, 24, 243, 76, 75, 21, 25, 29, 95, 12, 210, 31, 158, 178, 211, 4, 42, 173, 125, 113, 60, 1, 251, 81, 65, 191, 81, 112, 232, 126, 137, 207, 92, 250, 252, 144, 53, 61, 125, 248, 143, 20, 170, 156, 107, 67, 255, 202, 206, 6, 235, 37, 131},
	TxIn: []coins_txpayload.Input{
		{
			PrevOutpoint: p2p.OutPoint{TxHash: chainhash.DoubleHashH([]byte("row-3")), Index: 1},
			Sig:          [96]byte{152, 166, 106, 194, 194, 45, 240, 30, 78, 214, 31, 119, 135, 67, 140, 238, 240, 211, 2, 218, 11, 89, 79, 12, 206, 17, 218, 65, 145, 78, 12, 33, 85, 217, 179, 173, 202, 38, 142, 23, 126, 154, 2, 100, 24, 243, 76, 75, 21, 25, 29, 95, 12, 210, 31, 158, 178, 211, 4, 42, 173, 125, 113, 60, 1, 251, 81, 65, 191, 81, 112, 232, 126, 137, 207, 92, 250, 252, 144, 53, 61, 125, 248, 143, 20, 170, 156, 107, 67, 255, 202, 206, 6, 235, 37, 131},
			PubKey:       PubKey3,
		},
	},
	TxOut: []coins_txpayload.Output{
		{
			Value:   999 * 1e8,
			Address: "olpub1xyxzekz9f42e9wacwz4zrhcsxvzv7679u5msxptyfsh4jaw8y79q6hj79n",
		},
		{
			Value:   1 * 1e7,
			Address: "olpub16lc67dz0lcgxpvnwxn3jx7yqguds9wl2gxyvl33kgfn9r4tf0z7qynqglm",
		},
	},
}

func TestSigVerifyMockTransferValid1(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockPayloadValid1.Serialize(buf)
	if err != nil {
		t.Fatal("unable to serialize payload")
	}
	err = coinTxVerifier.MatchVerify(buf.Bytes(), p2p.Transfer)
	if err != nil {
		t.Fatalf("verification failed: %v", err.Error())
	}
	err = coinTxVerifier.SigVerify(buf.Bytes(), p2p.Transfer)
	if err != nil {
		t.Fatalf("verification failed: %v", err.Error())
	}
}

// Invalid Payload 1: Want to spend more than utxo
var mockPayloadInvalid1 = coins_txpayload.PayloadTransfer{
	AggSig: [96]byte{152, 166, 106, 194, 194, 45, 240, 30, 78, 214, 31, 119, 135, 67, 140, 238, 240, 211, 2, 218, 11, 89, 79, 12, 206, 17, 218, 65, 145, 78, 12, 33, 85, 217, 179, 173, 202, 38, 142, 23, 126, 154, 2, 100, 24, 243, 76, 75, 21, 25, 29, 95, 12, 210, 31, 158, 178, 211, 4, 42, 173, 125, 113, 60, 1, 251, 81, 65, 191, 81, 112, 232, 126, 137, 207, 92, 250, 252, 144, 53, 61, 125, 248, 143, 20, 170, 156, 107, 67, 255, 202, 206, 6, 235, 37, 131},
	TxIn: []coins_txpayload.Input{
		{
			PrevOutpoint: p2p.OutPoint{TxHash: chainhash.DoubleHashH([]byte("row-3")), Index: 1},
			Sig:          [96]byte{152, 166, 106, 194, 194, 45, 240, 30, 78, 214, 31, 119, 135, 67, 140, 238, 240, 211, 2, 218, 11, 89, 79, 12, 206, 17, 218, 65, 145, 78, 12, 33, 85, 217, 179, 173, 202, 38, 142, 23, 126, 154, 2, 100, 24, 243, 76, 75, 21, 25, 29, 95, 12, 210, 31, 158, 178, 211, 4, 42, 173, 125, 113, 60, 1, 251, 81, 65, 191, 81, 112, 232, 126, 137, 207, 92, 250, 252, 144, 53, 61, 125, 248, 143, 20, 170, 156, 107, 67, 255, 202, 206, 6, 235, 37, 131},
			PubKey:       [48]byte{134, 213, 17, 19, 3, 54, 137, 215, 44, 227, 142, 140, 201, 200, 156, 183, 116, 149, 80, 234, 1, 144, 145, 34, 138, 197, 76, 164, 33, 241, 184, 252, 46, 47, 107, 41, 180, 170, 88, 7, 93, 213, 180, 154, 141, 220, 222, 65},
		},
	},
	TxOut: []coins_txpayload.Output{
		{
			Value:   999 * 1e8,
			Address: "olpub1xyxzekz9f42e9wacwz4zrhcsxvzv7679u5msxptyfsh4jaw8y79q6hj79n",
		},
		{
			Value:   999 * 1e8,
			Address: "olpub16lc67dz0lcgxpvnwxn3jx7yqguds9wl2gxyvl33kgfn9r4tf0z7qynqglm",
		},
	},
}

func TestSigVerifyMockTransferInvalid1(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockPayloadInvalid1.Serialize(buf)
	if err != nil {
		t.Fatal("unable to serialize payload")
	}
	err = coinTxVerifier.MatchVerify(buf.Bytes(), p2p.Transfer)
	equal := reflect.DeepEqual(err, ErrorSpentTooMuch)
	if !equal {
		t.Error("errors should match")
	}
}

// Invalid Payload 2: AggSig empty
var mockPayloadInvalid2 = coins_txpayload.PayloadTransfer{
	AggSig: [96]byte{},
	TxIn: []coins_txpayload.Input{
		{
			PrevOutpoint: p2p.OutPoint{TxHash: chainhash.DoubleHashH([]byte("row-3")), Index: 1},
			Sig:          [96]byte{},
			PubKey:       [48]byte{134, 213, 17, 19, 3, 54, 137, 215, 44, 227, 142, 140, 201, 200, 156, 183, 116, 149, 80, 234, 1, 144, 145, 34, 138, 197, 76, 164, 33, 241, 184, 252, 46, 47, 107, 41, 180, 170, 88, 7, 93, 213, 180, 154, 141, 220, 222, 65},
		},
	},
	TxOut: []coins_txpayload.Output{
		{
			Value:   999 * 1e8,
			Address: "olpub1xyxzekz9f42e9wacwz4zrhcsxvzv7679u5msxptyfsh4jaw8y79q6hj79n",
		},
		{
			Value:   1 * 1e7,
			Address: "olpub16lc67dz0lcgxpvnwxn3jx7yqguds9wl2gxyvl33kgfn9r4tf0z7qynqglm",
		},
	},
}

func TestSigVerifyMockTransferInvalid2(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockPayloadInvalid2.Serialize(buf)
	if err != nil {
		t.Fatal("unable to serialize payload")
	}
	err = coinTxVerifier.SigVerify(buf.Bytes(), p2p.Transfer)
	equal := reflect.DeepEqual(err, ErrorGetSig)
	if !equal {
		t.Error("errors should match")
	}
}

// Invalid Payload 3: Input pubkey empty
var mockPayloadInvalid3 = coins_txpayload.PayloadTransfer{
	AggSig: [96]byte{152, 166, 106, 194, 194, 45, 240, 30, 78, 214, 31, 119, 135, 67, 140, 238, 240, 211, 2, 218, 11, 89, 79, 12, 206, 17, 218, 65, 145, 78, 12, 33, 85, 217, 179, 173, 202, 38, 142, 23, 126, 154, 2, 100, 24, 243, 76, 75, 21, 25, 29, 95, 12, 210, 31, 158, 178, 211, 4, 42, 173, 125, 113, 60, 1, 251, 81, 65, 191, 81, 112, 232, 126, 137, 207, 92, 250, 252, 144, 53, 61, 125, 248, 143, 20, 170, 156, 107, 67, 255, 202, 206, 6, 235, 37, 131},
	TxIn: []coins_txpayload.Input{
		{
			PrevOutpoint: p2p.OutPoint{TxHash: chainhash.DoubleHashH([]byte("row-3")), Index: 1},
			Sig:          [96]byte{},
			PubKey:       [48]byte{},
		},
	},
	TxOut: []coins_txpayload.Output{
		{
			Value:   999 * 1e8,
			Address: "",
		},
		{
			Value:   1 * 1e7,
			Address: "",
		},
	},
}

func TestSigVerifyMockTransferInvalid3(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockPayloadInvalid3.Serialize(buf)
	if err != nil {
		t.Fatal("unable to serialize payload")
	}
	err = coinTxVerifier.SigVerify(buf.Bytes(), p2p.Transfer)
	equal := reflect.DeepEqual(err, ErrorGetAggPubKey)
	if !equal {
		t.Error("errors should match")
	}
}

// Invalid Payload 4: Invalid PubKey
var mockPayloadInvalid4 = coins_txpayload.PayloadTransfer{
	AggSig: [96]byte{152, 166, 106, 194, 194, 45, 240, 30, 78, 214, 31, 119, 135, 67, 140, 238, 240, 211, 2, 218, 11, 89, 79, 12, 206, 17, 218, 65, 145, 78, 12, 33, 85, 217, 179, 173, 202, 38, 142, 23, 126, 154, 2, 100, 24, 243, 76, 75, 21, 25, 29, 95, 12, 210, 31, 158, 178, 211, 4, 42, 173, 125, 113, 60, 1, 251, 81, 65, 191, 81, 112, 232, 126, 137, 207, 92, 250, 252, 144, 53, 61, 125, 248, 143, 20, 170, 156, 107, 67, 255, 202, 206, 6, 235, 37, 131},
	TxIn: []coins_txpayload.Input{
		{
			PrevOutpoint: p2p.OutPoint{TxHash: chainhash.DoubleHashH([]byte("row-3")), Index: 1},
			Sig:          [96]byte{152, 166, 106, 194, 194, 45, 240, 30, 78, 214, 31, 119, 135, 67, 140, 238, 240, 211, 2, 218, 11, 89, 79, 12, 206, 17, 218, 65, 145, 78, 12, 33, 85, 217, 179, 173, 202, 38, 142, 23, 126, 154, 2, 100, 24, 243, 76, 75, 21, 25, 29, 95, 12, 210, 31, 158, 178, 211, 4, 42, 173, 125, 113, 60, 1, 251, 81, 65, 191, 81, 112, 232, 126, 137, 207, 92, 250, 252, 144, 53, 61, 125, 248, 143, 20, 170, 156, 107, 67, 255, 202, 206, 6, 235, 37, 131},
			PubKey:       PubKey2,
		},
	},
	TxOut: []coins_txpayload.Output{
		{
			Value:   999 * 1e8,
			Address: "",
		},
		{
			Value:   1 * 1e7,
			Address: "",
		},
	},
}

func TestSigVerifyMockTransferInvalid4(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockPayloadInvalid4.Serialize(buf)
	if err != nil {
		t.Fatal("unable to serialize payload")
	}
	err = coinTxVerifier.MatchVerify(buf.Bytes(), p2p.Transfer)
	equal := reflect.DeepEqual(err, ErrorDataNoMatch)
	if !equal {
		t.Error("errors should match")
	}
}

// Invalid Payload 5: Invalid AggSig
var mockPayloadInvalid5 = coins_txpayload.PayloadTransfer{
	AggSig: [96]byte{153, 164, 8, 77, 127, 200, 185, 184, 28, 161, 146, 85, 64, 132, 109, 10, 240, 68, 184, 229, 9, 16, 218, 238, 53, 12, 210, 144, 8, 43, 65, 50, 134, 145, 229, 228, 140, 190, 27, 12, 204, 96, 171, 110, 3, 141, 136, 132, 8, 25, 156, 28, 137, 92, 28, 205, 244, 137, 142, 179, 180, 68, 76, 122, 145, 110, 65, 83, 100, 103, 161, 94, 176, 75, 146, 218, 231, 129, 224, 1, 177, 48, 138, 206, 114, 189, 174, 214, 150, 70, 144, 130, 27, 192, 209, 250},
	TxIn: []coins_txpayload.Input{
		{
			PrevOutpoint: p2p.OutPoint{TxHash: chainhash.DoubleHashH([]byte("row-3")), Index: 1},
			Sig:          [96]byte{153, 164, 8, 77, 127, 200, 185, 184, 28, 161, 146, 85, 64, 132, 109, 10, 240, 68, 184, 229, 9, 16, 218, 238, 53, 12, 210, 144, 8, 43, 65, 50, 134, 145, 229, 228, 140, 190, 27, 12, 204, 96, 171, 110, 3, 141, 136, 132, 8, 25, 156, 28, 137, 92, 28, 205, 244, 137, 142, 179, 180, 68, 76, 122, 145, 110, 65, 83, 100, 103, 161, 94, 176, 75, 146, 218, 231, 129, 224, 1, 177, 48, 138, 206, 114, 189, 174, 214, 150, 70, 144, 130, 27, 192, 209, 250},
			PubKey:       PubKey3,
		},
	},
	TxOut: []coins_txpayload.Output{
		{
			Value:   999 * 1e8,
			Address: "olpub1xyxzekz9f42e9wacwz4zrhcsxvzv7679u5msxptyfsh4jaw8y79q6hj79n",
		},
		{
			Value:   1 * 1e7,
			Address: "olpub16lc67dz0lcgxpvnwxn3jx7yqguds9wl2gxyvl33kgfn9r4tf0z7qynqglm",
		},
	},
}

func TestSigVerifyMockTransferInvalid5(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockPayloadInvalid5.Serialize(buf)
	if err != nil {
		t.Fatal("unable to serialize payload")
	}
	err = coinTxVerifier.MatchVerify(buf.Bytes(), p2p.Transfer)
	if err != nil {
		t.Fatalf("verification failed: %v", err.Error())
	}
	err = coinTxVerifier.SigVerify(buf.Bytes(), p2p.Transfer)
	equal := reflect.DeepEqual(err, ErrorInvalidSignature)
	if !equal {
		t.Error("errors should match")
	}
}

var mockPayloadBatch = []coins_txpayload.PayloadTransfer{mockPayloadValid1}

func TestSigVerifyMockBatchTransfer(t *testing.T) {
	var payload [][]byte
	for _, mockPayload := range mockPayloadBatch {
		buf := bytes.NewBuffer([]byte{})
		err := mockPayload.Serialize(buf)
		if err != nil {
			t.Fatal("unable to serialize payload")
		}
		payload = append(payload, buf.Bytes())
	}
	err := coinTxVerifier.MatchVerifyBatch(payload, p2p.Transfer)
	if err != nil {
		t.Fatalf("batch verification failed: %v", err.Error())
	}
	err = coinTxVerifier.SigVerifyBatch(payload, p2p.Transfer)
	if err != nil {
		t.Fatalf("batch verification failed: %v", err.Error())
	}
}
