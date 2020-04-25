package wallet

import (
	"github.com/olympus-protocol/ogen/db/walletdb"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/bip39"
	"github.com/olympus-protocol/ogen/utils/hdwallets"
	"os"
	"os/exec"
	"strconv"
)

const currWalletVersion = 100000
const accounts = 100

type Config struct {
	Log      *logger.Logger
	Path     string
	Enabled  bool
	Gui      bool
}

type activeWallet struct {
	meta        *walletdb.WalletMetaData
	credentials *walletdb.WalletCredentials
	utxos       []walletdb.WalletUtxo
	txs         []walletdb.WalletTx
}
type WalletMan struct {
	log          *logger.Logger
	config       Config
	params       params.ChainParams
	wallet       *walletdb.WalletDB
	activeWallet activeWallet
}

func raw(start bool) error {
	r := "raw"
	if !start {
		r = "-raw"
	}
	rawMode := exec.Command("stty", r)
	rawMode.Stdin = os.Stdin
	err := rawMode.Run()
	if err != nil {
		return err
	}
	return rawMode.Wait()
}

func (wm *WalletMan) Start() error {
	wm.log.Info("Starting WalletMan instance")
start:
	walletMeta, err := wm.wallet.GetMetadata()
	if err != nil {
		// no meta bucket means the wallet is not initialized.
		// here we should create the wallet struct.
		if err == walletdb.ErrorNoMetaBucket {
			err = wm.initWallet("")
			goto start
		}
		wm.log.Warn("Unable to load wallet metadata. Possible wallet corruption")
		return err
	}
	walletCreds, err := wm.wallet.GetCredentials()
	if err != nil {
		wm.log.Fatal("Unable to load wallet credentials. Possible wallet corruption")
		return err
	}
	walletUtxos, err := wm.wallet.GetUtxos()
	if err != nil {
		wm.log.Fatal("Unable to load wallet utxos. Need wallet rescan")
		return err
	}
	walletTxs, err := wm.wallet.GetTxs()
	if err != nil {
		wm.log.Fatal("Unable to load wallet txs. Need wallet rescan")
		return err
	}
	wm.activeWallet = activeWallet{
		meta:        walletMeta,
		credentials: walletCreds,
		utxos:       walletUtxos,
		txs:         walletTxs,
	}

	return nil
}

func (wm *WalletMan) initWallet(pass string) error {
	newMeta := walletdb.WalletMetaData{
		Version:         currWalletVersion,
		Txs:             0,
		Utxos:           0,
		Accounts:        0,
		LastBlockHash:   wm.params.GenesisHash,
		LastBlockHeight: 0,
	}
	err := wm.wallet.StoreMetadata(newMeta)
	if err != nil {
		return err
	}
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return err
	}
	seed := bip39.NewSeed(mnemonic, pass)
	hdRoot, err := hdwallets.NewMaster(seed, &wm.params.HDPrefixes)
	if err != nil {
		return err
	}
	newCredentials := walletdb.WalletCredentials{
		Accounts: make(map[int32]walletdb.Account, accounts),
		Mnemonic: mnemonic,
	}
	purpose, err := hdRoot.Child(44 + hdwallets.HardenedKeyStart)
	if err != nil {
		return err
	}
	coin, err := purpose.Child(wm.params.HDCoinIndex + hdwallets.HardenedKeyStart)
	if err != nil {
		return err
	}
	for i := int32(0); i < accounts; i++ {
		acc, err := coin.Child(uint32(i) + hdwallets.HardenedKeyStart)
		if err != nil {
			return err
		}
		pubAcc, err := acc.Neuter(&wm.params.HDPrefixes)
		if err != nil {
			return err
		}
		newAcc := walletdb.Account{
			Number:            i,
			Path:              "m/44'/" + strconv.Itoa(int(wm.params.HDCoinIndex)) + "'/" + strconv.Itoa(int(i)) + "'",
			ExtendedPublicKey: pubAcc.String(),
		}
		newCredentials.AddAccount(newAcc)
		acc.Zero()
	}

	coin.Zero()
	purpose.Zero()
	hdRoot.Zero()
	err = wm.wallet.StoreCredentials(&newCredentials)
	if err != nil {
		return err
	}
	err = wm.wallet.InitUtxosBucket()
	if err != nil {
		return err
	}
	err = wm.wallet.InitTxBucket()
	if err != nil {
		return err
	}
	return nil
}

func (wm *WalletMan) Stop() error {
	wm.log.Info("Stoping WalletMan instance")
	err := wm.wallet.Close()
	if err != nil {
		return err
	}
	return nil
}

func NewWalletMan(config Config, params params.ChainParams) (*WalletMan, error) {
	walletDB := walletdb.NewWalletDB(config.Path + "/wallet.dat")
	wm := &WalletMan{
		log:    config.Log,
		config: config,
		params: params,
		wallet: walletDB,
	}
	return wm, nil
}
