package peers_test

import (
	"encoding/json"
	"fmt"
	"github.com/multiformats/go-multiaddr"
	"go.etcd.io/bbolt"
	"path"
	"testing"
)

func TestReadSavedPeers(t *testing.T) {

	var configBucketKey = []byte("config")
	var peersDbKey = []byte("peers")
	netDB, err := bbolt.Open(path.Join("/home/waychin/.config/ogen", "net.db"), 0600, nil)
	if err != nil {
		t.Error("could not open db")
	}
	//retrieve the saved addresses
	_ = netDB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(configBucketKey).Bucket(peersDbKey)
		if b == nil {
			b, err = tx.Bucket(configBucketKey).CreateBucketIfNotExists(peersDbKey)
			if err != nil {
				return nil
			}
		}

		_ = b.ForEach(func(k, v []byte) error {
			addr, _ := multiaddr.NewMultiaddrBytes(k)
			s, _ := json.MarshalIndent(addr, "", "\t")
			fmt.Println(string(s))
			return nil
		})
		return nil
	})

}
