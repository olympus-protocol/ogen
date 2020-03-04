package peers

import (
	"bytes"
	"errors"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/db/filedb"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/peers/peer"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"net"
	"reflect"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	Log          *logger.Logger
	Listen       bool
	AddNodes     []string
	ConnectNodes []string
	Port         int32
	MaxPeers     int32
	Path         string
}

type PeerMan struct {
	// Ogen Manager Properties
	log    *logger.Logger
	config Config
	params params.ChainParams

	// Custom PeerMan Properties
	addNodes     []string
	connectNodes []string
	peers        map[int]*peer.Peer
	peersLock    sync.RWMutex

	// Peer Communication Channels
	closePeerChan chan peer.Peer
	msgChan       chan p2p.Message

	// Peers DB
	peersDB     *filedb.FileDB
	bannedPeers *filedb.FileDB
	// Services Pointers
	chain *chain.Blockchain

	peersSyncLock sync.RWMutex
	peersAhead    map[int]*peer.Peer
	peersBehind   map[int]*peer.Peer
	peersEqual    map[int]*peer.Peer
}

func (pm *PeerMan) listener() {
	// TODO prevent connect to itself
	pm.log.Tracef("Starting peer listener on port %v", pm.config.Port)
	list, err := net.Listen("tcp", ":"+strconv.Itoa(int(pm.config.Port)))
	if err != nil {
		pm.log.Fatalf("Unable to bind addr: :%v. Error: %v", strconv.Itoa(int(pm.config.Port)), err.Error())
	}
	for {
		conn, err := list.Accept()
		if conn.RemoteAddr().String() == "127.0.0.1:"+strconv.Itoa(int(pm.config.Port)) {
			// Prevent self connections
			_ = conn.Close()
		}
		pm.log.Infof("received connection request from %v", conn.RemoteAddr().String())
		if err != nil {
			pm.log.Fatalf("Unable to bind connect to peer")
		}
		ip, port, err := net.SplitHostPort(conn.RemoteAddr().String())

		portParse, _ := strconv.Atoi(port)
		newAddr := serializer.NetAddress{
			IP:        net.ParseIP(ip),
			Port:      uint16(portParse),
			Timestamp: time.Now().Unix(),
		}
		newPeer := peer.NewPeer(len(pm.peers)+1, conn, newAddr, true, time.Now(), make(chan interface{}), pm.log)
		pm.syncNewPeer(newPeer)
	}
}

func (pm *PeerMan) Start() error {
	pm.log.Info("Starting PeersMan instance")
	if pm.config.Listen && len(pm.connectNodes) == 0 {
		go pm.listener()
	}
	var initialPeers []string
	// If no known peers, load databases
	if len(pm.config.AddNodes) == 0 && len(pm.config.ConnectNodes) == 0 {
		// Load peers database and banlist, compare both and get a sanitizied initial peer list
		peersFound, err := pm.loadDatabase()
		if err != nil {
			pm.log.Warn("error loading peers and banlist database")
		}
		initialPeers = peersFound
		if len(initialPeers) == 0 {
			// Load peers database is empty, query seeders
			initialPeers = pm.querySeeders()
		}
	}
	if len(pm.connectNodes) != 0 {
		// If there are connect nodes configured, ignore everything and use them
		initialPeers = pm.connectNodes
	}
	if len(pm.addNodes) != 0 {
		// If there are add nodes configured, ignore everything and use them
		initialPeers = pm.addNodes
	}
	pm.log.Infof("%v known peer(s)", len(initialPeers))

	// Run initial peer connect
	pm.initialPeersConnection(initialPeers)
	// Run connection routine for constant peer number maintaining
	go pm.peersLockup()
	// Run possible sync routine for constantly catch the chain from peers
	go pm.peersSync()
	return nil
}

