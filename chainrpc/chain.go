package chainrpc

import (
	"context"
	"encoding/hex"
	"errors"
	"reflect"

	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/chainrpc/proto"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

type chainServer struct {
	chain *chain.Blockchain
	proto.UnimplementedChainServer
}

func (s *chainServer) GetChainInfo(ctx context.Context, _ *proto.Empty) (*proto.ChainInfo, error) {
	state := s.chain.State()
	return &proto.ChainInfo{
		BlockHash:   state.Tip().Hash.String(),
		BlockHeight: state.Height(),
		Validators:  uint64(len(state.TipState().ValidatorRegistry)),
	}, nil
}

func (s *chainServer) GetRawBlock(ctx context.Context, in *proto.Hash) (*proto.Block, error) {
	hash, err := chainhash.NewHashFromStr(in.Hash)
	if err != nil {
		return nil, err
	}
	block, err := s.chain.GetRawBlock(*hash)
	if err != nil {
		return nil, err
	}
	return &proto.Block{RawBlock: hex.EncodeToString(block)}, nil
}

func (s *chainServer) GetBlock(ctx context.Context, in *proto.Hash) (*proto.Block, error) {
	hash, err := chainhash.NewHashFromStr(in.Hash)
	if err != nil {
		return nil, err
	}
	block, err := s.chain.GetBlock(*hash)
	if err != nil {
		return nil, err
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
			Timestamp:                  block.Header.Timestamp.Unix(),
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

func (s *chainServer) GetBlockHash(ctx context.Context, in *proto.Height) (*proto.Hash, error) {
	blockRow, exists := s.chain.State().Chain().GetNodeByHeight(in.Height)
	if !exists {
		return nil, errors.New("block not found")
	}
	return &proto.Hash{
		Hash: blockRow.Hash.String(),
	}, nil
}

func (s *chainServer) Sync(in *proto.Hash, stream proto.Chain_SyncServer) error {
	// Define starting point
	blockRow := new(index.BlockRow)

	// If user is on tip, silently close the channel
	if reflect.DeepEqual(in.Hash, s.chain.State().Tip().Hash.String()) {
		return nil
	}

	ok := true
	hash, err := chainhash.NewHashFromStr(in.Hash)
	if err != nil {
		return errors.New("unable to decode hash from string")
	}
	currBlockRow, ok := s.chain.State().GetRowByHash(*hash)
	if !ok {
		return errors.New("block starting point doesnt exist")
	}
	blockRow, ok = s.chain.State().Chain().Next(currBlockRow)
	if !ok {
		return errors.New("there is no next blockrow")
	}
	for {
		ok := true
		rawBlock, err := s.chain.GetRawBlock(blockRow.Hash)
		if err != nil {
			return errors.New("unable get raw block")
		}
		response := &proto.RawData{
			Data: hex.EncodeToString(rawBlock),
		}
		stream.Send(response)
		blockRow, ok = s.chain.State().Chain().Next(blockRow)
		if blockRow == nil || !ok {
			break
		}
	}
	return nil
}

func (s *chainServer) SubscribeBlock(in *proto.BlockSubscribe, stream proto.Chain_SubscribeBlocksServer) error {
	return nil
}

func (s *chainServer) SubscribeAccount(in *proto.Account, stream proto.Chain_SubscribeAccountServer) error {
	return nil
}
