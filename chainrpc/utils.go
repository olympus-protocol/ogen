package chainrpc

import (
	"context"

	"github.com/olympus-protocol/ogen/proto"
)

type utilsServer struct {
	proto.UnimplementedUtilsServer
}

func (s *utilsServer) GenValidatorKey(context.Context, *proto.Empty) (*proto.KeyPair, error) {
	return nil, nil
}
func (s *utilsServer) SendRawTransaction(context.Context, *proto.RawData) (*proto.Success, error) {
	return nil, nil
}
func (s *utilsServer) DecodeRawTransaction(context.Context, *proto.RawData) (*proto.Tx, error) {
	return nil, nil
}
func (s *utilsServer) DecodeRawBlock(context.Context, *proto.RawData) (*proto.Block, error) {
	return nil, nil
}
