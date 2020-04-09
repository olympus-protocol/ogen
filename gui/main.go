package main

import (
	"crypto/rand"
	"flag"
	"os"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/config"
	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/miner"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/server"
)

const preferenceCurrentTab = "currentTab"

func main() {
	var dataDirPath = flag.String("datadir", "", "Directory to store Ogen data")
	flag.Parse()
	conf := config.LoadConfig(*dataDirPath)
	f, err := os.OpenFile(conf.DataFolder+"/debug.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {

	}
	log := logger.New(f).WithoutColor().WithoutTimestamp()
	if conf.Debug {
		log = log.WithDebug()
	}
	log.Infof("Starting Ogen v%v", config.OgenVersion())
	log.Trace("loading log on debug mode")
	err = loadOgen(conf, log)
	if err != nil {
		log.Fatal(err)
	}
}

const numTestValidators = 128

func getTestInitializationParameters() (*primitives.InitializationParameters, []bls.SecretKey) {
	vals := make([]primitives.ValidatorInitialization, numTestValidators)
	keys := make([]bls.SecretKey, numTestValidators)
	for i := range vals {
		k, err := bls.RandSecretKey(rand.Reader)
		if err != nil {
			panic(err)
		}

		keys[i] = *k

		vals[i] = primitives.ValidatorInitialization{
			PubKey:       keys[i].DerivePublicKey().Serialize(),
			PayeeAddress: "",
		}
	}

	return &primitives.InitializationParameters{
		InitialValidators: vals,
	}, keys
}

func loadOgen(configParams *config.Config, log *logger.Logger) error {
	var currParams params.ChainParams
	switch configParams.NetworkName {
	case "mainnet":
		currParams = params.Mainnet
	default:
		currParams = params.TestNet
	}
	loadGuiApp()
	db, err := blockdb.NewBlockDB(configParams.DataFolder, currParams, log)
	if err != nil {
		return err
	}
	// TODO: replace this with something better
	testParams, keys := getTestInitializationParameters()
	s, err := server.NewServer(configParams, log, currParams, db, true, *testParams, miner.NewBasicKeystore(keys))
	if err != nil {
		return err
	}
	go s.Start()
	defer func() {
		_ = s.Stop()
		db.Close()
	}()
	return nil
}

func loadGuiApp() {

	return
}
