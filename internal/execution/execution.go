package execution

import (
	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/internal/csmt"
	"github.com/olympus-protocol/ogen/internal/logger"
	"path"
)

type Execution interface {
}

type execution struct {
	path string
	db   csmt.Tree
	log  logger.Logger
}

var _ Execution = &execution{}

func NewExecutionInstance(dbdir string, log logger.Logger) (Execution, error) {
	opts := badger.DefaultOptions(path.Join(dbdir, "modules"))
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	tree := csmt.NewTree(csmt.NewBadgerTreeDB(db))
	return &execution{
		path: dbdir,
		db:   tree,
		log:  log,
	}, nil
}
