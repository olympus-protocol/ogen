package db

import (
	"errors"
	"github.com/olympus-protocol/ogen/internal/state"
)

var ErrorNoState = errors.New("unable to fetch database state")

func (d *Database) GetState() (state.State, error) {

}

func (d *Database) SetState() error {

}