func (pm *PeerMan) loadDatabase() ([]string, error) {
	banListMap := make(map[string]serializer.NetAddress)
	var peers []serializer.NetAddress
	banListRaw, err := pm.bannedPeers.GetAll()
	if err != nil {
		return nil, err
	}
	if len(banListRaw) > 0 {
		for _, rawBanPeer := range banListRaw {
			var na serializer.NetAddress
			buf := bytes.NewBuffer(rawBanPeer)
			err = serializer.ReadNetAddress(buf, &na)
			if err != nil {
				return nil, err
			}
			banListMap[na.IP.String()] = na
		}
	}
	peersRawList, err := pm.peersDB.GetAll()
	if err != nil {
		return nil, err
	}
	if len(peersRawList) > 0 {
		for _, peerRaw := range peersRawList {
			var na serializer.NetAddress
			buf := bytes.NewBuffer(peerRaw)
			err = serializer.ReadNetAddress(buf, &na)
			if err != nil {
				return nil, err
			}
			peers = append(peers, na)
		}
	}
	cleanPeersList := make(map[string]interface{})
	for _, peerNet := range peers {
		_, ok := banListMap[peerNet.IP.String()]
		if !ok {
			_, notUsed := cleanPeersList[peerNet.IP.String()]
			if !notUsed {
				cleanPeersList[peerNet.IP.String()] = nil
			}
		}
	}
	var cleanPeersArray []string
	for k, _ := range cleanPeersList {
		cleanPeersArray = append(cleanPeersArray, k+":"+pm.params.DefaultP2PPort)
	}
	return cleanPeersArray, nil
}

func (pm *PeerMan) querySeeders() []string {
	return nil
}

func (pm *PeerMan) initialPeersConnection(initialPeers []string) {
	for _, peerIP := range initialPeers {
		if peerIP == "127.0.0.1:"+strconv.Itoa(int(pm.config.Port)) || peerIP == "localhost:"+strconv.Itoa(int(pm.config.Port)) {
			// Prevent self connections
			continue
		}
		conn, err := pm.dial(peerIP)
		if err != nil {
			pm.log.Tracef("Unable to dial peer %v", peerIP)
			continue
		}
		ip, port, err := net.SplitHostPort(conn.RemoteAddr().String())

		portParse, _ := strconv.Atoi(port)
		newAddr := serializer.NetAddress{
			IP:        net.ParseIP(ip),
			Port:      uint16(portParse),
			Timestamp: time.Now().Unix(),
		}
		newPeer := peer.NewPeer(len(pm.peers)+1, conn, newAddr, false, time.Now(), make(chan interface{}), pm.log)
		pm.syncNewPeer(newPeer)
	}
}

func (pm *PeerMan) syncNewPeer(p *peer.Peer) {
	go pm.peerChan(p)
	p.Start(pm.chain.StateSnapshot().Height)
}

func (pm *PeerMan) peerChan(p *peer.Peer) {
	for {
		select {
		// TODO close chan
		default:
			rmsg := <-p.ManChan
			switch msg := rmsg.(type) {
			case *peer.PeerMsg:
				// Ignore error, the only reason it can fail is
				// because of data storage.
				_ = pm.handlePeerMsg(msg)
			case *peer.BlockMsg:
				err := pm.handleBlockMsg(msg)
				if err != nil {
					// Add Ban Score
				}
			case *peer.TxMsg:
				err := pm.handleTxMsg(msg)
				if err != nil {
					// Add Ban Score
				}
			case *peer.BlocksInvMsg:
				err := pm.handleBlockInvMsg(msg)
				if err != nil {

				}
			case *peer.DataReqMsg:
				err := pm.handleDataRequestMsg(msg)
				if err != nil {
					pm.closePeerChan <- *msg.Peer
				}
			}
		}
	}
}

