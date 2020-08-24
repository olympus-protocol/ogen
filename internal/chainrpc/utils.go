package chainrpc

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/pkg/p2p"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/internal/proposer"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type utilsServer struct {
	proposer     proposer.Proposer
	hostnode     hostnode.HostNode
	txTopic      *pubsub.Topic
	depositTopic *pubsub.Topic
	exitTopic    *pubsub.Topic
	proto.UnimplementedUtilsServer
}

func (s *utilsServer) StartProposer(ctx context.Context, in *proto.Empty) (*proto.Success, error) {
	err := s.proposer.Start()
	if err != nil {
		return &proto.Success{Success: false, Error: err.Error()}, nil
	}
	return &proto.Success{Success: true}, nil
}
func (s *utilsServer) StopProposer(ctx context.Context, _ *proto.Empty) (*proto.Success, error) {
	s.proposer.Stop()
	err := s.proposer.Keystore().Close()
	if err != nil {
		return &proto.Success{Success: false, Error: err.Error()}, nil
	}
	return &proto.Success{Success: true}, nil
}

func (s *utilsServer) GenValidatorKey(ctx context.Context, in *proto.GenValidatorKeys) (*proto.KeyPairs, error) {
	key, err := s.proposer.Keystore().GenerateNewValidatorKey(in.Keys)
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
	dataBytes, err := hex.DecodeString(data.Data)
	if err != nil {
		return nil, err
	}
	switch data.Type {
	case "tx":

		tx := new(primitives.Tx)

		err := tx.Unmarshal(dataBytes)
		if err != nil {
			return nil, errors.New("unable to decode raw data")
		}

		msg := &p2p.MsgTx{Data: tx}

		buf := bytes.NewBuffer([]byte{})
		err = p2p.WriteMessage(buf, msg, s.hostnode.GetNetMagic())
		if err != nil {
			return nil, err
		}

		err = s.txTopic.Publish(ctx, buf.Bytes())
		if err != nil {
			return nil, err
		}

		return &proto.Success{Success: true, Data: tx.Hash().String()}, nil

	case "deposit":

		deposit := new(primitives.Deposit)

		err := deposit.Unmarshal(dataBytes)
		if err != nil {
			return nil, errors.New("unable to decode raw data")
		}

		msg := &p2p.MsgDeposit{Data: deposit}

		buf := bytes.NewBuffer([]byte{})
		err = p2p.WriteMessage(buf, msg, s.hostnode.GetNetMagic())
		if err != nil {
			return nil, err
		}

		err = s.depositTopic.Publish(ctx, buf.Bytes())
		if err != nil {
			return nil, err
		}

		return &proto.Success{Success: true, Data: deposit.Hash().String()}, nil

	case "exit":

		exit := new(primitives.Exit)

		err := exit.Unmarshal(dataBytes)
		if err != nil {
			return nil, errors.New("unable to decode raw data")
		}

		msg := &p2p.MsgExit{Data: exit}

		buf := bytes.NewBuffer([]byte{})
		err = p2p.WriteMessage(buf, msg, s.hostnode.GetNetMagic())
		if err != nil {
			return nil, err
		}

		err = s.exitTopic.Publish(ctx, buf.Bytes())
		if err != nil {
			return nil, err
		}

		return &proto.Success{Success: true, Data: exit.Hash().String()}, nil

	default:
		return &proto.Success{Success: false, Error: "unknown raw data type"}, nil
	}
}

func (s *utilsServer) DecodeRawTransaction(ctx context.Context, data *proto.RawData) (*proto.Tx, error) {
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
