package burnproof

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"io"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

type CoinsProofSerializable struct {
	MerkleIndex   uint64
	MerkleBranch  [][32]byte `ssz-max:"64"`
	PkScript      [25]byte
	Transaction   [192]byte
	RedeemAccount [44]byte
}

func (c *CoinsProofSerializable) Marshal() ([]byte, error) {
	return c.MarshalSSZ()
}

func (c *CoinsProofSerializable) Unmarshal(b []byte) error {
	return c.UnmarshalSSZ(b)
}

func (c *CoinsProofSerializable) Hash() chainhash.Hash {
	b, _ := c.Marshal()
	return chainhash.HashH(b)
}

func (c *CoinsProofSerializable) RedeemAccountHash() ([20]byte, error) {
	accStr := string(c.RedeemAccount[:])
	_, dec, err := bech32.Decode(accStr)
	if err != nil {
		return [20]byte{}, err
	}
	if len(dec) != 20 {
		return [20]byte{}, errors.New("invalid account hash")
	}

	var acc [20]byte
	copy(acc[:], dec)

	return acc, nil
}

func (c *CoinsProofSerializable) ToCoinProof() (*CoinsProof, error) {
	merkle := make([]chainhash.Hash, len(c.MerkleBranch))

	for i := range c.MerkleBranch {
		merkle[i] = c.MerkleBranch[i]
	}

	tx := wire.MsgTx{}

	buf := bytes.NewBuffer(c.Transaction[:])

	err := tx.Deserialize(buf)
	if err != nil {
		return nil, err
	}

	cp := &CoinsProof{
		MerkleIndex:  c.MerkleIndex,
		MerkleBranch: merkle,
		PkScript:     c.PkScript,
		Transaction:  tx,
	}

	return cp, nil
}

// CoinsProof is a proof of coins on the old blockchain.
type CoinsProof struct {
	MerkleIndex  uint64
	MerkleBranch []chainhash.Hash
	PkScript     [25]byte
	Transaction  wire.MsgTx
}

func (c *CoinsProof) ToSerializable(acc [44]byte) (*CoinsProofSerializable, error) {
	merkles := make([][32]byte, len(c.MerkleBranch))

	for i := range merkles {
		merkles[i] = c.MerkleBranch[i]
	}

	buf := bytes.NewBuffer([]byte{})

	err := c.Transaction.Serialize(buf)
	if err != nil {
		return nil, err
	}

	var tx [192]byte
	copy(tx[:], buf.Bytes())

	cps := &CoinsProofSerializable{
		MerkleIndex:   c.MerkleIndex,
		MerkleBranch:  merkles,
		PkScript:      c.PkScript,
		Transaction:   tx,
		RedeemAccount: acc,
	}

	return cps, nil
}

// Unmarshal decodes the proof from a byte slice.
func (c *CoinsProof) Unmarshal(r io.Reader) error {
	indexBytes := make([]byte, 4)
	if _, err := io.ReadFull(r, indexBytes); err != nil {
		return err
	}

	c.MerkleIndex = uint64(binary.LittleEndian.Uint32(indexBytes))

	merkleCount, err := wire.ReadVarInt(r, 0)
	if err != nil {
		return err
	}

	c.MerkleBranch = make([]chainhash.Hash, merkleCount)
	for i := range c.MerkleBranch {
		_, err := io.ReadFull(r, c.MerkleBranch[i][:])
		if err != nil {
			return err
		}
	}

	if err := c.Transaction.BtcDecode(r, 0, wire.BaseEncoding); err != nil {
		return err
	}

	script, err := wire.ReadVarBytes(r, 0, 10000, "script")
	if err != nil {
		return err
	}

	copy(c.PkScript[:], script)

	return nil
}

var cache = txscript.NewSigCache(1000)

// CalcMerkleRoot calculates the merkle root under some assumptions.
func (c *CoinsProof) CalcMerkleRoot(leafHash chainhash.Hash) chainhash.Hash {
	index := c.MerkleIndex
	for _, h := range c.MerkleBranch {
		if index&1 > 0 {
			leafHash = chainhash.DoubleHashH(append(h[:], leafHash[:]...))
		} else {
			leafHash = chainhash.DoubleHashH(append(leafHash[:], h[:]...))
		}
		index >>= 1
	}

	return leafHash
}

