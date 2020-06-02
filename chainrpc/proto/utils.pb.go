// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.24.0
// 	protoc        v3.11.4
// source: chainrpc/proto/utils.proto

package proto

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

var File_chainrpc_proto_utils_proto protoreflect.FileDescriptor

var file_chainrpc_proto_utils_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72, 0x70, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x75, 0x74, 0x69, 0x6c, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x63, 0x68,
	0x61, 0x69, 0x6e, 0x72, 0x70, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xa9, 0x01, 0x0a, 0x05, 0x55, 0x74,
	0x69, 0x6c, 0x73, 0x12, 0x25, 0x0a, 0x0f, 0x47, 0x65, 0x6e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x6f, 0x72, 0x4b, 0x65, 0x79, 0x12, 0x06, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x08,
	0x2e, 0x4b, 0x65, 0x79, 0x50, 0x61, 0x69, 0x72, 0x22, 0x00, 0x12, 0x2a, 0x0a, 0x12, 0x53, 0x65,
	0x6e, 0x64, 0x52, 0x61, 0x77, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x08, 0x2e, 0x52, 0x61, 0x77, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x08, 0x2e, 0x53, 0x75, 0x63,
	0x63, 0x65, 0x73, 0x73, 0x22, 0x00, 0x12, 0x27, 0x0a, 0x14, 0x44, 0x65, 0x63, 0x6f, 0x64, 0x65,
	0x52, 0x61, 0x77, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x08,
	0x2e, 0x52, 0x61, 0x77, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x03, 0x2e, 0x54, 0x78, 0x22, 0x00, 0x12,
	0x24, 0x0a, 0x0e, 0x44, 0x65, 0x63, 0x6f, 0x64, 0x65, 0x52, 0x61, 0x77, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x12, 0x08, 0x2e, 0x52, 0x61, 0x77, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x06, 0x2e, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x22, 0x00, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x6f, 0x6c, 0x79, 0x6d, 0x70, 0x75, 0x73, 0x2d, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x2f, 0x6f, 0x67, 0x65, 0x6e, 0x2f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72,
	0x70, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_chainrpc_proto_utils_proto_goTypes = []interface{}{
	(*Empty)(nil),   // 0: Empty
	(*RawData)(nil), // 1: RawData
	(*KeyPair)(nil), // 2: KeyPair
	(*Success)(nil), // 3: Success
	(*Tx)(nil),      // 4: Tx
	(*Block)(nil),   // 5: Block
}
var file_chainrpc_proto_utils_proto_depIdxs = []int32{
	0, // 0: Utils.GenValidatorKey:input_type -> Empty
	1, // 1: Utils.SendRawTransaction:input_type -> RawData
	1, // 2: Utils.DecodeRawTransaction:input_type -> RawData
	1, // 3: Utils.DecodeRawBlock:input_type -> RawData
	2, // 4: Utils.GenValidatorKey:output_type -> KeyPair
	3, // 5: Utils.SendRawTransaction:output_type -> Success
	4, // 6: Utils.DecodeRawTransaction:output_type -> Tx
	5, // 7: Utils.DecodeRawBlock:output_type -> Block
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_chainrpc_proto_utils_proto_init() }
func file_chainrpc_proto_utils_proto_init() {
	if File_chainrpc_proto_utils_proto != nil {
		return
	}
	file_chainrpc_proto_common_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_chainrpc_proto_utils_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_chainrpc_proto_utils_proto_goTypes,
		DependencyIndexes: file_chainrpc_proto_utils_proto_depIdxs,
	}.Build()
	File_chainrpc_proto_utils_proto = out.File
	file_chainrpc_proto_utils_proto_rawDesc = nil
	file_chainrpc_proto_utils_proto_goTypes = nil
	file_chainrpc_proto_utils_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// UtilsClient is the client API for Utils service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type UtilsClient interface {
	//*
	//Method: GenValidatorKey
	//Input: message Empty
	//Response: message KeyPair
	//Description: Returns a validator public key and stores the private on the node keychain.
	GenValidatorKey(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*KeyPair, error)
	//*
	//Method: SendRawTransaction
	//Input: message RawData
	//Response: message Success
	//Description: Broadcast a transaction to the network.
	SendRawTransaction(ctx context.Context, in *RawData, opts ...grpc.CallOption) (*Success, error)
	//*
	//Method: DecodeRawTransaction
	//Input: message RawData
	//Response: message Tx
	//Description: Returns a raw transaction on human readable format.
	DecodeRawTransaction(ctx context.Context, in *RawData, opts ...grpc.CallOption) (*Tx, error)
	//*
	//Method: DecodeRawBlock
	//Input: message RawData
	//Response: message Block
	//Description: Returns a raw block on human readable format.
	DecodeRawBlock(ctx context.Context, in *RawData, opts ...grpc.CallOption) (*Block, error)
}

