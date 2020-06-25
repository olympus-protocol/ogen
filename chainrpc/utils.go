package chainrpc

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/olympus-protocol/ogen/keystore"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/proto"
)

type utilsServer struct {
	keystore     *keystore.Keystore
	txTopic      *pubsub.Topic
	depositTopic *pubsub.Topic
	exitTopic    *pubsub.Topic
	proto.UnimplementedUtilsServer
}

func (s *utilsServer) GenValidatorKey(ctx context.Context, in *proto.Number) (*proto.KeyPairs, error) {
	key, err := s.keystore.GenerateNewValidatorKey(in.Number)
	if err != nil {
		return nil, err
	}
	keys := make([]string, in.Number)
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
	buf := bytes.NewBuffer(dataBytes)
	switch data.Type {
	case "tx":
		tx := new(primitives.Tx)
		err := tx.Unmarshal(dataBytes)
		if err != nil {
			return nil, errors.New("unable to decode raw data")
		}
		err = s.txTopic.Publish(ctx, dataBytes)
		if err != nil {
			return nil, err
		}
		return &proto.Success{Success: true, Data: tx.Hash().String()}, nil
	case "deposit":
		deposit := new(primitives.Deposit)
		err := deposit.Decode(buf)
		if err != nil {
			return nil, errors.New("unable to decode raw data")
		}
		err = s.depositTopic.Publish(ctx, dataBytes)
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
		err = s.exitTopic.Publish(ctx, dataBytes)
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
		Hash:    tx.Hash().String(),
		Version: tx.Version,
		Type:    tx.Type,
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
			TxMerkleRoot:               block.Header.TxMerkleRoot.String(),
			VoteMerkleRoot:             block.Header.VoteMerkleRoot.String(),
			DepositMerkleRoot:          block.Header.DepositMerkleRoot.String(),
			ExitMerkleRoot:             block.Header.ExitMerkleRoot.String(),
			VoteSlashingMerkleRoot:     block.Header.VoteSlashingMerkleRoot.String(),
			RandaoSlashingMerkleRoot:   block.Header.RANDAOSlashingMerkleRoot.String(),
			ProposerSlashingMerkleRoot: block.Header.ProposerSlashingMerkleRoot.String(),
			PrevBlockHash:              block.Header.PrevBlockHash.String(),
			Timestamp:                  block.Header.Timestamp,
			Slot:                       block.Header.Slot,
			StateRoot:                  block.Header.StateRoot.String(),
			FeeAddress:                 hex.EncodeToString(block.Header.FeeAddress[:]),
		},
		Txs:             block.GetTxs(),
		Signature:       hex.EncodeToString(block.Signature),
		RandaoSignature: hex.EncodeToString(block.RandaoSignature),
	}
	return blockParse, nil
}
