package hostnode_test

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var db hostnode.Database
var peerId peer.ID

func init() {
	_ = os.Mkdir(testdata.Node1Folder, 0777)
	db, _ = hostnode.NewDatabase(testdata.Node1Folder)
}

func TestDatabase_Initialize(t *testing.T) {
	err := db.Initialize()
	assert.NoError(t, err)
}

func TestDatabase_SavePeer(t *testing.T) {
	peer1, err := multiaddr.NewMultiaddr("/ip4/10.0.2.15/tcp/25000/p2p/12D3KooWPzn8FgE4hbvmTvwdDRCZ2zz69mumw17fsPquPscjTWPS")
	assert.NoError(t, err)
	err = db.SavePeer(peer1)
	assert.NoError(t, err)
}

func TestDatabase_GetPrivKey(t *testing.T) {
	priv, err := db.GetPrivKey()
	assert.NoError(t, err)
	assert.NotNil(t, priv)
}

func TestDatabase_BanPeers(t *testing.T) {
	// add a peer to ban
	peer1, err := multiaddr.NewMultiaddr("/ip4/10.0.2.16/tcp/25000/p2p/12D3KooWCnt52MYKVLn6fhKCoKy6HsNejEtxUt9MUwcpj1LYU2N1")
	assert.NoError(t, err)
	err = db.SavePeer(peer1)
	assert.NoError(t, err)

	peerInfo, err := peer.AddrInfoFromP2pAddr(peer1)
	assert.NoError(t, err)
	assert.NotNil(t, peerInfo)

	for i := 0; i < 5; i++ {
		_, err := db.BanscorePeer(peerInfo.ID, 100/5)
		assert.NoError(t, err)
	}
	peerId = peerInfo.ID
}

func TestDatabase_IsPeerBanned(t *testing.T) {
	// requires the BanPeers test to pass
	isBanned, err := db.IsPeerBanned(peerId)
	assert.NoError(t, err)
	assert.Equal(t, true, isBanned)
}

func TestDatabase_IsIPBanned(t *testing.T) {
	isBanned, _, err := db.IsIPBanned("ip4/10.0.2.16")
	assert.NoError(t, err)
	assert.Equal(t, true, isBanned)
}

func TestDatabase_GetSavedPeers(t *testing.T) {
	// added 1 correct peer and banned another one in previous tests, so it should have 1 element saved
	savedAddresses, err := db.GetSavedPeers()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(savedAddresses))
	cleanFolder1()
}

func cleanFolder1() {
	_ = os.RemoveAll(testdata.Node1Folder)
}
