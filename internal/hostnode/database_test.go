package hostnode_test

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-core/peer"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestDatabase(t *testing.T) {
	mockNet := mocknet.New(context.Background())

	p1, err := mockNet.GenPeer()
	assert.NoError(t, err)

	pinfo1 := &peer.AddrInfo{
		ID:    p1.ID(),
		Addrs: p1.Addrs(),
	}

	ctrl := gomock.NewController(t)

	hn := hostnode.NewMockHostNode(ctrl)
	hn.EXPECT().DisconnectPeer(p1.ID()).Return(nil)

	pathDir, _ := filepath.Abs("./")

	db, err := hostnode.NewDatabase(pathDir, hn)
	assert.NoError(t, err)

	assert.NoError(t, err)
	priv1, err := db.GetPrivKey()
	assert.NoError(t, err)
	priv2, err := db.GetPrivKey()
	assert.NoError(t, err)

	// Priv1 and Priv2 should be the same, this means the db is generating a privkey only once
	assert.Equal(t, priv1, priv2)

	// Peers should be empty
	peers, err := db.GetSavedPeers()
	assert.NoError(t, err)
	assert.Equal(t, []*peer.AddrInfo(nil), peers)

	err = db.SavePeer(pinfo1)
	assert.NoError(t, err)

	peers, err = db.GetSavedPeers()
	assert.NoError(t, err)
	assert.Equal(t, []*peer.AddrInfo{pinfo1}, peers)

	err = db.BanscorePeer(pinfo1, 10)

	peers, err = db.GetSavedPeers()
	assert.NoError(t, err)
	assert.Equal(t, []*peer.AddrInfo{pinfo1}, peers)

	err = db.BanscorePeer(pinfo1, 90)

	peers, err = db.GetSavedPeers()
	assert.NoError(t, err)
	assert.Equal(t, []*peer.AddrInfo(nil), peers)

	err = db.SavePeer(pinfo1)
	assert.Equal(t, hostnode.ErrorPeerBanned, err)

	_ = os.Remove("./net.db")
}
