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

func (k *keystoreServer) GenerateKeys(_ context.Context, in *proto.Number) (*proto.KeystoreKeys, error) {
	keys, err := k.keystore.GenerateNewValidatorKey(in.Number)
	if err != nil {
		return nil, err
	}

	protoKeys := make([]*proto.KeystoreKey, len(keys))
	for i := range keys {
		protoKeys[i] = &proto.KeystoreKey{
			Key:     hex.EncodeToString(keys[i].Secret.Marshal()),
			Enabled: keys[i].Enable,
			Path:    keys[i].Path,
		}
	}

	return &proto.KeystoreKeys{Keys: protoKeys}, nil
}

func (k *keystoreServer) GetMnemonic(context.Context, *proto.Empty) (*proto.Mnemonic, error) {
	mnemonic := k.keystore.GetMnemonic()
	return &proto.Mnemonic{Mnemonic: mnemonic}, nil
}

func (k *keystoreServer) GetKey(_ context.Context, in *proto.PublicKey) (*proto.KeystoreKey, error) {
	var pub [48]byte
	pubRaw, err := hex.DecodeString(in.Key)
	if err != nil {
		return nil, err
	}
	copy(pub[:], pubRaw)
	key, ok := k.keystore.GetValidatorKey(pub)
	if !ok {
		return nil, keystore.ErrorKeyNotOnKeystore
	}

	return &proto.KeystoreKey{
		Key:     hex.EncodeToString(key.Secret.Marshal()),
		Enabled: key.Enable,
		Path:    key.Path,
	}, nil
}

func (k *keystoreServer) GetKeys(_ context.Context, _ *proto.Empty) (*proto.KeystoreKeys, error) {

	keys, err := k.keystore.GetValidatorKeys()
	if err != nil {
		return nil, err
	}

	protoKeys := make([]*proto.KeystoreKey, len(keys))
	for i := range keys {
		protoKeys[i] = &proto.KeystoreKey{
			Key:     hex.EncodeToString(keys[i].Secret.Marshal()),
			Enabled: keys[i].Enable,
			Path:    keys[i].Path,
		}
	}

	return &proto.KeystoreKeys{Keys: protoKeys}, nil
}

func (k *keystoreServer) ToggleKey(_ context.Context, in *proto.ToggleKeyMsg) (*proto.KeystoreKey, error) {
	var pub [48]byte
	pubRaw, err := hex.DecodeString(in.PublicKey)
	if err != nil {
		return nil, err
	}
	copy(pub[:], pubRaw)

	err = k.keystore.ToggleKey(pub, in.Enabled)
	if err != nil {
		return nil, err
	}
	key, ok := k.keystore.GetValidatorKey(pub)
	if !ok {
		return nil, keystore.ErrorKeyNotOnKeystore
	}

	return &proto.KeystoreKey{
		Key:     hex.EncodeToString(key.Secret.Marshal()),
		Enabled: key.Enable,
		Path:    key.Path,
	}, nil
}
