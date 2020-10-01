// +build chain_test

package chain_test

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/cmd/ogen/initialization"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/internal/server"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"math/rand"
	"os"
	"path"
	"strconv"
	"sync"
	"testing"
	"time"
)

const NumNodes = 6
const NumValidators = 10

var folders = make([]string, NumNodes)
var loggers = make([]logger.Logger, NumNodes)

var validatorsKeys = make(map[int][]*bls.SecretKey)
var initParams initialization.InitializationParameters

var keystores = make([]keystore.Keystore, NumNodes)
var servers = make([]server.Server, NumNodes)

var premineBytes, _ = hex.DecodeString("464725989655873131a985e94febf059523278c483d2b3e21434fd6bd3720537")
var premineAddr, _ = bls.SecretKeyFromBytes(premineBytes)

var receivingBytes, _ = hex.DecodeString("1bc06361dfd5a9cb4817b878d9921d340aac483813197653caca0ecdf7744b75")
var receivingAddr, _ = bls.SecretKeyFromBytes(receivingBytes)

var walletsPass = "wallet_secure_password"

var params = &testdata.TestParams
var delaySeconds int64 = 30

const dataFolder = "./chain_test"
const nodeFolderPrefix = "data_folder_"

func createTestEnvironment() {
	_ = os.RemoveAll(dataFolder)
	_ = os.MkdirAll(dataFolder, 0777)
	for i := range folders {
		strfolder := path.Join(dataFolder, nodeFolderPrefix+strconv.Itoa(i))
		folders[i] = strfolder
		_ = os.Mkdir(strfolder, 0777)
	}
}

func createKeystoresAndValidators() {

	// Create loggers and keystore instances
	var folderWg sync.WaitGroup
	folderWg.Add(len(folders))
	for i, folder := range folders {
		go func(index int, folder string, wg *sync.WaitGroup) {
			defer wg.Done()
			logPath := path.Join(folder, "logger.log")
			logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_RDWR, 0755)
			if err != nil {
				panic(err)
			}
			loggers[index] = logger.New(logFile)
			keystores[index] = keystore.NewKeystore(folder)
		}(i, folder, &folderWg)
	}
	folderWg.Wait()

	// Initialize each keystore with NumValidators
	var keystoreWg sync.WaitGroup
	keystoreWg.Add(len(keystores))
	var keysLock sync.RWMutex
	for i, ks := range keystores {
		go func(index int, keystore keystore.Keystore, wg *sync.WaitGroup, lock *sync.RWMutex) {
			lock.Lock()
			defer wg.Done()
			defer lock.Unlock()
			err := keystore.CreateKeystore()
			if err != nil {
				panic(err)
			}
			ksvalidators, err := keystore.GenerateNewValidatorKey(NumValidators)
			if err != nil {
				panic(err)
			}
			validatorsKeys[index] = ksvalidators
			err = keystore.Close()
			if err != nil {
				panic(err)
			}
		}(i, ks, &keystoreWg, &keysLock)
	}

	keystoreWg.Wait()
}

func createInitializationParams() {

	var valInit []initialization.ValidatorInitialization
	for _, kslice := range validatorsKeys {
		for _, key := range kslice {
			val := initialization.ValidatorInitialization{
				PubKey:       hex.EncodeToString(key.PublicKey().Marshal()),
				PayeeAddress: premineAddr.PublicKey().ToAccount(),
			}
			valInit = append(valInit, val)
		}

	}

	initParams = initialization.InitializationParameters{
		InitialValidators: valInit,
		PremineAddress:    premineAddr.PublicKey().ToAccount(),
		GenesisTime:       time.Unix(time.Now().Unix()+delaySeconds, 0),
	}
}

