package chain_test

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/state"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"testing"
)

type testChain struct {
	prevHash       chainhash.Hash
	workerQueue    []chainhash.Hash
	workerRegistry state.WorkerRegistry
}

func newTestChain(workers []state.Worker) testChain {
	workerQueue := make([]chainhash.Hash, len(workers))
	registry := state.NewWorkerRegistry()

	for i := range workers {
		h := workers[i].ID()
		workerQueue[i] = *h
		registry.Add(workers[i])
	}
	return testChain{
		prevHash:       chainhash.Hash{},
		workerQueue:    workerQueue,
		workerRegistry: *registry,
	}
}

func (t *testChain) Valid(block primitives.Block) error {
	if !block.Header.PrevBlockHash.IsEqual(&t.prevHash) {
		return errors.New("invalid block hash")
	}
	return nil
}

func (t *testChain) GetValue() (*primitives.Block, error) {
	// block signatures aren't checked yet, so...
	return &primitives.Block{
		Header: primitives.BlockHeader{
			PrevBlockHash: t.prevHash,
		},
		Txs:       nil,
		PubKey:    [48]byte{},
		Signature: [96]byte{},
	}, nil
}

func (t *testChain) GetProposer(round uint64) chainhash.Hash {
	return t.workerQueue[round%uint64(len(t.workerQueue))]
}

func (t *testChain) GetWorkerData(h chainhash.Hash) (state.Worker, bool) {
	return t.workerRegistry.Get(h)
}

func (t *testChain) NumWorkers() uint64 {
	return uint64(len(t.workerQueue))
}

func (t *testChain) Decide(h chainhash.Hash) {
	fmt.Println("decided on ", h)
	t.prevHash = h
}

type p2pLayer struct {
	workers []*testWorker
}

func (p *p2pLayer) broadcast(msg p2p.Message) {
	fmt.Println(msg)
	for _, w := range p.workers {
		w.handleMessage(msg)
	}
}

type testWorker struct {
	p2p      *p2pLayer
	worker   state.Worker
	messages chan p2p.Message

	key bls.SecretKey
}

func (t *testWorker) Sign(hash chainhash.Hash) (bls.Signature, error) {
	sig, err := bls.Sign(&t.key, hash[:])
	if err != nil {
		return bls.Signature{}, err
	}
	return *sig, nil
}

func (t *testWorker) ValidatorID() chainhash.Hash {
	h := t.worker.ID()
	return *h
}

func (t *testWorker) Broadcast(message p2p.Message) {
	t.p2p.broadcast(message)
}

func (t *testWorker) handleMessage(msg p2p.Message) {
	t.messages <- msg
}

func (t *testWorker) handleMessages(c *chain.Consensus) {
	for {
		select {
		case msg := <- t.messages:
			switch message := msg.(type) {
			case *p2p.MsgProposal:
				if err := c.OnMessageProposal(*message); err != nil {
					panic(err)
				}
				break
			case *p2p.MsgPrevote:
				if err := c.OnMessagePrevote(*message); err != nil {
					panic(err)
				}
				break
			case *p2p.MsgPrecommit:
				if err := c.OnMessagePrecommit(*message); err != nil {
					panic(err)
				}
				break
			}
		default:
			return
		}
	}
}

const numWorkers = 5

func TestConsensusRound(t *testing.T) {
	workers := make([]state.Worker, numWorkers)
	testWorkers := make([]*testWorker, numWorkers)
	keys := make([]bls.SecretKey, numWorkers)

	p := p2pLayer{
		workers: testWorkers,
	}

	for i := range workers {
		privKey, err := bls.RandSecretKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		keys[i] = *privKey
		workers[i] = state.Worker{
			Outpoint: p2p.OutPoint{
				TxHash: chainhash.HashH([]byte(fmt.Sprintf("worker %d", i))),
				Index:  0,
			},
			PubKey:       privKey.DerivePublicKey().Serialize(),
			PayeeAddress: "",
		}
	}

	testChain := newTestChain(workers)

	consensuses := make([]*chain.Consensus, len(workers))

	for i := range testWorkers {
		testWorkers[i] = &testWorker{
			p2p:      &p,
			worker:   workers[i],
			key:      keys[i],
			messages: make(chan p2p.Message, 100),
		}

		c, err := chain.NewConsensus(&testChain, testWorkers[i], testWorkers[i])
		if err != nil {
			t.Fatal(err)
		}

		consensuses[i] = c
	}

	for _, c := range consensuses {
		err := c.Initialize()
		if err != nil {
			t.Fatal(err)
		}
	}

	for j := 0; j < 10; j++ {
		for i := range consensuses {
			testWorkers[i].handleMessages(consensuses[i])
		}
	}
}
