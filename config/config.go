package config

import (
	"errors"
	"github.com/BurntSushi/toml"
	"log"
	"os"
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
	Port             int32
	MaxPeers         int32
	Mode             string
	Wallet           bool
}

func OgenVersion() string {
	return version
}
