package chainrpc

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/wallet"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/shopspring/decimal"
)

type walletServer struct {
	wallet    wallet.Wallet
	chain     chain.Blockchain
	netParams *params.ChainParams
	proto.UnimplementedWalletServer
}

func (s *walletServer) ListWallets(ctx context.Context, _ *proto.Empty) (*proto.Wallets, error) {
	defer ctx.Done()

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
func (s *walletServer) CreateWallet(ctx context.Context, ref *proto.WalletReference) (*proto.NewWalletInfo, error) {
	defer ctx.Done()

	err := s.wallet.NewWallet(ref.Name, "", ref.Password)
	if err != nil {
		return nil, err
	}

	pubKey, err := s.wallet.GetAccount()
	if err != nil {
		return nil, err
	}

	mnemonic, err := s.wallet.GetMnemonic()
	if err != nil {
		return nil, err
	}

	return &proto.NewWalletInfo{Name: ref.Name, Account: pubKey, Mnemonic: mnemonic}, nil
}

func (s *walletServer) OpenWallet(ctx context.Context, ref *proto.WalletReference) (*proto.Success, error) {
	defer ctx.Done()

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

func (s *walletServer) CloseWallet(ctx context.Context, _ *proto.Empty) (*proto.Success, error) {
	defer ctx.Done()

	err := s.wallet.CloseWallet()
	if err != nil {
		return &proto.Success{Success: false, Error: err.Error()}, nil
	}

	return &proto.Success{Success: true}, nil
}

func (s *walletServer) ImportWallet(ctx context.Context, in *proto.ImportWalletData) (*proto.KeyPair, error) {
	defer ctx.Done()

	name := in.Name
	if name == "" {
		return nil, errors.New("please specify a name for the wallet")
	}

	err := s.wallet.NewWallet(name, in.Mnemonic, in.Password)
	if err != nil {
		return nil, err
	}

	acc, err := s.wallet.GetAccount()
	if err != nil {
		return nil, err
	}

	return &proto.KeyPair{Public: acc}, nil
}

func (s *walletServer) DumpWallet(ctx context.Context, _ *proto.Empty) (*proto.KeyPair, error) {
	defer ctx.Done()

	priv, err := s.wallet.GetSecret()
	if err != nil {
		return nil, err
	}

	return &proto.KeyPair{Private: priv.ToWIF()}, nil
}

func (s *walletServer) DumpHDWallet(ctx context.Context, _ *proto.Empty) (*proto.DumpHDWalletInfo, error) {
	defer ctx.Done()

	mnemonic, err := s.wallet.GetMnemonic()
	if err != nil {
		return nil, err
	}

	return &proto.DumpHDWalletInfo{Mnemonic: mnemonic}, nil
}

func (s *walletServer) GetBalance(ctx context.Context, _ *proto.Empty) (*proto.Balance, error) {
	defer ctx.Done()

	acc, err := s.wallet.GetAccountRaw()
	if err != nil {
		return nil, err
	}

	balance, err := s.wallet.GetBalance()
	if err != nil {
		return nil, err
	}

	validators := s.getValidators(acc)
	lock := decimal.NewFromInt(0)
	for _, v := range validators.Validators {
		b, err := decimal.NewFromString(v.Balance)
		if err != nil {
			return nil, err
		}
		lock = lock.Add(b)
	}

	return &proto.Balance{
			Confirmed:   decimal.NewFromInt(int64(balance.Confirmed)).DivRound(decimal.NewFromInt(1e8), 8).StringFixed(8),
			Unconfirmed: decimal.NewFromInt(int64(balance.Pending)).DivRound(decimal.NewFromInt(1e8), 8).StringFixed(8),
			Locked:      lock.StringFixed(8),
			Total:       decimal.NewFromInt(int64(balance.Confirmed)).DivRound(decimal.NewFromInt(1e8), 8).Add(lock).StringFixed(8)},
		nil
}

func (s *walletServer) GetValidators(ctx context.Context, _ *proto.Empty) (*proto.ValidatorsRegistry, error) {
	defer ctx.Done()

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
			Balance:          decimal.NewFromInt(int64(v.Balance)).Div(decimal.NewFromInt(int64(s.netParams.UnitsPerCoin))).StringFixed(3),
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

func (s *walletServer) GetAccount(ctx context.Context, _ *proto.Empty) (*proto.KeyPair, error) {
	defer ctx.Done()

	account, err := s.wallet.GetAccount()
	if err != nil {
		return nil, err
	}

	return &proto.KeyPair{Public: account}, nil
}

func (s *walletServer) SendTransaction(ctx context.Context, send *proto.SendTransactionInfo) (*proto.Hash, error) {
	defer ctx.Done()

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
	defer ctx.Done()

	privKeyBytes, err := hex.DecodeString(key.Private)
	if err != nil {
		return &proto.Success{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	privKeyBls, err := bls.SecretKeyFromBytes(privKeyBytes)
	if err != nil {
		return &proto.Success{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	success, err := s.wallet.StartValidator(privKeyBls)
	if err != nil {
		return &proto.Success{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	return &proto.Success{
		Success: success,
	}, nil
}

func (s *walletServer) StartValidatorBulk(ctx context.Context, keys *proto.KeyPairs) (*proto.Success, error) {
	defer ctx.Done()

	keysStr := keys.Keys
	blsKeys := make([]*bls.SecretKey, len(keysStr))

	for i := range blsKeys {
		privKeyDecode, err := hex.DecodeString(keysStr[i])
		if err != nil {
			return &proto.Success{
				Success: false,
				Error:   err.Error(),
				Data:    "",
			}, nil
		}
		key, err := bls.SecretKeyFromBytes(privKeyDecode)
		if err != nil {
			return &proto.Success{
				Success: false,
				Error:   err.Error(),
				Data:    "",
			}, nil
		}
		blsKeys[i] = key
	}

	success, err := s.wallet.StartValidatorBulk(blsKeys)
	if err != nil {
		return &proto.Success{
			Success: false,
			Error:   err.Error(),
			Data:    "",
		}, nil
	}
	return &proto.Success{
		Success: success,
	}, nil
}

func (s *walletServer) ExitValidator(ctx context.Context, key *proto.KeyPair) (*proto.Success, error) {
	defer ctx.Done()

	pubKeyBytes, err := hex.DecodeString(key.Public)
	if err != nil {
		return &proto.Success{
			Success: false,
			Error:   err.Error(),
			Data:    "",
		}, nil
	}

	pubKeyBls, err := bls.PublicKeyFromBytes(pubKeyBytes)
	if err != nil {
		return &proto.Success{
			Success: false,
			Error:   err.Error(),
			Data:    "",
		}, nil
	}

	success, err := s.wallet.ExitValidator(pubKeyBls)
	if err != nil {
		return &proto.Success{
			Success: false,
			Error:   err.Error(),
			Data:    "",
		}, nil
	}
	return &proto.Success{
		Success: success,
	}, nil
}

func (s *walletServer) ExitValidatorBulk(ctx context.Context, keys *proto.KeyPairs) (*proto.Success, error) {
	defer ctx.Done()

	keysStr := keys.Keys
	blsKeys := make([]*bls.PublicKey, len(keysStr))

	for i := range blsKeys {
		pubKeyBytes, err := hex.DecodeString(keysStr[i])
		if err != nil {
			return &proto.Success{
				Success: false,
				Error:   err.Error(),
			}, nil
		}
		key, err := bls.PublicKeyFromBytes(pubKeyBytes)
		if err != nil {
			return &proto.Success{
				Success: false,
				Error:   err.Error(),
			}, nil
		}
		blsKeys[i] = key
	}

	success, err := s.wallet.ExitValidatorBulk(blsKeys)
	if err != nil {
		return &proto.Success{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	return &proto.Success{
		Success: success,
	}, nil
}
