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
	Pubkey            [48]byte
	Balance           uint64
	Status            primitives.WorkerStatus
	HavePrivateKey    bool
	HaveWithdrawalKey bool
}

// ValidatorListReponse is the respons the wallet sends for a list of validators.
type ValidatorListReponse struct {
	Validators []ValidatorResponse
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
		ok, err := w.wallet.ValidatorWallet.HasValidatorKey(v.PubKey)
		if err != nil {
			return err
		}
		hasPrivateKey := ok

		validators = append(validators, ValidatorResponse{
			Pubkey:            v.PubKey,
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
