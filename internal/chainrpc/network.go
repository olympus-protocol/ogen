package chainrpc

import (
	"context"
	"encoding/hex"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/internal/host"
)

type networkServer struct {
	host host.Host
	proto.UnimplementedNetworkServer
}

func (s *networkServer) GetNetworkInfo(_ context.Context, _ *proto.Empty) (*proto.NetworkInfo, error) {
	return &proto.NetworkInfo{Peers: int32(s.host.ConnectedPeers()), ID: s.host.ID().String()}, nil
}

func (s *networkServer) GetPeersInfo(_ context.Context, _ *proto.Empty) (*proto.Peers, error) {

	stats := s.host.GetPeersInfo()
	peersStats := make([]*proto.Peer, len(stats))
	for i := range peersStats {
		s := &proto.Peer{
			Id:            stats[i].ID.String(),
			Direction:     "",
			BytesReceived: int64(stats[i].BytesReceived),
			BytesSent:     int64(stats[i].BytesSent),
			BadMessages:   int64(stats[i].BadMessages),
			BanScore:      int64(stats[i].BanScore),
			ChainStats: &proto.PeerChainStats{
				TipSlot:         int64(stats[i].ChainStats.TipSlot),
				TipHeight:       int64(stats[i].ChainStats.TipHeight),
				TipHash:         hex.EncodeToString(stats[i].ChainStats.TipHash[:]),
				JustifiedSlot:   int64(stats[i].ChainStats.JustifiedSlot),
				JustifiedHeight: int64(stats[i].ChainStats.JustifiedHeight),
				JustifiedHash:   hex.EncodeToString(stats[i].ChainStats.JustifiedHash[:]),
				FinalizedSlot:   int64(stats[i].ChainStats.FinalizedSlot),
				FinalizedHeight: int64(stats[i].ChainStats.FinalizedHeight),
				FinalizedHash:   hex.EncodeToString(stats[i].ChainStats.FinalizedHash[:]),
			},
		}
		peersStats[i] = s
	}

	return &proto.Peers{Peers: peersStats}, nil
}

func (s *networkServer) AddPeer(_ context.Context, peerAddr *proto.IP) (*proto.Success, error) {

	maddr, err := multiaddr.NewMultiaddr(peerAddr.Host)
	if err != nil {
		return nil, err
	}

	pinfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return nil, err
	}

	err = s.host.Connect(*pinfo)
	if err != nil {
		return &proto.Success{Success: false, Error: err.Error()}, nil
	}

	return &proto.Success{Success: true}, nil
}
