package host

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"io"
	"strings"
)

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

func (h *host) receiveMessages(id peer.ID, r io.Reader) {
	err := processMessages(h.ctx, h.netMagic, r, func(message p2p.Message) error {
		cmd := message.Command()

		h.log.Tracef("processing message %s from peer %s", cmd, id)

		h.messageHandlersLock.Lock()
		defer h.messageHandlersLock.Unlock()

		if handler, found := h.messageHandler[cmd]; found {
			err := handler(id, message)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		if !strings.Contains(err.Error(), "stream reset") {
			h.stats.IncreaseWrongMsgCount(id)
			h.log.Errorf("error receiving messages from peer %s: %s", id, err)
		}
	}
}

func (h *host) sendMessages(id peer.ID, w io.Writer) {
	msgChan := make(chan p2p.Message)

	h.outgoingMessagesLock.Lock()
	h.outgoingMessages[id] = msgChan
	h.outgoingMessagesLock.Unlock()

	go func() {
		for msg := range msgChan {
			err := p2p.WriteMessage(w, msg, h.netMagic)
			if err != nil {
				h.log.Errorf("error sending message to peer %s: %s", id, err)
			}
		}
	}()
}

func (h *host) SendMessage(id peer.ID, msg p2p.Message) error {
	h.outgoingMessagesLock.Lock()
	defer h.outgoingMessagesLock.Unlock()
	msgChan, found := h.outgoingMessages[id]
	if !found {
		return fmt.Errorf("not tracking peer %s", id)
	}
	msgChan <- msg
	h.stats.IncreasePeerSentBytes(id, msg.PayloadLength())
	return nil
}

func (h *host) handleStream(s network.Stream) {
	if s != nil {
		h.sendMessages(s.Conn().RemotePeer(), s)
		h.log.Tracef("handling messages from peer %s for protocol %s", s.Conn().RemotePeer(), s.Protocol())
		go h.receiveMessages(s.Conn().RemotePeer(), s)
	}
}
