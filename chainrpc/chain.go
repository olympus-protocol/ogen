package chainrpc

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/chainrpc/proto"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

type chainServer struct {
	chain *chain.Blockchain
	proto.UnimplementedChainServer
}

func (s *chainServer) GetChainInfo(ctx context.Context, _ *proto.Empty) (*proto.ChainInfoResponse, error) {
	state := s.chain.State()
	return &proto.ChainInfoResponse{
		BlockHash:   state.Tip().Hash.String(),
		BlockHeight: state.Height(),
		Validators:  uint64(len(state.TipState().ValidatorRegistry)),
	}, nil
}

func (s *chainServer) GetRawBlock(ctx context.Context, in *proto.GetBlockInfo) (*proto.GetBlockRawResponse, error) {
	hash, err := chainhash.NewHashFromStr(in.BlockHash)
	if err != nil {
		return nil, err
	}
	block, err := s.chain.GetRawBlock(*hash)
	if err != nil {
		return nil, err
	}
	return &proto.GetBlockRawResponse{RawBlock: hex.EncodeToString(block)}, nil
}

func (s *chainServer) GetBlock(ctx context.Context, in *proto.GetBlockInfo) (*proto.GetBlockResponse, error) {
	hash, err := chainhash.NewHashFromStr(in.BlockHash)
	if err != nil {
		return nil, err
	}
	block, err := s.chain.GetBlock(*hash)
	if err != nil {
		return nil, err
	}
	blockParse := &proto.GetBlockResponse{
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

func (s *chainServer) GetBlockHash(ctx context.Context, in *proto.GetBlockHashInfo) (*proto.GetBlockHashResponse, error) {
	blockRow, exists := s.chain.State().Chain().GetNodeByHeight(in.BlockHeigth)
	if !exists {
		return nil, errors.New("block not found")
	}
	return &proto.GetBlockHashResponse{
		BlockHash: blockRow.Hash.String(),
	}, nil
}

func (s *chainServer) Sync(in *proto.SyncInfo, stream proto.Chain_SyncServer) error {
	blockRow := new(index.BlockRow)
	if in.FullSync {
		var exist bool
		blockRow, exist = s.chain.State().Chain().GetNodeByHeight(0)
		if !exist {
			return errors.New("unable to get genesis block row")
		}
	} else {
		var exist bool
		hash, err := chainhash.NewHashFromStr(in.FromBlockHash)
		if err != nil {
			return errors.New("unable to decode block from hash")
		}
		blockRow, exist = s.chain.State().GetRowByHash(*hash)
		if !exist {
			return errors.New("unable to get genesis block row")
		}
	}
	if !in.FullSync {
		firstRawBlock, err := s.chain.GetRawBlock(blockRow.Hash)
		if err != nil {
			return errors.New("unable get first raw block")
		}
		firsBlock := &proto.SyncStreamResponse{
			RawBlock: hex.EncodeToString(firstRawBlock),
		}
		stream.Send(firsBlock)
	}
	for {
		blockRow = blockRow.Parent
		if blockRow == nil {
			break
		}
		rawBlock, err := s.chain.GetRawBlock(blockRow.Hash)
		if err != nil {
			return errors.New("unable get first raw block")
		}
		response := &proto.SyncStreamResponse{
			RawBlock: hex.EncodeToString(rawBlock),
		}
		stream.Send(response)
	}
	return nil
}

func (s *chainServer) Subscribe(in *proto.SubscribeInfo, stream proto.Chain_SubscribeServer) error {
	return nil
}
