package chainrpc

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/olympus-protocol/ogen/internal/state"
	"reflect"

	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/shopspring/decimal"
)

type chainServer struct {
	chain chain.Blockchain
	proto.UnimplementedChainServer
}

func (s *chainServer) GetChainInfo(ctx context.Context, _ *proto.Empty) (*proto.ChainInfo, error) {
	st := s.chain.State()
	tip := st.Tip()
	validators := st.TipState().GetValidators()
	return &proto.ChainInfo{
		BlockHash:   tip.Hash.String(),
		BlockHeight: tip.Height,
		Validators: &proto.ValidatorsInfo{
			Active:      validators.Active,
			PendingExit: validators.PendingExit,
			PenaltyExit: validators.PenaltyExit,
			Exited:      validators.Exited,
			Starting:    validators.Starting,
		},
	}, nil
}

func (s *chainServer) GetRawBlock(ctx context.Context, in *proto.Hash) (*proto.Block, error) {
	hash, err := chainhash.NewHashFromStr(in.Hash)
	if err != nil {
		return nil, err
	}
	block, err := s.chain.GetRawBlock(*hash)
	if err != nil {
		return nil, err
	}
	return &proto.Block{RawBlock: hex.EncodeToString(block)}, nil
}

func (s *chainServer) GetBlock(ctx context.Context, in *proto.Hash) (*proto.Block, error) {
	hash, err := chainhash.NewHashFromStr(in.Hash)
	if err != nil {
		return nil, err
	}
	block, err := s.chain.GetBlock(*hash)
	if err != nil {
		return nil, err
	}
	blockParse := &proto.Block{
		Hash: block.Hash().String(),
		Header: &proto.BlockHeader{
			Version:                    block.Header.Version,
			Nonce:                      block.Header.Nonce,
			TxMerkleRoot:               hex.EncodeToString(block.Header.TxMerkleRoot[:]),
			VoteMerkleRoot:             hex.EncodeToString(block.Header.VoteMerkleRoot[:]),
			DepositMerkleRoot:          hex.EncodeToString(block.Header.DepositMerkleRoot[:]),
			ExitMerkleRoot:             hex.EncodeToString(block.Header.ExitMerkleRoot[:]),
			VoteSlashingMerkleRoot:     hex.EncodeToString(block.Header.VoteSlashingMerkleRoot[:]),
			RandaoSlashingMerkleRoot:   hex.EncodeToString(block.Header.RANDAOSlashingMerkleRoot[:]),
			ProposerSlashingMerkleRoot: hex.EncodeToString(block.Header.ProposerSlashingMerkleRoot[:]),
			PrevBlockHash:              hex.EncodeToString(block.Header.PrevBlockHash[:]),
			Timestamp:                  block.Header.Timestamp,
			Slot:                       block.Header.Slot,
			StateRoot:                  hex.EncodeToString(block.Header.StateRoot[:]),
			FeeAddress:                 hex.EncodeToString(block.Header.FeeAddress[:]),
		},
		Txs:             block.GetTxs(),
		Signature:       hex.EncodeToString(block.Signature[:]),
		RandaoSignature: hex.EncodeToString(block.RandaoSignature[:]),
	}
	return blockParse, nil
}

func (s *chainServer) GetBlockHash(ctx context.Context, in *proto.Number) (*proto.Hash, error) {
	blockRow, exists := s.chain.State().Chain().GetNodeByHeight(in.Number)
	if !exists {
		return nil, errors.New("block not found")
	}
	return &proto.Hash{
		Hash: blockRow.Hash.String(),
	}, nil
}

func (s *chainServer) Sync(in *proto.Hash, stream proto.Chain_SyncServer) error {
	_, cancel := context.WithCancel(stream.Context())
	// Define starting point
	blockRow := new(chainindex.BlockRow)
	defer cancel()
	// If user is on tip, silently close the channel
	if reflect.DeepEqual(in.Hash, s.chain.State().Tip().Hash.String()) {
		return nil
	}

	hash, err := hex.DecodeString(in.Hash)
	if err != nil {
		return errors.New("unable to decode hash from string")
	}

	var hashB [32]byte
	copy(hashB[:], hash)

	currBlockRow, ok := s.chain.State().GetRowByHash(hashB)
	if !ok {
		return errors.New("block starting point doesnt exist")
	}
	blockRow, ok = s.chain.State().Chain().Next(currBlockRow)
	if !ok {
		return errors.New("there is no next blockrow")
	}
	for {
		rawBlock, err := s.chain.GetRawBlock(blockRow.Hash)
		if err != nil {
			return errors.New("unable get raw block")
		}
		response := &proto.RawData{
			Data: hex.EncodeToString(rawBlock),
		}
		err = stream.Send(response)
		if err != nil {
			return err
		}
		blockRow, ok = s.chain.State().Chain().Next(blockRow)
		if blockRow == nil || !ok {
			break
		}
	}
	return nil
}

