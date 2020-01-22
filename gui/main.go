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
	s, err := server.NewServer(configParams, log, currParams, db, true)
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
