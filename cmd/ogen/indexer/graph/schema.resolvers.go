package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
)

func (r *queryResolver) Block(ctx context.Context) (*Block, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Blocks(ctx context.Context) ([]*Block, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Tx(ctx context.Context) (*TxSingle, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Txs(ctx context.Context) ([]*TxSingle, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Account(ctx context.Context) (*Account, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Accounts(ctx context.Context) ([]*Account, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Deposit(ctx context.Context) (*Deposit, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Deposits(ctx context.Context) ([]*Deposit, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Exit(ctx context.Context) (*Exit, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Exits(ctx context.Context) ([]*Exit, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Vote(ctx context.Context) (*Votes, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Votes(ctx context.Context) ([]*Votes, error) {
	panic(fmt.Errorf("not implemented"))
}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
