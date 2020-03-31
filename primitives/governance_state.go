package primitives

import (
	"io"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

type GovObject struct {
	GovID         chainhash.Hash
	Amount        int16
	Cycles        int32
	PayedCycles   int32
	BurnedUtxo    OutPoint
	Name          string
	URL           string
	PayoutAddress string
	Votes         map[OutPoint]Vote
}

func (g *GovObject) Copy() GovObject {
	g2 := *g

	g2.Votes = make(map[OutPoint]Vote)
	for i, v := range g.Votes {
		g2.Votes[i] = v.Copy()
	}

	return g2
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
	for outpoint, vote := range g.Votes {
		err = outpoint.Serialize(w)
		if err != nil {
			return err
		}
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
	g.Votes = make(map[OutPoint]Vote, voteCount)
	for i := uint64(0); i < voteCount; i++ {
		var outpoint OutPoint
		if err := outpoint.Deserialize(r); err != nil {
			return err
		}
		var vote Vote
		err = vote.Deserialize(r)
		if err != nil {
			return err
		}
		g.Votes[outpoint] = vote
	}
	return nil
}

type Vote struct {
	GovID    chainhash.Hash
	Approval bool
	WorkerID OutPoint
}

func (v *Vote) Copy() Vote {
	return *v
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

type GovernanceProposal struct {
	OutPoint OutPoint
	GovData  GovObject
}

// Serialize serializes a GovRow to the passed writer.
func (gr *GovernanceProposal) Serialize(w io.Writer) error {
	err := gr.OutPoint.Serialize(w)
	if err != nil {
		return err
	}
	err = gr.GovData.Serialize(w)
	if err != nil {
		return err
	}
	return nil
}

// Deserialize deserialized a GovRow from the passed reader.
func (gr *GovernanceProposal) Deserialize(r io.Reader) error {
	err := gr.OutPoint.Deserialize(r)
	if err != nil {
		return err
	}
	err = gr.GovData.Deserialize(r)
	if err != nil {
		return err
	}
	return nil
}

func (gr *GovernanceProposal) Copy() GovernanceProposal {
	return GovernanceProposal{
		OutPoint: gr.OutPoint,
		GovData:  gr.GovData.Copy(),
	}
}

type GovernanceState struct {
	Proposals map[chainhash.Hash]GovernanceProposal
}

func (g *GovernanceState) Copy() GovernanceState {
	g2 := *g
	g2.Proposals = make(map[chainhash.Hash]GovernanceProposal)
	for i, c := range g.Proposals {
		g2.Proposals[i] = c.Copy()
	}
	return g2
}

// Have checks if the governance state contains a specific proposal hash.
func (g *GovernanceState) Have(c chainhash.Hash) bool {
	_, ok := g.Proposals[c]
	return ok
}

// Get gets a governance proposal from the governance state.
func (g *GovernanceState) Get(c chainhash.Hash) GovernanceProposal {
	return g.Proposals[c]
}

func (g *GovernanceState) Serialize(w io.Writer) error {
	if err := serializer.WriteVarInt(w, uint64(len(g.Proposals))); err != nil {
		return err
	}

	for h, proposal := range g.Proposals {
		if _, err := w.Write(h[:]); err != nil {
			return err
		}

		if err := proposal.Serialize(w); err != nil {
			return err
		}
	}

	return nil
}

func (g *GovernanceState) Deserialize(r io.Reader) error {
	if g.Proposals == nil {
		g.Proposals = make(map[chainhash.Hash]GovernanceProposal)
	}

	numProposals, err := serializer.ReadVarInt(r)

	if err != nil {
		return err
	}

	for i := uint64(0); i < numProposals; i++ {
		var hash chainhash.Hash
		if _, err := r.Read(hash[:]); err != nil {
			return err
		}

		var proposal GovernanceProposal
		if err := proposal.Deserialize(r); err != nil {
			return err
		}

		g.Proposals[hash] = proposal
	}

	return nil
}