type blockNotifee struct {
	blocks chan blockAndReceipts
}

type blockAndReceipts struct {
	block    *primitives.Block
	receipts []*primitives.EpochReceipt
	state    state.State
}

func newBlockNotifee(ctx context.Context, chain chain.Blockchain) blockNotifee {
	bn := blockNotifee{
		blocks: make(chan blockAndReceipts),
	}

	go func() {
		chain.Notify(&bn)

		<-ctx.Done()

		chain.Unnotify(&bn)
	}()

	return bn
}

func (bn *blockNotifee) NewTip(row *chainindex.BlockRow, block *primitives.Block, newState state.State, receipts []*primitives.EpochReceipt) {
	toSend := blockAndReceipts{block: block, receipts: receipts, state: newState}
	select {
	case bn.blocks <- toSend:
	default:
	}
}

func (bn *blockNotifee) ProposerSlashingConditionViolated(slashing *primitives.ProposerSlashing) {}

func (s *chainServer) SubscribeBlocks(_ *proto.Empty, stream proto.Chain_SubscribeBlocksServer) error {
	bn := newBlockNotifee(stream.Context(), s.chain)

	for {
		select {
		case bl := <-bn.blocks:
			buf, err := bl.block.Marshal()
			if err != nil {
				return err
			}
			err = stream.Send(&proto.RawData{
				Data: hex.EncodeToString(buf),
			})
			if err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (s *chainServer) SubscribeTransactions(in *proto.KeyPairs, stream proto.Chain_SubscribeTransactionsServer) error {
	bn := newBlockNotifee(stream.Context(), s.chain)
	accounts := make(map[[20]byte]struct{})
	for _, a := range in.Keys {
		account, err := hex.DecodeString(a)
		if err != nil {
			return err
		}
		if len(account) != 20 {
			return fmt.Errorf("expected public key hashes to be 20 bytes but got %d", len(a))
		}

		var acc [20]byte
		copy(acc[:], a)

		accounts[acc] = struct{}{}
	}

	for {
		select {
		case bl := <-bn.blocks:
			err := stream.Send(&proto.RawData{
				Data: hex.EncodeToString(bl.block.SerializedTx(accounts)),
			})
			if err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (s *chainServer) SubscribeValidatorTransaction(in *proto.KeyPairs, stream proto.Chain_SubscribeValidatorTransactionsServer) error {
	bn := newBlockNotifee(stream.Context(), s.chain)
	accounts := make(map[[48]byte]struct{})
	for _, a := range in.Keys {
		pubkey, err := hex.DecodeString(a)
		if err != nil {
			return err
		}
		if len(pubkey) != 48 {
			return fmt.Errorf("expected public key to be 48 bytes but got %d", len(a))
		}

		var acc [48]byte
		copy(acc[:], a)

		accounts[acc] = struct{}{}
	}

	for {
		select {
		case bl := <-bn.blocks:
			err := stream.Send(&proto.RawData{
				Data: hex.EncodeToString(bl.block.SerializedEpochs(accounts)),
			})
			if err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (s *chainServer) GetAccountInfo(ctx context.Context, data *proto.Account) (*proto.AccountInfo, error) {
	var account [20]byte
	_, decoded, err := bech32.Decode(data.Account)
	if err != nil {
		return nil, err
	}
	copy(account[:], decoded)
	nonce := s.chain.State().TipState().GetCoinsState().Nonces[account]
	confirmed := decimal.NewFromInt(int64(s.chain.State().TipState().GetCoinsState().Balances[account])).DivRound(decimal.NewFromInt(1e8), 8)
	lock := decimal.NewFromInt(0)
	for _, v := range s.chain.State().TipState().GetValidatorRegistry() {
		if v.PayeeAddress == account {
			lock = lock.Add(decimal.NewFromInt(int64(v.Balance)))
		}
	}
	balance := &proto.Balance{
		Confirmed: confirmed.String(),
		Locked:    lock.String(),
		Total:     decimal.Zero.Add(confirmed).Add(lock).String(),
	}
	accInfo := &proto.AccountInfo{
		Account: data.Account,
		Balance: balance,
		Nonce:   nonce,
	}
	return accInfo, nil
}

var _ proto.ChainServer = &chainServer{}