func createServers() {
	var wg sync.WaitGroup
	wg.Add(NumNodes)
	for i, f := range folders {
		go func(index int, folder string, wg *sync.WaitGroup) {
			defer wg.Done()
			log := loggers[index]
			params.SlotDuration = 4
			db, err := blockdb.NewBadgerDB(folder, params, log)
			if err != nil {
				panic(err)
			}
			config := &server.GlobalConfig{
				DataFolder:   folder,
				NetworkName:  "",
				Port:         strconv.Itoa(24000 + index),
				InitConfig:   initialization.InitializationParameters{},
				RPCProxy:     false,
				RPCProxyPort: strconv.Itoa(8080 + index),
				RPCProxyAddr: "",
				RPCPort:      strconv.Itoa(25000 + index),
				RPCWallet:    true,
				RPCAuthToken: "",
				Debug:        true,
				LogFile:      false,
			}
			s, err := server.NewServer(context.Background(), config, log, params, db, initParams)
			if err != nil {
				panic(err)
			}
			servers[index] = s
		}(i, f, &wg)
	}
	wg.Wait()
}

func TestMain(t *testing.M) {
	createTestEnvironment()
	createKeystoresAndValidators()
	createInitializationParams()
	createServers()
	os.Exit(t.Run())
}

func TestStartNodes(t *testing.T) {
	for _, s := range servers {
		assert.NotPanics(t, func() {
			s.Start()
		})
	}
}

func TestConnectNodes(t *testing.T) {

	var peersInfo []multiaddr.Multiaddr

	for i, s := range servers {
		netID := s.HostNode().GetHost().ID()
		ma, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/" + strconv.Itoa(24000+i) + "/p2p/" + netID.String())
		assert.NoError(t, err)
		peersInfo = append(peersInfo, ma)
	}

	peersConnect := getRandPeers(peersInfo)

	for i := range servers {

		client, err := rpcClient("127.0.0.1:" + strconv.Itoa(25000+i))
		assert.NoError(t, err)

		for _, p := range peersConnect {

			success, err := client.network.AddPeer(context.Background(), &proto.IP{
				Host: p.String(),
			})

			assert.NoError(t, err)
			if success.Error != "dial to self attempted" {
				assert.True(t, success.Success)
			}
		}
	}

}

func getRandPeers(peers []multiaddr.Multiaddr) []multiaddr.Multiaddr {
	peersLength := len(peers)
	peersCalc := make([]multiaddr.Multiaddr, peersLength/3)
	for i := range peersCalc {
		r := rand.Intn(peersLength)
		peersCalc[i] = peers[r]
	}
	return peersCalc
}

func TestCheckNodeConnections(t *testing.T) {
	for i := range servers {

		client, err := rpcClient("127.0.0.1:" + strconv.Itoa(25000+i))
		assert.NoError(t, err)

		peers, err := client.network.GetPeersInfo(context.Background(), &proto.Empty{})
		assert.NoError(t, err)

		assert.GreaterOrEqual(t, len(peers.Peers), (NumNodes-1)/2/2)
		assert.LessOrEqual(t, len(peers.Peers), NumNodes)
	}
}

type client struct {
	network proto.NetworkClient
	wallet  proto.WalletClient
}

func rpcClient(addr string) (*client, error) {
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTransportCredentials(creds))
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}
	return &client{
		network: proto.NewNetworkClient(conn),
		wallet:  proto.NewWalletClient(conn),
	}, nil
}

type notify struct {
	num           int
	lastJustified uint64
	lastFinalized uint64
	slashed       bool
}

func (n *notify) NewTip(r *chainindex.BlockRow, b *primitives.Block, s state.State, receipts []*primitives.EpochReceipt) {
	n.lastFinalized = s.GetFinalizedEpoch()
	n.lastJustified = s.GetJustifiedEpoch()
	if len(receipts) > 0 {
		msg := "\nEpoch Receipts\n----------\n"
		receiptTypes := make(map[string]int64)

		for _, r := range receipts {
			if _, ok := receiptTypes[r.TypeString()]; !ok {
				receiptTypes[r.TypeString()] = r.Amount
			} else {
				receiptTypes[r.TypeString()] += r.Amount
			}
		}

		for rt, amount := range receiptTypes {
			if amount > 0 {
				msg += fmt.Sprintf("rewarded %d for %s\n", amount, rt)
			} else if amount < 0 {
				msg += fmt.Sprintf("penalized %d for %s\n", -amount, rt)
			} else {
				msg += fmt.Sprintf("neutral increments for %s\n", rt)
			}
		}

		fmt.Println(msg)
	}
	fmt.Printf("Validator Registry: Active %d Starting %d Pending Exit %d Penalty Exit %d Exited %d \n", s.GetValidators().Active, s.GetValidators().Starting, s.GetValidators().PendingExit, s.GetValidators().PenaltyExit, s.GetValidators().Exited)
	fmt.Printf("Node %d: received block %d at slot %d Justified: %d Finalized: %d \n", n.num, r.Height, r.Slot, n.lastJustified, n.lastFinalized)
}

