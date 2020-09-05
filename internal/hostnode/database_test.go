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

func init() {
	_ = os.Remove("./test")
	_ = os.MkdirAll("./test", 0777)
}

func TestDatabase(t *testing.T) {

	mockNet := mocknet.New(context.Background())

	ctrl := gomock.NewController(t)

	hn := hostnode.NewMockHostNode(ctrl)

	pathDir, _ := filepath.Abs("./test")

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

	p1, err := mockNet.GenPeer()
	assert.NoError(t, err)

	//p2, err := mockNet.GenPeer()
	//assert.NoError(t, err)

	pinfo1 := &peer.AddrInfo{
		ID:    p1.ID(),
		Addrs: p1.Addrs(),
	}

	//pinfo2 := &peer.AddrInfo{
	//	ID:    p2.ID(),
	//	Addrs: p2.Addrs(),
	//}

	err = db.SavePeer(pinfo1)
	assert.NoError(t, err)

	//err = db.SavePeer(pinfo2)
	//assert.NoError(t, err)

	//peers, err = db.GetSavedPeers()
	//assert.NoError(t, err)
	//assert.Equal(t, []*peer.AddrInfo{pinfo1, pinfo2}, peers)

}
