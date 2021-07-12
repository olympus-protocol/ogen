package server

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/debug"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"strings"
)

func (s *server) apis() []rpc.API {
	return []rpc.API{
		{
			Namespace: "admin",
			Version:   "1.0",
			Service:   &privateAdminAPI{s.log, s},
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   &publicAdminAPI{s.log, s},
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   debug.Handler,
		}, {
			Namespace: "web3",
			Version:   "1.0",
			Service:   &publicWeb3API{s.log, s},
			Public:    true,
		},
	}
}

// privateAdminAPI is the collection of administrative API methods exposed only
// over a secure RPC channel.
type privateAdminAPI struct {
	log  logger.Logger
	node *server
}

// AddPeer requests connecting to a remote node, and also maintaining the new
// connection at all times, even reconnecting if it is lost.
func (api *privateAdminAPI) AddPeer(url string) (bool, error) {
	// Make sure the server is running, fail otherwise
	/*server := api.node.Server()
	if server == nil {
		return false, ErrNodeStopped
	}
	// Try to add the url as a static peer and return
	node, err := enode.Parse(enode.ValidSchemes, url)
	if err != nil {
		return false, fmt.Errorf("invalid enode: %v", err)
	}
	server.AddPeer(node)*/
	return true, nil
}

// RemovePeer disconnects from a remote node if the connection exists
func (api *privateAdminAPI) RemovePeer(url string) (bool, error) {
	// Make sure the server is running, fail otherwise
	/*server := api.node.Server()
	if server == nil {
		return false, ErrNodeStopped
	}
	// Try to remove the url as a static peer and return
	node, err := enode.Parse(enode.ValidSchemes, url)
	if err != nil {
		return false, fmt.Errorf("invalid enode: %v", err)
	}
	server.RemovePeer(node)*/
	return true, nil
}

// AddTrustedPeer allows a remote node to always connect, even if slots are full
func (api *privateAdminAPI) AddTrustedPeer(url string) (bool, error) {
	// Make sure the server is running, fail otherwise
	/*server := api.node.Server()
	if server == nil {
		return false, ErrNodeStopped
	}
	node, err := enode.Parse(enode.ValidSchemes, url)
	if err != nil {
		return false, fmt.Errorf("invalid enode: %v", err)
	}
	server.AddTrustedPeer(node)*/
	return true, nil
}

// RemoveTrustedPeer removes a remote node from the trusted peer set, but it
// does not disconnect it automatically.
func (api *privateAdminAPI) RemoveTrustedPeer(url string) (bool, error) {
	// Make sure the server is running, fail otherwise
	/*server := api.node.Server()
	if server == nil {
		return false, ErrNodeStopped
	}
	node, err := enode.Parse(enode.ValidSchemes, url)
	if err != nil {
		return false, fmt.Errorf("invalid enode: %v", err)
	}
	server.RemoveTrustedPeer(node)*/
	return true, nil
}

// PeerEvents creates an RPC subscription which receives peer events from the
// node's p2p.Server
func (api *privateAdminAPI) PeerEvents(ctx context.Context) (*rpc.Subscription, error) {
	// Make sure the server is running, fail otherwise
	/*server := api.node.Server()
	if server == nil {
		return nil, ErrNodeStopped
	}*/

	// Create the subscription
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return nil, rpc.ErrNotificationsUnsupported
	}
	rpcSub := notifier.CreateSubscription()

	/*go func() {
		events := make(chan *p2p.PeerEvent)
		sub := server.SubscribeEvents(events)
		defer sub.Unsubscribe()

		for {
			select {
			case event := <-events:
				notifier.Notify(rpcSub.ID, event)
			case <-sub.Err():
				return
			case <-rpcSub.Err():
				return
			case <-notifier.Closed():
				return
			}
		}
	}()*/

	return rpcSub, nil
}

