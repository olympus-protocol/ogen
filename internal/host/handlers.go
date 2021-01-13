package host

import (
	"bytes"
	"context"
	"errors"
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

		var handler MessageHandler
		switch cmd {
		case p2p.MsgVersionCmd:
			handler = h.handleVersionMsg
		case p2p.MsgGetBlocksCmd:
			handler = h.handleGetBlocksMsg
		case p2p.MsgBlockCmd:
			handler = h.handleBlockMsg
		default:
			h.log.Tracef("received unknown msg %s", cmd)
			return nil
		}

		err := handler(id, message)
		if err != nil {
			return err
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

func (h *host) handleVersionMsg(id peer.ID, msg p2p.Message) error {
	theirVersion, ok := msg.(*p2p.MsgVersion)
	if !ok {
		return fmt.Errorf("did not receive version message")
	}

	h.log.Infof("received version message from %s", id)

	// Send our version message if required
	ourVersion := h.Version()
	direction := h.GetPeerDirection(id)

	h.AddPeerStats(id, theirVersion, direction)

	if direction == network.DirInbound {
		if err := h.SendMessage(id, ourVersion); err != nil {
			return err
		}

	}

	return nil
}

func (h *host) handleGetBlocksMsg(id peer.ID, rawMsg p2p.Message) error {
	msg, ok := rawMsg.(*p2p.MsgGetBlocks)
	if !ok {
		return errors.New("did not receive get blocks message")
	}

	h.log.Debug("received getblocks")

	// Get the announced last block to make sure we have a common point
	firstCommon, ok := h.chain.State().Index().Get(msg.LastBlockHash)
	if !ok {
		err := fmt.Sprintf("unable to find common point for peer %s", id)
		h.log.Error(err)
		return nil
	}

	blockRow, ok := h.chain.State().Chain().Next(firstCommon)
	if !ok {
		err := fmt.Sprintf("unable to next block from common point for peer %s", id)
		h.log.Error(err)
		return nil
	}

	for {

		block, err := h.chain.GetBlock(blockRow.Hash)
		if err != nil {
			return nil
		}

		err = h.SendMessage(id, &p2p.MsgBlock{
			Data: block,
		})

		if err != nil {
			return nil
		}

		blockRow, ok = h.chain.State().Chain().Next(blockRow)
		if !ok {
			break
		}

	}

	return nil
}

func (h *host) handleBlockMsg(id peer.ID, msg p2p.Message) error {
	block, ok := msg.(*p2p.MsgBlock)
	if !ok {
		return errors.New("non block msg")
	}

	h.IncreasePeerReceivedBytes(id, msg.PayloadLength())

	err := h.synchronizer.handleBlock(id, block.Data)
	if err != nil {
		return err
	}

	return nil
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

func (h *host) listenTopics() {
	for {
		msg, err := h.topicSub.Next(h.ctx)
		if err != nil {
			continue
		}

		if msg.GetFrom() == h.host.ID() {
			continue
		}

		buf := bytes.NewBuffer(msg.Data)

		msgData, err := p2p.ReadMessage(buf, h.netMagic)
		if err != nil {
			h.log.Warnf("unable to decode message: %s", err)
			continue
		}

		cmd := msgData.Command()

		h.topicHandlersLock.Lock()
		handler, found := h.topicHandlers[cmd]
		h.topicHandlersLock.Unlock()
		if !found {
			continue
		}

		err = handler(msg.GetFrom(), msgData)
		if err != nil {
			h.log.Error(err)
		}

	}
}

func (h *host) handleFinalizationMsg(id peer.ID, msg p2p.Message) error {

	fin, ok := msg.(*p2p.MsgFinalization)
	if !ok {
		return errors.New("non block msg")
	}

	h.IncreasePeerReceivedBytes(id, msg.PayloadLength())

	if h.ID() == id {
		return nil
	}

	return h.stats.Update(id, fin)
}
