package hostnode

import (
	"context"
	"fmt"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"io"
	"strings"
	"sync"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
)

// MessageHandler is a handler for a specific message.
type MessageHandler func(id peer.ID, msg p2p.Message) error

// handler handles all of the peers messages.
type handler struct {
	// ID is the protocol being handled.
	ID protocol.ID

	// host is the host to connect to.
	host HostNode

	messageHandler      map[string]MessageHandler
	messageHandlersLock sync.Mutex

	topicHandlers     map[string]MessageHandler
	topicHandlersLock sync.Mutex

	outgoingMessages     map[peer.ID]chan p2p.Message
	outgoingMessagesLock sync.Mutex

	ctx context.Context

	log logger.Logger
}

// newHandler constructs a new handler for a specific protocol ID.
func newHandler(id protocol.ID, host HostNode) (*handler, error) {
	ph := &handler{
		ID:               id,
		host:             host,
		topicHandlers:    make(map[string]MessageHandler),
		messageHandler:   make(map[string]MessageHandler),
		outgoingMessages: make(map[peer.ID]chan p2p.Message),
		ctx:              config.GlobalParams.Context,
		log:              config.GlobalParams.Logger,
	}

	host.GetHost().SetStreamHandler(id, ph.handleStream)

	return ph, nil
}

// RegisterHandler registers a handler for a protocol.
func (p *handler) RegisterHandler(messageName string, handler MessageHandler) error {
	p.messageHandlersLock.Lock()
	defer p.messageHandlersLock.Unlock()
	if _, found := p.messageHandler[messageName]; found {
		return fmt.Errorf("handler for message name %s already exists", messageName)
	}

	p.messageHandler[messageName] = handler
	return nil
}

// RegisterTopicHandler registers a handler for a protocol.
func (p *handler) RegisterTopicHandler(messageName string, handler MessageHandler) error {
	p.topicHandlersLock.Lock()
	defer p.topicHandlersLock.Unlock()
	if _, found := p.topicHandlers[messageName]; found {
		return fmt.Errorf("handler for message name %s already exists", messageName)
	}

	p.topicHandlers[messageName] = handler
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

func (p *handler) receiveMessages(id peer.ID, r io.Reader) {
	err := processMessages(p.ctx, p.host.GetNetMagic(), r, func(message p2p.Message) error {
		cmd := message.Command()

		p.log.Tracef("processing message %s from peer %s", cmd, id)

		p.messageHandlersLock.Lock()
		defer p.messageHandlersLock.Unlock()

		if handler, found := p.messageHandler[cmd]; found {
			err := handler(id, message)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		if !strings.Contains(err.Error(), "stream reset") {
			p.log.Errorf("error receiving messages from peer %s: %s", id, err)
		}

	}
}

func (p *handler) sendMessages(id peer.ID, w io.Writer) {
	msgChan := make(chan p2p.Message)

	p.outgoingMessagesLock.Lock()
	p.outgoingMessages[id] = msgChan
	p.outgoingMessagesLock.Unlock()

	go func() {
		for msg := range msgChan {
			err := p2p.WriteMessage(w, msg, p.host.GetNetMagic())
			if err != nil {
				p.log.Errorf("error sending message to peer %s: %s", id, err)
				_ = p.host.DisconnectPeer(id)
			}
		}
	}()
}

func (p *handler) handleStream(s network.Stream) {
	if s != nil {
		p.sendMessages(s.Conn().RemotePeer(), s)
		p.log.Tracef("handling messages from peer %s for protocol %s", s.Conn().RemotePeer(), p.ID)
		go p.receiveMessages(s.Conn().RemotePeer(), s)
	}
}

// SendMessage writes a message to a peer.
func (p *handler) SendMessage(id peer.ID, msg p2p.Message) error {
	p.outgoingMessagesLock.Lock()
	msgsChan, found := p.outgoingMessages[id]
	p.outgoingMessagesLock.Unlock()
	if !found {
		return fmt.Errorf("not tracking peer %s", id)
	}
	msgsChan <- msg
	return nil
}
