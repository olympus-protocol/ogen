//+build backup_test

package backup_test

import (
	"context"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/config"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/server"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/olympus-protocol/ogen/utils/logger"
)

// Backup Validator Test
// 1. Create one node that is creating a chain.
// 2. The second node should share the same keystore than the initial node but shouldn't vote or propose blocks.
// 3. The first node should stop proposing blocks and the second node should take the place of the first node.
func TestMain(m *testing.M) {
	firstNode()
}

func firstNode() {
	// Create datafolder
	os.Mkdir(testdata.Node1Folder, 0777)
	log := logger.New(os.Stdin)
	log.WithDebug()
	addr, err := testdata.PremineAddr.PublicKey().ToAddress(testdata.IntTestParams.AddrPrefix.Public)
	if err != nil {
		log.Fatal(err)
	}
	// Create a keystore
	log.Info("Creating keystore")
	keystore, err := keystore.NewKeystore(testdata.Node1Folder, log, testdata.KeystorePass)
	if err != nil {
		log.Fatal(err)
	}
	valData, err := keystore.GenerateNewValidatorKey(128, testdata.KeystorePass)
	if err != nil {
		log.Fatal(err)
	}
	validators := []primitives.ValidatorInitialization{}
	for _, vk := range valData {
		val := primitives.ValidatorInitialization{
			PubKey:       hex.EncodeToString(vk.PublicKey().Marshal()),
			PayeeAddress: addr,
		}
		validators = append(validators, val)
	}
	// Create the initialization parameters
	ip := primitives.InitializationParameters{
		GenesisTime:       time.Unix(time.Now().Unix()+10, 0),
		PremineAddress:    addr,
		InitialValidators: validators,
	}
	// Load the block database
	bdb, err := bdb.NewBlockDB(testdata.Node1Folder, testdata.IntTestParams, log)
	if err != nil {
		log.Fatal(err)
	}

	// Create the server instance
	ctx, cancel := context.WithCancel(context.Background())
	config.InterruptListener(log, cancel)
	c := testdata.Conf
	c.DataFolder = testdata.Node1Folder
	s, err := server.NewServer(ctx, &c, log, testdata.IntTestParams, bdb, ip)
	if err != nil {
		log.Fatal(err)
	}
	go s.Start()
	s.Proposer.OpenKeystore(testdata.KeystorePass)
	s.Proposer.Start()
	<-ctx.Done()
	bdb.Close()
	err = s.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

func secondNode() {

}