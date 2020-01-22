package gov

import (
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
)

type Config struct {
	Log *logger.Logger
}

type GovMan struct {
	log    *logger.Logger
	config Config
	params params.ChainParams
}

func NewGovMan(config Config, params params.ChainParams) *GovMan {
	govMan := &GovMan{
		config: config,
		log:    config.Log,
		params: params,
	}
	return govMan
}
