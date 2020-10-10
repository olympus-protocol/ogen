package execution

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/csmt"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"path"
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

	opts := badger.DefaultOptions(path.Join(datapath, "modules"))

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	tree := csmt.NewTree(csmt.NewBadgerTreeDB(db))

	return &execution{
		datapath: datapath,
		db:       tree,
		log:      log,
	}, nil
}
