package hostnode

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
	unreachablePeerTimePenalization = time.Minute * 5
	banPeerTimePenalization         = time.Minute * 60
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
	ChainStats    *peerChainStats
	Direction     network.Direction
	BytesReceived uint64
	BytesSent     uint64
	BadMessages   int
	BanScore      uint64
}

type statsService struct {
	banPeersCache  *fastcache.Cache
	peersStats     map[peer.ID]*peerStats
	peersStatsLock sync.Mutex
	count          int
	host           HostNode
}

// IsBanned returns if a known peer is banned for bad behaviour
func (s *statsService) IsBanned(p peer.ID) (bool, error) {
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

func (s *statsService) GetPeerStats(p peer.ID) (*peerStats, bool) {
	s.peersStatsLock.Lock()
	ps, ok := s.peersStats[p]
	s.peersStatsLock.Unlock()
	if !ok {
		return nil, false
	}
	return ps, true
}

func (s *statsService) SetPeerBan(p peer.ID, until time.Duration) {
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
func (s *statsService) FindBestPeer() (peer.ID, bool) {

	verMsg := s.host.VersionMsg()

	var peersAhead []*peerStats
	var peersBehind []*peerStats
	var peersEqual []*peerStats

	s.peersStatsLock.Lock()
	for _, p := range s.peersStats {
		if p.ChainStats.TipHeight > verMsg.FinalizedHeight {
			peersAhead = append(peersAhead, p)
		}

		if p.ChainStats.TipHeight == verMsg.FinalizedHeight {
			peersEqual = append(peersEqual, p)
		}

		if p.ChainStats.TipHeight < verMsg.FinalizedHeight {
			peersBehind = append(peersBehind, p)
		}
	}
	s.peersStatsLock.Unlock()

	if len(peersAhead) == 0 {
		return "", false
	}

	r := rand.Intn(len(peersAhead))
	peerSelected := peersAhead[r]

	return peerSelected.ID, true
}

func (s *statsService) Count() int {
	return s.count
}

func (s *statsService) Add(p peer.ID, ver *p2p.MsgVersion, dir network.Direction) {
	peerStats := &peerStats{
		ID: p,
		ChainStats: &peerChainStats{
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
	s.peersStatsLock.Lock()
	s.peersStats[p] = peerStats
	s.peersStatsLock.Unlock()
	s.count += 1
}

func (s *statsService) Remove(p peer.ID) {
	s.peersStatsLock.Lock()
	delete(s.peersStats, p)
	s.peersStatsLock.Unlock()
	s.count -= 1
}

func (s *statsService) Close() {
	datapath := config.GlobalFlags.DataPath
	_ = s.banPeersCache.SaveToFile(datapath + "/badpeers")
}

func (s *statsService) IncreaseWrongMsgCount(p peer.ID) {
	s.peersStatsLock.Lock()
	_, ok := s.peersStats[p]
	if !ok {
		return
	}
	s.peersStats[p].BadMessages += 1
	s.peersStats[p].BanScore += 10

	if s.peersStats[p].BanScore >= 500 {
		s.SetPeerBan(p, banPeerTimePenalization)
		_ = s.host.DisconnectPeer(p)
	}
	s.peersStatsLock.Unlock()
	return

}

func (s *statsService) IncreasePeerReceivedBytes(p peer.ID, amount uint64) {
	s.peersStatsLock.Lock()
	_, ok := s.peersStats[p]
	if !ok {
		return
	}
	s.peersStats[p].BytesReceived += amount
	s.peersStatsLock.Unlock()
	return
}

func (s *statsService) IncreasePeerSentBytes(p peer.ID, amount uint64) {
	s.peersStatsLock.Lock()
	_, ok := s.peersStats[p]
	if !ok {
		return
	}
	s.peersStats[p].BytesSent += amount
	s.peersStatsLock.Unlock()
	return
}

func (s *statsService) handleFinalizationMsg(id peer.ID, msg p2p.Message) (uint64, error) {

	fin, ok := msg.(*p2p.MsgFinalization)
	if !ok {
		return 0, errors.New("non finalization msg")
	}

	if s.host.GetHost().ID() == id {
		return 0, nil
	}

	s.peersStatsLock.Lock()

	ps, ok := s.peersStats[id]
	if !ok {
		return msg.PayloadLength(), nil
	}

	ps.ChainStats = &peerChainStats{
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

	s.peersStats[id] = ps

	s.peersStatsLock.Unlock()

	return msg.PayloadLength(), nil
}

func NewPeersStatsService(host HostNode) (*statsService, error) {
	datapath := config.GlobalFlags.DataPath

	cache := fastcache.LoadFromFileOrNew(datapath+"/badpeers", 50*1024*1024)

	ss := &statsService{
		banPeersCache: cache,
		count:         0,
		host:          host,
		peersStats: make(map[peer.ID]*peerStats),
	}
	if err := host.RegisterTopicHandler(p2p.MsgFinalizationCmd, ss.handleFinalizationMsg); err != nil {
		return nil, err
	}

	return ss, nil
}
