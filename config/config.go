package config

import (
	"errors"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

var (
	ErrorPathDontExist   = errors.New("the specified path for datadir doesn't exists")
	ErrorConfigDontExist = errors.New("unable to load config.toml from datadir")
)

const (
	FileName = "config.toml"
	version  = "0.1.0"
)

type Config struct {
	DataFolder       string
	Debug            bool
	Listen           bool
	NetworkName      string
	ConnectNodes     []string
	AddNodes         []string
	Port             int32
	MaxPeers         int32
	Mode             string
	Wallet           bool
	AddrGap          int32
	AccountsGenerate int32
}

var defaultConfig = Config{
	DataFolder:       AppDataDir("Ogen", false),
	Debug:            false,
	Listen:           true,
	NetworkName:      "testnet",
	Port:             24126,
	MaxPeers:         9,
	Mode:             "node",
	Wallet:           true,
	AddrGap:          20,
	AccountsGenerate: 10,
}

func OgenVersion() string {
	return version
}

func LoadConfig(dataDirPath string) *Config {
	confDefault := defaultConfig
	// Check if there is a config file flag, if don't use default
	var dataDir string
	// If dataDir is specified, check if folder exists.
	if dataDirPath != "" {
		if _, err := os.Stat(dataDirPath); os.IsNotExist(err) {
			log.Panic(ErrorPathDontExist)
		}
		dataDir = dataDirPath
	} else {
		// If is not specified, create the folder and config.toml file
		if _, err := os.Stat(confDefault.DataFolder); os.IsNotExist(err) {
			_ = os.Mkdir(confDefault.DataFolder, 0700)
		}
		dataDir = confDefault.DataFolder
		_, _ = os.OpenFile(confDefault.DataFolder+"/"+FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0700)
	}
	var loadedConfig Config
	_, err := toml.DecodeFile(dataDir+"/"+FileName, &loadedConfig)
	if err != nil {
		log.Panic(ErrorConfigDontExist)
	}
	// TODO optimize automatic reload
	if dataDirPath == "" {
		loadedConfig.DataFolder = defaultConfig.DataFolder
	} else {
		loadedConfig.DataFolder = dataDirPath
	}
	// Check differences for omitted values on critical values
	if loadedConfig.Port == 0 {
		loadedConfig.Port = defaultConfig.Port
	}
	if loadedConfig.NetworkName == "" {
		loadedConfig.NetworkName = defaultConfig.NetworkName
	}
	if loadedConfig.MaxPeers == 0 {
		loadedConfig.MaxPeers = defaultConfig.MaxPeers
	}
	if loadedConfig.Mode == "" {
		loadedConfig.Mode = defaultConfig.Mode
	}
	if loadedConfig.AccountsGenerate == 0 {
		loadedConfig.AccountsGenerate = defaultConfig.AccountsGenerate
	}
	if loadedConfig.AddrGap == 0 {
		loadedConfig.AddrGap = defaultConfig.AddrGap
	}
	return &loadedConfig
}
