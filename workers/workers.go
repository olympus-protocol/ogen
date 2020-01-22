package workers

import (
	"github.com/grupokindynos/ogen/logger"
	"github.com/grupokindynos/ogen/params"
)

type Config struct {
	Log *logger.Logger
}

type WorkerMan struct {
	config Config
	log    *logger.Logger
	params params.ChainParams
}

func NewWorkersMan(config Config, params params.ChainParams) *WorkerMan {
	workersMan := &WorkerMan{
		config: config,
		log:    config.Log,
		params: params,
	}
	return workersMan
}
