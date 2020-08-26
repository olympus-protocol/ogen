package chainrpc

import (
	"context"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/wallet"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/shopspring/decimal"
)

type walletServer struct {
	wallet wallet.Wallet
	chain  chain.Blockchain
	params *params.ChainParams
	proto.UnimplementedWalletServer
}

func (s *walletServer) ListWallets(context.Context, *proto.Empty) (*proto.Wallets, error) {
	files, err := s.wallet.GetAvailableWallets()
	if err != nil {
		return nil, err
	}
	var walletFiles []string
	for k := range files {
		walletFiles = append(walletFiles, k)
	}
	return &proto.Wallets{Wallets: walletFiles}, nil
}
func (s *walletServer) CreateWallet(ctx context.Context, ref *proto.WalletReference) (*proto.KeyPair, error) {
	err := s.wallet.NewWallet(ref.Name, nil, ref.Password)
	if err != nil {
		return nil, err
	}
	pubKey, err := s.wallet.GetAccount()
	if err != nil {
		return nil, err
	}
	return &proto.KeyPair{Public: pubKey}, nil
}

func (s *walletServer) OpenWallet(ctx context.Context, ref *proto.WalletReference) (*proto.Success, error) {
	ok := s.wallet.HasWallet(ref.Name)
	if !ok {
		return nil, errors.New("the is no wallet with the current name specified")
	}
	err := s.wallet.OpenWallet(ref.Name, ref.Password)
	if err != nil {
		return &proto.Success{Success: false, Error: err.Error()}, nil
	}
	return &proto.Success{Success: true}, nil
}

func (s *walletServer) CloseWallet(context.Context, *proto.Empty) (*proto.Success, error) {
	err := s.wallet.CloseWallet()
	if err != nil {
		return &proto.Success{Success: false, Error: err.Error()}, nil
	}
	return &proto.Success{Success: true}, nil
}

func (s *walletServer) ImportWallet(ctx context.Context, in *proto.ImportWalletData) (*proto.KeyPair, error) {
	name := in.Name
	if name == "" {
		return nil, errors.New("please specify a name for the wallet")
	}
	prefix, priv, err := bech32.Decode(in.Key.Private)
	if err != nil {
		return nil, err
	}
	if prefix != s.params.AccountPrefixes.Private {
		return nil, errors.New("wrong wallet format for current network")
	}
	blsPriv, err := bls.SecretKeyFromBytes(priv)
	if err != nil {
		return nil, err
	}
	err = s.wallet.NewWallet(name, blsPriv, in.Password)
	if err != nil {
		return nil, err
	}
	acc, err := s.wallet.GetAccount()
	if err != nil {
		return nil, err
	}
	return &proto.KeyPair{Public: acc}, nil
}

func (s *walletServer) DumpWallet(context.Context, *proto.Empty) (*proto.KeyPair, error) {
	priv, err := s.wallet.GetSecret()
	if err != nil {
		return nil, err
	}
	return &proto.KeyPair{Private: priv.ToWIF()}, nil
}

func (s *walletServer) GetBalance(context.Context, *proto.Empty) (*proto.Balance, error) {
	acc, err := s.wallet.GetAccountRaw()
	if err != nil {
		return nil, err
	}
	balance, err := s.wallet.GetBalance()
	if err != nil {
		return nil, err
	}
	balanceStr := decimal.NewFromInt(int64(balance)).DivRound(decimal.NewFromInt(1e8), 8).String()
	validators := s.getValidators(acc)
	lock := decimal.NewFromInt(0)
	for _, v := range validators.Validators {
		b, err := decimal.NewFromString(v.Balance)
		if err != nil {
			return nil, err
		}
		lock = lock.Add(b)
	}
	return &proto.Balance{Confirmed: balanceStr, Locked: lock.StringFixed(3), Total: decimal.NewFromInt(int64(balance)).DivRound(decimal.NewFromInt(1e8), 8).Add(lock).StringFixed(3)}, nil
}

func (s *walletServer) GetValidators(ctx context.Context, _ *proto.Empty) (*proto.ValidatorsRegistry, error) {
	acc, err := s.wallet.GetAccountRaw()
	if err != nil {
		return nil, err
	}
	return s.getValidators(acc), nil
}

func (s *walletServer) getValidators(acc [20]byte) *proto.ValidatorsRegistry {
	validators := s.chain.State().TipState().GetValidatorsForAccount(acc[:])
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
	}}
}

func (s *walletServer) GetAccount(context.Context, *proto.Empty) (*proto.KeyPair, error) {
	account, err := s.wallet.GetAccount()
	if err != nil {
		return nil, err
	}
	return &proto.KeyPair{Public: account}, nil
}

func (s *walletServer) SendTransaction(ctx context.Context, send *proto.SendTransactionInfo) (*proto.Hash, error) {
	amount, err := decimal.NewFromString(send.Amount)
	if err != nil {
		return nil, err
	}
	amountFixed := amount.Mul(decimal.NewFromInt(1e8)).Round(0)
	hash, err := s.wallet.SendToAddress(send.Account, uint64(amountFixed.IntPart()))
	if err != nil {
		return nil, err
	}
	return &proto.Hash{Hash: hash.String()}, nil
}
func (s *walletServer) StartValidator(ctx context.Context, key *proto.KeyPair) (*proto.Success, error) {
	var privKeyBytes [32]byte
	privKeyDecode, err := hex.DecodeString(key.Private)
	if err != nil {
		return nil, err
	}
	copy(privKeyBytes[:], privKeyDecode)
	_, err = s.wallet.StartValidator(privKeyBytes)
	if err != nil {
		return nil, err
	}
	return &proto.Success{Success: true}, nil
}

func (s *walletServer) StartValidatorBulk(ctx context.Context, keys *proto.KeyPairs) (*proto.Success, error) {
	var wg sync.WaitGroup
	for _, key := range keys.Keys {
		wg.Add(1)
		go func(wg *sync.WaitGroup, keyStr string) {
			var privKeyBytes [32]byte
			privKeyDecode, err := hex.DecodeString(keyStr)
			if err != nil {
				return
			}
			copy(privKeyBytes[:], privKeyDecode)

			_, err = s.wallet.StartValidator(privKeyBytes)
			if err != nil {
				return
			}
		}(&wg, key)

	}
	return &proto.Success{Success: true}, nil
}

func (s *walletServer) ExitValidator(ctx context.Context, key *proto.KeyPair) (*proto.Success, error) {
	var pubKeyBytes [48]byte
	pubKeyDecode, err := hex.DecodeString(key.Public)
	if err != nil {
		return nil, err
	}
	copy(pubKeyBytes[:], pubKeyDecode)
	_, err = s.wallet.ExitValidator(pubKeyBytes)
	if err != nil {
		return &proto.Success{Success: false, Error: err.Error()}, nil
	}
	return &proto.Success{Success: true}, nil
}

func (s *walletServer) ExitValidatorBulk(ctx context.Context, keys *proto.KeyPairs) (*proto.Success, error) {
	for i := range keys.Keys {
		var pubKeyBytes [48]byte
		pubKeyDecode, err := hex.DecodeString(keys.Keys[i])
		if err != nil {
			return nil, err
		}
		copy(pubKeyBytes[:], pubKeyDecode)
		_, err = s.wallet.ExitValidator(pubKeyBytes)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

	}
	return &proto.Success{Success: true}, nil
}
