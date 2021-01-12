package host

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"time"
)

type synchronizer struct {
	host Host
	ctx  context.Context
	log  logger.Logger

	chain chain.Blockchain

	sync            bool
	withPeer        peer.ID
	blockStallTimer *time.Timer
}

func (sp *synchronizer) handleVersionMsg(id peer.ID, msg p2p.Message) error {
	theirVersion, ok := msg.(*p2p.MsgVersion)
	if !ok {
		return fmt.Errorf("did not receive version message")
	}

	sp.log.Infof("received version message from %s", id)

	// Send our version message if required
	ourVersion := sp.host.Version()
	direction := sp.host.GetPeerDirection(id)

	sp.host.AddPeerStats(id, theirVersion, direction)

	if direction == network.DirInbound {
		if err := sp.host.SendMessage(id, ourVersion); err != nil {
			return err
		}

	}

	return nil
}

// NewSynchronizer constructs a new sync protocol with a given host and chain.
func NewSynchronizer(host Host, chain chain.Blockchain) (*synchronizer, error) {

	sp := &synchronizer{
		host:  host,
		log:   config.GlobalParams.Logger,
		ctx:   config.GlobalParams.Context,
		chain: chain,
		sync:  false,
	}

	host.RegisterHandler(p2p.MsgVersionCmd, sp.handleVersionMsg)
	/*host.RegisterHandler(p2p.MsgGetBlocksCmd, sp.handleGetBlocksMsg)
	host.RegisterTopicHandler(p2p.MsgBlockCmd, sp.handleBlockMsg)
	host.RegisterHandler(p2p.MsgBlockCmd, sp.handleBlockMsg)

	go sp.initialBlockDownload()*/

	return sp, nil
}
