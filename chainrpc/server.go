package chainrpc

import (
	"context"
	"encoding/hex"
	"errors"
	"net"

	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chainrpc/proto"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/logger"
	"google.golang.org/grpc"
)

type Config struct {
	Network string
	Address string
	Log     *logger.Logger
}

type chainServer struct {
	chain *chain.Blockchain
	proto.UnimplementedChainServer
}

type RPCServer struct {
	log    *logger.Logger
	config Config
	rpc    *grpc.Server

	chainServer *chainServer
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

func (s *RPCServer) registerServices() {
	proto.RegisterChainServer(s.rpc, s.chainServer)
}

func (s *RPCServer) Stop() {
	s.log.Info("stoping gRPC Server")
	s.rpc.GracefulStop()
}

func (s *RPCServer) Start() error {
	s.registerServices()
	s.log.Info("Starting gRPC Server")
	lis, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return err
	}
	err = s.rpc.Serve(lis)
	if err != nil {
		return err
	}
	return nil
}

// NewRPCServer Returns an RPC server instance
func NewRPCServer(config Config, chain *chain.Blockchain) *RPCServer {
	return &RPCServer{
		rpc:    grpc.NewServer(),
		config: config,
		log:    config.Log,
		chainServer: &chainServer{
			chain: chain,
		},
	}
}
