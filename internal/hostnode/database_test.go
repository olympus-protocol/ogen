package hostnode_test

import (
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestDatabase(t *testing.T) {

	pathDir, _ := filepath.Abs("./test")
	db, err := hostnode.NewDatabase(pathDir)
	assert.NoError(t, err)
	err = db.Initialize()
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

	assert.Equal(t, []multiaddr.Multiaddr(nil), peers)

}
