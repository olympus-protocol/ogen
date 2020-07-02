package main

import (
	"context"
	"encoding/hex"
	"os"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/config"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/server"
	"github.com/olympus-protocol/ogen/utils/logger"

)

var premineAddr = bls.RandKey()

var conf = config.Config{
	DataFolder: folder,
	NetworkName: "integration tests net",
	AddNodes: []peer.AddrInfo{},
	MaxPeers: 10,
	Port: "24126",
	MiningEnabled: true,
	RPCProxy: false,
	RPCProxyPort: "8080",
	RPCPort: "24130",
	RPCWallet: false,
	Debug: true,
	LogFile: false,
	Pprof: true,
}

var folder = "./data"

var secondaryNodeFolder = "./data_secondary"

var thirdNodeFolder = "./data_third"

var hostMultiAddr peer.AddrInfo

var testParams = params.ChainParams{
	Name:           "testnet",
	DefaultP2PPort: "25126",
	AddrPrefix: params.AddrPrefixes{
		Public:   "tlpub",
		Private:  "tlprv",
		Multisig: "tlmul",
	},
	GovernanceBudgetQuotient:     5, // 20%
	BaseRewardPerBlock:           2600,
	IncluderRewardQuotient:       8,
	EpochLength:                  5,
	EjectionBalance:              80,
	MaxBalanceChurnQuotient:      32,
	MaxVotesPerBlock:             32,
	LatestBlockRootsLength:       64,
	MinAttestationInclusionDelay: 1,
	DepositAmount:                100,
	UnitsPerCoin:                 1000,
	InactivityPenaltyQuotient:    17179869184,
	SlotDuration:                 1,
	MaxTxsPerBlock:               1000,
	MaxDepositsPerBlock:          32,
	MaxExitsPerBlock:             32,
	MaxRANDAOSlashingsPerBlock:   20,
	MaxProposerSlashingsPerBlock: 2,
	MaxVoteSlashingsPerBlock:     10,
	WhistleblowerRewardQuotient:  2,
	GovernancePercentages: []uint8{
		30, // tech
		10, // community
		20, // business
		20, // marketing
		20, // adoption
	},
	MinVotingBalance:          100,
	CommunityOverrideQuotient: 3,
	VotingPeriodSlots:         20160, // minutes in a week
	InitialManagers: [][20]byte{
		{252, 94, 117, 132, 63, 93, 202, 26, 36, 23, 195, 26, 169, 95, 74, 147, 72, 184, 66, 20},        // tlpub1l308tpplth9p5fqhcvd2jh62jdytsss54nt6d4
		{192, 13, 158, 167, 115, 190, 56, 51, 43, 11, 156, 43, 27, 145, 143, 61, 40, 209, 114, 238},     // tlpub1cqxeafmnhcurx2ctns43hyv0855dzuhwnllx6w
		{88, 192, 115, 125, 142, 126, 244, 13, 253, 225, 139, 36, 184, 34, 71, 31, 69, 205, 216, 125},   // tlpub1trq8xlvw0m6qml0p3vjtsgj8razumkrawvwzza
		{143, 17, 152, 250, 184, 122, 141, 208, 109, 72, 148, 187, 248, 89, 83, 127, 113, 217, 23, 144}, // tlpub13uge374c02xaqm2gjjalsk2n0acaj9uswmr687
		{162, 207, 33, 52, 96, 81, 17, 131, 72, 175, 180, 222, 125, 41, 3, 108, 43, 47, 231, 7},         // tlpub15t8jzdrq2ygcxj90kn0862grds4jlec8tjcg6j
	},
}

var initializationParams primitives.InitializationParameters

