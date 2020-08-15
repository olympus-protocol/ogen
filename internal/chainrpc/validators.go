package chainrpc

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/shopspring/decimal"
)

type validatorsServer struct {
	keystore *keystore.Keystore
	params   *params.ChainParams
	chain    chain.Blockchain
	proto.UnimplementedValidatorsServer
}

func (s *validatorsServer) GetValidatorsList(context.Context, *proto.Empty) (*proto.ValidatorsRegistry, error) {
	validators := s.chain.State().TipState().GetValidators()
	parsedValidators := make([]*proto.ValidatorRegistry, len(validators.Validators))
	for i, v := range validators.Validators {
		newValidator := &proto.ValidatorRegistry{
			PublicKey:        hex.EncodeToString(v.PubKey[:]),
			Status:           v.StatusString(),
			Balance:          decimal.NewFromInt(int64(v.Balance)).Div(decimal.NewFromInt(int64(s.params.UnitsPerCoin))).StringFixed(3),
			FirstActiveEpoch: v.FirstActiveEpoch,
			LastActiveEpoch:  v.LastActiveEpoch,
		}
		parsedValidators[i] = newValidator
	}
	return &proto.ValidatorsRegistry{Validators: parsedValidators, Info: &proto.ValidatorsInfo{
		Active:      validators.Active,
		PendingExit: validators.PendingExit,
		PenaltyExit: validators.PenaltyExit,
		Exited:      validators.Exited,
		Starting:    validators.Starting,
	}}, nil
}

func (s *validatorsServer) GetAccountValidators(ctx context.Context, acc *proto.Account) (*proto.ValidatorsRegistry, error) {
	var account []byte
	_, account, err := bech32.Decode(acc.Account)
	if err != nil {
		account, err = hex.DecodeString(acc.Account)
		if err != nil {
			return nil, errors.New("unable to decode account")
		}
	}
	validators := s.chain.State().TipState().GetValidatorsForAccount(account)
	parsedValidators := make([]*proto.ValidatorRegistry, len(validators.Validators))
	for i, v := range validators.Validators {
		newValidator := &proto.ValidatorRegistry{
			PublicKey:        hex.EncodeToString(v.PubKey[:]),
			Status:           v.StatusString(),
			Balance:          decimal.NewFromInt(int64(v.Balance)).Div(decimal.NewFromInt(int64(s.params.UnitsPerCoin))).StringFixed(3),
			FirstActiveEpoch: v.FirstActiveEpoch,
			LastActiveEpoch:  v.LastActiveEpoch,
		}
		parsedValidators[i] = newValidator
	}
	return &proto.ValidatorsRegistry{Validators: parsedValidators, Info: &proto.ValidatorsInfo{
		Active:      validators.Active,
		PendingExit: validators.PendingExit,
		PenaltyExit: validators.PenaltyExit,
		Exited:      validators.Exited,
		Starting:    validators.Starting,
	}}, nil
}
