package hdwallets_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/hdwallets"
)

type XORShift struct {
	state uint64
}

func NewXORShift(state uint64) *XORShift {
	return &XORShift{state}
}

func (xor *XORShift) Read(b []byte) (int, error) {
	for i := range b {
		x := xor.state
		x ^= x << 13
		x ^= x >> 7
		x ^= x << 17
		b[i] = uint8(x)
		xor.state = x
	}
	return len(b), nil
}

var polisNetPrefix = &hdwallets.NetPrefix{
	ExtPub:  []byte{0x1f, 0x74, 0x90, 0xf0},
	ExtPriv: []byte{0x11, 0x24, 0xd9, 0x70},
}

func TestExtendedPrivateKey(t *testing.T) {
	// Checks to make sure HD.child(10).toPub() == HD.toPub().child(10)

	x := NewXORShift(200)

	var key [64]byte
	x.Read(key[:])
	esk, err := hdwallets.NewMaster(key[:], polisNetPrefix)

	if err != nil {
		t.Fatal(err)
	}

	child10, err := esk.Child(10)
	if err != nil {
		t.Fatal(err)
	}

	epk, err := esk.Neuter(polisNetPrefix)
	if err != nil {
		t.Fatal(err)
	}

	childXPub, err := epk.Child(10)
	if err != nil {
		t.Fatal(err)
	}

	childPub, err := childXPub.BlsPubKey()
	if err != nil {
		t.Fatal(err)
	}

	priv, err := child10.BlsPrivKey()
	if err != nil {
		t.Fatal(err)
	}
	expectedPub := priv.PublicKey()

	if !bytes.Equal(expectedPub.Marshal(), childPub.Marshal()) {
		t.Fatalf("expected child priv key to match child pub key")
	}
}

func TestExtended(t *testing.T) {
	// Checks to make sure HD.child(10).toPub() == HD.toPub().child(10)
	x := NewXORShift(200)

	var key [64]byte
	x.Read(key[:])
	esk, err := hdwallets.NewMaster(key[:], polisNetPrefix)

	if err != nil {
		t.Fatal(err)
	}

	child10, err := esk.Child(10)
	if err != nil {
		t.Fatal(err)
	}

	epk, err := esk.Neuter(polisNetPrefix)
	if err != nil {
		t.Fatal(err)
	}

	childXPub, err := epk.Child(10)
	if err != nil {
		t.Fatal(err)
	}

	childPub, err := childXPub.BlsPubKey()
	if err != nil {
		t.Fatal(err)
	}

	priv, err := child10.BlsPrivKey()
	if err != nil {
		t.Fatal(err)
	}
	expectedPub := priv.PublicKey()

	if !bytes.Equal(expectedPub.Marshal(), childPub.Marshal()) {
		t.Fatalf("expected child priv key to match child pub key")
	}
}

func TestBasicProperties(t *testing.T) {
	x := NewXORShift(200)

	var key [64]byte
	x.Read(key[:])
	esk, err := hdwallets.NewMaster(key[:], polisNetPrefix)

	if err != nil {
		t.Fatal(err)
	}

	epk, err := esk.Neuter(polisNetPrefix)
	if err != nil {
		t.Fatal(err)
	}

	if !esk.IsPrivate() {
		t.Fatal("expected extended priv key to be private")
	}

	if epk.IsPrivate() {
		t.Fatal("expected extended public key to be public")
	}

	esk10, err := esk.Child(10)
	if err != nil {
		t.Fatal(err)
	}

	if esk10.Depth() != 1 {
		t.Fatal("expected derived key to be depth 1")
	}

	epk10, err := epk.Child(10)
	if err != nil {
		t.Fatal(err)
	}

	if epk10.Depth() != 1 {
		t.Fatal("expected derived key to be depth 1")
	}

	_, err = esk.Child(10 + 0x80000000)
	if err != nil {
		t.Fatal(err)
	}

	epkHardened, err := esk.Neuter(polisNetPrefix)
	if err != nil {
		t.Fatal(err)
	}

	_, err = epkHardened.Child(10 + 0x80000000)
	if err != hdwallets.ErrDeriveHardFromPublic {
		t.Fatal("incorrectly derived child from harded pubkey")
	}
}

func TestExtendedKeyToFromString(t *testing.T) {
	x := NewXORShift(200)

	var key [64]byte
	x.Read(key[:])
	esk, err := hdwallets.NewMaster(key[:], polisNetPrefix)

	if err != nil {
		t.Fatal(err)
	}

	epk, err := esk.Neuter(polisNetPrefix)
	if err != nil {
		t.Fatal(err)
	}

	eskStr := esk.String()

	eskFromStr, err := hdwallets.NewKeyFromString(eskStr)
	if err != nil {
		t.Fatal(err)
	}

	gotPriv, err := eskFromStr.BlsPrivKey()
	if err != nil {
		t.Fatal(err)
	}

	expectedPriv, err := esk.BlsPrivKey()
	if err != nil {
		t.Fatal(err)
	}

	gotPrivBytes := gotPriv.Marshal()
	expectedPrivBytes := expectedPriv.Marshal()

	if !bytes.Equal(gotPrivBytes[:], expectedPrivBytes[:]) {
		t.Fatal("expected extended keys to match after serializing/deserializing")
	}

	if eskFromStr.Depth() != esk.Depth() {
		t.Fatal("expected extended key depths to match after serializing/deserializing")
	}

	if eskFromStr.ParentFingerprint() != esk.ParentFingerprint() {
		t.Fatal("expected extended key parent FP to match after serializing/deserializing")
	}

	epkStr := epk.String()

	epkFromStr, err := hdwallets.NewKeyFromString(epkStr)
	if err != nil {
		t.Fatal(err)
	}

	gotPub, err := epkFromStr.BlsPubKey()
	if err != nil {
		t.Fatal(err)
	}

	expectedPub, err := epk.BlsPubKey()
	if err != nil {
		t.Fatal(err)
	}

	gotPubBytes := gotPub.Marshal()
	expectedPubBytes := expectedPub.Marshal()

	if !bytes.Equal(gotPubBytes[:], expectedPubBytes[:]) {
		t.Fatal("expected extended keys to match after serializing/deserializing")
	}

	if epkFromStr.Depth() != epk.Depth() {
		t.Fatal("expected extended key depths to match after serializing/deserializing")
	}

	if epkFromStr.ParentFingerprint() != epk.ParentFingerprint() {
		t.Fatal("expected extended key parent FP to match after serializing/deserializing")
	}

	if !strings.HasPrefix(epkStr, "ppub") {
		t.Fatal("expected public key to have prefix ppub")
	}

	if !strings.HasPrefix(eskStr, "pprv") {
		t.Fatal("expected secret key to have prefix pprv")
	}
}

const DeriveIterations = 1000

func TestDeriveKey(t *testing.T) {

	for i := 0; i < DeriveIterations; i++ {
		hash := chainhash.DoubleHashB([]byte(fmt.Sprintf("%d", i)))
		masterKey, err := hdwallets.NewMaster(hash, polisNetPrefix)
		if err != nil {
			t.Fatal(err)
		}
		_, err = masterKey.Child(44 + 0x80000000)
		if err != nil {
			t.Fatal(err)
		}
	}
}
