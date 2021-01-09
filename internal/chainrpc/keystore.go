package chainrpc

import (
	"context"
	"encoding/hex"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/pkg/params"

	"github.com/olympus-protocol/ogen/api/proto"
)

type keystoreServer struct {
	netParams *params.ChainParams
	keystore  keystore.Keystore
	proto.UnimplementedKeystoreServer
}

func (s *keystoreServer) GenValidatorKey(ctx context.Context, in *proto.GenValidatorKeys) (*proto.KeyPairs, error) {
	defer ctx.Done()

	key, err := s.keystore.GenerateNewValidatorKey(in.Keys)
	if err != nil {
		return nil, err
	}
	keys := make([]string, in.Keys)
	for i := range keys {
		keys[i] = hex.EncodeToString(key[i].Secret.Marshal())
	}
	return &proto.KeyPairs{Keys: keys}, nil
}
