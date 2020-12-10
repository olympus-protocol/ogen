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

// ChainClient is the client API for Chain service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChainClient interface {
	GetChainInfo(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ChainInfo, error)
	GetRawBlock(ctx context.Context, in *Hash, opts ...grpc.CallOption) (*Block, error)
	GetBlock(ctx context.Context, in *Hash, opts ...grpc.CallOption) (*Block, error)
	GetBlockHash(ctx context.Context, in *Number, opts ...grpc.CallOption) (*Hash, error)
	GetAccountInfo(ctx context.Context, in *Account, opts ...grpc.CallOption) (*AccountInfo, error)
	Sync(ctx context.Context, in *Hash, opts ...grpc.CallOption) (Chain_SyncClient, error)
	SubscribeBlocks(ctx context.Context, in *Empty, opts ...grpc.CallOption) (Chain_SubscribeBlocksClient, error)
}

type chainClient struct {
	cc grpc.ClientConnInterface
}

func NewChainClient(cc grpc.ClientConnInterface) ChainClient {
	return &chainClient{cc}
}

func (c *chainClient) GetChainInfo(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ChainInfo, error) {
	out := new(ChainInfo)
	err := c.cc.Invoke(ctx, "/Chain/GetChainInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chainClient) GetRawBlock(ctx context.Context, in *Hash, opts ...grpc.CallOption) (*Block, error) {
	out := new(Block)
	err := c.cc.Invoke(ctx, "/Chain/GetRawBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chainClient) GetBlock(ctx context.Context, in *Hash, opts ...grpc.CallOption) (*Block, error) {
	out := new(Block)
	err := c.cc.Invoke(ctx, "/Chain/GetBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chainClient) GetBlockHash(ctx context.Context, in *Number, opts ...grpc.CallOption) (*Hash, error) {
	out := new(Hash)
	err := c.cc.Invoke(ctx, "/Chain/GetBlockHash", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chainClient) GetAccountInfo(ctx context.Context, in *Account, opts ...grpc.CallOption) (*AccountInfo, error) {
	out := new(AccountInfo)
	err := c.cc.Invoke(ctx, "/Chain/GetAccountInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chainClient) Sync(ctx context.Context, in *Hash, opts ...grpc.CallOption) (Chain_SyncClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Chain_serviceDesc.Streams[0], "/Chain/Sync", opts...)
	if err != nil {
		return nil, err
	}
	x := &chainSyncClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Chain_SyncClient interface {
	Recv() (*RawData, error)
	grpc.ClientStream
}

type chainSyncClient struct {
	grpc.ClientStream
}

func (x *chainSyncClient) Recv() (*RawData, error) {
	m := new(RawData)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *chainClient) SubscribeBlocks(ctx context.Context, in *Empty, opts ...grpc.CallOption) (Chain_SubscribeBlocksClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Chain_serviceDesc.Streams[1], "/Chain/SubscribeBlocks", opts...)
	if err != nil {
		return nil, err
	}
	x := &chainSubscribeBlocksClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Chain_SubscribeBlocksClient interface {
	Recv() (*RawData, error)
	grpc.ClientStream
}

type chainSubscribeBlocksClient struct {
	grpc.ClientStream
}

func (x *chainSubscribeBlocksClient) Recv() (*RawData, error) {
	m := new(RawData)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ChainServer is the server API for Chain service.
// All implementations must embed UnimplementedChainServer
// for forward compatibility
type ChainServer interface {
	GetChainInfo(context.Context, *Empty) (*ChainInfo, error)
	GetRawBlock(context.Context, *Hash) (*Block, error)
	GetBlock(context.Context, *Hash) (*Block, error)
	GetBlockHash(context.Context, *Number) (*Hash, error)
	GetAccountInfo(context.Context, *Account) (*AccountInfo, error)
	Sync(*Hash, Chain_SyncServer) error
	SubscribeBlocks(*Empty, Chain_SubscribeBlocksServer) error
	mustEmbedUnimplementedChainServer()
}

// UnimplementedChainServer must be embedded to have forward compatible implementations.
type UnimplementedChainServer struct {
}

func (UnimplementedChainServer) GetChainInfo(context.Context, *Empty) (*ChainInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetChainInfo not implemented")
}
func (UnimplementedChainServer) GetRawBlock(context.Context, *Hash) (*Block, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRawBlock not implemented")
}
func (UnimplementedChainServer) GetBlock(context.Context, *Hash) (*Block, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlock not implemented")
}
func (UnimplementedChainServer) GetBlockHash(context.Context, *Number) (*Hash, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlockHash not implemented")
}
func (UnimplementedChainServer) GetAccountInfo(context.Context, *Account) (*AccountInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccountInfo not implemented")
}
func (UnimplementedChainServer) Sync(*Hash, Chain_SyncServer) error {
	return status.Errorf(codes.Unimplemented, "method Sync not implemented")
}
func (UnimplementedChainServer) SubscribeBlocks(*Empty, Chain_SubscribeBlocksServer) error {
	return status.Errorf(codes.Unimplemented, "method SubscribeBlocks not implemented")
}
func (UnimplementedChainServer) mustEmbedUnimplementedChainServer() {}

// UnsafeChainServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChainServer will
// result in compilation errors.
type UnsafeChainServer interface {
	mustEmbedUnimplementedChainServer()
}

func RegisterChainServer(s *grpc.Server, srv ChainServer) {
	s.RegisterService(&_Chain_serviceDesc, srv)
}

func _Chain_GetChainInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChainServer).GetChainInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chain/GetChainInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChainServer).GetChainInfo(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chain_GetRawBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Hash)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChainServer).GetRawBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chain/GetRawBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChainServer).GetRawBlock(ctx, req.(*Hash))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chain_GetBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Hash)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChainServer).GetBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chain/GetBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChainServer).GetBlock(ctx, req.(*Hash))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chain_GetBlockHash_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Number)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChainServer).GetBlockHash(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chain/GetBlockHash",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChainServer).GetBlockHash(ctx, req.(*Number))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chain_GetAccountInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Account)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChainServer).GetAccountInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chain/GetAccountInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChainServer).GetAccountInfo(ctx, req.(*Account))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chain_Sync_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Hash)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ChainServer).Sync(m, &chainSyncServer{stream})
}

