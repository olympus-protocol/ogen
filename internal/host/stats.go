package host

import (
	"errors"
	"github.com/VictoriaMetrics/fastcache"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"math/rand"
	"sync"
	"time"
)

const (
	banPeerTimePenalization = time.Minute * 60
)

type peerChainStats struct {
	TipSlot         uint64
	TipHeight       uint64
	TipHash         chainhash.Hash
	JustifiedSlot   uint64
	JustifiedHeight uint64
	JustifiedHash   chainhash.Hash
	FinalizedSlot   uint64
	FinalizedHeight uint64
	FinalizedHash   chainhash.Hash
}

type peerStats struct {
	ID            peer.ID
	ChainStats    peerChainStats
	Direction     network.Direction
	BytesReceived uint64
	BytesSent     uint64
	BadMessages   int
	BanScore      uint64
}

type stats struct {
	banPeersCache *fastcache.Cache
	peersStats    sync.Map
	count         int
	h             Host
}

// IsBanned returns if a known peer is banned for bad behaviour
func (s *stats) IsBanned(p peer.ID) (bool, error) {
	ip, err := p.MarshalBinary()
	if err != nil {
		return false, err
	}

	data, ok := s.banPeersCache.HasGet(nil, ip)
	if !ok {
		return false, nil
	}

	var t time.Time
	err = t.UnmarshalBinary(data)
	if err != nil {
		return false, err
	}

	if time.Now().Unix() > t.Unix() {
		s.banPeersCache.Del(ip)
		return false, nil
	}

	return true, nil
}

func (s *stats) GetPeerStats(p peer.ID) (*peerStats, bool) {
	ps, ok := s.peersStats.Load(p)
	if !ok {
		return nil, false
	}
	stats, ok := ps.(peerStats)
	if !ok {
		return nil, false
	}
	return &stats, true
}

func (s *stats) SetPeerBan(p peer.ID, until time.Duration) {
	ip, err := p.MarshalBinary()
	if err != nil {
		return
	}
	t := time.Now().Add(until)
	tb, err := t.MarshalBinary()
	if err != nil {
		return
	}
	s.banPeersCache.Set(ip, tb)
}

// FindBestPeer will perform a contextual check for peers and return a random peer ahead if we need to sync.
func (s *stats) FindBestPeer() (peer.ID, bool) {
	verMsg := s.h.Version()

	var peersAhead []*peerStats
	var peersBehind []*peerStats
	var peersEqual []*peerStats

	s.peersStats.Range(func(key, value interface{}) bool {
		p, ok := value.(peerStats)
		if !ok {
			return true
		}
		if p.ChainStats.TipHeight > verMsg.FinalizedHeight {
			peersAhead = append(peersAhead, &p)
		}

		if p.ChainStats.TipHeight == verMsg.FinalizedHeight {
			peersEqual = append(peersEqual, &p)
		}

		if p.ChainStats.TipHeight < verMsg.FinalizedHeight {
			peersBehind = append(peersBehind, &p)
		}
		return true
	})

	if len(peersAhead) == 0 {
		return "", false
	}

	r := rand.Intn(len(peersAhead))
	peerSelected := peersAhead[r]

	return peerSelected.ID, true
}

func (s *stats) Count() int {
	return s.count
}

func (s *stats) Add(p peer.ID, ver *p2p.MsgVersion, dir network.Direction) {
	peerStats := peerStats{
		ID: p,
		ChainStats: peerChainStats{
			TipSlot:         ver.TipSlot,
			TipHeight:       ver.Tip,
			TipHash:         ver.TipHash,
			JustifiedSlot:   ver.JustifiedSlot,
			JustifiedHeight: ver.JustifiedHeight,
			JustifiedHash:   ver.JustifiedHash,
			FinalizedSlot:   ver.FinalizedSlot,
			FinalizedHeight: ver.FinalizedHeight,
			FinalizedHash:   ver.FinalizedHash,
		},
		Direction:     dir,
		BytesReceived: 0,
		BytesSent:     0,
		BadMessages:   0,
		BanScore:      0,
	}
	s.peersStats.Store(p, peerStats)
	s.count += 1
}

func (s *stats) Remove(p peer.ID) {
	s.peersStats.Delete(p)
	s.count -= 1
}

func (s *stats) Close() {
	datapath := config.GlobalFlags.DataPath
	_ = s.banPeersCache.SaveToFile(datapath + "/badpeers")
}

func (s *stats) IncreaseWrongMsgCount(p peer.ID) {
	ps, ok := s.peersStats.Load(p)
	if !ok {
		return
	}
	stats, ok := ps.(peerStats)
	if !ok {
		return
	}

	stats.BadMessages += 1
	stats.BanScore += 10

	if stats.BanScore >= 500 {
		s.SetPeerBan(p, banPeerTimePenalization)
		_ = s.h.Disconnect(p)
	}

	s.peersStats.Store(p, stats)
}

func (s *stats) IncreasePeerReceivedBytes(p peer.ID, amount uint64) {
	ps, ok := s.peersStats.Load(p)
	if !ok {
		return
	}
	stats, ok := ps.(peerStats)
	if !ok {
		return
	}

	stats.BytesReceived += amount

	s.peersStats.Store(p, stats)
}

func (s *stats) IncreasePeerSentBytes(p peer.ID, amount uint64) {
	ps, ok := s.peersStats.Load(p)
	if !ok {
		return
	}
	stats, ok := ps.(peerStats)
	if !ok {
		return
	}

	stats.BytesSent += amount

	s.peersStats.Store(p, stats)
}

func (s *stats) handleFinalizationMsg(id peer.ID, msg p2p.Message) (uint64, error) {

	fin, ok := msg.(*p2p.MsgFinalization)
	if !ok {
		return 0, errors.New("non block msg")
	}

	if s.h.ID() == id {
		return 0, nil
	}

	ps, ok := s.peersStats.Load(id)
	if !ok {
		return msg.PayloadLength(), nil
	}

	peerStats, ok := ps.(peerStats)
	if !ok {
		return msg.PayloadLength(), nil
	}

	peerStats.ChainStats = peerChainStats{
		TipSlot:         fin.TipSlot,
		TipHeight:       fin.Tip,
		TipHash:         fin.TipHash,
		JustifiedSlot:   fin.JustifiedSlot,
		JustifiedHeight: fin.JustifiedHeight,
		JustifiedHash:   fin.JustifiedHash,
		FinalizedSlot:   fin.FinalizedSlot,
		FinalizedHeight: fin.FinalizedHeight,
		FinalizedHash:   fin.FinalizedHash,
	}

	s.peersStats.Store(id, peerStats)

	return msg.PayloadLength(), nil
}

func NewStatsService(h Host) (*stats, error) {
	datapath := config.GlobalFlags.DataPath

	cache := fastcache.LoadFromFileOrNew(datapath+"/badpeers", 50*1024*1024)

	ss := &stats{
		banPeersCache: cache,
		count:         0,
		h:             h,
	}

	h.RegisterTopicHandler(p2p.MsgFinalizationCmd, ss.handleFinalizationMsg)

	return ss, nil
}
