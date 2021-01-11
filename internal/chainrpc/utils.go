package chainrpc

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/host"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/burnproof"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/params"

	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type utilsServer struct {
	netParams *params.ChainParams
	keystore  keystore.Keystore
	host      host.Host
	pool      mempool.Pool
	chain     chain.Blockchain
	proto.UnimplementedUtilsServer
}

func (s *utilsServer) GenKeyPair(_ context.Context, _ *proto.Empty) (*proto.KeyPair, error) {

	k, err := bls.RandKey()
	if err != nil {
		return nil, err
	}

	return &proto.KeyPair{Private: k.ToWIF(&s.netParams.AccountPrefixes), Public: k.PublicKey().ToAccount(&s.netParams.AccountPrefixes)}, nil
}

func (s *utilsServer) SubmitRawData(_ context.Context, data *proto.RawData) (*proto.Success, error) {
	dataBytes, err := hex.DecodeString(data.Data)
	if err != nil {
		return nil, err
	}
	switch data.Type {
	case "tx":
		tx := new(primitives.Tx)

		err := tx.Unmarshal(dataBytes)
		if err != nil {
			return &proto.Success{Success: false, Error: "unable to decode raw data"}, nil
		}

		msg := &p2p.MsgTx{Data: tx}

		err = s.pool.AddTx(tx)

		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		err = s.host.Broadcast(msg)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		return &proto.Success{Success: true, Data: tx.Hash().String()}, nil

	case "deposit":

		deposit := new(primitives.Deposit)

		err := deposit.Unmarshal(dataBytes)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		err = s.pool.AddDeposit(deposit)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		msg := &p2p.MsgDeposit{Data: deposit}

		err = s.host.Broadcast(msg)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		return &proto.Success{Success: true}, nil

	case "exit":

		exit := new(primitives.Exit)

		err := exit.Unmarshal(dataBytes)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		err = s.pool.AddExit(exit)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		msg := &p2p.MsgExit{Data: exit}

		err = s.host.Broadcast(msg)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		return &proto.Success{Success: true}, nil

	case "deposits_bulk":
		deposits := new(p2p.MsgDeposits)

		err := deposits.Unmarshal(dataBytes)
		if err != nil {
			return nil, errors.New("unable to decode raw data")
		}

		for _, d := range deposits.Data {
			err = s.pool.AddDeposit(d)
			if err != nil {
				return &proto.Success{Success: false, Error: err.Error()}, nil
			}
		}

		err = s.host.Broadcast(deposits)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		return &proto.Success{Success: true}, nil

	case "exits_bulk":
		exits := new(p2p.MsgExits)

		err := exits.Unmarshal(dataBytes)
		if err != nil {
			return nil, errors.New("unable to decode raw data")
		}

		for _, d := range exits.Data {
			err = s.pool.AddExit(d)
			if err != nil {
				return &proto.Success{Success: false, Error: err.Error()}, nil
			}
		}

		err = s.host.Broadcast(exits)
		if err != nil {
			return &proto.Success{Success: false, Error: err.Error()}, nil
		}

		return &proto.Success{Success: true}, nil

	default:
		return &proto.Success{Success: false, Error: "unknown raw data type"}, nil
	}
}

func (s *utilsServer) DecodeRawTransaction(_ context.Context, data *proto.RawData) (*proto.Tx, error) {

	dataBytes, err := hex.DecodeString(data.Data)
	if err != nil {
		return nil, err
	}
	tx := new(primitives.Tx)
	err = tx.Unmarshal(dataBytes)
	if err != nil {
		return nil, errors.New("unable to decode block raw data")
	}
	txParse := &proto.Tx{
		Hash:          tx.Hash().String(),
		To:            hex.EncodeToString(tx.To[:]),
		FromPublicKey: hex.EncodeToString(tx.FromPublicKey[:]),
		Amount:        tx.Amount,
		Nonce:         tx.Nonce,
		Fee:           tx.Fee,
		Signature:     hex.EncodeToString(tx.Signature[:]),
	}
	return txParse, nil
}

