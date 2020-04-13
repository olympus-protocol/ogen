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

var log = logger.New(os.Stdin)

func main() {
	var dataDirPath = flag.String("datadir", "", "Directory to store Ogen data")
	flag.Parse()
	conf := config.LoadConfig(*dataDirPath)
	log := logger.New(os.Stdin)
	if conf.Debug {
		log = log.WithDebug()
	}
	log.Infof("Starting Ogen v%v", config.OgenVersion())
	log.Trace("loading log on debug mode")
	err := loadOgen(conf, log)
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

// loadOgen is the main function to run ogen.
func loadOgen(configParams *config.Config, log *logger.Logger) error {
	var currParams params.ChainParams
	switch configParams.NetworkName {
	case "mainnet":
		currParams = params.Mainnet
	default:
		currParams = params.TestNet
	}
	db, err := blockdb.NewBlockDB(configParams.DataFolder, currParams, log)
	if err != nil {
		return err
	}
	ip, keys := getTestInitializationParameters()
	listenChan := config.InterruptListener(log)
	s, err := server.NewServer(configParams, log, currParams, db, false, *ip, miner.NewBasicKeystore(keys))
	if err != nil {
		return err
	}
	go s.Start()
	<-listenChan
	err = s.Stop()
	if err != nil {

	}
	db.Close()
	return nil
}
