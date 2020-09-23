package hostnode

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
)

// ProtocolHandler is an interface for the ProtocolHandler
type ProtocolHandler interface {
	RegisterHandler(messageName string, handler MessageHandler) error
	receiveMessages(id peer.ID, r io.Reader)
	SendMessage(toPeer peer.ID, msg p2p.Message) error
	Listen(network.Network, multiaddr.Multiaddr)
	ListenClose(network.Network, multiaddr.Multiaddr)
	Connected(net network.Network, conn network.Conn)
	Disconnected(net network.Network, conn network.Conn)
	OpenedStream(network.Network, network.Stream)
	ClosedStream(network.Network, network.Stream)
	Notify(n ConnectionManagerNotifee)
	StopNotify(n ConnectionManagerNotifee)
	HandleStream(s network.Stream)
}

var _ ProtocolHandler = &protocolHandler{}

// MessageHandler is a handler for a specific message.
type MessageHandler func(id peer.ID, msg p2p.Message) error

// protocolHandler handles all of the messages, discovery, and shut down for each protocol.
type protocolHandler struct {
	// ID is the protocol being handled.
	ID protocol.ID

	// host is the host to connect to.
	host HostNode

	discovery discovery.Discovery

	messageHandlers     map[string]MessageHandler
	messageHandlersLock sync.RWMutex

	outgoingMessages     map[peer.ID]chan p2p.Message
	outgoingMessagesLock sync.RWMutex

	ctx context.Context

	notifees    []ConnectionManagerNotifee
	notifeeLock sync.Mutex

	log logger.Logger
}

// ConnectionManagerNotifee is a notifee for the connection manager.
type ConnectionManagerNotifee interface {
	PeerConnected(peer.ID, network.Direction)
	PeerDisconnected(peer.ID)
}

// newProtocolHandler constructs a new protocol handler for a specific protocol ID.
func newProtocolHandler(ctx context.Context, id protocol.ID, host HostNode, config Config) ProtocolHandler {
	ph := &protocolHandler{
		ID:               id,
		host:             host,
		messageHandlers:  make(map[string]MessageHandler),
		outgoingMessages: make(map[peer.ID]chan p2p.Message),
		ctx:              ctx,
		notifees:         make([]ConnectionManagerNotifee, 0),
		log:              config.Log,
	}

	host.SetStreamHandler(id, ph.HandleStream)
	host.Notify(ph)

	return ph
}

// RegisterHandler registers a handler for a protocol.
func (p *protocolHandler) RegisterHandler(messageName string, handler MessageHandler) error {
	p.messageHandlersLock.Lock()
	defer p.messageHandlersLock.Unlock()
	if _, found := p.messageHandlers[messageName]; found {
		return fmt.Errorf("handler for message name %s already exists", messageName)
	}

	p.messageHandlers[messageName] = handler
	return nil
}

// processMessages continuously reads from stream and handles any protobuf messages.
func processMessages(ctx context.Context, net uint32, stream io.Reader, handler func(p2p.Message) error) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			break
		}
		msg, err := p2p.ReadMessage(stream, net)
		if err != nil {
			return err
		}

		if err := handler(msg); err != nil {
			return err
		}
	}
}

func (p *protocolHandler) receiveMessages(id peer.ID, r io.Reader) {
	err := processMessages(p.ctx, p.host.GetNetMagic(), r, func(message p2p.Message) error {
		cmd := message.Command()

		p.log.Tracef("processing message %s from peer %s", cmd, id)

		p.messageHandlersLock.RLock()
		if handler, found := p.messageHandlers[cmd]; found {
			p.messageHandlersLock.RUnlock()
			err := handler(id, message)
			if err != nil {
				return err
			}
		} else {
			p.messageHandlersLock.RUnlock()
		}
		return nil
	})
	if err != nil {
		p.notifeeLock.Lock()
		for _, n := range p.notifees {
			n.PeerDisconnected(id)
		}
		p.notifeeLock.Unlock()
		if !strings.Contains(err.Error(), "stream reset") {
			p.log.Errorf("error receiving messages from peer %s: %s", id, err)
		}

	}
}

func (p *protocolHandler) sendMessages(id peer.ID, w io.Writer) {
	msgChan := make(chan p2p.Message)

	p.outgoingMessagesLock.Lock()
	p.outgoingMessages[id] = msgChan
	p.outgoingMessagesLock.Unlock()

	go func() {
		for msg := range msgChan {
			err := p2p.WriteMessage(w, msg, p.host.GetNetMagic())
			if err != nil {
				p.log.Errorf("error sending message to peer %s: %s", id, err)

				p.notifeeLock.Lock()
				for _, n := range p.notifees {
					n.PeerDisconnected(id)
				}
				p.notifeeLock.Unlock()

				_ = p.host.DisconnectPeer(id)
			}
		}
	}()
}

func (p *protocolHandler) HandleStream(s network.Stream) {
	p.sendMessages(s.Conn().RemotePeer(), s)

	p.log.Tracef("handling messages from peer %s for protocol %s", s.Conn().RemotePeer(), p.ID)
	go p.receiveMessages(s.Conn().RemotePeer(), s)

	p.notifeeLock.Lock()
	for _, n := range p.notifees {
		n.PeerConnected(s.Conn().RemotePeer(), s.Stat().Direction)
	}
	p.notifeeLock.Unlock()
}

// SendMessage writes a message to a peer.
func (p *protocolHandler) SendMessage(toPeer peer.ID, msg p2p.Message) error {
	p.outgoingMessagesLock.RLock()
	msgsChan, found := p.outgoingMessages[toPeer]
	p.outgoingMessagesLock.RUnlock()
	if !found {
		return fmt.Errorf("not tracking peer %s", toPeer)
	}
	msgsChan <- msg
	return nil
}

// Listen is called when we start listening on an address.
func (p *protocolHandler) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when we stop listening on an address.
func (p *protocolHandler) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when we connect to a peer.
func (p *protocolHandler) Connected(net network.Network, conn network.Conn) {}

// Disconnected is called when we disconnect to a peer.
func (p *protocolHandler) Disconnected(net network.Network, conn network.Conn) {
	peerID := conn.RemotePeer()

	if net.Connectedness(peerID) == network.NotConnected {
		p.outgoingMessagesLock.Lock()
		defer p.outgoingMessagesLock.Unlock()
		if handler, found := p.outgoingMessages[peerID]; found {
			close(handler)
			delete(p.outgoingMessages, peerID)
		}
	}
}

// OpenedStream is called when we open a stream to a peer.
func (p *protocolHandler) OpenedStream(network.Network, network.Stream) {}

// ClosedStream is called when we close a stream to a peer.
func (p *protocolHandler) ClosedStream(network.Network, network.Stream) {}

// Notify notifies a specific notifier when certain events happen.
func (p *protocolHandler) Notify(n ConnectionManagerNotifee) {
	p.notifeeLock.Lock()
	p.notifees = append(p.notifees, n)
	p.notifeeLock.Unlock()
}

// StopNotify stops notifying a certain notifee about certain events.
func (p *protocolHandler) StopNotify(n ConnectionManagerNotifee) {
	p.notifeeLock.Lock()
	found := -1
	for i, notif := range p.notifees {
		if notif == n {
			found = i
			break
		}
	}
	if found != -1 {
		p.notifees = append(p.notifees[:found], p.notifees[found+1:]...)
	}
	p.notifeeLock.Unlock()
}
