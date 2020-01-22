package gov

import (
	"github.com/grupokindynos/ogen/logger"
	"github.com/grupokindynos/ogen/params"
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
