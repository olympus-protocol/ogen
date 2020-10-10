package execution

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/csmt"
	"github.com/olympus-protocol/ogen/pkg/logger"
)

type Execution interface {
}

type execution struct {
	datapath string
	db       csmt.Tree
	log      logger.Logger
}

var _ Execution = &execution{}

func NewExecutionInstance() (Execution, error) {
	datapath := config.GlobalFlags.DataPath
	log := config.GlobalParams.Logger

	tree := csmt.NewTree(csmt.NewInMemoryTreeDB())

	return &execution{
		datapath: datapath,
		db:       tree,
		log:      log,
	}, nil
}
