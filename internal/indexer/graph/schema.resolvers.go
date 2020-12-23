package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/olympus-protocol/ogen/internal/indexer/db"
	"github.com/olympus-protocol/ogen/internal/indexer/graph/generated"
	"github.com/olympus-protocol/ogen/internal/indexer/graph/model"
)

func (r *queryResolver) Accounts(ctx context.Context) ([]*model.Account, error) {
	var data []*db.Account
	r.DB.DB.Find(&data)
	gql := make([]*model.Account, len(data))
	for i := range gql {
		gql[i] = data[i].ToGQL()
	}
	return gql, nil
}

func (r *queryResolver) BlockHeaders(ctx context.Context) ([]*model.BlockHeader, error) {
	var data []*db.BlockHeader
	r.DB.DB.Find(&data)
	gql := make([]*model.BlockHeader, len(data))
	for i := range gql {
		gql[i] = data[i].ToGQL()
	}
	return gql, nil
}

func (r *queryResolver) Blocks(ctx context.Context) ([]*model.Block, error) {
	var data []*db.Block
	r.DB.DB.Find(&data)
	gql := make([]*model.Block, len(data))
	for i := range gql {
		gql[i] = data[i].ToGQL()
	}
	return gql, nil
}

func (r *queryResolver) Deposits(ctx context.Context) ([]*model.Deposit, error) {
	var data []*db.Deposit
	r.DB.DB.Find(&data)
	gql := make([]*model.Deposit, len(data))
	for i := range gql {
		gql[i] = data[i].ToGQL()
	}
	return gql, nil
}

func (r *queryResolver) Epoches(ctx context.Context) ([]*model.Epoch, error) {
	var data []*db.Epoch
	r.DB.DB.Find(&data)
	gql := make([]*model.Epoch, len(data))
	for i := range gql {
		gql[i] = data[i].ToGQL()
	}
	return gql, nil
}

func (r *queryResolver) Exits(ctx context.Context) ([]*model.Exit, error) {
	var data []*db.Exit
	r.DB.DB.Find(&data)
	gql := make([]*model.Exit, len(data))
	for i := range gql {
		gql[i] = data[i].ToGQL()
	}
	return gql, nil
}

func (r *queryResolver) Slots(ctx context.Context) ([]*model.Slot, error) {
	var data []*db.Slot
	r.DB.DB.Find(&data)
	gql := make([]*model.Slot, len(data))
	for i := range gql {
		gql[i] = data[i].ToGQL()
	}
	return gql, nil
}

func (r *queryResolver) Txes(ctx context.Context) ([]*model.Tx, error) {
	var data []*db.Tx
	r.DB.DB.Find(&data)
	gql := make([]*model.Tx, len(data))
	for i := range gql {
		gql[i] = data[i].ToGQL()
	}
	return gql, nil
}

func (r *queryResolver) Validators(ctx context.Context) ([]*model.Validator, error) {
	var data []*db.Validator
	r.DB.DB.Find(&data)
	gql := make([]*model.Validator, len(data))
	for i := range gql {
		gql[i] = data[i].ToGQL()
	}
	return gql, nil
}

func (r *queryResolver) Votes(ctx context.Context) ([]*model.Vote, error) {
	var data []*db.Vote
	r.DB.DB.Find(&data)
	gql := make([]*model.Vote, len(data))
	for i := range gql {
		gql[i] = data[i].ToGQL()
	}
	return gql, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
