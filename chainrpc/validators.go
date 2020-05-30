package chainrpc

import (
	"context"
	"encoding/hex"

	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chainrpc/proto"
	"github.com/olympus-protocol/ogen/wallet"
	"github.com/shopspring/decimal"
)

type validatorsServer struct {
	wallet *wallet.Wallet
	chain  *chain.Blockchain
	proto.UnimplementedValidatorsServer
}

func (s *validatorsServer) GetValidatorsList(context.Context, *proto.Empty) (*proto.GetValidatorsListResponse, error) {
	validators := s.chain.State().TipState().ValidatorRegistry
	validatorsResponse := make([]*proto.ValidatorRegistry, len(validators))
	for i, v := range validators {
		newValidator := &proto.ValidatorRegistry{
			PublicKey:        hex.EncodeToString(v.PubKey),
			Status:           v.Status.String(),
			Balance:          decimal.NewFromInt(int64(v.Balance)).StringFixed(3),
			FirstActiveEpoch: v.FirstActiveEpoch,
			LastActiveEpoch:  v.LastActiveEpoch,
		}
		validatorsResponse[i] = newValidator
	}
	return &proto.GetValidatorsListResponse{Validators: validatorsResponse}, nil
}

func (s *validatorsServer) GenerateValidatorKey(context.Context, *proto.Empty) (*proto.GenerateValidatorKeyResponse, error) {
	key, err := s.wallet.GenerateNewValidatorKey()
	if err != nil {
		return nil, err
	}
	return &proto.GenerateValidatorKeyResponse{Key: hex.EncodeToString(key.Marshal())}, nil
}

func (s *validatorsServer) ExitValidator(context.Context, *proto.ExitValidatorInfo) (*proto.ExitValidatorResponse, error) {
	return nil, nil
}

func (s *validatorsServer) StartValidator(context.Context, *proto.StartValidatorInfo) (*proto.StartValidatorResponse, error) {
	return nil, nil
}
