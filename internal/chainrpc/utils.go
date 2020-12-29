package chainrpc

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/burnproof"
	"github.com/olympus-protocol/ogen/pkg/p2p"

	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type utilsServer struct {
	keystore       keystore.Keystore
	host           hostnode.HostNode
	coinsMempool   mempool.CoinsMempool
	actionsMempool mempool.ActionMempool
	chain          chain.Blockchain
	proto.UnimplementedUtilsServer
}

func (s *utilsServer) GenKeyPair(ctx context.Context, _ *proto.Empty) (*proto.KeyPair, error) {
	defer ctx.Done()

	k, err := bls.RandKey()
	if err != nil {
		return nil, err
	}

	return &proto.KeyPair{Private: k.ToWIF(), Public: k.PublicKey().ToAccount()}, nil
}

func (s *utilsServer) GenValidatorKey(ctx context.Context, in *proto.GenValidatorKeys) (*proto.KeyPairs, error) {
	defer ctx.Done()

	key, err := s.keystore.GenerateNewValidatorKey(in.Keys)
	if err != nil {
		return nil, err
	}
	keys := make([]string, in.Keys)
	for i := range keys {
		keys[i] = hex.EncodeToString(key[i].Marshal())
	}
	return &proto.KeyPairs{Keys: keys}, nil
}

func (s *utilsServer) SubmitRawData(ctx context.Context, data *proto.RawData) (*proto.Success, error) {
	defer ctx.Done()
	dataBytes, err := hex.DecodeString(data.Data)
	if err != nil {
		return nil, err
	}
	switch data.Type {
	case "tx":
		tx := new(primitives.Tx)

		err := tx.Unmarshal(dataBytes)
		if err != nil {
			return &proto.Success{Success: false, Error: "unable to decode raw data"}, nil
		}

		msg := &p2p.MsgTx{Data: tx}

		err = s.coinsMempool.Add(tx)

		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		err = s.host.Broadcast(msg)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		return &proto.Success{Success: true, Data: tx.Hash().String()}, nil

	case "deposit":

		deposit := new(primitives.Deposit)

		err := deposit.Unmarshal(dataBytes)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		err = s.actionsMempool.AddDeposit(deposit)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		msg := &p2p.MsgDeposit{Data: deposit}

		err = s.host.Broadcast(msg)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		return &proto.Success{Success: true}, nil

	case "exit":

		exit := new(primitives.Exit)

		err := exit.Unmarshal(dataBytes)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		err = s.actionsMempool.AddExit(exit)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		msg := &p2p.MsgExit{Data: exit}

		err = s.host.Broadcast(msg)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		return &proto.Success{Success: true}, nil

	case "deposits_bulk":
		deposits := new(p2p.MsgDeposits)

		err := deposits.Unmarshal(dataBytes)
		if err != nil {
			return nil, errors.New("unable to decode raw data")
		}

		for _, d := range deposits.Data {
			err = s.actionsMempool.AddDeposit(d)
			if err != nil {
				return &proto.Success{Success: false, Error: err.Error()}, nil
			}
		}

		err = s.host.Broadcast(deposits)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		return &proto.Success{Success: true}, nil

	case "exits_bulk":
		exits := new(p2p.MsgExits)

		err := exits.Unmarshal(dataBytes)
		if err != nil {
			return nil, errors.New("unable to decode raw data")
		}

		for _, d := range exits.Data {
			err = s.actionsMempool.AddExit(d)
			if err != nil {
				return &proto.Success{Success: false, Error: err.Error()}, nil
			}
		}

		err = s.host.Broadcast(exits)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		return &proto.Success{Success: true}, nil

	default:
		return &proto.Success{Success: false, Error: "unknown raw data type"}, nil
	}
}

func (s *utilsServer) DecodeRawTransaction(ctx context.Context, data *proto.RawData) (*proto.Tx, error) {
	defer ctx.Done()

	dataBytes, err := hex.DecodeString(data.Data)
	if err != nil {
		return nil, err
	}
	tx := new(primitives.Tx)
	err = tx.Unmarshal(dataBytes)
	if err != nil {
		return nil, errors.New("unable to decode block raw data")
	}
	txParse := &proto.Tx{
		Hash:          tx.Hash().String(),
		To:            hex.EncodeToString(tx.To[:]),
		FromPublicKey: hex.EncodeToString(tx.FromPublicKey[:]),
		Amount:        tx.Amount,
		Nonce:         tx.Nonce,
		Fee:           tx.Fee,
		Signature:     hex.EncodeToString(tx.Signature[:]),
	}
	return txParse, nil
}

