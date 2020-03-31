package primitives

import (
	"bytes"
	"testing"

	"github.com/go-test/deep"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

func TestUserCopy(t *testing.T) {
	user := User{
		OutPoint: OutPoint{
			TxHash: chainhash.Hash{1},
			Index:  2,
		},
		PubKey: [48]byte{3},
		Name:   "test!",
	}
	user2 := user.Copy()

	user.OutPoint.TxHash[0] = 2
	if user2.OutPoint.TxHash[0] == 2 {
		t.Fatal("mutating outpoint mutates copy")
	}

	user.PubKey[0] = 1
	if user2.PubKey[0] == 1 {
		t.Fatal("mutating pubkey mutates copy")
	}

	user.Name = "test2"
	if user2.Name == "test2" {
		t.Fatal("mutating name mutates copy")
	}
}

func TestUserDeserializeSerialize(t *testing.T) {
	user := User{
		OutPoint: OutPoint{
			TxHash: chainhash.Hash{1},
			Index:  2,
		},
		PubKey: [48]byte{3},
		Name:   "test!",
	}

	buf := bytes.NewBuffer([]byte{})
	err := user.Serialize(buf)
	if err != nil {
		t.Fatal(err)
	}

	var user2 User
	if err := user2.Deserialize(buf); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(user2, user); diff != nil {
		t.Fatal(diff)
	}
}

func TestUserStateCopy(t *testing.T) {
	userState := UserState{
		Users: map[chainhash.Hash]User{
			chainhash.Hash{14}: {
				OutPoint: OutPoint{
					TxHash: chainhash.Hash{1},
					Index:  2,
				},
				PubKey: [48]byte{3},
				Name:   "4",
			},
		},
	}

	userState2 := userState.Copy()

	userState.Users[chainhash.Hash{14}] = User{
		Name: "5",
	}

	if userState2.Users[chainhash.Hash{14}].Name == "5" {
		t.Fatal("mutating users mutates copy")
	}
}

func TestUserStateDeserializeSerialize(t *testing.T) {
	userState := UserState{
		Users: map[chainhash.Hash]User{
			chainhash.Hash{14}: {
				OutPoint: OutPoint{
					TxHash: chainhash.Hash{1},
					Index:  2,
				},
				PubKey: [48]byte{3},
				Name:   "4",
			},
		},
	}

	buf := bytes.NewBuffer([]byte{})
	err := userState.Serialize(buf)
	if err != nil {
		t.Fatal(err)
	}

	var userState2 UserState
	if err := userState2.Deserialize(buf); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(userState2, userState); diff != nil {
		t.Fatal(diff)
	}
}
