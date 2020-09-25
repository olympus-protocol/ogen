package hostnode

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"

	"github.com/libp2p/go-libp2p-core/protocol"

	"github.com/libp2p/go-libp2p-core/peer"
)

const discoveryProtocolID = protocol.ID("/ogen/discovery/" + OgenVersion)

// DiscoveryProtocol is an interface for discoveryProtocol
type DiscoveryProtocol interface {
	Start() error
	Listen(network.Network, multiaddr.Multiaddr)
	ListenClose(network.Network, multiaddr.Multiaddr)
	Connected(net network.Network, conn network.Conn)
	Disconnected(net network.Network, conn network.Conn)
	OpenedStream(network.Network, network.Stream)
	ClosedStream(network.Network, network.Stream)
}

var _ DiscoveryProtocol = &discoveryProtocol{}

// discoveryProtocol is the service to discover other hostnode.
type discoveryProtocol struct {
	host   HostNode
	config Config
	ctx    context.Context
	log    logger.Logger

	lastConnect     map[peer.ID]time.Time
	lastConnectLock sync.RWMutex

	protocolHandler ProtocolHandler

	nonWorkingPeersLock sync.RWMutex
	nonWorkingPeers     map[peer.ID]*peer.AddrInfo
}

// NewDiscoveryProtocol creates a new discovery service.
func NewDiscoveryProtocol(ctx context.Context, host HostNode, config Config) (DiscoveryProtocol, error) {
	ph := newProtocolHandler(ctx, discoveryProtocolID, host, config)
	dp := &discoveryProtocol{
		host:            host,
		ctx:             ctx,
		config:          config,
		lastConnect:     make(map[peer.ID]time.Time),
		protocolHandler: ph,
		log:             config.Log,
		nonWorkingPeers: make(map[peer.ID]*peer.AddrInfo),
	}
	if err := ph.RegisterHandler(p2p.MsgGetAddrCmd, dp.handleGetAddr); err != nil {
		return nil, err
	}
	if err := ph.RegisterHandler(p2p.MsgAddrCmd, dp.handleAddr); err != nil {
		return nil, err
	}
	host.Notify(dp)
	return dp, nil
}

const connectionCooldown = 60 * time.Second

func shufflePeers(peers []peer.AddrInfo) []peer.AddrInfo {
	rand.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})

	return peers
}

func (cm *discoveryProtocol) handleAddr(_ peer.ID, msg p2p.Message) error {
	msgAddr, ok := msg.(*p2p.MsgAddr)
	if !ok {
		return fmt.Errorf("message received is not addr")
	}

	peers := msgAddr.Addr

	for _, pb := range peers {
		offset := binary.LittleEndian.Uint64(pb[:])
		maBytes := pb[8 : 8+offset]
		pma, err := multiaddr.NewMultiaddrBytes(maBytes)
		if err != nil {
			continue
		}
		p, err := peer.AddrInfoFromP2pAddr(pma)
		if err != nil {
			continue
		}
		if p.ID == cm.host.GetHost().ID() {
			continue
		}
		cm.nonWorkingPeersLock.Lock()
		_, ok := cm.nonWorkingPeers[p.ID]
		if !ok {
			if err := cm.host.SavePeer(*p); err != nil {
				cm.log.Errorf("error saving peer: %s", err)
				cm.nonWorkingPeers[p.ID] = p
				cm.nonWorkingPeersLock.Unlock()
				continue
			}
		}
		cm.nonWorkingPeersLock.Unlock()
	}

	return nil
}

func (cm *discoveryProtocol) handleGetAddr(id peer.ID, msg p2p.Message) error {
	_, ok := msg.(*p2p.MsgGetAddr)
	if !ok {
		return fmt.Errorf("message received is not get addr")
	}
	var peers [][256]byte
	peersIDs := cm.host.GetHost().Peerstore().Peers()
	var peersInfo []peer.AddrInfo
	for _, id := range peersIDs {
		pInfo := cm.host.GetHost().Peerstore().PeerInfo(id)
		peersInfo = append(peersInfo, pInfo)
	}
	peersData := shufflePeers(peersInfo)

	for i, p := range peersData {
		if i < p2p.MaxAddrPerMsg {
			peerMulti, err := peer.AddrInfoToP2pAddrs(&p)
			if err != nil {
				continue
			}
			var pb [256]byte
			offset := len(peerMulti[0].Bytes())
			binary.LittleEndian.PutUint64(pb[:], uint64(offset))
			copy(pb[8:], peerMulti[0].Bytes())
			peers = append(peers, pb)
		}
	}

	return cm.protocolHandler.SendMessage(id, &p2p.MsgAddr{
		Addr: peers,
	})
}

const askForPeersCycle = 60 * time.Second

func (cm *discoveryProtocol) Start() error {
	go func() {
		storedPeers, err := cm.host.Database().GetSavedPeers()
		if err != nil {
			cm.log.Errorf("unable to load stored peers")
		}
		var initialPeers []peer.AddrInfo
		initialPeers = append(initialPeers, storedPeers...)
		initialPeers = append(initialPeers, cm.config.InitialNodes...)
		for _, addr := range initialPeers {
			if err := cm.connect(&addr); err != nil {
				cm.log.Errorf("error connecting to add node %s: %s", addr, err)
				cm.host.Database()
			}
		}
	}()

	go func() {
		askForPeersTicker := time.NewTicker(askForPeersCycle)
		for {
			select {
			case <-askForPeersTicker.C:
				possiblePeersToAsk := cm.host.GetPeerList()
				if len(possiblePeersToAsk) == 0 {
					continue
				}
				peerIdxToAsk := rand.Int() % len(possiblePeersToAsk)
				peerToAsk := possiblePeersToAsk[peerIdxToAsk]

				if err := cm.protocolHandler.SendMessage(peerToAsk, &p2p.MsgGetAddr{}); err != nil {
					cm.log.Errorf("error sending getaddr: %s", err)
					return
				}
			case <-cm.ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Connect connects to a peer.
func (cm *discoveryProtocol) connect(pi *peer.AddrInfo) error {
	cm.lastConnectLock.Lock()
	defer cm.lastConnectLock.Unlock()

	lastConnect, found := cm.lastConnect[pi.ID]
	if !found || time.Since(lastConnect) > connectionCooldown {
		cm.lastConnect[pi.ID] = time.Now()
		return cm.host.SavePeer(*pi)
	}
	return nil
}

// Listen is called when we start listening on a multiaddr.
func (cm *discoveryProtocol) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when we stop listening on a multiaddr.
func (cm *discoveryProtocol) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when we connect to a peer.
func (cm *discoveryProtocol) Connected(_ network.Network, conn network.Conn) {
	if conn.Stat().Direction != network.DirOutbound {
		return
	}

	// open a stream for the discovery protocol:
	s, err := cm.host.GetHost().NewStream(cm.ctx, conn.RemotePeer(), discoveryProtocolID)
	if err != nil {
		cm.log.Errorf("could not open stream for connection: %s", err)
	}

	cm.protocolHandler.HandleStream(s)
}

// Disconnected is called when we disconnect from a peer.
func (cm *discoveryProtocol) Disconnected(network.Network, network.Conn) {}

// OpenedStream is called when we open a stream.
func (cm *discoveryProtocol) OpenedStream(network.Network, network.Stream) {}

// ClosedStream is called when we close a stream.
func (cm *discoveryProtocol) ClosedStream(network.Network, network.Stream) {}
