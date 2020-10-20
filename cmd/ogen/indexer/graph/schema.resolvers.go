package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/olympus-protocol/ogen/cmd/ogen/indexer/graph/generated"
	"github.com/olympus-protocol/ogen/cmd/ogen/indexer/graph/model"
)

func (r *queryResolver) Block(ctx context.Context) (*model.Block, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Txs(ctx context.Context) (*model.TxSingle, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Account(ctx context.Context) (*model.Account, error) {
	panic(fmt.Errorf("not implemented"))
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
