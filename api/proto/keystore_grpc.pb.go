// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// KeystoreClient is the client API for Keystore service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type KeystoreClient interface {
	//*
	//Method: GenerateKeys
	//Input: message GenValidatorKeys
	//Response: message KeyPairs
	//Description: Returns private keys generated for validators start.
	GenerateKeys(ctx context.Context, in *Number, opts ...grpc.CallOption) (*Keys, error)
	//*
	//Method: GetMnemonic
	//Input: message Empty
	//Response: message Mnemonic
	//Description: Returns the mnemonic key of the keystore.
	GetMnemonic(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Mnemonic, error)
	//*
	//Method: GetKey
	//Input: message PublicKey
	//Response: message KeystoreKey
	//Description: Returns the keystore information for a specific key.
	GetKey(ctx context.Context, in *PublicKey, opts ...grpc.CallOption) (*KeystoreKey, error)
	//*
	//Method: GetKeys
	//Input: message Empty
	//Response: message KeystoreKeys
	//Description: Returns all the keystore keys.
	GetKeys(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*KeystoreKeys, error)
	//*
	//Method: ToggleKey
	//Input: message ToggleKey
	//Response: message KeystoreKey
	//Description: Enables/Disable a keystore key.
	ToggleKey(ctx context.Context, in *ToggleKeyMsg, opts ...grpc.CallOption) (*KeystoreKey, error)
}

type keystoreClient struct {
	cc grpc.ClientConnInterface
}

func NewKeystoreClient(cc grpc.ClientConnInterface) KeystoreClient {
	return &keystoreClient{cc}
}

func (c *keystoreClient) GenerateKeys(ctx context.Context, in *Number, opts ...grpc.CallOption) (*Keys, error) {
	out := new(Keys)
	err := c.cc.Invoke(ctx, "/Keystore/GenerateKeys", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keystoreClient) GetMnemonic(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Mnemonic, error) {
	out := new(Mnemonic)
	err := c.cc.Invoke(ctx, "/Keystore/GetMnemonic", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keystoreClient) GetKey(ctx context.Context, in *PublicKey, opts ...grpc.CallOption) (*KeystoreKey, error) {
	out := new(KeystoreKey)
	err := c.cc.Invoke(ctx, "/Keystore/GetKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keystoreClient) GetKeys(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*KeystoreKeys, error) {
	out := new(KeystoreKeys)
	err := c.cc.Invoke(ctx, "/Keystore/GetKeys", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keystoreClient) ToggleKey(ctx context.Context, in *ToggleKeyMsg, opts ...grpc.CallOption) (*KeystoreKey, error) {
	out := new(KeystoreKey)
	err := c.cc.Invoke(ctx, "/Keystore/ToggleKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KeystoreServer is the server API for Keystore service.
// All implementations must embed UnimplementedKeystoreServer
// for forward compatibility
type KeystoreServer interface {
	//*
	//Method: GenerateKeys
	//Input: message GenValidatorKeys
	//Response: message KeyPairs
	//Description: Returns private keys generated for validators start.
	GenerateKeys(context.Context, *Number) (*Keys, error)
	//*
	//Method: GetMnemonic
	//Input: message Empty
	//Response: message Mnemonic
	//Description: Returns the mnemonic key of the keystore.
	GetMnemonic(context.Context, *Empty) (*Mnemonic, error)
	//*
	//Method: GetKey
	//Input: message PublicKey
	//Response: message KeystoreKey
	//Description: Returns the keystore information for a specific key.
	GetKey(context.Context, *PublicKey) (*KeystoreKey, error)
	//*
	//Method: GetKeys
	//Input: message Empty
	//Response: message KeystoreKeys
	//Description: Returns all the keystore keys.
	GetKeys(context.Context, *Empty) (*KeystoreKeys, error)
	//*
	//Method: ToggleKey
	//Input: message ToggleKey
	//Response: message KeystoreKey
	//Description: Enables/Disable a keystore key.
	ToggleKey(context.Context, *ToggleKeyMsg) (*KeystoreKey, error)
	mustEmbedUnimplementedKeystoreServer()
}

// UnimplementedKeystoreServer must be embedded to have forward compatible implementations.
type UnimplementedKeystoreServer struct {
}

func (UnimplementedKeystoreServer) GenerateKeys(context.Context, *Number) (*Keys, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateKeys not implemented")
}
func (UnimplementedKeystoreServer) GetMnemonic(context.Context, *Empty) (*Mnemonic, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMnemonic not implemented")
}
func (UnimplementedKeystoreServer) GetKey(context.Context, *PublicKey) (*KeystoreKey, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetKey not implemented")
}
func (UnimplementedKeystoreServer) GetKeys(context.Context, *Empty) (*KeystoreKeys, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetKeys not implemented")
}
func (UnimplementedKeystoreServer) ToggleKey(context.Context, *ToggleKeyMsg) (*KeystoreKey, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ToggleKey not implemented")
}
func (UnimplementedKeystoreServer) mustEmbedUnimplementedKeystoreServer() {}

// UnsafeKeystoreServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to KeystoreServer will
// result in compilation errors.
type UnsafeKeystoreServer interface {
	mustEmbedUnimplementedKeystoreServer()
}

func RegisterKeystoreServer(s *grpc.Server, srv KeystoreServer) {
	s.RegisterService(&_Keystore_serviceDesc, srv)
}

func _Keystore_GenerateKeys_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Number)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeystoreServer).GenerateKeys(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Keystore/GenerateKeys",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeystoreServer).GenerateKeys(ctx, req.(*Number))
	}
	return interceptor(ctx, in, info, handler)
}

func _Keystore_GetMnemonic_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeystoreServer).GetMnemonic(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Keystore/GetMnemonic",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeystoreServer).GetMnemonic(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Keystore_GetKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PublicKey)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeystoreServer).GetKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Keystore/GetKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeystoreServer).GetKey(ctx, req.(*PublicKey))
	}
	return interceptor(ctx, in, info, handler)
}

func _Keystore_GetKeys_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeystoreServer).GetKeys(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Keystore/GetKeys",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeystoreServer).GetKeys(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Keystore_ToggleKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ToggleKeyMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeystoreServer).ToggleKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Keystore/ToggleKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeystoreServer).ToggleKey(ctx, req.(*ToggleKeyMsg))
	}
	return interceptor(ctx, in, info, handler)
}

var _Keystore_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Keystore",
	HandlerType: (*KeystoreServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GenerateKeys",
			Handler:    _Keystore_GenerateKeys_Handler,
		},
		{
			MethodName: "GetMnemonic",
			Handler:    _Keystore_GetMnemonic_Handler,
		},
		{
			MethodName: "GetKey",
			Handler:    _Keystore_GetKey_Handler,
		},
		{
			MethodName: "GetKeys",
			Handler:    _Keystore_GetKeys_Handler,
		},
		{
			MethodName: "ToggleKey",
			Handler:    _Keystore_ToggleKey_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "keystore.proto",
}