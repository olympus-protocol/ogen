package chainrpc

import (
	"context"

	"github.com/olympus-protocol/ogen/chainrpc/proto"
)

type utilsServer struct {
	proto.UnimplementedUtilsServer
}

func (s *validatorsServer) GenValidatorKey(context.Context, *proto.Empty) (*proto.KeyPair, error) {
	return nil, nil
}

func (s *validatorsServer) GenKeyPair(context.Context, *proto.Empty) (*proto.KeyPair, error) {
	return nil, nil
}

func (s *validatorsServer) GenRawKeyPair(context.Context, *proto.Empty) (*proto.KeyPair, error) {
	return nil, nil
}

func (s *validatorsServer) SendRawTransaction(context.Context, *proto.RawData) (*proto.Success, error) {
	return nil, nil
}

func (s *validatorsServer) DecodeRawTransaction(context.Context, *proto.RawData) (*proto.Tx, error) {
	return nil, nil
}

func (s *validatorsServer) DecodeRawBlock(context.Context, *proto.RawData) (*proto.Block, error) {
	return nil, nil
}
