package chainrpc

import (
	"bytes"
	"net/http"

	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/wallet"
)

// Wallet is the wallet RPC.
type Wallet struct {
	config *Config
	log    *logger.Logger

	wallet *wallet.Wallet
	chain  *chain.Blockchain
}

// NewRPCWallet constructs an RPC wallet.
func NewRPCWallet(wallet *wallet.Wallet, ch *chain.Blockchain) *Wallet {
	return &Wallet{
		wallet: wallet,
		chain:  ch,
	}
}

// Empty is an RPC method that does not take arguments.
type Empty struct{}

// GetAddress gets the address of the wallet.
func (w *Wallet) GetAddress(req *http.Request, args *interface{}, reply *string) error {
	*reply = w.wallet.GetAddress()
	return nil
}

// GetBalance gets the balance of the wallet or another address.
func (w *Wallet) GetBalance(req *http.Request, args *string, reply *uint64) error {
	bal, err := w.wallet.GetBalance(*args)
	if err != nil {
		return err
	}

	*reply = bal
	return nil
}

// SendToAddressRequest is the request to send money to someone.
type SendToAddressRequest struct {
	Password  []byte
	ToAddress string
	Amount    uint64
}

// SendToAddress sends money to an address.
func (w *Wallet) SendToAddress(req *http.Request, args *SendToAddressRequest, reply *chainhash.Hash) error {
	reply, err := w.wallet.SendToAddress(args.Password, args.ToAddress, args.Amount)
	return err
}

// ValidatorResponse is the response the wallet sends for a validator.
type ValidatorResponse struct {
	Pubkey            [48]byte                   `json: "pubkey"`
	Balance           uint64                     `json: "balance"`
	Status            primitives.ValidatorStatus `json: "status"`
	HavePrivateKey    bool                       `json: "have_private_key"`
	HaveWithdrawalKey bool                       `json: "have_withdrawal_key"`
}

// ValidatorListReponse is the response the wallet sends for a list of validators.
type ValidatorListReponse struct {
	Validators []ValidatorResponse `json: "validators"`
}

// ListValidators lists all validators the user owns or controls.
func (w *Wallet) ListValidators(req *http.Request, args *interface{}, reply *ValidatorListReponse) error {
	currentState := w.chain.State().TipState()
	walletAddress, err := w.wallet.GetAddressRaw()
	if err != nil {
		return err
	}

	validators := make([]ValidatorResponse, 0)

	for _, v := range currentState.ValidatorRegistry {
		hasWithdrawalKey := false
		if bytes.Equal(v.PayeeAddress[:], walletAddress[:]) {
			hasWithdrawalKey = true
		}
		var pub [48]byte
		copy(pub[:], v.PubKey)
		ok, err := w.wallet.ValidatorWallet.HasValidatorKey(pub)
		var hasPrivateKey bool
		if err != nil {
			hasPrivateKey = false
		} else {
			hasPrivateKey = ok
		}

		validators = append(validators, ValidatorResponse{
			Pubkey:            pub,
			Balance:           v.Balance,
			Status:            v.Status,
			HavePrivateKey:    hasPrivateKey,
			HaveWithdrawalKey: hasWithdrawalKey,
		})
	}

	*reply = ValidatorListReponse{
		Validators: validators,
	}

	return nil
}

// ValidatorKeyResponse is the response to a generate key request.
type ValidatorKeyResponse struct {
	PrivateKey []byte
}

// GenerateValidatorKey generates a validator key and adds it to the wallet.
func (w *Wallet) GenerateValidatorKey(req *http.Request, args *interface{}, reply *ValidatorKeyResponse) error {
	secKey, err := w.wallet.GenerateNewValidatorKey()
	if err != nil {
		return err
	}

	*reply = ValidatorKeyResponse{
		PrivateKey: secKey.Marshal(),
	}
	return nil
}

// StartValidatorRequest is the request to start a validator.
type StartValidatorRequest struct {
	PrivateKey [32]byte
	Password   []byte
}

// StartValidatorResponse is the response to starting a validator.
type StartValidatorResponse struct {
	PublicKey []byte
}

// StartValidator starts a validator.
func (w *Wallet) StartValidator(req *http.Request, args *StartValidatorRequest, reply *StartValidatorResponse) error {
	deposit, err := w.wallet.StartValidator(args.Password, args.PrivateKey)
	if err != nil {
		return err
	}

	*reply = StartValidatorResponse{
		PublicKey: deposit.Data.PublicKey.Marshal(),
	}
	return nil
}

// ExitValidatorRequest is the request to exit a validator.
type ExitValidatorRequest struct {
	ValidatorPubKey [48]byte
	Password        []byte
}

// ExitValidatorResponse is the response to exiting a validator.
type ExitValidatorResponse struct {
	Success bool
}

// ExitValidator exits the validator.
func (w *Wallet) ExitValidator(req *http.Request, args *ExitValidatorRequest, reply *ExitValidatorResponse) error {
	_, err := w.wallet.ExitValidator(args.Password, args.ValidatorPubKey)
	if err != nil {
		return err
	}

	*reply = ExitValidatorResponse{
		Success: true,
	}

	return nil
}