// StartHTTP starts the HTTP RPC API server.
func (api *privateAdminAPI) StartHTTP(host *string, port *int, cors *string, apis *string, vhosts *string) (bool, error) {
	api.node.lock.Lock()
	defer api.node.lock.Unlock()

	// Determine host and port.
	if host == nil {
		h := "localhost"
		if config.GlobalFlags.HTTPHost != "" {
			h = config.GlobalFlags.HTTPHost
		}
		host = &h
	}
	if port == nil {
		port = &config.GlobalFlags.HTTPPort
	}

	// Determine config.
	cfg := httpConfig{
		CorsAllowedOrigins: config.GlobalFlags.HTTPCors,
		Vhosts:             config.GlobalFlags.HTTPVirtualHosts,
		Modules:            config.GlobalFlags.HTTPModules,
	}
	if cors != nil {
		cfg.CorsAllowedOrigins = nil
		for _, origin := range strings.Split(*cors, ",") {
			cfg.CorsAllowedOrigins = append(cfg.CorsAllowedOrigins, strings.TrimSpace(origin))
		}
	}
	if vhosts != nil {
		cfg.Vhosts = nil
		for _, vhost := range strings.Split(*host, ",") {
			cfg.Vhosts = append(cfg.Vhosts, strings.TrimSpace(vhost))
		}
	}
	if apis != nil {
		cfg.Modules = nil
		for _, m := range strings.Split(*apis, ",") {
			cfg.Modules = append(cfg.Modules, strings.TrimSpace(m))
		}
	}

	if err := api.node.http.setListenAddr(*host, *port); err != nil {
		return false, err
	}
	if err := api.node.http.enableRPC(api.node.rpcAPIs, cfg); err != nil {
		return false, err
	}
	if err := api.node.http.start(); err != nil {
		return false, err
	}
	return true, nil
}

// StopHTTP shuts down the HTTP server.
func (api *privateAdminAPI) StopHTTP() (bool, error) {
	api.node.http.stop()
	return true, nil
}

// StartWS starts the websocket RPC API server.
func (api *privateAdminAPI) StartWS(host *string, port *int, allowedOrigins *string, apis *string) (bool, error) {
	api.node.lock.Lock()
	defer api.node.lock.Unlock()

	// Determine host and port.
	if host == nil {
		h := "localhost"
		if config.GlobalFlags.WSHost != "" {
			h = config.GlobalFlags.WSHost
		}
		host = &h
	}
	if port == nil {
		port = &config.GlobalFlags.WSPort
	}

	// Determine config.
	cfg := wsConfig{
		Modules: config.GlobalFlags.WSModules,
		Origins: config.GlobalFlags.WSOrigins,
		// ExposeAll: api.node.config.WSExposeAll,
	}
	if apis != nil {
		cfg.Modules = nil
		for _, m := range strings.Split(*apis, ",") {
			cfg.Modules = append(cfg.Modules, strings.TrimSpace(m))
		}
	}
	if allowedOrigins != nil {
		cfg.Origins = nil
		for _, origin := range strings.Split(*allowedOrigins, ",") {
			cfg.Origins = append(cfg.Origins, strings.TrimSpace(origin))
		}
	}

	// Enable WebSocket on the server.
	server := api.node.wsServerForPort(*port)
	if err := server.setListenAddr(*host, *port); err != nil {
		return false, err
	}
	if err := server.enableWS(api.node.rpcAPIs, cfg); err != nil {
		return false, err
	}
	if err := server.start(); err != nil {
		return false, err
	}
	return true, nil
}

// StopWS terminates all WebSocket servers.
func (api *privateAdminAPI) StopWS() (bool, error) {
	api.node.http.stopWS()
	api.node.ws.stop()
	return true, nil
}

// publicAdminAPI is the collection of administrative API methods exposed over
// both secure and unsecure RPC channels.
type publicAdminAPI struct {
	log  logger.Logger
	node *server // Node interfaced by this API
}

// Peers retrieves all the information we know about each individual peer at the
// protocol granularity.
func (api *publicAdminAPI) Peers() ([]*p2p.PeerInfo, error) {
	return []*p2p.PeerInfo{}, nil
}

// NodeInfo retrieves all the information we know about the host node at the
// protocol granularity.
func (api *publicAdminAPI) NodeInfo() (*p2p.NodeInfo, error) {
	return &p2p.NodeInfo{}, nil
}

// Datadir retrieves the current data directory the node is using.
func (api *publicAdminAPI) Datadir() string {
	return config.GlobalFlags.DataPath
}

// publicWeb3API offers helper utils
type publicWeb3API struct {
	log   logger.Logger
	stack *server
}

// ClientVersion returns the node name
func (s *publicWeb3API) ClientVersion() string {
	return config.GlobalParams.NetParams.Name
}

// Sha3 applies the ethereum sha3 implementation on the input.
// It assumes the input is hex encoded.
func (s *publicWeb3API) Sha3(input hexutil.Bytes) hexutil.Bytes {
	return crypto.Keccak256(input)
}
