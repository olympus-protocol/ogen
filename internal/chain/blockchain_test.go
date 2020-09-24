package chain_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/golang/mock/gomock"
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

// validatorKeys is a slice of signatures that match the validators index
var validatorKeys1 []*bls.SecretKey
var validatorKeys2 []*bls.SecretKey

// validators are the initial validators on the realState
var validators1 []*primitives.Validator
var validators2 []*primitives.Validator
var validatorsGlobal []*primitives.Validator

// genesisHash is just a random hash to set as genesis hash.
var genesisHash chainhash.Hash

// params are the params used on the test
var param = &testdata.TestParams

// init params used on the test
var stateParams state.InitializationParameters

//bc hold the initialized blockchain
var bc chain.Blockchain

func init() {
	f := fuzz.New().NilChance(0)
	f.Fuzz(&genesisHash)
	priv := testdata.PremineAddr

	addrByte, _ := priv.PublicKey().Hash()
	addr := testdata.PremineAddr.PublicKey().ToAccount()

	for i := 0; i < 100; i++ {
		if i < 50 {
			key := bls.RandKey()
			validatorKeys1 = append(validatorKeys1, bls.RandKey())
			val := &primitives.Validator{
				Balance:          100 * 1e8,
				PayeeAddress:     addrByte,
				Status:           primitives.StatusActive,
				FirstActiveEpoch: 0,
				LastActiveEpoch:  0,
			}
			copy(val.PubKey[:], key.PublicKey().Marshal())
			validators1 = append(validators1, val)
		} else {
			key := bls.RandKey()
			validatorKeys2 = append(validatorKeys2, bls.RandKey())
			val := &primitives.Validator{
				Balance:          100 * 1e8,
				PayeeAddress:     addrByte,
				Status:           primitives.StatusActive,
				FirstActiveEpoch: 0,
				LastActiveEpoch:  0,
			}
			copy(val.PubKey[:], key.PublicKey().Marshal())
			validators2 = append(validators2, val)
		}

	}
	validatorsGlobal = append(validators1, validators2...)
	stateParams.GenesisTime = time.Unix(time.Now().Unix(), 0)
	stateParams.InitialValidators = []state.ValidatorInitialization{}
	// Convert the validators to initialization params.
	for _, vk := range validatorKeys1 {
		val := state.ValidatorInitialization{
			PubKey:       hex.EncodeToString(vk.PublicKey().Marshal()),
			PayeeAddress: addr,
		}
		stateParams.InitialValidators = append(stateParams.InitialValidators, val)
	}
	stateParams.PremineAddress = addr
}

// create a blockchain instance and test its methods
func TestBlockchain_Instance(t *testing.T) {
	//f := fuzz.New().NilChance(0)
	ctrl := gomock.NewController(t)
	log := logger.New(os.Stdin)

	db := blockdb.NewMockDatabase(ctrl)
	var c chain.Config
	c.Log = log
	c.Datadir = testdata.Conf.DataFolder
	var err error
	bc, err = chain.NewBlockchain(c, param, db, stateParams)
	assert.NoError(t, err)
	assert.NotNil(t, bc)
	err = bc.Start()
	assert.NoError(t, err)

	genTime := bc.GenesisTime()
	assert.NotNil(t, genTime)

	//block-related methods
	genblock := primitives.GetGenesisBlock()
	genesisHash = genblock.Hash()

	//get signature of genesis hash
	currState, _ := bc.State().GetStateForHash(genesisHash)

	b := primitives.Block{
		Header: &primitives.BlockHeader{
			Version:                    0,
			Nonce:                      0,
			TxMerkleRoot:               chainhash.Hash{},
			VoteMerkleRoot:             chainhash.Hash{},
			DepositMerkleRoot:          chainhash.Hash{},
			ExitMerkleRoot:             chainhash.Hash{},
			VoteSlashingMerkleRoot:     chainhash.Hash{},
			RANDAOSlashingMerkleRoot:   chainhash.Hash{},
			ProposerSlashingMerkleRoot: chainhash.Hash{},
			GovernanceVotesMerkleRoot:  chainhash.Hash{},
			PrevBlockHash:              genesisHash,
			Timestamp:                  uint64(time.Now().Unix()),
			Slot:                       1,
			StateRoot:                  chainhash.Hash{},
			FeeAddress:                 [20]byte{},
		},
		Txs: []*primitives.Tx{},
	}
	// sign the block with the next validator
	valPub, err := currState.GetProposerPublicKey(&b, param)
	assert.NoError(t, err)
	var priv *bls.SecretKey
	for _, element := range validatorKeys1 {
		if bytes.Equal(element.PublicKey().Marshal(), valPub.Marshal()) {
			priv = element
		}
	}
	assert.NotNil(t, priv)
	randaoHash := chainhash.HashH([]byte(fmt.Sprintf("%d", 1)))
	randaoSig := priv.Sign(randaoHash[:])

	bH := b.Hash()
	blockSig := priv.Sign(bH[:])
	var ds [96]byte
	var rs [96]byte
	copy(ds[:], blockSig.Marshal())
	copy(rs[:], randaoSig.Marshal())
	b.Signature = ds
	b.RandaoSignature = rs

	// ProcessBlock
	err = bc.ProcessBlock(&b)
	assert.NoError(t, err)

	_ = os.Remove("./tx.db")

}
