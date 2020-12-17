package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/olympus-protocol/ogen/internal/indexer/graph/generated"
	"github.com/olympus-protocol/ogen/internal/indexer/graph/model"
)

func (r *queryResolver) Accounts(ctx context.Context) ([]*model.Account, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) BlockHeaders(ctx context.Context) ([]*model.BlockHeader, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Blocks(ctx context.Context) ([]*model.Block, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Deposits(ctx context.Context) ([]*model.Deposit, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Epoches(ctx context.Context) ([]*model.Epoch, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Exits(ctx context.Context) ([]*model.Exit, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Slots(ctx context.Context) ([]*model.Slot, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Txes(ctx context.Context) ([]*model.Tx, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Validators(ctx context.Context) ([]*model.Validator, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Votes(ctx context.Context) ([]*model.Vote, error) {
	panic(fmt.Errorf("not implemented"))
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }