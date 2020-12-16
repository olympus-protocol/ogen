package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/olympus-protocol/ogen/internal/indexer/graph/generated"
	"github.com/olympus-protocol/ogen/internal/indexer/graph/model"
)

func (r *queryResolver) Blocks(ctx context.Context) ([]*model.Block, error) {
	var blocks []*model.Block
	r.Resolver.DB.DB.Preload("blocks").Find(&blocks)
	return blocks, nil
}

func (r *queryResolver) BlockHeaders(ctx context.Context) ([]*model.BlockHeader, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Txs(ctx context.Context) ([]*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Account(ctx context.Context) ([]*model.Account, error) {
	panic(fmt.Errorf("not implemented"))
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
