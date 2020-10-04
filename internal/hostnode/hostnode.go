package hostnode

import (
	"context"
	"crypto/rand"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/params"
	"io/ioutil"
	"os"
	"path"
	"time"

	dsbadger "github.com/ipfs/go-ds-badger"
	"github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	"github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-peerstore/pstoreds"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/pkg/logger"
)

// HostNode is an interface for hostNode
type HostNode interface {
	GetHost() host.Host
	GetNetMagic() uint32
	DisconnectPeer(p peer.ID) error
	GetPeerInfos() []peer.AddrInfo
	GetPeerDirection(id peer.ID) network.Direction
	GetPeerInfo(id peer.ID) *peer.AddrInfo
	RegisterHandler(message string, handler MessageHandler) error
	HandleStream(s network.Stream)
	SendMessage(id peer.ID, msg p2p.Message) error
	BroadcastMessage(msg p2p.Message)
}

var _ HostNode = &hostNode{}

// HostNode is the node for p2p host
// It's the low level P2P communication layer, the App class handles high level protocols
// The RPC communication is hanlded by App, not HostNode
type hostNode struct {
	host     host.Host
	ctx      context.Context
	datapath string
	netMagic uint32
	log      logger.Logger

	discover     *discover
	synchronizer *synchronizer
	handler      *handler
}

// NewHostNode creates a host node
func NewHostNode(blockchain chain.Blockchain) (HostNode, error) {
	ctx := config.GlobalParams.Context
	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams

	node := &hostNode{
		ctx:      ctx,
		log:      log,
		netMagic: netParams.NetMagic,
		datapath: config.GlobalFlags.DataPath,
	}

	ds, err := dsbadger.NewDatastore(path.Join(node.datapath, "peerstore"), nil)
	if err != nil {
		return nil, err
	}

	ps, err := pstoreds.NewPeerstore(node.ctx, ds, pstoreds.DefaultOpts())
	if err != nil {
		return nil, err
	}

	priv, err := node.loadPrivateKey()
	if err != nil {
		return nil, err
	}

	listenAddress, err := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/" + netParams.DefaultP2PPort)
	if err != nil {
		return nil, err
	}

	connman := connmgr.NewConnManager(2, 64, time.Second*60)
	h, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs([]ma.Multiaddr{listenAddress}...),
		libp2p.Identity(priv),
		libp2p.EnableRelay(circuit.OptActive, circuit.OptHop),
		libp2p.Peerstore(ps),
		libp2p.ConnectionManager(connman),
	)
	if err != nil {
		return nil, err
	}

	node.host = h

	addrs, err := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{
		ID:    h.ID(),
		Addrs: []ma.Multiaddr{listenAddress},
	})
	if err != nil {
		return nil, err
	}

	for _, a := range addrs {
		log.Infof("binding to address: %s", a)
	}

	synchronizer, err := NewSyncronizer(node, blockchain)
	if err != nil {
		return nil, err
	}
	node.synchronizer = synchronizer

	discovery, err := NewDiscover(node)
	if err != nil {
		return nil, err
	}
	node.discover = discovery

	handler, err := newHandler(params.ProtocolID, node)
	if err != nil {
		return nil, err
	}
	node.handler = handler

	return node, nil
}

// GetHost returns the host
func (node *hostNode) GetHost() host.Host {
	return node.host
}

func (node *hostNode) GetNetMagic() uint32 {
	return node.netMagic
}

// DisconnectPeer disconnects a peer
func (node *hostNode) DisconnectPeer(p peer.ID) error {
	return node.host.Network().ClosePeer(p)
}

// GetPeerInfos gets peer infos of connected hostnode.
func (node *hostNode) GetPeerInfos() []peer.AddrInfo {
	peers := node.host.Network().Peers()
	infos := make([]peer.AddrInfo, 0, len(peers))
	for _, p := range peers {
		addrInfo := node.host.Peerstore().PeerInfo(p)
		infos = append(infos, addrInfo)
	}

	return infos
}

// GetPeerDirection gets the direction of the peer.
func (node *hostNode) GetPeerDirection(id peer.ID) network.Direction {
	conns := node.host.Network().ConnsToPeer(id)

	if len(conns) != 1 {
		return network.DirUnknown
	}
	return conns[0].Stat().Direction
}

func (node *hostNode) GetPeerInfo(id peer.ID) *peer.AddrInfo {
	pinfo := node.host.Peerstore().PeerInfo(id)
	return &pinfo
}

func (node *hostNode) RegisterHandler(message string, handler MessageHandler) error {
	return node.handler.RegisterHandler(message, handler)
}

func (node *hostNode) HandleStream(s network.Stream) {
	node.handler.handleStream(s)
}

func (node *hostNode) SendMessage(id peer.ID, msg p2p.Message) error {
	return node.handler.SendMessage(id, msg)
}

func (node *hostNode) BroadcastMessage(msg p2p.Message)  {
	node.handler.BroadcastMessage(msg)
}

func (node *hostNode) loadPrivateKey() (crypto.PrivKey, error) {
	keyBytes, err := ioutil.ReadFile(path.Join(node.datapath, "node_key.dat"))
	if err != nil {
		return node.createPrivateKey()
	}

	key, err := crypto.UnmarshalPrivateKey(keyBytes)
	if err != nil {
		return node.createPrivateKey()
	}
	return key, nil
}

func (node *hostNode) createPrivateKey() (crypto.PrivKey, error) {
	_ = os.RemoveAll(path.Join(node.datapath, "node_key.dat"))

	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}

	keyBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(path.Join(node.datapath, "node_key.dat"), keyBytes, 0700)
	if err != nil {
		return nil, err
	}

	return priv, nil
}
