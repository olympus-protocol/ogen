package state

import (
	"bytes"
	"github.com/olympus-protocol/ogen/gov"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/users"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

type State struct {
	UtxoState       UtxoState
	GovernanceState GovernanceState
	UserState       UserState
	WorkerRegistry  WorkerRegistry
	WorkerQueue     []chainhash.Hash
}

func (s *State) Serialize(w io.Writer) error {
	if err := s.UtxoState.Serialize(w); err != nil {
		return err
	}
	if err := s.GovernanceState.Serialize(w); err != nil {
		return err
	}
	if err := s.UserState.Serialize(w); err != nil {
		return err
	}
	if err := s.WorkerRegistry.Serialize(w); err != nil {
		return err
	}
	return nil
}

func (s *State) Deserialize(r io.Reader) error {
	if err := s.UtxoState.Deserialize(r); err != nil {
		return err
	}
	if err := s.GovernanceState.Deserialize(r); err != nil {
		return err
	}
	if err := s.UserState.Deserialize(r); err != nil {
		return err
	}
	if err := s.WorkerRegistry.Deserialize(r); err != nil {
		return err
	}
	return nil
}

type GovernanceProposal struct {
	OutPoint p2p.OutPoint
	GovData  gov.GovObject
}

// Encode serializes a GovRow to the passed writer.
func (gr *GovernanceProposal) Serialize(w io.Writer) error {
	err := gr.OutPoint.Encode(w)
	if err != nil {
		return err
	}
	err = gr.GovData.Serialize(w)
	if err != nil {
		return err
	}
	return nil
}

// Decode deserialized a GovRow from the passed reader.
func (gr *GovernanceProposal) Deserialize(r io.Reader) error {
	err := gr.OutPoint.Decode(r)
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

type Utxo struct {
	OutPoint          p2p.OutPoint
	PrevInputsPubKeys [][48]byte
	Owner             string
	Amount            int64
}

// Encode serializes the UtxoRow to a writer.
func (l *Utxo) Serialize(w io.Writer) error {
	err := l.OutPoint.Encode(w)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, l.Owner)
	if err != nil {
		return err
	}
	err = serializer.WriteVarInt(w, uint64(len(l.PrevInputsPubKeys)))
	if err != nil {
		return err
	}
	for _, pub := range l.PrevInputsPubKeys {
		err = serializer.WriteElements(w, pub)
		if err != nil {
			return err
		}
	}
	err = serializer.WriteVarInt(w, uint64(l.Amount))
	if err != nil {
		return err
	}
	return nil
}

// Decode deserializes a UtxoRow from a reader.
func (l *Utxo) Deserialize(r io.Reader) error {
	err := l.OutPoint.Decode(r)
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
	l.PrevInputsPubKeys = make([][48]byte, 0, count)
	for i := uint64(0); i < count; i++ {
		var pubKey [48]byte
		err = serializer.ReadElement(r, &pubKey)
		if err != nil {
			return err
		}
		l.PrevInputsPubKeys = append(l.PrevInputsPubKeys, pubKey)
	}
	amount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	l.Amount = int64(amount)
	return nil
}

func (l *Utxo) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = l.OutPoint.Encode(buf)
	return chainhash.DoubleHashH(buf.Bytes())
}

type UtxoState struct {
	UTXOs map[chainhash.Hash]Utxo
}

// Have checks if a UTXO exists.
func (u *UtxoState) Have(c chainhash.Hash) bool {
	_, found := u.UTXOs[c]
	return found
}

// Get gets the UTXO from state.
func (u *UtxoState) Get(c chainhash.Hash) Utxo {
	return u.UTXOs[c]
}

func (u *UtxoState) Serialize(w io.Writer) error {
	if err := serializer.WriteVarInt(w, uint64(len(u.UTXOs))); err != nil {
		return err
	}

	for h, utxo := range u.UTXOs {
		if _, err := w.Write(h[:]); err != nil {
			return err
		}

		if err := utxo.Serialize(w); err != nil {
			return err
		}
	}

	return nil
}

func (u *UtxoState) Deserialize(r io.Reader) error {
	if u.UTXOs == nil {
		u.UTXOs = make(map[chainhash.Hash]Utxo)
	}

	numUtxos, err := serializer.ReadVarInt(r)

	if err != nil {
		return err
	}

	for i := uint64(0); i < numUtxos; i++ {
		var hash chainhash.Hash
		if _, err := r.Read(hash[:]); err != nil {
			return err
		}

		var utxo Utxo
		if err := utxo.Deserialize(r); err != nil {
			return err
		}

		u.UTXOs[hash] = utxo
	}

	return nil
}

