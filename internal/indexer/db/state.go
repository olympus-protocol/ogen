package db

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/olympus-protocol/ogen/internal/state"
)

func (d *Database) SetState() error {

	return nil
}

func (d *Database) GetState() error {
	s := state.NewEmptyState()

	return nil
}
