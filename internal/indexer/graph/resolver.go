package graph

import "github.com/olympus-protocol/ogen/internal/indexer/db"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB *db.Database
}
