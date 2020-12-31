package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/olympus-protocol/ogen/internal/indexer/db"
	"github.com/olympus-protocol/ogen/internal/indexer/graph/generated"
	"github.com/olympus-protocol/ogen/internal/indexer/graph/model"
)

func (r *queryResolver) Account(ctx context.Context, account string) (*model.Account, error) {
	var acc db.Account
	res := r.DB.DB.Where(&db.Account{Account: account}).First(&acc)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return &model.Account{
				Account: account,
				Balance: 0,
				Nonce:   0,
			}, nil
		}
		return nil, res.Error
	}

	return acc.ToGQL(), nil
}

func (r *queryResolver) AccountValidators(ctx context.Context, account string) ([]*model.Validator, error) {
	var validators []db.Validator

	res := r.DB.DB.Where(db.Validator{PayeeAddress: account}).Find(&validators)

	if res.Error != nil {
		return nil, res.Error
	}

	if len(validators) == 0 {
		return []*model.Validator{}, nil
	}

	vals := make([]*model.Validator, len(validators))

	for i := range vals {
		vals[i] = validators[i].ToGQL()
	}

	return vals, nil
}

func (r *queryResolver) AccountCoinProofs(ctx context.Context, account string) ([]*model.CoinProofs, error) {
	var coinProofs []db.CoinProofs
	res := r.DB.DB.Where(db.CoinProofs{RedeemAccount: account}).Find(&coinProofs)

	if res.Error != nil {
		return nil, res.Error
	}

	if len(coinProofs) == 0 {
		return []*model.CoinProofs{}, nil
	}

	proofs := make([]*model.CoinProofs, len(coinProofs))

	for i := range proofs {
		proofs[i] = coinProofs[i].ToGQL()
	}

	return proofs, nil
}

func (r *queryResolver) AccountTxs(ctx context.Context, account string) ([]*model.Tx, error) {
	var txs []*model.Tx

	var sentTxs []db.Tx
	var receiveTxs []db.Tx

	res := r.DB.DB.Where(db.Tx{FromPublicKeyHash: account}).Find(&sentTxs)

	if res.Error != nil {
		return nil, res.Error
	}

	res = r.DB.DB.Where(db.Tx{ToAddress: account}).Find(&receiveTxs)

	if res.Error != nil {
		return nil, res.Error
	}

	if len(sentTxs) > 0 {
		for _, tx := range sentTxs {
			txs = append(txs, tx.ToGQL())
		}
	}

	if len(receiveTxs) > 0 {
		for _, tx := range sentTxs {
			txs = append(txs, tx.ToGQL())
		}
	}

	return txs, nil
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

func (r *queryResolver) Validator(ctx context.Context, pubkey string) (*model.Validator, error) {
	pub, err := hex.DecodeString(pubkey)
	if err != nil {
		return nil, err
	}

	var validator db.Validator

	res := r.DB.DB.Where(&db.Validator{PubKey: pub}).First(&validator)

	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, errors.New("no validator found")
		}
		return nil, res.Error
	}

	return validator.ToGQL(), nil
}

func (r *queryResolver) Slot(ctx context.Context, slot int) (*model.Slot, error) {
	var s db.Slot

	res := r.DB.DB.Where(&db.Slot{Slot: uint64(slot)}).First(&s)

	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, errors.New("no slot found")
		}
		return nil, res.Error
	}

	return s.ToGQL(), nil
}

func (r *queryResolver) Epoch(ctx context.Context, epoch int) (*model.Epoch, error) {
	var e db.Epoch

	res := r.DB.DB.Where(&db.Epoch{Epoch: uint64(epoch)}).First(&e)

	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, errors.New("no epoch found")
		}
		return nil, res.Error
	}

	return e.ToGQL(), nil
}

func (r *queryResolver) Tx(ctx context.Context, hash string) (*model.Tx, error) {
	h, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	var tx db.Tx

	res := r.DB.DB.Where(&db.Tx{Hash: h}).First(&tx)

	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, errors.New("no tx found")
		}
		return nil, res.Error
	}

	return tx.ToGQL(), nil
}

func (r *queryResolver) BlockBySlot(ctx context.Context, slot int) (*model.Block, error) {
	var block db.Block

	res := r.DB.DB.Where(&db.Block{Slot: uint64(slot)}).First(&block)

	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, errors.New("no block found")
		}
		return nil, res.Error
	}

	return block.ToGQL(), nil
}

func (r *queryResolver) BlockByHash(ctx context.Context, hash string) (*model.Block, error) {
	h, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	var block db.Block

	res := r.DB.DB.Where(&db.Block{Hash: h}).First(&block)

	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, errors.New("no block found")
		}
		return nil, res.Error
	}

	return block.ToGQL(), nil
}

func (r *queryResolver) BlockByHeight(ctx context.Context, height int) (*model.Block, error) {
	var block db.Block

	res := r.DB.DB.Where(&db.Block{Height: uint64(height)}).First(&block)

	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, errors.New("no block found")
		}
		return nil, res.Error
	}

	return block.ToGQL(), nil
}

func (r *queryResolver) Tip(ctx context.Context) (*model.Tip, error) {
	validators, err := r.Validators(ctx)
	if err != nil {
		return nil, err
	}

	var slot db.Slot
	res := r.DB.DB.Select(&db.Slot{}, "max(slot)").Scan(&slot)

	if res.Error != nil || res.RowsAffected == -1 {
		return nil, errors.New("error trying to load latest slot")
	}

	var epoch db.Epoch
	res = r.DB.DB.Select(&db.Epoch{}, "max(epoch)").Scan(&epoch)

	if res.Error != nil || res.RowsAffected == -1 {
		return nil, errors.New("error trying to load latest epoch")
	}

	var block db.Block
	res = r.DB.DB.Select(&db.Block{}, "max(height)").Scan(&block)

	if res.Error != nil || res.RowsAffected == -1 {
		return nil, errors.New("error trying to load latest block")
	}

	return &model.Tip{
		Slot:       slot.ToGQL(),
		Epoch:      epoch.ToGQL(),
		Block:      block.ToGQL(),
		Validators: validators,
	}, nil
}

func (r *subscriptionResolver) Account(ctx context.Context, account string) (<-chan *model.Account, error) {
	accChannel := make(chan *model.Account)

	go func() {
		var initAccData db.Account
		res := r.DB.DB.Where(&db.Account{Account: account}).First(&initAccData)

		if res.Error != nil {

			if res.Error == gorm.ErrRecordNotFound {

				accChannel <- &model.Account{
					Account: account,
					Balance: 0,
					Nonce:   0,
				}

			}

		} else {
			accChannel <- initAccData.ToGQL()
		}
		u := uuid.New()
		accNotify := db.NewAccountBalanceNotify(account, accChannel, r.DB)

		r.DB.AddAccountBalanceNotifier(account, u, accNotify)
		<-ctx.Done()
		r.DB.RemoveAccountBalanceNotifier(account, u)
		return
	}()

	return accChannel, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
