package peers

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/logger"

	"github.com/libp2p/go-libp2p-core/protocol"

	"github.com/libp2p/go-libp2p-core/peer"
)

const DiscoveryProtocolID = protocol.ID("/ogen/discovery/0.0.1")

// DiscoveryProtocol is the service to discover other peers.
type DiscoveryProtocol struct {
	host   *HostNode
	config Config
	ctx    context.Context
	log    *logger.Logger

	lastConnect     map[peer.ID]time.Time
	lastConnectLock sync.RWMutex

	protocolHandler *ProtocolHandler
}

// NewDiscoveryProtocol creates a new discovery service.
func NewDiscoveryProtocol(ctx context.Context, host *HostNode, config Config) (*DiscoveryProtocol, error) {
	ph := newProtocolHandler(ctx, DiscoveryProtocolID, host, config)
	dp := &DiscoveryProtocol{
		host:            host,
		ctx:             ctx,
		config:          config,
		lastConnect:     make(map[peer.ID]time.Time),
		protocolHandler: ph,
		log:             config.Log,
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

const connectionTimeout = 10 * time.Second
const connectionCooldown = 60 * time.Second

func shufflePeers(peers []peer.AddrInfo) []peer.AddrInfo {
	rand.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})

	return peers
}

func (cm *DiscoveryProtocol) handleAddr(id peer.ID, msg p2p.Message) error {
	msgAddr, ok := msg.(*p2p.MsgAddr)
	if !ok {
		return fmt.Errorf("message received is not addr")
	}

	peers := msgAddr.Addr

	// let's set a very short timeout so we can connect faster
	timeout := time.Second * 5

	for _, pb := range peers {
		pma, err := multiaddr.NewMultiaddrBytes(pb)
		if err != nil {
			continue
		}
		p, err := peer.AddrInfoFromP2pAddr(pma)
		if err != nil {
			continue
		}
		if p.ID == cm.host.host.ID() {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		if err := cm.host.host.Connect(ctx, *p); err != nil {
			cm.log.Tracef("error connecting to suggested peer %s: %s", p, err)
			cancel()
			continue
		}
		cancel()
	}

	return nil
}

func (cm *DiscoveryProtocol) handleGetAddr(id peer.ID, msg p2p.Message) error {
	_, ok := msg.(*p2p.MsgGetAddr)
	if !ok {
		return fmt.Errorf("message received is not get addr")
	}
	peers := [][]byte{}
	peersData := shufflePeers(cm.host.GetPeerInfos())

	for _, p := range peersData {
		if len(peers) < p2p.MaxAddrPerMsg {
			peers = append(peers, p.Addrs[0].Bytes())
		}
	}

	if len(peers) > p2p.MaxAddrPerMsg {
		peers = peers[:p2p.MaxAddrPerMsg]
	}

	return cm.protocolHandler.SendMessage(id, &p2p.MsgAddr{
		Addr: peers,
	})
}

const askForPeersCycle = 60 * time.Second

func (cm *DiscoveryProtocol) Start() error {
	go func() {
		for _, addr := range cm.config.AddNodes {
			if err := cm.connect(addr); err != nil {
				cm.log.Errorf("error connecting to add node %s: %s", addr, err)
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
func (cm *DiscoveryProtocol) connect(pi peer.AddrInfo) error {
	cm.lastConnectLock.Lock()
	defer cm.lastConnectLock.Unlock()

	if cm.host.PeersConnected() < cm.config.MaximumPeers {
	}

	lastConnect, found := cm.lastConnect[pi.ID]
	if !found || time.Since(lastConnect) > connectionCooldown {
		cm.lastConnect[pi.ID] = time.Now()
		ctx, cancel := context.WithTimeout(cm.ctx, connectionTimeout)
		defer cancel()
		return cm.host.host.Connect(ctx, pi)
	}
	return nil
}

// Listen is called when we start listening on a multiaddr.
func (cm *DiscoveryProtocol) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when we stop listening on a multiaddr.
func (cm *DiscoveryProtocol) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when we connect to a peer.
func (cm *DiscoveryProtocol) Connected(net network.Network, conn network.Conn) {
	if conn.Stat().Direction != network.DirOutbound {
		return
	}

	// open a stream for the discovery protocol:
	s, err := cm.host.host.NewStream(cm.ctx, conn.RemotePeer(), DiscoveryProtocolID)
	if err != nil {
		cm.log.Errorf("could not open stream for connection: %s", err)
	}

	cm.protocolHandler.handleStream(s)
}

// Disconnected is called when we disconnect from a peer.
func (cm *DiscoveryProtocol) Disconnected(net network.Network, conn network.Conn) {}

// OpenedStream is called when we open a stream.
func (cm *DiscoveryProtocol) OpenedStream(network.Network, network.Stream) {}

// ClosedStream is called when we close a stream.
func (cm *DiscoveryProtocol) ClosedStream(network.Network, network.Stream) {}
