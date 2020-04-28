package peers

import (
	"bytes"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/db/filedb"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/serializer"
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
	peers        map[int]*Peer
	peersLock    sync.RWMutex

	// Peer Communication Channels
	closePeerChan chan Peer
	msgChan       chan p2p.Message

	// Peers DB
	peersDB     *filedb.FileDB
	bannedPeers *filedb.FileDB
	// Services Pointers
	chain   *chain.Blockchain
	mempool *Mempool
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
		newPeer := NewPeer(len(pm.peers)+1, conn, newAddr, true, time.Now(), pm.log, pm)
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
	} else if len(pm.addNodes) != 0 {
		// If there are add nodes configured, ignore everything and use them
		initialPeers = pm.addNodes
	}
	pm.log.Infof("%v known peer(s)", len(initialPeers))

	// Run initial peer connect
	pm.initialPeersConnection(initialPeers)
	// Run connection routine for constant peer number maintaining
	go pm.peersLockup()
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
	for k := range cleanPeersList {
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
		newPeer := NewPeer(len(pm.peers)+1, conn, newAddr, false, time.Now(), pm.log, pm)
		pm.syncNewPeer(newPeer)
	}
}

func (pm *PeerMan) syncNewPeer(p *Peer) {
	p.Start(pm.chain.State().Height())
}

func (pm *PeerMan) Disconnect(p *Peer) {
	pm.log.Infof("removing peer addr=%v:%v ", p.GetAddr().IP, p.GetAddr().Port)
	pm.removePeer(p)
}

func (pm *PeerMan) SubmitVote(vote *primitives.SingleValidatorVote) error {
	pm.peersLock.Lock()
	defer pm.peersLock.Unlock()
	for _, p := range pm.peers {
		if err := p.submitVote(vote); err != nil {
			return err
		}
	}
	return nil
}

func (pm *PeerMan) Handshake(p *Peer) {
	pm.log.Infof("new peer handshaked addr=%v:%v ", p.GetAddr().IP, p.GetAddr().Port)
	pm.addPeer(p)
	rawPeerData, err := p.GetSerializedData()
	if err != nil {
		pm.log.Errorf("error serializing peer: %s", err)
		return
	}
	err = pm.peersDB.Add(rawPeerData)
	if err != nil {
		pm.log.Errorf("error adding peer: %s", err)
	}
}

func (pm *PeerMan) Ban(p *Peer) {
	pm.log.Infof("banning peer addr=%v:%v ", p.GetAddr().IP, p.GetAddr().Port)
	pm.removePeer(p)
	rawPeerData, err := p.GetSerializedData()
	if err != nil {
		pm.log.Errorf("error serializing peer: %s", err)
		return
	}
	err = pm.bannedPeers.Add(rawPeerData)
	if err != nil {
		pm.log.Errorf("error banning peer: %s", err)
		return
	}
}

func (pm *PeerMan) addPeer(p *Peer) {
	pm.peersLock.Lock()
	pm.peers[len(pm.peers)+1] = p
	pm.peersLock.Unlock()
}

func (pm *PeerMan) removePeer(p *Peer) {
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

func NewPeersMan(config Config, params params.ChainParams, chain *chain.Blockchain, mempool *Mempool) (*PeerMan, error) {
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
		peers:        make(map[int]*Peer),
		peersDB:      peersdb,
		bannedPeers:  bansDB,
		chain:        chain,
		mempool:      mempool,
	}
	return peersMan, nil
}
