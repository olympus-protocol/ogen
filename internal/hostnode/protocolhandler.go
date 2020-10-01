package hostnode

import (
	"bytes"
	"context"
	"fmt"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"io"
	"strings"
	"sync"

	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
)

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

	log      logger.Logger
	finTopic *pubsub.Topic
}

// newProtocolHandler constructs a new protocol handler for a specific protocol ID.
func newProtocolHandler(id protocol.ID, host HostNode) (*protocolHandler, error) {
	finTopic, err := host.Topic(p2p.MsgFinalizationCmd)
	if err != nil {
		return nil, err
	}
	ph := &protocolHandler{
		ID:               id,
		host:             host,
		messageHandlers:  make(map[string]MessageHandler),
		outgoingMessages: make(map[peer.ID]chan p2p.Message),
		ctx:              config.GlobalParams.Context,
		log:              config.GlobalParams.Logger,
		finTopic:         finTopic,
	}

	host.SetStreamHandler(id, ph.HandleStream)

	return ph, nil
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
				_ = p.host.DisconnectPeer(id)
			}
		}
	}()
}

func (p *protocolHandler) HandleStream(s network.Stream) {
	if s != nil {
		p.sendMessages(s.Conn().RemotePeer(), s)
		p.log.Tracef("handling messages from peer %s for protocol %s", s.Conn().RemotePeer(), p.ID)
		go p.receiveMessages(s.Conn().RemotePeer(), s)
	}
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

func (p *protocolHandler) SendFinalizedMessage(msg *p2p.MsgFinalization) error {
	buf := bytes.NewBuffer([]byte{})
	err := p2p.WriteMessage(buf, msg, p.host.GetNetMagic())
	if err != nil {
		return err
	}
	return p.finTopic.Publish(p.ctx, buf.Bytes())
}
