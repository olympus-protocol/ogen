package primitives

import (
	"bytes"
	"testing"

	"github.com/go-test/deep"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

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
