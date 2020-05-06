package peers

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/olympus-protocol/ogen/bloom"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/mempool"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"github.com/pkg/errors"
)

type Peer struct {
	// Ogen main config
	NetMagic p2p.NetMagic

	// Pass the log to debug easily
	log *logger.Logger

	// Version Msg Data
	protocol  int32
	services  p2p.ServiceFlag
	lastBlock uint64
	userAgent string

	// Peer properties
	id             int
	inbound        bool
	outbound       bool
	address        *serializer.NetAddress
	conn           net.Conn
	connectionTime time.Time
	lastPingTime   int64
	lastPingNonce  uint64
	banScore       int32

	// Peer dinamic properties
	bytesReceived  uint64
	verackReceived bool

	// Internal Chan
	ctx   context.Context
	Close context.CancelFunc

	// Chans for PeerMan Communication
	mainCloseChan chan interface{}

	// Locks for safe usage
	peerLock sync.RWMutex

	// For syncMan sync peer selector
	selectedForSync bool

	blockchain       *chain.Blockchain
	peerman          *PeerMan
	voteBloomFilter  *bloom.BloomFilter
	blockBloomFilter *bloom.BloomFilter
}

func (p *Peer) Start(lastBlockHeight uint64) {
	// First version handshake
	err := p.versionHandshake(lastBlockHeight)
	if err != nil {
		p.log.Errorf("disconnecting peer %v because of error %s", p.GetID(), err)
		p.peerman.Disconnect(p)
		return
	}
	p.peerman.Handshake(p)
	if p.peerman.needsPeers() {
		fmt.Println("asking for some peers")
		p.writeMessage(&p2p.MsgGetAddr{})
	}
	// Init message listener
	go func() {
		err := p.messageListener()
		if err != nil {
			p.log.Errorf("disconnecting peer %v because of %s", p.GetID(), err)
			p.peerman.Disconnect(p)
			return
		}
	}()
	// Init ping/pong
	go func() {
		err := p.peerRoutine()
		if err != nil {
			p.log.Errorf("disconnecting peer %v because of %s", p.GetID(), err)
			p.peerman.Disconnect(p)
		}
	}()
}

func (p *Peer) versionHandshake(lastBlockHeight uint64) error {
	if p.inbound {
		p.log.Debug("running inbound version handshake")
		err := p.handleInboundPeerHandshake(lastBlockHeight)
		if err != nil {
			p.log.Errorf("unable to perform inbound handshake for peer: %v", p.GetID())
			return err
		}
	}
	if !p.inbound {
		p.log.Debug("running outbound version handshake")
		err := p.handleOutboundPeerHandshake(lastBlockHeight)
		if err != nil {
			p.log.Errorf("unable to perform outbound handshake for peer: %v", p.GetID())
			return err
		}
	}
	return nil
}

var zeroHash = chainhash.Hash{}

