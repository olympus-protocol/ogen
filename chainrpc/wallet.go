package chainrpc

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/olympus-protocol/ogen/chainrpc/proto"
	"github.com/olympus-protocol/ogen/wallet"
	"github.com/shopspring/decimal"
)

type walletServer struct {
	wallet *wallet.Wallet
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
func (s *walletServer) CreateWallet(ctx context.Context, name *proto.Name) (*proto.KeyPair, error) {
	err := s.wallet.OpenWallet(name.Name)
	if err != nil {
		return nil, err
	}
	pubKey, err := s.wallet.GetAccount()
	if err != nil {
		return nil, err
	}
	return &proto.KeyPair{Public: pubKey}, nil
}

func (s *walletServer) OpenWallet(ctx context.Context, name *proto.Name) (*proto.Success, error) {
	ok := s.wallet.HasWallet(name.Name)
	if !ok {
		return nil, errors.New("the is no wallet with the current name specified")
	}
	err := s.wallet.OpenWallet(name.Name)
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

func (s *walletServer) GetBalance(context.Context, *proto.Empty) (*proto.Balance, error) {
	balance, err := s.wallet.GetBalance()
	if err != nil {
		return nil, err
	}
	return &proto.Balance{Confirmed: balance}, nil
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
	amountFixed := amount.Mul(decimal.NewFromInt(1e3)).Round(0)
	hash, err := s.wallet.SendToAddress(send.Account, uint64(amountFixed.IntPart()))
	if err != nil {
		return nil, err
	}
	return &proto.Hash{Hash: hash[:]}, nil
}
func (s *walletServer) StartValidator(ctx context.Context, key *proto.KeyPair) (*proto.KeyPair, error) {
	var privKeyBytes [32]byte
	privKeyDecode, err := hex.DecodeString(key.Private)
	if err != nil {
		return nil, err
	}
	copy(privKeyBytes[:], privKeyDecode)
	deposit, err := s.wallet.StartValidator(privKeyBytes)
	if err != nil {
		return nil, err
	}
	pubKeyStr := hex.EncodeToString(deposit.PublicKey.Marshal())
	return &proto.KeyPair{Public: pubKeyStr}, nil
}

func (s *walletServer) ExitValidator(ctx context.Context, key *proto.KeyPair) (*proto.Success, error) {
	var pubKeyBytes [48]byte
	pubKeyDecode, err := hex.DecodeString(key.Private)
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