func (s *State) TransitionBlock(block *primitives.Block) (State, error) {
	return *s, nil
}

type User struct {
	OutPoint p2p.OutPoint
	UserData users.User
}

// Encode serializes the UserRow to a writer.
func (ur *User) Serialize(w io.Writer) error {
	err := ur.OutPoint.Encode(w)
	if err != nil {
		return err
	}
	err = ur.UserData.Serialize(w)
	if err != nil {
		return err
	}
	return nil
}

// Decode deserializes a user from the writer.
func (ur *User) Deserialize(r io.Reader) error {
	err := ur.OutPoint.Decode(r)
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

func (u *UserState) Serialize(w io.Writer) error {
	if err := serializer.WriteVarInt(w, uint64(len(u.Users))); err != nil {
		return err
	}

	for h, user := range u.Users {
		if _, err := w.Write(h[:]); err != nil {
			return err
		}

		if err := user.Serialize(w); err != nil {
			return err
		}
	}

	return nil
}

func (u *UserState) Deserialize(r io.Reader) error {
	if u.Users == nil {
		u.Users = make(map[chainhash.Hash]User)
	}

	numUsers, err := serializer.ReadVarInt(r)

	if err != nil {
		return err
	}

	for i := uint64(0); i < numUsers; i++ {
		var hash chainhash.Hash
		if _, err := r.Read(hash[:]); err != nil {
			return err
		}

		var user User
		if err := user.Deserialize(r); err != nil {
			return err
		}

		u.Users[hash] = user
	}

	return nil
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
	Outpoint     p2p.OutPoint
	PubKey       [48]byte
	PayeeAddress string
}

func (wk *Worker) ID() *chainhash.Hash {
	return serializer.Hash(&wk.Outpoint)
}

func (wk *Worker) Serialize(w io.Writer) error {
	err := wk.Outpoint.Encode(w)
	if err != nil {
		return err
	}
	err = serializer.WriteElements(w, wk.PubKey)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, wk.PayeeAddress)
	if err != nil {
		return err
	}
	return nil
}

func (wk *Worker) Deserialize(r io.Reader) error {
	err := wk.Outpoint.Decode(r)
	if err != nil {
		return err
	}
	err = serializer.ReadElements(r, &wk.PubKey)
	if err != nil {
		return err
	}
	wk.PayeeAddress, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	return nil
}

type WorkerRegistry struct {
	Workers map[chainhash.Hash]Worker
}

func NewWorkerRegistry() *WorkerRegistry {
	return &WorkerRegistry{
		Workers: map[chainhash.Hash]Worker{},
	}
}

// Have checks if a Worker exists.
func (w *WorkerRegistry) Have(c chainhash.Hash) bool {
	_, found := w.Workers[c]
	return found
}

// Get gets a Worker from state.
func (w *WorkerRegistry) Get(c chainhash.Hash) (Worker, bool) {
	wor, found := w.Workers[c]
	return wor, found
}

// Add adds a worker to the registry.
func (w *WorkerRegistry) Add(worker Worker) {
	h := worker.ID()
	w.Workers[*h] = worker
}

func (w *WorkerRegistry) Serialize(wr io.Writer) error {
	if err := serializer.WriteVarInt(wr, uint64(len(w.Workers))); err != nil {
		return err
	}

	for h, utxo := range w.Workers {
		if _, err := wr.Write(h[:]); err != nil {
			return err
		}

		if err := utxo.Serialize(wr); err != nil {
			return err
		}
	}

	return nil
}

func (w *WorkerRegistry) Deserialize(r io.Reader) error {
	if w.Workers == nil {
		w.Workers = make(map[chainhash.Hash]Worker)
	}

	numWorkers, err := serializer.ReadVarInt(r)

	if err != nil {
		return err
	}

	for i := uint64(0); i < numWorkers; i++ {
		var hash chainhash.Hash
		if _, err := r.Read(hash[:]); err != nil {
			return err
		}

		var worker Worker
		if err := worker.Deserialize(r); err != nil {
			return err
		}

		w.Workers[hash] = worker
	}

	return nil
}