func (p *Peer) handleInboundPeerHandshake(lastBlockHeight uint64) error {
	// first, we read their version
	remoteMsgVersion, _, err := p.readMessage()
	if err != nil {
		return errors.Wrap(err, "error reading from remote")
	}
	msgVersion, ok := remoteMsgVersion.(*p2p.MsgVersion)
	if !ok {
		return ErrorNoVersionFirst
	}
	p.updateStats(msgVersion)

	// acknowledge their version message
	verack := p2p.NewMsgVerack()
	err = p.writeMessage(verack)
	if err != nil {
		return ErrorWriteRemote
	}

	// send our version
	err = p.writeMessage(p.versionMsg(lastBlockHeight, p.peerman.ListenPort()))
	if err != nil {
		return ErrorWriteRemote
	}

	// wait for them to acknowledge our version message
	remoteMsgVerack, _, err := p.readMessage()
	if err != nil {
		return errors.Wrap(err, "error reading from remote")
	}
	_, ok = remoteMsgVerack.(*p2p.MsgVerack)
	if !ok {
		return ErrorNoVerackAfterVersion
	}

	// mark verack as received
	p.verackReceived = true

	p.log.Debugf("their last block: %d our last block: %d", msgVersion.LastBlock, lastBlockHeight)
	if msgVersion.LastBlock > lastBlockHeight {
		err := p.writeMessage(&p2p.MsgGetBlocks{
			LocatorHashes: p.blockchain.GetLocatorHashes(),
			HashStop:      zeroHash,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Peer) handleOutboundPeerHandshake(lastBlockHeight uint64) error {
	err := p.writeMessage(p.versionMsg(lastBlockHeight, p.peerman.ListenPort()))
	if err != nil {
		return ErrorWriteRemote
	}
	remoteVerack, _, _ := p.readMessage()
	_, ok := remoteVerack.(*p2p.MsgVerack)
	if !ok {
		return ErrorNoVerackAfterVersion
	}
	p.verackReceived = true
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
	p.log.Debugf("their last block: %d our last block: %d", msgVersion.LastBlock, lastBlockHeight)
	if msgVersion.LastBlock > lastBlockHeight {
		err := p.writeMessage(&p2p.MsgGetBlocks{
			LocatorHashes: p.blockchain.GetLocatorHashes(),
			HashStop:      zeroHash,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Peer) updateStats(msgVersion *p2p.MsgVersion) {
	p.peerLock.Lock()
	p.protocol = msgVersion.ProtocolVersion
	p.services = msgVersion.Services
	p.lastBlock = msgVersion.LastBlock
	p.userAgent = msgVersion.UserAgent
	p.address = &msgVersion.AddrMe
	p.peerLock.Unlock()
}

func (p *Peer) sendMempool() error {
	possibleVotes := p.peerman.mempool.GetVotesNotInBloom(p.voteBloomFilter)
	// send 50% of these
	votesToSend := mempool.PickPercentVotes(possibleVotes, 0.5)
	for _, v := range votesToSend {
		p.voteBloomFilter.Add(v.Hash())
	}

	return p.writeMessage(&p2p.MsgVotes{
		Votes: votesToSend,
	})
}

func (p *Peer) peerRoutine() error {
	p.log.Tracef("starting peer routine for peer %v", p.GetID())
	pingTicker := time.NewTicker(15 * time.Second)
	mempoolTicker := time.NewTicker(5 * time.Minute)
outer:
	for {
		select {
		case <-p.ctx.Done():
			break outer
		case <-pingTicker.C:
			err := p.ping()
			if err != nil {
				p.log.Errorf("unable to ping peer %v: %s", p.GetID(), err)
				return err
			}
			if p.peerman.needsPeers() {
				p.writeMessage(&p2p.MsgGetAddr{})
			}
		case <-mempoolTicker.C:
			err := p.sendMempool()
			if err != nil {
				p.log.Errorf("unable to send mempool: %s", err)
				return err
			}
		}
	}
	return nil
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

func (p *Peer) versionMsg(lastBlockHeight uint64, listenPort uint16) *p2p.MsgVersion {
	meInformation := strings.Split(p.conn.LocalAddr().String(), ":")
	meIP, _ := meInformation[0], meInformation[1]

	youInformation := strings.Split(p.conn.RemoteAddr().String(), ":")
	youIP, youPortString := youInformation[0], youInformation[1]
	youPort, _ := strconv.Atoi(youPortString)

	me := serializer.NewNetAddress(time.Now(), net.ParseIP(meIP), listenPort)
	me.Timestamp = time.Now().Unix()

	you := serializer.NewNetAddress(time.Now(), net.ParseIP(youIP), uint16(youPort))
	you.Timestamp = time.Now().Unix()

	nonce, _ := serializer.RandomUint64()
	msg := p2p.NewMsgVersion(*me, *you, nonce, lastBlockHeight)
	return msg
}

func (p *Peer) Stop() {
	p.Close()
	_ = p.conn.Close()
}

const maxBlocksPerMessage = 500

func (p *Peer) submitBlock(block *primitives.Block) error {
	bh := block.Hash()
	if p.blockBloomFilter.Has(bh) {
		return nil
	}
	p.blockBloomFilter.Add(bh)
	return p.writeMessage(&p2p.MsgBlocks{
		Blocks: []primitives.Block{*block},
	})
}

func (p *Peer) sendBlocksToPeer(msg *p2p.MsgGetBlocks) error {
	// first block is tip, so we check each block in order and check if the block matches
	firstCommon := p.blockchain.State().Chain().Genesis()
	locatorHashesGenesis := &msg.LocatorHashes[len(msg.LocatorHashes)-1]

	if !firstCommon.Hash.IsEqual(locatorHashesGenesis) {
		return fmt.Errorf("incorrect genesis block (got: %s, expected: %s)", locatorHashesGenesis, firstCommon.Hash)
	}

	for _, b := range msg.LocatorHashes {
		if b, found := p.blockchain.State().Index().Get(b); found {
			firstCommon = b
			break
		}
	}

	p.log.Debugf("found first common block %s", firstCommon.Hash)

	blocksToSend := make([]primitives.Block, 0, 500)

	if firstCommon.Hash.IsEqual(locatorHashesGenesis) {
		fc, ok := p.blockchain.State().Chain().Next(firstCommon)
		if !ok {
			return nil
		}
		firstCommon = fc
	}

	for firstCommon != nil && len(blocksToSend) < maxBlocksPerMessage {
		block, err := p.blockchain.GetBlock(firstCommon.Hash)
		if err != nil {
			return err
		}

		blocksToSend = append(blocksToSend, *block)
		p.blockBloomFilter.Add(firstCommon.Hash)

		if firstCommon.Hash.IsEqual(&msg.HashStop) {
			break
		}
		var ok bool
		firstCommon, ok = p.blockchain.State().Chain().Next(firstCommon)
		if !ok {
			break
		}
	}

	p.log.Debugf("sending %d blocks", len(blocksToSend))

	return p.writeMessage(&p2p.MsgBlocks{
		Blocks: blocksToSend,
	})
}

func (p *Peer) submitVote(vote *primitives.SingleValidatorVote) error {
	vh := vote.Hash()
	if p.voteBloomFilter.Has(vh) || true {
		// already sent it
		return nil
	}

	p.voteBloomFilter.Add(vh)

	return p.writeMessage(&p2p.MsgVotes{
		Votes: []primitives.SingleValidatorVote{*vote},
	})
}

func (p *Peer) messageListener() error {
	p.log.Tracef("starting message listener for peer %v", p.GetID())
outer:
	for {
		select {
		case <-p.ctx.Done():
			break outer
		default:
			p.log.Debug("reading...")
			rmsg, _, err := p.readMessage()
			if err != nil {
				return errors.Wrap(err, "error reading from remote")
			}
			p.log.Debugf("read message %s", rmsg.Command())
			go func() {
				switch msg := rmsg.(type) {

				// Initial connection handlers
				case *p2p.MsgVersion:
					p.log.Errorf("handshake already received, duplicated version from peer %v", p.GetID())
					p.peerman.Disconnect(p)
				case *p2p.MsgVerack:
					p.log.Errorf("handshake already received, duplicated verack from peer %v", p.GetID())
					p.peerman.Disconnect(p)
				case *p2p.MsgPing:
					p.log.Tracef("received ping msg from peer %v", p.GetID())
					err := p.pong(msg.Nonce)
					if err != nil {
						p.log.Errorf("Error processing ping: %s", err)
					}
				case *p2p.MsgPong:
					p.log.Tracef("received pong msg from peer %v", p.GetID())
					if msg.Nonce != p.lastPingNonce {
						p.log.Tracef("received pong msg from peer %v", p.GetID())
						p.log.Errorf("Error processing ping: %s", ErrorPingNonceMismatch)
					}

				// Address sharing handlers
				case *p2p.MsgGetAddr:
					p.log.Tracef("received getaddr msg from peer %v", p.GetID())
					knownPeers := make([]*serializer.NetAddress, 0)
					for _, peer := range p.peerman.Peers() {
						if p != peer && peer.address != nil {
							knownPeers = append(knownPeers, peer.address)
						}
					}
					if len(knownPeers) > 0 {
						if err := p.writeMessage(&p2p.MsgAddr{
							AddrList: knownPeers,
						}); err != nil {
							p.log.Errorf("error responding to get addrs: %s", err)
						}
					}
				case *p2p.MsgAddr:
					p.log.Tracef("received addr msg from peer %v", p.GetID())
					if err := p.peerman.receiveAddrs(msg.AddrList); err != nil {
						p.log.Error(err)
					}

				// Blocks handlers
				case *p2p.MsgGetBlocks:
					p.log.Tracef("received getblocks msg from peer %v", p.GetID())
					if err := p.sendBlocksToPeer(msg); err != nil {
						p.log.Errorf("error sending blocks to peer: %s", err)
					}
					// TODO: fix
				case *p2p.MsgBlocks:
					p.log.Tracef("received blocks msg from peer %v", p.GetID())
					for _, b := range msg.Blocks {
						if !p.blockchain.State().Index().Have(b.Header.PrevBlockHash) {
							err = p.writeMessage(&p2p.MsgGetBlocks{
								LocatorHashes: p.blockchain.GetLocatorHashes(),
								HashStop:      b.Hash(),
							})
							if err != nil {
								p.log.Error(err)
							}
							break
						}

						bh := b.Hash()
						p.blockBloomFilter.Add(bh)
						p.log.Debugf("processing block %s", bh)
						if err := p.blockchain.ProcessBlock(&b); err != nil {
							p.log.Errorf("error processing block from peer: %s", err)
							break
						}
						if err := p.peerman.SubmitBlock(&b); err != nil {
							p.log.Error(err)
						}
					}

				// Tx handlers
				case *p2p.MsgVotes:
					// p.log.Tracef("received votes msg from peer %v with %d votes", p.GetID(), len(msg.Votes))
					for _, v := range msg.Votes {
						p.voteBloomFilter.Add(v.Hash())
						p.peerman.mempool.Add(&v, v.OutOf)
					}
				case *p2p.MsgTx:
					p.log.Tracef("received tx msg from peer %v", p.GetID())
				}
			}()
		}
	}
	return nil
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

func (p *Peer) GetAddr() *serializer.NetAddress {
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
	err := serializer.WriteNetAddress(buf, p.address)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Peer) GetLastBlock() uint64 {
	p.peerLock.RLock()
	lastBlock := p.lastBlock
	p.peerLock.RUnlock()
	return lastBlock
}

func (p *Peer) SetLastBlock(lastBlock uint64) {
	p.peerLock.RLock()
	p.lastBlock = lastBlock
	p.peerLock.RUnlock()
}

func (p *Peer) IsSelectedForSync() bool {
	return p.selectedForSync
}

func (p *Peer) SetPeerSync(syncing bool) {
	p.selectedForSync = syncing
	return
}

const (
	voteBloomFilterSize  = 1024 * 1024
	blockBloomFilterSize = 1024 * 1024
)

func NewPeer(id int, conn net.Conn, inbound bool, time time.Time, log *logger.Logger, peerMgr *PeerMan) *Peer {
	ctx, cancel := context.WithCancel(context.Background())
	peer := &Peer{
		id:               id,
		log:              log,
		conn:             conn,
		inbound:          inbound,
		outbound:         !inbound,
		connectionTime:   time,
		NetMagic:         p2p.MainNet,
		selectedForSync:  false,
		blockchain:       peerMgr.chain,
		peerman:          peerMgr,
		voteBloomFilter:  bloom.NewBloomFilter(voteBloomFilterSize),
		blockBloomFilter: bloom.NewBloomFilter(blockBloomFilterSize),
		ctx:              ctx,
		Close:            cancel,
	}
	return peer
}
