package main

import (
	"flag"
	"github.com/grupokindynos/ogen/config"
	"github.com/grupokindynos/ogen/db/blockdb"
	"github.com/grupokindynos/ogen/logger"
	"github.com/grupokindynos/ogen/params"
	"github.com/grupokindynos/ogen/server"
	"os"
)

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
	listenChan := config.InterruptListener(log)
	s, err := server.NewServer(configParams, log, currParams, db, false)
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