func (pm *PeerMan) handlePeerMsg(msg *peer.PeerMsg) error {
	switch msg.Status {
	case peer.Handshaked:
		pm.log.Infof("new peer handshaked addr=%v:%v ", msg.Peer.GetAddr().IP, msg.Peer.GetAddr().Port)
		pm.addPeer(msg.Peer)
		pm.organizePeer(msg.Peer)
		rawPeerData, err := msg.Peer.GetSerializedData()
		if err != nil {
			return err
		}
		err = pm.peersDB.Add(rawPeerData)
		if err != nil {
			return err
		}
		break
	case peer.Syncing:
		reqMsg := p2p.NewMsgGetBlock(pm.chain.StateSnapshot().Hash)
		p := msg.Peer
		p.SetPeerSync(true)
		err := p.SendGetBlocks(reqMsg)
		if err != nil {
			return err
		}
	case peer.Disconnect:
		pm.log.Infof("removing peer addr=%v:%v ", msg.Peer.GetAddr().IP, msg.Peer.GetAddr().Port)
		pm.removePeer(msg.Peer)
	case peer.Ban:
		pm.log.Infof("banning peer addr=%v:%v ", msg.Peer.GetAddr().IP, msg.Peer.GetAddr().Port)
		pm.removePeer(msg.Peer)
		rawPeerData, err := msg.Peer.GetSerializedData()
		if err != nil {
			return err
		}
		err = pm.bannedPeers.Add(rawPeerData)
		if err != nil {
			return err
		}
		break
	}
	return nil
}

func (pm *PeerMan) handleBlockInvMsg(msg *peer.BlocksInvMsg) error {
	if !msg.Peer.IsSelectedForSync() {
		pm.log.Infof("block inv msg for non-requested peer")
		return errors.New("non requested block inv msg")
	}
	pm.log.Infof("new block inv with %v blocks", len(msg.Blocks.GetBlocks()))
	for _, block := range msg.Blocks.GetBlocks() {
		newBlock, err := primitives.NewBlockFromMsg(block, uint32(pm.chain.State().Snapshot().Height+1))
		if err != nil {
			return err
		}
		err = pm.chain.ProcessBlock(newBlock)
		if err != nil {
			return err
		}
	}
	pm.organizePeer(msg.Peer)
	return nil
}

func (pm *PeerMan) handleBlockMsg(msg *peer.BlockMsg) error {
	blockHash := msg.Block.Header.Hash()
	pm.log.Infof("new block received hash: %v", blockHash)
	if pm.chain.State().IsSync() {
		newBlock, err := primitives.NewBlockFromMsg(msg.Block, uint32(pm.chain.State().Snapshot().Height+1))
		err = pm.chain.ProcessBlock(newBlock)
		if err != nil {
			return err
		}
		return nil
	}
	pm.log.Infof("ignored block, we are not synced yet")
	return nil
}

func (pm *PeerMan) RelayBlockMsg(msg *p2p.MsgBlock) {
	for _, p := range pm.peers {
		err := p.SendBlock(msg)
		if err != nil {

		}
	}
}

func (pm *PeerMan) handleTxMsg(msg *peer.TxMsg) error {
	return nil
}