type utilsClient struct {
	cc grpc.ClientConnInterface
}

func NewUtilsClient(cc grpc.ClientConnInterface) UtilsClient {
	return &utilsClient{cc}
}

func (c *utilsClient) GenValidatorKey(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*KeyPair, error) {
	out := new(KeyPair)
	err := c.cc.Invoke(ctx, "/Utils/GenValidatorKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *utilsClient) SendRawTransaction(ctx context.Context, in *RawData, opts ...grpc.CallOption) (*Success, error) {
	out := new(Success)
	err := c.cc.Invoke(ctx, "/Utils/SendRawTransaction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *utilsClient) DecodeRawTransaction(ctx context.Context, in *RawData, opts ...grpc.CallOption) (*Tx, error) {
	out := new(Tx)
	err := c.cc.Invoke(ctx, "/Utils/DecodeRawTransaction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *utilsClient) DecodeRawBlock(ctx context.Context, in *RawData, opts ...grpc.CallOption) (*Block, error) {
	out := new(Block)
	err := c.cc.Invoke(ctx, "/Utils/DecodeRawBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UtilsServer is the server API for Utils service.
type UtilsServer interface {
	//*
	//Method: GenValidatorKey
	//Input: message Empty
	//Response: message KeyPair
	//Description: Returns a validator public key and stores the private on the node keychain.
	GenValidatorKey(context.Context, *Empty) (*KeyPair, error)
	//*
	//Method: SendRawTransaction
	//Input: message RawData
	//Response: message Success
	//Description: Broadcast a transaction to the network.
	SendRawTransaction(context.Context, *RawData) (*Success, error)
	//*
	//Method: DecodeRawTransaction
	//Input: message RawData
	//Response: message Tx
	//Description: Returns a raw transaction on human readable format.
	DecodeRawTransaction(context.Context, *RawData) (*Tx, error)
	//*
	//Method: DecodeRawBlock
	//Input: message RawData
	//Response: message Block
	//Description: Returns a raw block on human readable format.
	DecodeRawBlock(context.Context, *RawData) (*Block, error)
}

// UnimplementedUtilsServer can be embedded to have forward compatible implementations.
type UnimplementedUtilsServer struct {
}

func (*UnimplementedUtilsServer) GenValidatorKey(context.Context, *Empty) (*KeyPair, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenValidatorKey not implemented")
}
func (*UnimplementedUtilsServer) SendRawTransaction(context.Context, *RawData) (*Success, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendRawTransaction not implemented")
}
func (*UnimplementedUtilsServer) DecodeRawTransaction(context.Context, *RawData) (*Tx, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DecodeRawTransaction not implemented")
}
func (*UnimplementedUtilsServer) DecodeRawBlock(context.Context, *RawData) (*Block, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DecodeRawBlock not implemented")
}

func RegisterUtilsServer(s *grpc.Server, srv UtilsServer) {
	s.RegisterService(&_Utils_serviceDesc, srv)
}

func _Utils_GenValidatorKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UtilsServer).GenValidatorKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Utils/GenValidatorKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UtilsServer).GenValidatorKey(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Utils_SendRawTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RawData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UtilsServer).SendRawTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Utils/SendRawTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UtilsServer).SendRawTransaction(ctx, req.(*RawData))
	}
	return interceptor(ctx, in, info, handler)
}

func _Utils_DecodeRawTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RawData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UtilsServer).DecodeRawTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Utils/DecodeRawTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UtilsServer).DecodeRawTransaction(ctx, req.(*RawData))
	}
	return interceptor(ctx, in, info, handler)
}

func _Utils_DecodeRawBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RawData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UtilsServer).DecodeRawBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Utils/DecodeRawBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UtilsServer).DecodeRawBlock(ctx, req.(*RawData))
	}
	return interceptor(ctx, in, info, handler)
}

var _Utils_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Utils",
	HandlerType: (*UtilsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GenValidatorKey",
			Handler:    _Utils_GenValidatorKey_Handler,
		},
		{
			MethodName: "SendRawTransaction",
			Handler:    _Utils_SendRawTransaction_Handler,
		},
		{
			MethodName: "DecodeRawTransaction",
			Handler:    _Utils_DecodeRawTransaction_Handler,
		},
		{
			MethodName: "DecodeRawBlock",
			Handler:    _Utils_DecodeRawBlock_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "chainrpc/proto/utils.proto",
}
