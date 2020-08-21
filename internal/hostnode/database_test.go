package hostnode_test

//
//import (
//	"github.com/libp2p/go-libp2p-core/peer"
//	"github.com/multiformats/go-multiaddr"
//	"github.com/olympus-protocol/ogen/internal/hostnode"
//	"github.com/stretchr/testify/assert"
//	"go.etcd.io/bbolt"
//	"path"
//	"path/filepath"
//	"testing"
//)
//
//const banLimit = 100
//
//var bansDbKey = []byte("bans")
//var ipDbKey = []byte("ips")
//
//func TestInitDb(t *testing.T) {
//	pathDir, _ := filepath.Abs("")
//	db, err := hostnode.NewDatabase(pathDir)
//	if err == nil {
//		err := db.Initialize()
//		assert.Nil(t, err)
//	}
//}
//
//func TestGetPrivKey(t *testing.T) {
//	pathDir, _ := filepath.Abs("")
//	netDB, err := bbolt.Open(path.Join(pathDir, "net.db"), 0600, nil)
//	if err == nil {
//		_, err := hostnode.InitBuckets(netDB)
//		assert.Nil(t, err)
//		priv, err := hostnode.GetPrivKey(netDB)
//		assert.Nil(t, err)
//		assert.NotNil(t, priv)
//		defer netDB.Close()
//	}
//}
//
//func TestSavePeers(t *testing.T) {
//	pathDir, _ := filepath.Abs("")
//	netDB, err := bbolt.Open(path.Join(pathDir, "net.db"), 0600, nil)
//	if err == nil {
//		_, err := hostnode.InitBuckets(netDB)
//		assert.Nil(t, err)
//
//		peer1, err := multiaddr.NewMultiaddr("/ip4/10.0.2.15/tcp/25000/p2p/12D3KooWPzn8FgE4hbvmTvwdDRCZ2zz69mumw17fsPquPscjTWPS")
//		assert.Nil(t, err)
//		peer2, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/24126/p2p/12D3KooWCnt52MYKVLn6fhKCoKy6HsNejEtxUt9MUwcpj1LYU2N1")
//		assert.Nil(t, err)
//
//		err = hostnode.SavePeer(netDB, peer1)
//		assert.Nil(t, err)
//		err = hostnode.SavePeer(netDB, peer2)
//		assert.Nil(t, err)
//
//		savedAddresses, err := hostnode.GetSavedPeers(netDB)
//		assert.Nil(t, err)
//		assert.Equal(t, len(savedAddresses), 2)
//
//		defer netDB.Close()
//	}
//}
//
//func TestBanPeers(t *testing.T) {
//	pathDir, _ := filepath.Abs("")
//	netDB, err := bbolt.Open(path.Join(pathDir, "net.db"), 0600, nil)
//	if err == nil {
//		_, err := hostnode.InitBuckets(netDB)
//		assert.Nil(t, err)
//
//		peer1, err := multiaddr.NewMultiaddr("/ip4/10.0.2.15/tcp/25000/p2p/12D3KooWPzn8FgE4hbvmTvwdDRCZ2zz69mumw17fsPquPscjTWPS")
//		assert.Nil(t, err)
//
//		err = hostnode.SavePeer(netDB, peer1)
//		assert.Nil(t, err)
//
//		peerId, err := peer.AddrInfoFromP2pAddr(peer1)
//		assert.Nil(t, err)
//		assert.NotNil(t, peerId)
//
//		for i := 0; i < 5; i++ {
//			_, err = hostnode.BanscorePeer(netDB, peerId.ID, banLimit/5)
//			assert.NoError(t, err)
//		}
//
//		isBanned, err := hostnode.IsPeerBanned(netDB, peerId.ID)
//
//		assert.NoError(t, err)
//
//		assert.Equal(t, isBanned, true)
//
//		// add same peer, and different peer with same ip address. In both cases it should not work, since it's an ip ban
//		err = hostnode.SavePeer(netDB, peer1)
//		assert.NoError(t, err)
//
//		peer2, err := multiaddr.NewMultiaddr("/ip4/10.0.2.15/tcp/25000/p2p/12D3KooWCnt52MYKVLn6fhKCoKy6HsNejEtxUt9MUwcpj1LYU2N1")
//		assert.NoError(t, err)
//		err = hostnode.SavePeer(netDB, peer2)
//		assert.NoError(t, err)
//
//		//remove ip from banList, it's just a test
//		peerByte, err := peerId.ID.MarshalBinary()
//		assert.NoError(t, err)
//		err = netDB.Update(func(tx *bbolt.Tx) error {
//			var err error
//			b := tx.Bucket(bansDbKey)
//			ipb := tx.Bucket(ipDbKey)
//			ipPeer := ipb.Get(peerByte)
//			err = b.Delete(ipPeer)
//			assert.Nil(t, err)
//			return nil
//		})
//		assert.NoError(t, err)
//		defer netDB.Close()
//	}
//}
