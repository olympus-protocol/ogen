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

func processMessages(ctx context.Context, net uint32, stream io.Reader, handler func(p2p.Message) (uint64, error)) (uint64, error) {
	for {
		select {
		case <-ctx.Done():
			return 0, nil
		default:
			break
		}

		msg, err := p2p.ReadMessage(stream, net)
		if err != nil {
			return 0, err
		}

		size, err := handler(msg)
		if err != nil {
			return 0, err
		}

		return size, nil
	}
}

func (h *host) receiveMessages(id peer.ID, r io.Reader) {
	size, err := processMessages(h.ctx, h.netMagic, r, func(message p2p.Message) (uint64, error) {
		cmd := message.Command()

		h.log.Tracef("processing message %s from peer %s", cmd, id)

		h.messageHandlersLock.Lock()
		defer h.messageHandlersLock.Unlock()

		handler, found := h.messageHandler[cmd]
		if !found {
			return 0, nil
		}

		size, err := handler(id, message)
		if err != nil {
			return 0, err
		}

		return size, nil
	})

	if err != nil {
		if !strings.Contains(err.Error(), "stream reset") {
			h.stats.IncreaseWrongMsgCount(id)
			h.log.Errorf("error receiving messages from peer %s: %s", id, err)
		}
	}

	h.stats.IncreasePeerReceivedBytes(id, size)
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

func (h *host) sendMessage(id peer.ID, msg p2p.Message) error {
	h.outgoingMessagesLock.Lock()
	msgChan, found := h.outgoingMessages[id]
	h.outgoingMessagesLock.Unlock()
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
		err := h.sendMessage(s.Conn().RemotePeer(), h.Version())
		if err != nil {
			h.log.Error(err)
		}
		go h.receiveMessages(s.Conn().RemotePeer(), s)
	}
}