func (n *notify) ProposerSlashingConditionViolated(*primitives.ProposerSlashing) {
	n.slashed = true
}

var notifies = make([]*notify, NumNodes)

func TestChainCorrectness(t *testing.T) {
	for i, s := range servers {
		n := &notify{
			num:           i,
			lastFinalized: 0,
			lastJustified: 0,
			slashed:       false,
		}
		s.Chain().Notify(n)
		notifies[i] = n
	}
	for {
		time.Sleep(time.Second * 1)
		if servers[0].Chain().State().TipState().GetSlot() == 26 {
			for _, n := range notifies {
				assert.Equal(t, n.lastJustified, uint64(3))
				assert.Equal(t, n.lastFinalized, uint64(2))
				assert.False(t, n.slashed)
			}
			break
		}
	}
}

func TestImportCreateNewWallet(t *testing.T) {

	clientPremine, err := rpcClient("127.0.0.1:" + strconv.Itoa(25000))
	assert.NoError(t, err)

	clientReceiving, err := rpcClient("127.0.0.1:" + strconv.Itoa(25001))
	assert.NoError(t, err)

	_, err = clientPremine.wallet.ImportWallet(context.Background(), &proto.ImportWalletData{
		Name: "premine_wallet",
		Key: &proto.KeyPair{
			Private: premineAddr.ToWIF(),
		},
		Password: walletsPass,
	})
	assert.NoError(t, err)

	w, err := clientReceiving.wallet.ImportWallet(context.Background(), &proto.ImportWalletData{
		Name: "receiving_wallet",
		Key: &proto.KeyPair{
			Private: receivingAddr.ToWIF(),
		},
		Password: walletsPass,
	})
	assert.NoError(t, err)

	_, err = clientPremine.wallet.SendTransaction(context.Background(), &proto.SendTransactionInfo{
		Account: w.Public,
		Amount:  "12800",
	})
	assert.NoError(t, err)
}

func TestValidatorsIncrease(t *testing.T) {
	time.Sleep(time.Second * 5)
	newValKeys, err := servers[0].Proposer().Keystore().GenerateNewValidatorKey(128)
	assert.NoError(t, err)
	keys := &proto.KeyPairs{Keys: make([]string, len(newValKeys))}
	for i := range keys.Keys {
		keys.Keys[i] = hex.EncodeToString(newValKeys[i].Marshal())
	}

	client, err := rpcClient("127.0.0.1:" + strconv.Itoa(25001))
	assert.NoError(t, err)

	success, err := client.wallet.StartValidatorBulk(context.Background(), keys)
	assert.NoError(t, err)
	assert.True(t, success.Success)
}

func TestChainCorrectnessWithMoreValidators(t *testing.T) {
	for {
		time.Sleep(time.Second * 1)
		if servers[0].Chain().State().TipState().GetSlot() == 56 {
			for _, n := range notifies {
				assert.Equal(t, n.lastJustified, uint64(9))
				assert.Equal(t, n.lastFinalized, uint64(8))
				assert.False(t, n.slashed)
			}
			break
		}
	}
}

func TestStopProposers(t *testing.T) {
	servers[0].Proposer().Stop()
	servers[NumNodes-1].Proposer().Stop()
}

func TestChainCorrectnessWithValidatorsPenalization(t *testing.T) {
	for {
		time.Sleep(time.Second * 1)
		if servers[0].Chain().State().TipState().GetSlot() == 86 {
			for _, n := range notifies {
				assert.Equal(t, n.lastJustified, uint64(15))
				assert.Equal(t, n.lastFinalized, uint64(14))
			}
			break
		}
	}
}
