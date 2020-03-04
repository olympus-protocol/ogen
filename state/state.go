package state

import (
	"bytes"
	"fmt"
	"github.com/olympus-protocol/ogen/gov"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/users"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"github.com/olympus-protocol/ogen/workers"
	"io"
)

type State struct {
	UtxoState       UtxoState
	GovernanceState GovernanceState
	UserState       UserState
	WorkerState     WorkerState
}

type GovernanceProposal struct {
	OutPoint p2p.OutPoint
	GovData  gov.GovObject
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

type GovernanceState struct {
	Proposals map[chainhash.Hash]GovernanceProposal
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

type Utxo struct {
	OutPoint          p2p.OutPoint
	PrevInputsPubKeys [][48]byte
	Owner             string
	Amount            int64
}

// Serialize serializes the UtxoRow to a writer.
func (l *Utxo) Serialize(w io.Writer) error {
	err := l.OutPoint.Serialize(w)
	if err != nil {
		return err
	}
	err = serializer.WriteVarInt(w, uint64(len(l.PrevInputsPubKeys)))
	if err != nil {
		return err
	}
	err = serializer.WriteElements(w, l.PrevInputsPubKeys)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, l.Owner)
	if err != nil {
		return err
	}
	err = serializer.WriteElement(w, l.Amount)
	if err != nil {
		return err
	}
	return nil
}

// Deserialize deserializes a UtxoRow from a reader.
func (l *Utxo) Deserialize(r io.Reader) error {
	err := l.OutPoint.Deserialize(r)
	if err != nil {
		return err
	}
	l.Owner, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	count, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	l.PrevInputsPubKeys = make([][48]byte, count)
	for i := uint64(0); i < count; i++ {
		var pubKey [48]byte
		err = serializer.ReadElement(r, &pubKey)
		if err != nil {
			return err
		}
		l.PrevInputsPubKeys = append(l.PrevInputsPubKeys, pubKey)
	}
	err = serializer.ReadElement(r, &l.Amount)
	if err != nil {
		return err
	}
	return nil
}

func (l *Utxo) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = l.OutPoint.Serialize(buf)
	return chainhash.DoubleHashH(buf.Bytes())
}

type UtxoState struct {
	UTXOs map[chainhash.Hash]Utxo
}

// Have checks if a UTXO exists.
func (u *UtxoState) Have(c chainhash.Hash) bool {
	fmt.Println(c, u.UTXOs)
	_, found := u.UTXOs[c]
	return found
}

// Get gets the UTXO from state.
func (u *UtxoState) Get(c chainhash.Hash) Utxo {
	return u.UTXOs[c]
}

func (s *State) TransitionBlock(block primitives.Block) State {
	return *s
}

type User struct {
	OutPoint p2p.OutPoint
	UserData users.User
}

// Serialize serializes the UserRow to a writer.
func (ur *User) Serialize(w io.Writer) error {
	err := ur.OutPoint.Serialize(w)
	if err != nil {
		return err
	}
	err = ur.UserData.Serialize(w)
	if err != nil {
		return err
	}
	return nil
}

// Deserialize deserializes a user from the writer.
func (ur *User) Deserialize(r io.Reader) error {
	err := ur.OutPoint.Deserialize(r)
	if err != nil {
		return err
	}
	err = ur.UserData.Deserialize(r)
	if err != nil {
		return err
	}
	return nil
}

func (ur *User) Hash() chainhash.Hash {
	return chainhash.DoubleHashH([]byte(ur.UserData.Name))
}

type UserState struct {
	Users map[chainhash.Hash]User
}

// Have checks if a User exists.
func (u *UserState) Have(c chainhash.Hash) bool {
	_, found := u.Users[c]
	return found
}

// Get gets a User from state.
func (u *UserState) Get(c chainhash.Hash) User {
	return u.Users[c]
}

type Worker struct {
	OutPoint   p2p.OutPoint
	WorkerData workers.Worker
}

// Serialize serializes a WorkerRow to the provided writer.
func (wr *Worker) Serialize(w io.Writer) error {
	err := wr.OutPoint.Serialize(w)
	if err != nil {
		return err
	}
	err = wr.WorkerData.Serialize(w)
	if err != nil {
		return err
	}
	return nil
}

// Deserialize deserializes a worker row from the provided reader.
func (wr *Worker) Deserialize(r io.Reader) error {
	err := wr.OutPoint.Deserialize(r)
	if err != nil {
		return err
	}
	err = wr.WorkerData.Deserialize(r)
	if err != nil {
		return err
	}
	return nil
}

type WorkerState struct {
	Workers map[chainhash.Hash]Worker
}

// Have checks if a Worker exists.
func (u *WorkerState) Have(c chainhash.Hash) bool {
	_, found := u.Workers[c]
	return found
}

// Get gets a Worker from state.
func (u *WorkerState) Get(c chainhash.Hash) Worker {
	return u.Workers[c]
}