func calcLeafHash(proof *CoinsProof) (*chainhash.Hash, error) {
	if len(proof.Transaction.TxOut) != 1 {
		return nil, fmt.Errorf("expected transaction to have 1 output, but got %d", len(proof.Transaction.TxOut))
	}

	if len(proof.Transaction.TxIn) != 1 {
		return nil, fmt.Errorf("expected transaction to have 1 output, but got %d", len(proof.Transaction.TxOut))
	}

	leafHashBytes := bytes.NewBuffer([]byte{})
	outpoint := proof.Transaction.TxIn[0].PreviousOutPoint

	_, err := leafHashBytes.Write(outpoint.Hash[:])
	if err != nil {
		return nil, err
	}

	indexBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(indexBytes, outpoint.Index)

	_, err = leafHashBytes.Write(indexBytes)
	if err != nil {
		return nil, err
	}

	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, uint64(proof.Transaction.TxOut[0].Value))

	_, err = leafHashBytes.Write(valueBytes)
	if err != nil {
		return nil, err
	}

	if err := wire.WriteVarBytes(leafHashBytes, 0, proof.PkScript[:]); err != nil {
		return nil, err
	}

	h := chainhash.DoubleHashH(leafHashBytes.Bytes())
	return &h, nil
}

func verifyMerkleRoot(merkleRoot chainhash.Hash, proof *CoinsProof) error {
	out, err := calcLeafHash(proof)
	if err != nil {
		return err
	}

	calculatedRoot := proof.CalcMerkleRoot(*out)

	if !calculatedRoot.IsEqual(&merkleRoot) {
		return fmt.Errorf("calculated merkle root %s, but expected %s", calculatedRoot, merkleRoot)
	}

	return nil
}

func verifyScript(proof *CoinsProof) error {
	if len(proof.Transaction.TxOut) != 1 {
		return fmt.Errorf("expected transaction to have 1 output, but got %d", len(proof.Transaction.TxOut))
	}

	eng, err := txscript.NewEngine(proof.PkScript[:], &proof.Transaction, 0,
		txscript.StandardVerifyFlags, cache,
		txscript.NewTxSigHashes(&proof.Transaction), proof.Transaction.TxOut[0].Value)
	if err != nil {
		return err
	}

	if err := eng.Execute(); err != nil {
		return err
	}

	return nil
}

func verifyPkhMatchesAddress(script []byte, address []byte) error {
	if len(script) != 25 {
		return fmt.Errorf("expected transaction pkscript to be 25, but got %d", len(script))
	}

	addrHash := script[3:23]

	addrBuf := new(bytes.Buffer)

	if err := wire.WriteVarInt(addrBuf, 0, uint64(len(address))); err != nil {
		return err
	}

	if _, err := addrBuf.Write(address[:]); err != nil {
		return err
	}

	expectedAddrHash := chainhash.DoubleHashB(addrBuf.Bytes())
	if !bytes.Equal(addrHash, expectedAddrHash[:20]) {
		return fmt.Errorf("expected addresses to match (expected: %x, got: %x)", expectedAddrHash, addrHash)
	}

	return nil
}

// VerifyBurn verifies a burn proof.
func VerifyBurn(proofBytes []byte, address []byte, merkleRoot chainhash.Hash) error {
	var proofs []*CoinsProof

	buf := bytes.NewBuffer(proofBytes)

	for {
		c := new(CoinsProof)
		err := c.Unmarshal(buf)
		if err != nil {
			return err
		}

		proofs = append(proofs, c)
		if buf.Len() <= 0 {
			break
		}
	}

	for _, c := range proofs {
		err := VerifyBurnProof(c, address, merkleRoot)
		if err != nil {
			return err
		}
	}

	return nil
}

// VerifyBurnProof verifies a single burn proof.
func VerifyBurnProof(p *CoinsProof, address []byte, merkleRoot chainhash.Hash) error {
	if err := verifyMerkleRoot(merkleRoot, p); err != nil {
		return err
	}

	if err := verifyScript(p); err != nil {
		return err
	}

	if err := verifyPkhMatchesAddress(p.Transaction.TxOut[0].PkScript, address); err != nil {
		return err
	}

	return nil
}