func (pm *PeerMan) handleDataRequestMsg(msg *peer.DataReqMsg) error {
	switch msg.Request {
	case "getblocks":
		var blocksInv p2p.MsgBlockInv
		for i := 0; i < p2p.MaxBlocksPerInv; i++ {
			// TODO refactor
		}
		err := msg.Peer.SendBlockInv(&blocksInv)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pm *PeerMan) addPeer(p *peer.Peer) {
	pm.peersLock.Lock()
	pm.peers[len(pm.peers)+1] = p
	pm.peersLock.Unlock()
}

func (pm *PeerMan) organizePeer(p *peer.Peer) {
	if p.GetLastBlock() > pm.chain.StateSnapshot().Height {
		pm.peersSyncLock.Lock()
		pm.peersAhead[p.GetID()] = p
		pm.peersSyncLock.Unlock()
	}
	if p.GetLastBlock() < pm.chain.StateSnapshot().Height {
		pm.peersSyncLock.Lock()
		pm.peersBehind[p.GetID()] = p
		pm.peersSyncLock.Unlock()
	}
	if p.GetLastBlock() == pm.chain.StateSnapshot().Height {
		pm.peersSyncLock.Lock()
		pm.peersEqual[p.GetID()] = p
		pm.peersSyncLock.Unlock()
	}
	if len(pm.peersAhead) > 0 {
		pm.chain.State().SetSyncStatus(false)
		keys := reflect.ValueOf(pm.peersAhead).MapKeys()
		// User the first peer on the map as a sync peer.
		err := pm.handlePeerMsg(&peer.PeerMsg{Peer: pm.peersAhead[int(keys[0].Int())], Status: peer.Syncing})
		if err != nil {
			return
		}
	} else {
		pm.chain.State().SetSyncStatus(true)
	}
}

func (pm *PeerMan) removePeer(p *peer.Peer) {
	pm.peersLock.Lock()
	delete(pm.peers, p.GetID())
	pm.peersLock.Unlock()
	p.Stop()
}

func (pm *PeerMan) peersLockup() {
lookup:
	for len(pm.peers) < int(pm.config.MaxPeers) {
		time.Sleep(time.Second * 30)
	}
	time.Sleep(time.Minute)
	goto lookup
}

func (pm *PeerMan) peersSync() {
	// Ask blocks for ahead peers
	go func() {
	start:
		for len(pm.peersAhead) > 0 {
			for id, _ := range pm.peersAhead {
				pm.peersSyncLock.Lock()
				delete(pm.peersAhead, id)
				pm.peersSyncLock.Unlock()
			}
		}
		time.Sleep(10 * time.Second)
		goto start
	}()
	// Send blocks for behind peers
	go func() {
	start:
		for len(pm.peersBehind) > 0 {
			for id, _ := range pm.peersBehind {
				pm.peersSyncLock.Lock()
				delete(pm.peersBehind, id)
				pm.peersSyncLock.Unlock()
			}
		}
		time.Sleep(10 * time.Second)
		goto start
	}()
	// Relay blocks to equal peers
	go func() {
	start:
		for len(pm.peersEqual) > 0 {
			for id, _ := range pm.peersEqual {
				pm.peersSyncLock.Lock()
				delete(pm.peersEqual, id)
				pm.peersSyncLock.Unlock()
			}
		}
		time.Sleep(10 * time.Second)
		goto start
	}()
}

func (pm *PeerMan) dial(addres string) (net.Conn, error) {
	return net.Dial("tcp", addres)
}

func (pm *PeerMan) Stop() {
	pm.log.Info("Stoping PeersMan instance")
}

func (pm *PeerMan) GetPeersCount() int32 {
	pm.peersLock.Lock()
	count := len(pm.peers)
	pm.peersLock.Unlock()
	return int32(count)
}

func NewPeersMan(config Config, params params.ChainParams, chain *chain.Blockchain) (*PeerMan, error) {
	peersDbMetaData := filedb.MetaData{
		Version:     100000,
		Timestamp:   time.Now().Unix(),
		Name:        "peers-database",
		MaxElemSize: 26,
	}
	peersdb, err := filedb.NewFileDB(config.Path+"/peers.dat", peersDbMetaData)
	if err != nil {
		return nil, err
	}
	bansDbMeta := filedb.MetaData{
		Version:     100000,
		Timestamp:   time.Now().Unix(),
		Name:        "banned-peers-database",
		MaxElemSize: 26,
	}
	bansDB, err := filedb.NewFileDB(config.Path+"/banlist.dat", bansDbMeta)
	if err != nil {
		return nil, err
	}
	peersMan := &PeerMan{
		log:          config.Log,
		config:       config,
		params:       params,
		addNodes:     config.AddNodes,
		connectNodes: config.ConnectNodes,
		peers:        make(map[int]*peer.Peer),
		peersDB:      peersdb,
		bannedPeers:  bansDB,
		chain:        chain,
		peersEqual:   make(map[int]*peer.Peer),
		peersBehind:  make(map[int]*peer.Peer),
		peersAhead:   make(map[int]*peer.Peer),
	}
	return peersMan, nil
}