func (s *utilsServer) DecodeRawBlock(_ context.Context, data *proto.RawData) (*proto.Block, error) {

	dataBytes, err := hex.DecodeString(data.Data)
	if err != nil {
		return nil, err
	}
	block := new(primitives.Block)
	err = block.Unmarshal(dataBytes)
	if err != nil {
		return nil, errors.New("unable to decode block raw data")
	}
	blockParse := &proto.Block{
		Hash: block.Hash().String(),
		Header: &proto.BlockHeader{
			Version:                     block.Header.Version,
			Nonce:                       block.Header.Nonce,
			PrevBlockHash:               hex.EncodeToString(block.Header.PrevBlockHash[:]),
			Timestamp:                   block.Header.Timestamp,
			Slot:                        block.Header.Slot,
			FeeAddress:                  hex.EncodeToString(block.Header.FeeAddress[:]),
			VotesMerkleRoot:             hex.EncodeToString(block.Header.VoteMerkleRoot[:]),
			DepositsMerkleRoot:          hex.EncodeToString(block.Header.DepositMerkleRoot[:]),
			ExitsMerkleRoot:             hex.EncodeToString(block.Header.ExitMerkleRoot[:]),
			PartialExitsMerkleRoot:      hex.EncodeToString(block.Header.PartialExitMerkleRoot[:]),
			CoinProofsMerkleRoot:        hex.EncodeToString(block.Header.CoinProofsMerkleRoot[:]),
			ExecutionsMerkleRoot:        hex.EncodeToString(block.Header.ExecutionsMerkleRoot[:]),
			TxsMerkleRoot:               hex.EncodeToString(block.Header.TxsMerkleRoot[:]),
			VoteSlashingsMerkleRoot:     hex.EncodeToString(block.Header.VoteSlashingMerkleRoot[:]),
			RandaoSlashingsMerkleRoot:   hex.EncodeToString(block.Header.RANDAOSlashingMerkleRoot[:]),
			ProposerSlashingsMerkleRoot: hex.EncodeToString(block.Header.ProposerSlashingMerkleRoot[:]),
			MultiSignatureTxsMerkleRoot: hex.EncodeToString(block.Header.MultiSignatureTxsMerkleRoot[:]),
		},
		Txs:             block.GetTxs(),
		Signature:       hex.EncodeToString(block.Signature[:]),
		RandaoSignature: hex.EncodeToString(block.RandaoSignature[:]),
	}
	return blockParse, nil
}

func (s *utilsServer) SubmitRedeemProof(_ context.Context, data *proto.RedeemProof) (*proto.Success, error) {
	proofBytes, err := hex.DecodeString(data.Proof)
	if err != nil {
		return nil, err
	}

	addrBytes := []byte(data.Address)

	if len(addrBytes) != 44 {
		return &proto.Success{Error: errors.New("invalid address size").Error()}, nil
	}

	var address [44]byte
	copy(address[:], addrBytes)

	proofs := make([]*burnproof.CoinsProof, 0)
	buf := bytes.NewBuffer(proofBytes)
	for {
		proof := new(burnproof.CoinsProof)
		err = proof.Unmarshal(buf)
		if err != nil {
			return &proto.Success{Error: err.Error()}, nil
		}
		if buf.Len() < 0 {
			break
		}
		proofs = append(proofs, proof)
	}

	if len(proofs) > 2048 {
		return &proto.Success{Error: "too many proofs submited, max number is 2048"}, nil
	}

	serializableProofs := make([]*burnproof.CoinsProofSerializable, len(proofs))

	for i, p := range proofs {
		pser, err := p.ToSerializable(address)
		if err != nil {
			return &proto.Success{Error: err.Error()}, nil
		}

		serializableProofs[i] = pser

		// Add to a mempool and broadcast
		err = s.pool.AddCoinProof(pser)
		if err != nil {
			return &proto.Success{Error: err.Error()}, nil
		}

	}

	msg := &p2p.MsgProofs{Proofs: serializableProofs}

	err = s.host.Broadcast(msg)
	if err != nil {
		return &proto.Success{Success: false, Error: err.Error()}, nil
	}

	return &proto.Success{Success: true}, nil
}
