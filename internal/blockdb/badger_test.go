package blockdb_test

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func init() {
	config.SetTestParams()
	config.SetTestFlags()
}

func TestBlockDB_Instance(t *testing.T) {
	err := os.Mkdir(testdata.Node1Folder, 0777)
	assert.NoError(t, err)

	db, err := blockdb.NewBadgerDB()

	assert.NoError(t, err)
	assert.NotNil(t, db)

	// testing the database by Saving a time and then retrieving it
	testTime := time.Now()
	err = db.SetGenesisTime(testTime)
	assert.NoError(t, err)
	var savedTime time.Time
	savedTime, err = db.GetGenesisTime()
	assert.NoError(t, err)
	assert.Equal(t, testTime.Unix(), savedTime.Unix())

	_ = os.RemoveAll(testdata.Node1Folder)
}