type Chain_SyncServer interface {
	Send(*RawData) error
	grpc.ServerStream
}

type chainSyncServer struct {
	grpc.ServerStream
}

func (x *chainSyncServer) Send(m *RawData) error {
	return x.ServerStream.SendMsg(m)
}

func _Chain_SubscribeBlocks_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ChainServer).SubscribeBlocks(m, &chainSubscribeBlocksServer{stream})
}

type Chain_SubscribeBlocksServer interface {
	Send(*RawData) error
	grpc.ServerStream
}

type chainSubscribeBlocksServer struct {
	grpc.ServerStream
}

func (x *chainSubscribeBlocksServer) Send(m *RawData) error {
	return x.ServerStream.SendMsg(m)
}

var _Chain_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Chain",
	HandlerType: (*ChainServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetChainInfo",
			Handler:    _Chain_GetChainInfo_Handler,
		},
		{
			MethodName: "GetRawBlock",
			Handler:    _Chain_GetRawBlock_Handler,
		},
		{
			MethodName: "GetBlock",
			Handler:    _Chain_GetBlock_Handler,
		},
		{
			MethodName: "GetBlockHash",
			Handler:    _Chain_GetBlockHash_Handler,
		},
		{
			MethodName: "GetAccountInfo",
			Handler:    _Chain_GetAccountInfo_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Sync",
			Handler:       _Chain_Sync_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "SubscribeBlocks",
			Handler:       _Chain_SubscribeBlocks_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "chain.proto",
}
