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

func (s *validatorsServer) ValidatorsList(context.Context, *proto.Empty) (*proto.ValidatorsRegistry, error) {
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
	return &proto.ValidatorsRegistry{Validators: validatorsResponse}, nil
}

func (s *validatorsServer) ExitValidator(context.Context, *proto.ExitValidatorInfo) (*proto.Success, error) {
	return nil, nil
}

func (s *validatorsServer) StartValidator(context.Context, *proto.StartValidatorInfo) (*proto.KeyPair, error) {
	return nil, nil
}

func (s *validatorsServer) GetAccountValidators(context.Context, *proto.Account) (*proto.ValidatorsRegistry, error) {
	return nil, nil
}
