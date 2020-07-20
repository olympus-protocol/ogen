package peers_test

import (
	"encoding/json"
	"fmt"
	"github.com/multiformats/go-multiaddr"
	"go.etcd.io/bbolt"
	"path"
	"path/filepath"
	"testing"
)

func TestReadSavedPeers(t *testing.T) {

	var configBucketKey = []byte("config")
	var peersDbKey = []byte("peers")
	//var banDbKey = []byte("bans")
	pathDir, _ := filepath.Abs("")
	netDB, err := bbolt.Open(path.Join(pathDir, "net.db"), 0600, nil)
	if err != nil {
		fmt.Println("add a valid path")
		return
	}
	//retrieve the saved addresses
	_ = netDB.Update(func(tx *bbolt.Tx) error {

		b := tx.Bucket(configBucketKey)
		if b == nil {
			b, err = tx.CreateBucketIfNotExists(configBucketKey)
			if err != nil {
				return nil
			}
		}
		b2 := b.Bucket(peersDbKey)
		if b2 == nil {
			b2, err = b.CreateBucketIfNotExists(peersDbKey)
			if err != nil {
				return nil
			}
		}
		_ = b2.ForEach(func(k, v []byte) error {
			addr, _ := multiaddr.NewMultiaddrBytes(k)
			s, _ := json.MarshalIndent(addr, "", "\t")
			fmt.Println(string(s))
			return nil
		})
		return nil
	})
	defer netDB.Close()

}