func main() {
	// Create datafolder
	os.Mkdir(folder, 0777)
	logfile, err := os.Create(folder + "/log.log")
	if err != nil {
		panic(err)
	}
	// Create logger
	log := logger.New(logfile)
	log.WithDebug()

	// Create a keystore
	log.Info("Creating keystore")
	keystore, err := keystore.NewKeystore(folder, log)
	if err != nil {
		log.Fatal(err)
	}
	validatorKeys, err := keystore.GenerateNewValidatorKey(128)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Generated %v keys", len(validatorKeys))
	keystore.Close()
	addr, err := premineAddr.PublicKey().ToAddress(testParams.AddrPrefix.Public)
	if err != nil {
		log.Fatal(err)
	}
	validators := []primitives.ValidatorInitialization{}
	for _, vk := range validatorKeys {
		val := primitives.ValidatorInitialization{
			PubKey: hex.EncodeToString(vk.PublicKey().Marshal()),
			PayeeAddress: addr,
		}
		validators = append(validators, val)
	}
	// Create the initialization parameters
	initializationParams = primitives.InitializationParameters{
		GenesisTime: time.Now(),
		PremineAddress: addr,
		InitialValidators: validators,
	}
	// Load the block database
	bdb, err := bdb.NewBlockDB(folder, testParams, log)
	if err != nil {
		log.Fatal(err)
	}

	// Create the server instance
	ctx, cancel := context.WithCancel(context.Background())
	config.InterruptListener(log, cancel)

	testServer, err := server.NewServer(ctx, &conf, log, testParams, bdb, initializationParams)
	if err != nil {
		log.Fatal(err)
	}
	
	// Generate the chain up to 50 blocks
	go testServer.Start()
	hostMultiAddr.Addrs = testServer.HostNode.GetHost().Addrs()
	hostMultiAddr.ID = testServer.HostNode.GetHost().ID()
	// Run test frameworks
	go runSecondNode()
	go runThirdNode()

	<-ctx.Done()
	bdb.Close()
	err = testServer.Stop()
	if err != nil {
		log.Fatal(err)
	}
	os.RemoveAll(folder)
}


// The second node is in charge of testing a "good" peer behaviour.
func runSecondNode() {
	os.Mkdir(secondaryNodeFolder, 0777)
	logfile, err := os.Create(secondaryNodeFolder + "/log.log")
	if err != nil {
		panic(err)
	}
	log := logger.New(logfile)
	log.WithDebug()
	ctx, cancel := context.WithCancel(context.Background())
	config.InterruptListener(log, cancel)

	bdb, err := bdb.NewBlockDB(secondaryNodeFolder, testParams, log)
	if err != nil {
		log.Fatal(err)
	}
	secondNodeConf := &config.Config{
		DataFolder: secondaryNodeFolder,
		NetworkName: "integration tests net",
		AddNodes: []peer.AddrInfo{hostMultiAddr},
		MaxPeers: 10,
		Port: "24000",
		MiningEnabled: false,
		RPCProxy: false,
		RPCProxyPort: "8080",
		RPCPort: "24001",
		RPCWallet: false,
		Debug: false,
		LogFile: false,
		Pprof: false,
	}
	testServer, err := server.NewServer(ctx, secondNodeConf, log, testParams, bdb, initializationParams)
	if err != nil {
		log.Fatal(err)
	}
	go testServer.Start()
	<-ctx.Done()
	bdb.Close()
	err = testServer.Stop()
	if err != nil {
		log.Fatal(err)
	}
	os.RemoveAll(secondaryNodeFolder)
}

// The third node is in charge of testing a "bad" peer behaviour.
func runThirdNode() {
	os.Mkdir(thirdNodeFolder, 0777)
	logfile, err := os.Create(thirdNodeFolder + "/log.log")
	if err != nil {
		panic(err)
	}
	log := logger.New(logfile)
	log.WithDebug()
	ctx, cancel := context.WithCancel(context.Background())
	config.InterruptListener(log, cancel)
	bdb, err := bdb.NewBlockDB(thirdNodeFolder, testParams, log)
	if err != nil {
		log.Fatal(err)
	}
	conf := config.Config{
		DataFolder: thirdNodeFolder,
		NetworkName: "integration tests net",
		AddNodes: []peer.AddrInfo{hostMultiAddr},
		MaxPeers: 10,
		Port: "25000",
		MiningEnabled: false,
		RPCProxy: false,
		RPCProxyPort: "8080",
		RPCPort: "25001",
		RPCWallet: false,
		Debug: false,
		LogFile: false,
		Pprof: false,
	}
	testServer, err := server.NewServer(ctx, &conf, log, testParams, bdb, initializationParams)
	if err != nil {
		log.Fatal(err)
	}
	go testServer.Start()
	<-ctx.Done()
	bdb.Close()
	err = testServer.Stop()
	if err != nil {
		log.Fatal(err)
	}
	os.RemoveAll(thirdNodeFolder)
}