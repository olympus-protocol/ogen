package mempool

import (
	"errors"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/pkg/p2p"
)

func (p *pool) handleVote(id peer.ID, msg p2p.Message) error {

	if id == p.host.ID() {
		return nil
	}

	p.host.IncreasePeerReceivedBytes(id, msg.PayloadLength())

	data, ok := msg.(*p2p.MsgVote)
	if !ok {
		return errors.New("wrong message on vote topic")
	}

	vote := data.Data

	firstSlotAllowedToInclude := vote.Data.Slot + p.netParams.MinAttestationInclusionDelay
	tip := p.chain.State().Tip()

	if tip.Slot+p.netParams.EpochLength*2 < firstSlotAllowedToInclude {
		return nil
	}

	view, err := p.chain.State().GetSubView(tip.Hash)
	if err != nil {
		p.log.Warnf("could not get block view representing current tip: %s", err)
		return err
	}

	currentState, _, err := p.chain.State().GetStateForHashAtSlot(tip.Hash, firstSlotAllowedToInclude, &view)
	if err != nil {
		p.log.Warnf("error updating chain to attestation inclusion slot: %s", err)
		return err
	}
	p.log.Debugf("received vote from %s with %d votes", id, len(data.Data.ParticipationBitfield.BitIndices()))
	err = p.AddVote(data.Data, currentState)
	if err != nil {
		return err
	}

	return nil
}

func (p *pool) handleDeposits(id peer.ID, msg p2p.Message) error {
	if id == p.host.ID() {
		return nil
	}

	p.host.IncreasePeerReceivedBytes(id, msg.PayloadLength())

	data, ok := msg.(*p2p.MsgDeposits)
	if !ok {
		return errors.New("wrong message on deposits topic")
	}

	for _, d := range data.Data {
		err := p.AddDeposit(d)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *pool) handleExits(id peer.ID, msg p2p.Message) error {

	if id == p.host.ID() {
		return nil
	}

	p.host.IncreasePeerReceivedBytes(id, msg.PayloadLength())

	data, ok := msg.(*p2p.MsgExits)
	if !ok {
		return errors.New("wrong message on exits topic")
	}

	for _, d := range data.Data {

		err := p.AddExit(d)
		if err != nil {
			return err
		}

	}

	return nil
}

func (p *pool) handlePartialExits(id peer.ID, msg p2p.Message) error {

	if id == p.host.ID() {
		return nil
	}

	p.host.IncreasePeerReceivedBytes(id, msg.PayloadLength())

	data, ok := msg.(*p2p.MsgPartialExits)
	if !ok {
		return errors.New("wrong message on proofs topic")
	}

	for _, d := range data.Data {
		err := p.AddPartialExit(d)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *pool) handleTx(id peer.ID, msg p2p.Message) error {
	if id == p.host.ID() {
		return nil
	}

	p.host.IncreasePeerReceivedBytes(id, msg.PayloadLength())

	data, ok := msg.(*p2p.MsgTx)
	if !ok {
		return errors.New("wrong message on tx topic")
	}

	err := p.AddTx(data.Data)

	if err != nil {
		return err
	}

	return nil
}
