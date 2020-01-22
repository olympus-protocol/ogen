package users

import (
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
)

type Config struct {
	Log *logger.Logger
}

type UserMan struct {
	config Config
	log    *logger.Logger
	params params.ChainParams
}

func NewUsersMan(config Config, params params.ChainParams) *UserMan {
	usersMan := &UserMan{
		config: config,
		log:    config.Log,
		params: params,
	}
	return usersMan
}
