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
	//Method: GenValidatorKey
	//Input: message GenValidatorKeys
	//Response: message KeyPairs
	//Description: Returns private keys generated for validators start.
	GenValidatorKey(ctx context.Context, in *GenValidatorKeys, opts ...grpc.CallOption) (*KeyPairs, error)
}

type keystoreClient struct {
	cc grpc.ClientConnInterface
}

func NewKeystoreClient(cc grpc.ClientConnInterface) KeystoreClient {
	return &keystoreClient{cc}
}

func (c *keystoreClient) GenValidatorKey(ctx context.Context, in *GenValidatorKeys, opts ...grpc.CallOption) (*KeyPairs, error) {
	out := new(KeyPairs)
	err := c.cc.Invoke(ctx, "/Keystore/GenValidatorKey", in, out, opts...)
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
	//Method: GenValidatorKey
	//Input: message GenValidatorKeys
	//Response: message KeyPairs
	//Description: Returns private keys generated for validators start.
	GenValidatorKey(context.Context, *GenValidatorKeys) (*KeyPairs, error)
	mustEmbedUnimplementedKeystoreServer()
}

// UnimplementedKeystoreServer must be embedded to have forward compatible implementations.
type UnimplementedKeystoreServer struct {
}

func (UnimplementedKeystoreServer) GenValidatorKey(context.Context, *GenValidatorKeys) (*KeyPairs, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenValidatorKey not implemented")
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

func _Keystore_GenValidatorKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenValidatorKeys)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeystoreServer).GenValidatorKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Keystore/GenValidatorKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeystoreServer).GenValidatorKey(ctx, req.(*GenValidatorKeys))
	}
	return interceptor(ctx, in, info, handler)
}

var _Keystore_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Keystore",
	HandlerType: (*KeystoreServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GenValidatorKey",
			Handler:    _Keystore_GenValidatorKey_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "keystore.proto",
}
