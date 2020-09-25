package blockdb_test

import (
	"github.com/golang/mock/gomock"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/logger"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

// params are the params used on the test
var param = &testdata.TestParams

func TestBlockDB_Instance(t *testing.T) {
	err := os.Mkdir(testdata.Node1Folder, 0777)
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	log := logger.NewMockLogger(ctrl)
	log.EXPECT().Info(gomock.Any()).AnyTimes()

	db, err := blockdb.NewBadgerDB(testdata.Node1Folder, *param, log)

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