func (s *utilsServer) DecodeRawBlock(ctx context.Context, data *proto.RawData) (*proto.Block, error) {
	defer ctx.Done()

	dataBytes, err := hex.DecodeString(data.Data)
	if err != nil {
		return nil, err
	}
	block := new(primitives.Block)
	err = block.Unmarshal(dataBytes)
	if err != nil {
		return nil, errors.New("unable to decode block raw data")
	}
	blockParse := &proto.Block{
		Hash: block.Hash().String(),
		Header: &proto.BlockHeader{
			Version:                    block.Header.Version,
			Nonce:                      block.Header.Nonce,
			TxMerkleRoot:               hex.EncodeToString(block.Header.TxMerkleRoot[:]),
			VoteMerkleRoot:             hex.EncodeToString(block.Header.VoteMerkleRoot[:]),
			DepositMerkleRoot:          hex.EncodeToString(block.Header.DepositMerkleRoot[:]),
			ExitMerkleRoot:             hex.EncodeToString(block.Header.ExitMerkleRoot[:]),
			VoteSlashingMerkleRoot:     hex.EncodeToString(block.Header.VoteSlashingMerkleRoot[:]),
			RandaoSlashingMerkleRoot:   hex.EncodeToString(block.Header.RANDAOSlashingMerkleRoot[:]),
			ProposerSlashingMerkleRoot: hex.EncodeToString(block.Header.ProposerSlashingMerkleRoot[:]),
			PrevBlockHash:              hex.EncodeToString(block.Header.PrevBlockHash[:]),
			Timestamp:                  block.Header.Timestamp,
			Slot:                       block.Header.Slot,
			StateRoot:                  hex.EncodeToString(block.Header.StateRoot[:]),
			FeeAddress:                 hex.EncodeToString(block.Header.FeeAddress[:]),
		},
		Txs:             block.GetTxs(),
		Signature:       hex.EncodeToString(block.Signature[:]),
		RandaoSignature: hex.EncodeToString(block.RandaoSignature[:]),
	}
	return blockParse, nil
}

func (s *utilsServer) SyncMempool(_ *proto.Empty, stream proto.Utils_SyncMempoolServer) error {
	_, cancel := context.WithCancel(stream.Context())
	defer cancel()
	txs := s.coinsMempool.GetWithoutApply()
	for _, tx := range txs {
		protoTx := &proto.Tx{
			Hash:          tx.Hash().String(),
			To:            hex.EncodeToString(tx.To[:]),
			FromPublicKey: hex.EncodeToString(tx.FromPublicKey[:]),
			Amount:        tx.Amount,
			Nonce:         tx.Nonce,
			Fee:           tx.Fee,
			Signature:     hex.EncodeToString(tx.Signature[:]),
		}
		err := stream.Send(protoTx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *utilsServer) SubscribeMempool(_ *proto.Empty, stream proto.Utils_SubscribeMempoolServer) error {
	txn := newCoinsNotifee()
	s.coinsMempool.Notify(txn)
	for {
		select {
		case tx := <-txn.tx:
			protoTx := &proto.Tx{
				Hash:          tx.Hash().String(),
				To:            hex.EncodeToString(tx.To[:]),
				FromPublicKey: hex.EncodeToString(tx.FromPublicKey[:]),
				Amount:        tx.Amount,
				Nonce:         tx.Nonce,
				Fee:           tx.Fee,
				Signature:     hex.EncodeToString(tx.Signature[:]),
			}
			err := stream.Send(protoTx)
			if err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}

}

type coinNotifee struct {
	tx chan *primitives.Tx
}

func (c *coinNotifee) NotifyTx(tx *primitives.Tx) {
	c.tx <- tx
}

func newCoinsNotifee() *coinNotifee {
	bn := &coinNotifee{
		tx: make(chan *primitives.Tx),
	}
	return bn
}

func (s *utilsServer) SubmitRedeemProof(ctx context.Context, data *proto.RedeemProof) (*proto.Success, error) {
	defer ctx.Done()
	proofBytes, err := hex.DecodeString(data.Proof)
	if err != nil {
		return nil, err
	}

	addrBytes := []byte(data.Address)

	if len(addrBytes) != 44 {
		return &proto.Success{Error: errors.New("invalid address size").Error()}, nil
	}

	var address [44]byte
	copy(address[:], addrBytes)

	proofs := make([]*burnproof.CoinsProof, 0)
	buf := bytes.NewBuffer(proofBytes)
	for {
		proof := new(burnproof.CoinsProof)
		err = proof.Unmarshal(buf)
		if err != nil {
			return &proto.Success{Error: err.Error()}, nil
		}
		if buf.Len() < 0 {
			break
		}
		proofs = append(proofs, proof)
	}

	if len(proofs) > 2048 {
		return &proto.Success{Error: "too many proofs submited, max number is 2048"}, nil
	}

	serializableProofs := make([]*burnproof.CoinsProofSerializable, len(proofs))

	for i, p := range proofs {
		pser, err := p.ToSerializable(address)
		if err != nil {
			return &proto.Success{Error: err.Error()}, nil
		}

		serializableProofs[i] = pser

		// Add to a mempool and broadcast
		err = s.actionsMempool.AddProof(pser)
		if err != nil {
			return &proto.Success{Error: err.Error()}, nil
		}

	}

	msg := &p2p.MsgProofs{Proofs: serializableProofs}

	err = s.host.Broadcast(msg)
	if err != nil {
		return &proto.Success{Success: false, Error: err.Error()}, nil
	}

	return &proto.Success{Success: true}, nil
}
