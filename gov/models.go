package gov

import (
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

type GovObject struct {
	GovID         chainhash.Hash
	Amount        int16
	Cycles        int32
	PayedCycles   int32
	BurnedUtxo    p2p.OutPoint
	Name          string
	URL           string
	PayoutAddress string
	Votes         map[p2p.OutPoint]Vote
}

func (g *GovObject) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, g.GovID, g.Amount, g.Cycles, g.PayedCycles)
	if err != nil {
		return err
	}
	err = g.BurnedUtxo.Serialize(w)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, g.Name)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, g.URL)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, g.PayoutAddress)
	if err != nil {
		return err
	}
	err = serializer.WriteVarInt(w, uint64(len(g.Votes)))
	for _, vote := range g.Votes {
		err = vote.Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GovObject) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &g.GovID, &g.Amount, &g.Cycles, &g.PayedCycles)
	if err != nil {
		return err
	}
	err = g.BurnedUtxo.Deserialize(r)
	if err != nil {
		return err
	}
	g.Name, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	g.URL, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	g.PayoutAddress, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	voteCount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	g.Votes = make(map[p2p.OutPoint]Vote, voteCount)
	for i := uint64(0); i < voteCount; i++ {
		var vote Vote
		err = vote.Deserialize(r)
		if err != nil {
			return err
		}
		g.Votes[vote.WorkerID] = vote
	}
	return nil
}

type Vote struct {
	GovID    chainhash.Hash
	Approval bool
	WorkerID p2p.OutPoint
}

func (v *Vote) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, v.GovID, v.Approval)
	if err != nil {
		return err
	}
	err = v.WorkerID.Serialize(w)
	if err != nil {
		return err
	}
	return nil
}

func (v *Vote) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &v.GovID, &v.Approval)
	if err != nil {
		return err
	}
	err = v.WorkerID.Deserialize(r)
	if err != nil {
		return err
	}
	return nil
}
