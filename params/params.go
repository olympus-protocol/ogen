package params

import (
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/amount"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/hdwallets"
	"time"
)

type Prefixes struct {
	PubKey          string
	PrivKey         string
	ContractPubKey  string
	ContractPrivKey string
}

type ChainParams struct {
	Name                       string
	DefaultP2PPort             string
	GenesisBlock               p2p.MsgBlock
	GenesisHash                chainhash.Hash
	HDPrefixes                 hdwallets.NetPrefix
	HDCoinIndex                uint32
	AddressPrefixes            Prefixes
	LastPreWorkersBlock        uint32
	PreWorkersPubKeyHash       string
	BlockTimeSpan              int64
	BlocksReductionCycle       uint32
	SuperBlockCycle            uint32
	SuperBlockStartHeight      uint32
	GovernanceBudgetPercentage float64
	ProfitSharingCycle         uint32
	ProfitSharingStartCycle    uint32
	GovernanceProposalFee      amount.AmountType
	BaseBlockReward            float64
	BlockReductionPercentage   float64
}

var NetworkNames = map[string]string{
	"mainnet": "Main Network",
	"test":    "Test Network",
}

var Mainnet = ChainParams{
	Name:           "polis",
	DefaultP2PPort: "24126",
	GenesisBlock:   mainNetGenesisBlock,
	GenesisHash:    mainNetGenesisHash,
	HDPrefixes: hdwallets.NetPrefix{
		ExtPub:  []byte{0x1f, 0x74, 0x90, 0xf0},
		ExtPriv: []byte{0x11, 0x24, 0xd9, 0x70},
	},
	HDCoinIndex: 1997,
	AddressPrefixes: Prefixes{
		PubKey:          "olpub",
		PrivKey:         "olprv",
		ContractPubKey:  "ctpub",
		ContractPrivKey: "ctprv",
	},
	LastPreWorkersBlock:        500,
	PreWorkersPubKeyHash:       "olpub12vjdayxm6eygqkxrtyvt0jnjxn8965wflynmf4d899pnkzp9glmslqcvce",
	BlockTimeSpan:              120,                   // 120 seconds
	BlocksReductionCycle:       262800,                // 1 year
	SuperBlockCycle:            21600,                 // 1 month
	SuperBlockStartHeight:      0,                     // TODO define
	ProfitSharingCycle:         21600,                 // 1 month
	ProfitSharingStartCycle:    0,                     // TODO define
	GovernanceBudgetPercentage: 0.2,                   // 20%
	BlockReductionPercentage:   0.2,                   // 20%
	BaseBlockReward:            20,                    // 20 POLIS
	GovernanceProposalFee:      amount.AmountType(50), // 50 POLIS
}

var mainNetGenesisCoinBaseTx = p2p.MsgTx{
	TxVersion: 1,
	TxType:    p2p.Coins,
	TxAction:  p2p.Transfer,
}

var mainNetGenesisMerkleRoot = chainhash.Hash([chainhash.HashSize]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
})

var mainNetGenesisHash = chainhash.Hash([chainhash.HashSize]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
})

var mainNetGenesisBlock = p2p.MsgBlock{
	Header: p2p.BlockHeader{
		Version:       1,
		PrevBlockHash: chainhash.Hash{},
		MerkleRoot:    mainNetGenesisMerkleRoot,
		Timestamp:     time.Unix(0x0, 0),
	},
	Txs: []*p2p.MsgTx{&mainNetGenesisCoinBaseTx},
}

var TestNet = ChainParams{
	Name:           "test",
	DefaultP2PPort: "24126",
	GenesisBlock:   testNetGenesisBlock,
	GenesisHash:    testNetGenesisHash,
	HDPrefixes: hdwallets.NetPrefix{
		ExtPub:  nil,
		ExtPriv: nil,
	},
	AddressPrefixes: Prefixes{
		PubKey:          "tolpub",
		PrivKey:         "tolprv",
		ContractPrivKey: "tctpub",
		ContractPubKey:  "tctprv",
	},
	LastPreWorkersBlock:        10,
	PreWorkersPubKeyHash:       "1HWfiw9Lbg2vh8A1sZDsp5BVLHeW41V13R", // 5JbK2h1P7BQTmwJgCPRonJzCqRMNpFPvsAPTwrHBdT7DmEzzsUK
	BlockTimeSpan:              60,                                   // 60 seconds
	BlocksReductionCycle:       259200,                               // 6 months
	SuperBlockCycle:            1440,                                 // 1 day
	GovernanceBudgetPercentage: 0.2,                                  // 20%
	BlockReductionPercentage:   0.2,                                  // 20%
	BaseBlockReward:            20,                                   // 20
}

var testNetGenesisCoinBaseTx = p2p.MsgTx{
	TxVersion: 1,
	TxType:    p2p.Coins,
	TxAction:  p2p.Transfer,
}

var testNetGenesisMerkleRoot = chainhash.Hash([chainhash.HashSize]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
})

var testNetGenesisHash = chainhash.Hash([chainhash.HashSize]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
})

var testNetGenesisBlock = p2p.MsgBlock{
	Header: p2p.BlockHeader{
		Version:       1,
		PrevBlockHash: chainhash.Hash{},
		MerkleRoot:    testNetGenesisMerkleRoot,
		Timestamp:     time.Unix(0x0, 0),
	},
	Txs: []*p2p.MsgTx{&testNetGenesisCoinBaseTx},
}
