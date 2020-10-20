package graph

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/indexer/db"
)

// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	db *db.Database
}

func NewResolver(db *db.Database) *Resolver {
	return &Resolver{db: db}
}
