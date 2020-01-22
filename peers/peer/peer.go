package peer

import (
	"bytes"
	"github.com/grupokindynos/ogen/logger"
	"github.com/grupokindynos/ogen/p2p"
	"github.com/grupokindynos/ogen/utils/chainhash"
	"github.com/grupokindynos/ogen/utils/serializer"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Status int

const (
	Handshaked Status = iota
	Syncing
	Disconnect
	Ban
)

type BlocksInvMsg struct {
	Peer   *Peer
	Blocks *p2p.MsgBlockInv
}

func newBlocksInvMsg(p *Peer, blocks *p2p.MsgBlockInv) *BlocksInvMsg {
	blockInvMsg := &BlocksInvMsg{
		Peer:   p,
		Blocks: blocks,
	}
	return blockInvMsg
}

type DataReqMsg struct {
	Peer    *Peer
	Request string
	Payload *chainhash.Hash
}

func newDataRequestMsg(p *Peer, req string, payload *chainhash.Hash) *DataReqMsg {
	dataReqMsg := &DataReqMsg{
		Peer:    p,
		Request: req,
		Payload: payload,
	}
	return dataReqMsg
}

type TxMsg struct {
	Peer *Peer
	Tx   *p2p.MsgTx
}

func newTxMsg(p *Peer, Tx *p2p.MsgTx) *TxMsg {
	txMsg := &TxMsg{
		Peer: p,
		Tx:   Tx,
	}
	return txMsg
}

type BlockMsg struct {
	Peer  *Peer
	Block *p2p.MsgBlock
}

func newBlockMsg(p *Peer, block *p2p.MsgBlock) *BlockMsg {
	blockMsg := &BlockMsg{
		Peer:  p,
		Block: block,
	}
	return blockMsg
}

type PeerMsg struct {
	Peer   *Peer
	Status Status
}

func newPeerMsg(p *Peer, status Status) *PeerMsg {
	peerMsg := &PeerMsg{
		Peer:   p,
		Status: status,
	}
	return peerMsg
}

type Peer struct {
	// Ogen main config
	NetMagic p2p.NetMagic

	// Pass the log to debug easily
	log *logger.Logger

	// Version Msg Data
	protocol  int32
	services  p2p.ServiceFlag
	lastBlock int32
	userAgent string

	// Peer properties
	id             int
	inbound        bool
	outbound       bool
	address        serializer.NetAddress
	conn           net.Conn
	connectionTime time.Time
	lastPingTime   int64
	lastPingNonce  uint64
	banScore       int32

	// Peer dinamic properties
	bytesReceived  uint64
	verackReceived bool

	// Internal Chan
	closeSignal chan interface{}

	// Chans for PeerMan Communication
	ManChan       chan interface{}
	mainCloseChan chan interface{}

	// Locks for safe usage
	peerLock sync.RWMutex

	// For syncMan sync peer selector
	selectedForSync bool
}

func (p *Peer) Start(lastBlockHeight int32) {
	// First version handshake
	err := p.versionHandshake(lastBlockHeight)
	if err != nil {
		p.log.Errorf("disconnecting peer %v", p.GetID())
		p.ManChan <- newPeerMsg(p, Disconnect)
	}
	p.ManChan <- newPeerMsg(p, Handshaked)
	// Init message listener
	go func() {
		err := p.messageListener()
		if err != nil {
			p.log.Errorf("disconnecting peer %v", p.GetID())
			p.ManChan <- newPeerMsg(p, Disconnect)
		}
	}()
	// Init ping/pong
	go func() {
		err := p.pingRoutine()
		if err != nil {
			p.log.Errorf("disconnecting peer %v", p.GetID())
			p.ManChan <- newPeerMsg(p, Disconnect)
		}
	}()
}

func (p *Peer) versionHandshake(lastBlockHeight int32) error {
	if p.inbound {
		err := p.handleInboundPeerHandshake(lastBlockHeight)
		if err != nil {
			p.log.Errorf("unable to perform inbound handshake for peer: %v", p.GetID())
			return err
		}
	}
	if !p.inbound {
		err := p.handleOutboundPeerHandshake(lastBlockHeight)
		if err != nil {
			p.log.Errorf("unable to perform inbound handshake for peer: %v", p.GetID())
			return err
		}
	}
	return nil
}

func (p *Peer) handleInboundPeerHandshake(lastBlockHeight int32) error {
	remoteMsgVersion, _, err := p.readMessage()
	if err != nil {
		return ErrorReadRemote
	}
	msgVersion, ok := remoteMsgVersion.(*p2p.MsgVersion)
	if !ok {
		return ErrorNoVersionFirst
	}
	p.updateStats(msgVersion)
	verack := p2p.NewMsgVerack()
	err = p.writeMessage(verack)
	if err != nil {
		return ErrorWriteRemote
	}
	err = p.writeMessage(p.versionMsg(lastBlockHeight))
	if err != nil {
		return ErrorWriteRemote
	}
	remoteMsgVerack, _, err := p.readMessage()
	if err != nil {
		return ErrorReadRemote
	}
	_, ok = remoteMsgVerack.(*p2p.MsgVerack)
	if !ok {
		return ErrorNoVerackAfterVersion
	}
	p.verackReceived = true
	return nil
}

func (p *Peer) handleOutboundPeerHandshake(lastBlockHeight int32) error {
	err := p.writeMessage(p.versionMsg(lastBlockHeight))
	if err != nil {
		return ErrorWriteRemote
	}
	remoteVerack, _, _ := p.readMessage()
	_, ok := remoteVerack.(*p2p.MsgVerack)
	if !ok {
		return ErrorNoVerackAfterVersion
	}
	remoteMsgVersion, _, _ := p.readMessage()
	msgVersion, ok := remoteMsgVersion.(*p2p.MsgVersion)
	if !ok {
		return ErrorNoVersionAfterVerack
	}
	p.updateStats(msgVersion)
	verack := p2p.NewMsgVerack()
	err = p.writeMessage(verack)
	if err != nil {
		return ErrorWriteRemote
	}
	return nil
}

func (p *Peer) updateStats(msgVersion *p2p.MsgVersion) {
	p.peerLock.Lock()
	p.protocol = msgVersion.ProtocolVersion
	p.services = msgVersion.Services
	p.lastBlock = msgVersion.LastBlock
	p.userAgent = msgVersion.UserAgent
	p.peerLock.Unlock()
}

func (p *Peer) pingRoutine() error {
	p.log.Tracef("starting ping routine for peer %v", p.GetID())
	for {
		select {
		case <-p.closeSignal:
			break
		default:
			time.Sleep(15 * time.Second)
			err := p.ping()
			if err != nil {
				p.log.Errorf("unable to ping peer %v", p.GetID())
				return err
			}
		}
	}
}

func (p *Peer) ping() error {
	msgPing := p2p.NewMsgPing()
	p.lastPingTime = time.Now().Unix()
	p.lastPingNonce = msgPing.Nonce
	p.log.Tracef("sending ping msg to peer %v", p.GetID())
	err := p.writeMessage(msgPing)
	if err != nil {
		return ErrorWriteRemote
	}
	return nil
}

func (p *Peer) pong(nonce uint64) error {
	p.log.Tracef("sent pong msg to peer %v", p.GetID())
	msgPong := p2p.NewMsgPong(nonce)
	err := p.writeMessage(msgPong)
	if err != nil {
		return ErrorWriteRemote
	}
	return nil
}

func (p *Peer) versionMsg(lastBlockHeight int32) *p2p.MsgVersion {
	meInformation := strings.Split(p.conn.LocalAddr().String(), ":")
	meIP, mePortString := meInformation[0], meInformation[1]
	mePort, _ := strconv.Atoi(mePortString)
	youInformation := strings.Split(p.conn.RemoteAddr().String(), ":")
	youIP, youPortString := youInformation[0], youInformation[1]
	youPort, _ := strconv.Atoi(youPortString)
	me := serializer.NewNetAddress(time.Now(), net.ParseIP(meIP), uint16(mePort))
	me.Timestamp = time.Now().Unix()
	you := serializer.NewNetAddress(time.Now(), net.ParseIP(youIP), uint16(youPort))
	you.Timestamp = time.Now().Unix()
	nonce, _ := serializer.RandomUint64()
	msg := p2p.NewMsgVersion(*me, *you, nonce, lastBlockHeight)
	return msg
}

func (p *Peer) Stop() {
	p.closeSignal <- struct{}{}
	_ = p.conn.Close()
}

func (p *Peer) messageListener() error {
	p.log.Tracef("starting message listener for peer %v", p.GetID())
	for {
		select {
		case <-p.closeSignal:
			break
		default:
			rmsg, _, err := p.readMessage()
			if err != nil {
				return ErrorReadRemote
			}
			switch msg := rmsg.(type) {

			// Initial connection handlers
			case *p2p.MsgVersion:
				p.log.Errorf("handshake already received, duplicated version from peer %v", p.GetID())
				p.ManChan <- newPeerMsg(p, Disconnect)
			case *p2p.MsgVerack:
				p.log.Errorf("handshake already received, duplicated verack from peer %v", p.GetID())
				p.ManChan <- newPeerMsg(p, Disconnect)
			case *p2p.MsgPing:
				p.log.Tracef("received ping msg from peer %v", p.GetID())
				err := p.pong(msg.Nonce)
				if err != nil {
					return err
				}
			case *p2p.MsgPong:
				p.log.Tracef("received pong msg from peer %v", p.GetID())
				if msg.Nonce != p.lastPingNonce {
					p.log.Tracef("received pong msg from peer %v", p.GetID())
					return ErrorPingNonceMismatch
				}

			// Address sharing handlers
			case *p2p.MsgGetAddr:
				p.log.Tracef("received getaddr msg from peer %v", p.GetID())
				p.ManChan <- newDataRequestMsg(p, "getaddr", nil)
			case *p2p.MsgAddr:
				p.log.Tracef("received addr msg from peer %v", p.GetID())

			// Blocks handlers
			case *p2p.MsgBlock:
				p.log.Tracef("received block msg from peer %v", p.GetID())
				p.ManChan <- newBlockMsg(p, msg)
			case *p2p.MsgGetBlocks:
				p.log.Tracef("received getblocks msg from peer %v", p.GetID())
				p.ManChan <- newDataRequestMsg(p, "getblocks", &msg.LastBlockHash)
			case *p2p.MsgBlockInv:
				p.log.Tracef("received blockinv msg from peer %v", p.GetID())
				p.ManChan <- newBlocksInvMsg(p, msg)

			// Tx handlers
			case *p2p.MsgTx:
				p.log.Tracef("received tx msg from peer %v", p.GetID())
				p.ManChan <- newTxMsg(p, msg)
			}
		}
	}
}

func (p *Peer) readMessage() (p2p.Message, []byte, error) {
	n, msg, buf, err := p2p.ReadMessageWithEncodingN(p.conn, p.NetMagic)
	atomic.AddUint64(&p.bytesReceived, uint64(n))
	return msg, buf, err
}

func (p *Peer) writeMessage(msg p2p.Message) error {
	_, err := p2p.WriteMessageWithEncodingN(p.conn, msg, p.NetMagic)
	return err
}

func (p *Peer) GetAddr() serializer.NetAddress {
	p.peerLock.RLock()
	address := p.address
	p.peerLock.RUnlock()
	return address
}

func (p *Peer) GetID() int {
	p.peerLock.RLock()
	id := p.id
	p.peerLock.RUnlock()
	return id
}

func (p *Peer) GetSerializedData() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := serializer.WriteNetAddress(buf, &p.address)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Peer) GetLastBlock() int32 {
	p.peerLock.RLock()
	lastBlock := p.lastBlock
	p.peerLock.RUnlock()
	return lastBlock
}

func (p *Peer) SetLastBlock(lastBlock int32) {
	p.peerLock.RLock()
	p.lastBlock = lastBlock
	p.peerLock.RUnlock()
}

func (p *Peer) SendGetBlocks(getBlocks *p2p.MsgGetBlocks) error {
	return p.writeMessage(getBlocks)
}

func (p *Peer) SendBlock(blocks *p2p.MsgBlock) error {
	return p.writeMessage(blocks)
}

func (p *Peer) SendBlockInv(blockInv *p2p.MsgBlockInv) error {
	return p.writeMessage(blockInv)
}

func (p *Peer) IsSelectedForSync() bool {
	return p.selectedForSync
}

func (p *Peer) SetPeerSync(syncing bool) {
	p.selectedForSync = syncing
	return
}

func NewPeer(id int, conn net.Conn, addr serializer.NetAddress, inbound bool, time time.Time, messageChan chan interface{}, log *logger.Logger) *Peer {
	peer := &Peer{
		id:              id,
		log:             log,
		conn:            conn,
		address:         addr,
		inbound:         inbound,
		outbound:        !inbound,
		connectionTime:  time,
		ManChan:         messageChan,
		NetMagic:        p2p.MainNet,
		selectedForSync: false,
	}
	return peer
}
