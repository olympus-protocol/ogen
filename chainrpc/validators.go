package chainrpc

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"

	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chainrpc/proto"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/shopspring/decimal"
)

type validatorsServer struct {
	keystore *keystore.Keystore
	params   *params.ChainParams
	chain    *chain.Blockchain
	proto.UnimplementedValidatorsServer
}

func (s *validatorsServer) GetValidatorsList(context.Context, *proto.Empty) (*proto.ValidatorsRegistry, error) {
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

func (s *validatorsServer) GetAccountValidators(ctx context.Context, acc *proto.Account) (*proto.ValidatorsRegistry, error) {
	var account []byte
	_, account, err := bech32.Decode(acc.Account)
	if err != nil {
		account, err = hex.DecodeString(acc.Account)
		if err != nil {
			return nil, errors.New("unable to decode account")
		}
	}
	var accountValidators []*proto.ValidatorRegistry
	validators := s.chain.State().TipState().ValidatorRegistry
	for _, v := range validators {
		if bytes.EqualFold(account, v.PayeeAddress[:]) {
			validator := &proto.ValidatorRegistry{
				Balance:      decimal.NewFromInt(int64(v.Balance)).StringFixed(3),
				PublicKey:    hex.EncodeToString(v.PubKey),
				PayeeAddress: bech32.Encode(s.params.AddrPrefix.Public, v.PayeeAddress[:]),
				Status:       v.Status.String(),
			}
			accountValidators = append(accountValidators, validator)
		}
	}
	return &proto.ValidatorsRegistry{Validators: accountValidators}, nil
}
