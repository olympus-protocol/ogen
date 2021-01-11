package hostnode

import (
	"github.com/VictoriaMetrics/fastcache"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"sync"
	"time"
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
	BadMsgs       int
	Banscore      uint64
}

type statsService struct {
	banPeersCache *fastcache.Cache
	badPeersStats sync.Map
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

func (s *statsService) Close() {
	datapath := config.GlobalFlags.DataPath
	_ = s.banPeersCache.SaveToFile(datapath + "/badpeers")
}

func NewPeersStatsService() *statsService {
	datapath := config.GlobalFlags.DataPath

	cache := fastcache.LoadFromFileOrNew(datapath+"/badpeers", 50*1024*1024)

	return &statsService{
		banPeersCache: cache,
	}
}
