// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.24.0
// 	protoc        v3.11.4
// source: chainrpc/proto/chain.proto

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
	sync "sync"
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

type GetBlockInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BlockHash string `protobuf:"bytes,1,opt,name=block_hash,json=blockHash,proto3" json:"block_hash,omitempty"`
}

func (x *GetBlockInfo) Reset() {
	*x = GetBlockInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_proto_chain_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBlockInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBlockInfo) ProtoMessage() {}

func (x *GetBlockInfo) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_proto_chain_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBlockInfo.ProtoReflect.Descriptor instead.
func (*GetBlockInfo) Descriptor() ([]byte, []int) {
	return file_chainrpc_proto_chain_proto_rawDescGZIP(), []int{0}
}

func (x *GetBlockInfo) GetBlockHash() string {
	if x != nil {
		return x.BlockHash
	}
	return ""
}

type GetBlockHashInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BlockHeigth uint64 `protobuf:"varint,1,opt,name=block_heigth,json=blockHeigth,proto3" json:"block_heigth,omitempty"`
}

func (x *GetBlockHashInfo) Reset() {
	*x = GetBlockHashInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_proto_chain_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBlockHashInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBlockHashInfo) ProtoMessage() {}

func (x *GetBlockHashInfo) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_proto_chain_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBlockHashInfo.ProtoReflect.Descriptor instead.
func (*GetBlockHashInfo) Descriptor() ([]byte, []int) {
	return file_chainrpc_proto_chain_proto_rawDescGZIP(), []int{1}
}

func (x *GetBlockHashInfo) GetBlockHeigth() uint64 {
	if x != nil {
		return x.BlockHeigth
	}
	return 0
}

type GetBlockResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hash            string       `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	Header          *BlockHeader `protobuf:"bytes,2,opt,name=header,proto3" json:"header,omitempty"`
	Txs             []string     `protobuf:"bytes,3,rep,name=txs,proto3" json:"txs,omitempty"`
	Signature       string       `protobuf:"bytes,4,opt,name=signature,proto3" json:"signature,omitempty"`
	RandaoSignature string       `protobuf:"bytes,5,opt,name=randao_signature,json=randaoSignature,proto3" json:"randao_signature,omitempty"`
}

func (x *GetBlockResponse) Reset() {
	*x = GetBlockResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_proto_chain_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBlockResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBlockResponse) ProtoMessage() {}

func (x *GetBlockResponse) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_proto_chain_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBlockResponse.ProtoReflect.Descriptor instead.
func (*GetBlockResponse) Descriptor() ([]byte, []int) {
	return file_chainrpc_proto_chain_proto_rawDescGZIP(), []int{2}
}

func (x *GetBlockResponse) GetHash() string {
	if x != nil {
		return x.Hash
	}
	return ""
}

func (x *GetBlockResponse) GetHeader() *BlockHeader {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *GetBlockResponse) GetTxs() []string {
	if x != nil {
		return x.Txs
	}
	return nil
}

func (x *GetBlockResponse) GetSignature() string {
	if x != nil {
		return x.Signature
	}
	return ""
}

func (x *GetBlockResponse) GetRandaoSignature() string {
	if x != nil {
		return x.RandaoSignature
	}
	return ""
}

type GetBlockRawResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RawBlock string `protobuf:"bytes,1,opt,name=raw_block,json=rawBlock,proto3" json:"raw_block,omitempty"`
}

func (x *GetBlockRawResponse) Reset() {
	*x = GetBlockRawResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_proto_chain_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBlockRawResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBlockRawResponse) ProtoMessage() {}

func (x *GetBlockRawResponse) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_proto_chain_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBlockRawResponse.ProtoReflect.Descriptor instead.
func (*GetBlockRawResponse) Descriptor() ([]byte, []int) {
	return file_chainrpc_proto_chain_proto_rawDescGZIP(), []int{3}
}

func (x *GetBlockRawResponse) GetRawBlock() string {
	if x != nil {
		return x.RawBlock
	}
	return ""
}

type GetBlockHashResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BlockHash string `protobuf:"bytes,1,opt,name=block_hash,json=blockHash,proto3" json:"block_hash,omitempty"`
}

func (x *GetBlockHashResponse) Reset() {
	*x = GetBlockHashResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_proto_chain_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBlockHashResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBlockHashResponse) ProtoMessage() {}

func (x *GetBlockHashResponse) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_proto_chain_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBlockHashResponse.ProtoReflect.Descriptor instead.
func (*GetBlockHashResponse) Descriptor() ([]byte, []int) {
	return file_chainrpc_proto_chain_proto_rawDescGZIP(), []int{4}
}

func (x *GetBlockHashResponse) GetBlockHash() string {
	if x != nil {
		return x.BlockHash
	}
	return ""
}

type BlockHeader struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version                    int32  `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Nonce                      int32  `protobuf:"varint,2,opt,name=nonce,proto3" json:"nonce,omitempty"`
	TxMerkleRoot               string `protobuf:"bytes,3,opt,name=tx_merkle_root,json=txMerkleRoot,proto3" json:"tx_merkle_root,omitempty"`
	VoteMerkleRoot             string `protobuf:"bytes,4,opt,name=vote_merkle_root,json=voteMerkleRoot,proto3" json:"vote_merkle_root,omitempty"`
	DepositMerkleRoot          string `protobuf:"bytes,5,opt,name=deposit_merkle_root,json=depositMerkleRoot,proto3" json:"deposit_merkle_root,omitempty"`
	ExitMerkleRoot             string `protobuf:"bytes,6,opt,name=exit_merkle_root,json=exitMerkleRoot,proto3" json:"exit_merkle_root,omitempty"`
	VoteSlashingMerkleRoot     string `protobuf:"bytes,7,opt,name=vote_slashing_merkle_root,json=voteSlashingMerkleRoot,proto3" json:"vote_slashing_merkle_root,omitempty"`
	RandaoSlashingMerkleRoot   string `protobuf:"bytes,8,opt,name=randao_slashing_merkle_root,json=randaoSlashingMerkleRoot,proto3" json:"randao_slashing_merkle_root,omitempty"`
	ProposerSlashingMerkleRoot string `protobuf:"bytes,9,opt,name=proposer_slashing_merkle_root,json=proposerSlashingMerkleRoot,proto3" json:"proposer_slashing_merkle_root,omitempty"`
	PrevBlockHash              string `protobuf:"bytes,10,opt,name=prev_block_hash,json=prevBlockHash,proto3" json:"prev_block_hash,omitempty"`
	Timestamp                  int64  `protobuf:"varint,11,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Slot                       uint64 `protobuf:"varint,12,opt,name=slot,proto3" json:"slot,omitempty"`
	StateRoot                  string `protobuf:"bytes,13,opt,name=state_root,json=stateRoot,proto3" json:"state_root,omitempty"`
	FeeAddress                 string `protobuf:"bytes,14,opt,name=fee_address,json=feeAddress,proto3" json:"fee_address,omitempty"`
}

func (x *BlockHeader) Reset() {
	*x = BlockHeader{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_proto_chain_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockHeader) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockHeader) ProtoMessage() {}

func (x *BlockHeader) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_proto_chain_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockHeader.ProtoReflect.Descriptor instead.
func (*BlockHeader) Descriptor() ([]byte, []int) {
	return file_chainrpc_proto_chain_proto_rawDescGZIP(), []int{5}
}

func (x *BlockHeader) GetVersion() int32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *BlockHeader) GetNonce() int32 {
	if x != nil {
		return x.Nonce
	}
	return 0
}

func (x *BlockHeader) GetTxMerkleRoot() string {
	if x != nil {
		return x.TxMerkleRoot
	}
	return ""
}

func (x *BlockHeader) GetVoteMerkleRoot() string {
	if x != nil {
		return x.VoteMerkleRoot
	}
	return ""
}

func (x *BlockHeader) GetDepositMerkleRoot() string {
	if x != nil {
		return x.DepositMerkleRoot
	}
	return ""
}

func (x *BlockHeader) GetExitMerkleRoot() string {
	if x != nil {
		return x.ExitMerkleRoot
	}
	return ""
}

func (x *BlockHeader) GetVoteSlashingMerkleRoot() string {
	if x != nil {
		return x.VoteSlashingMerkleRoot
	}
	return ""
}

func (x *BlockHeader) GetRandaoSlashingMerkleRoot() string {
	if x != nil {
		return x.RandaoSlashingMerkleRoot
	}
	return ""
}

func (x *BlockHeader) GetProposerSlashingMerkleRoot() string {
	if x != nil {
		return x.ProposerSlashingMerkleRoot
	}
	return ""
}

func (x *BlockHeader) GetPrevBlockHash() string {
	if x != nil {
		return x.PrevBlockHash
	}
	return ""
}

func (x *BlockHeader) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *BlockHeader) GetSlot() uint64 {
	if x != nil {
		return x.Slot
	}
	return 0
}

func (x *BlockHeader) GetStateRoot() string {
	if x != nil {
		return x.StateRoot
	}
	return ""
}

func (x *BlockHeader) GetFeeAddress() string {
	if x != nil {
		return x.FeeAddress
	}
	return ""
}

type ChainInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BlockHash   string `protobuf:"bytes,1,opt,name=block_hash,json=blockHash,proto3" json:"block_hash,omitempty"`
	BlockHeight uint64 `protobuf:"varint,2,opt,name=block_height,json=blockHeight,proto3" json:"block_height,omitempty"`
	Validators  uint64 `protobuf:"varint,3,opt,name=validators,proto3" json:"validators,omitempty"`
}

func (x *ChainInfoResponse) Reset() {
	*x = ChainInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_proto_chain_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChainInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChainInfoResponse) ProtoMessage() {}

func (x *ChainInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_proto_chain_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChainInfoResponse.ProtoReflect.Descriptor instead.
func (*ChainInfoResponse) Descriptor() ([]byte, []int) {
	return file_chainrpc_proto_chain_proto_rawDescGZIP(), []int{6}
}

func (x *ChainInfoResponse) GetBlockHash() string {
	if x != nil {
		return x.BlockHash
	}
	return ""
}

func (x *ChainInfoResponse) GetBlockHeight() uint64 {
	if x != nil {
		return x.BlockHeight
	}
	return 0
}

func (x *ChainInfoResponse) GetValidators() uint64 {
	if x != nil {
		return x.Validators
	}
	return 0
}

var File_chainrpc_proto_chain_proto protoreflect.FileDescriptor

var file_chainrpc_proto_chain_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72, 0x70, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x63, 0x68,
	0x61, 0x69, 0x6e, 0x72, 0x70, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x2d, 0x0a, 0x0c, 0x47, 0x65, 0x74,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1d, 0x0a, 0x0a, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68, 0x22, 0x35, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x21, 0x0a, 0x0c,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x68, 0x65, 0x69, 0x67, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67, 0x74, 0x68, 0x22,
	0xa7, 0x01, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x12, 0x24, 0x0a, 0x06, 0x68, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b,
	0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x10,
	0x0a, 0x03, 0x74, 0x78, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x03, 0x74, 0x78, 0x73,
	0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x29,
	0x0a, 0x10, 0x72, 0x61, 0x6e, 0x64, 0x61, 0x6f, 0x5f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x72, 0x61, 0x6e, 0x64, 0x61, 0x6f,
	0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x22, 0x32, 0x0a, 0x13, 0x47, 0x65, 0x74,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x61, 0x77, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x1b, 0x0a, 0x09, 0x72, 0x61, 0x77, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x61, 0x77, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x35, 0x0a,
	0x14, 0x47, 0x65, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x68,
	0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x48, 0x61, 0x73, 0x68, 0x22, 0xbe, 0x04, 0x0a, 0x0b, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x14,
	0x0a, 0x05, 0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6e,
	0x6f, 0x6e, 0x63, 0x65, 0x12, 0x24, 0x0a, 0x0e, 0x74, 0x78, 0x5f, 0x6d, 0x65, 0x72, 0x6b, 0x6c,
	0x65, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x74, 0x78,
	0x4d, 0x65, 0x72, 0x6b, 0x6c, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x28, 0x0a, 0x10, 0x76, 0x6f,
	0x74, 0x65, 0x5f, 0x6d, 0x65, 0x72, 0x6b, 0x6c, 0x65, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x76, 0x6f, 0x74, 0x65, 0x4d, 0x65, 0x72, 0x6b, 0x6c, 0x65,
	0x52, 0x6f, 0x6f, 0x74, 0x12, 0x2e, 0x0a, 0x13, 0x64, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x5f,
	0x6d, 0x65, 0x72, 0x6b, 0x6c, 0x65, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x11, 0x64, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x4d, 0x65, 0x72, 0x6b, 0x6c, 0x65,
	0x52, 0x6f, 0x6f, 0x74, 0x12, 0x28, 0x0a, 0x10, 0x65, 0x78, 0x69, 0x74, 0x5f, 0x6d, 0x65, 0x72,
	0x6b, 0x6c, 0x65, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e,
	0x65, 0x78, 0x69, 0x74, 0x4d, 0x65, 0x72, 0x6b, 0x6c, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x39,
	0x0a, 0x19, 0x76, 0x6f, 0x74, 0x65, 0x5f, 0x73, 0x6c, 0x61, 0x73, 0x68, 0x69, 0x6e, 0x67, 0x5f,
	0x6d, 0x65, 0x72, 0x6b, 0x6c, 0x65, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x16, 0x76, 0x6f, 0x74, 0x65, 0x53, 0x6c, 0x61, 0x73, 0x68, 0x69, 0x6e, 0x67, 0x4d,
	0x65, 0x72, 0x6b, 0x6c, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x3d, 0x0a, 0x1b, 0x72, 0x61, 0x6e,
	0x64, 0x61, 0x6f, 0x5f, 0x73, 0x6c, 0x61, 0x73, 0x68, 0x69, 0x6e, 0x67, 0x5f, 0x6d, 0x65, 0x72,
	0x6b, 0x6c, 0x65, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x18,
	0x72, 0x61, 0x6e, 0x64, 0x61, 0x6f, 0x53, 0x6c, 0x61, 0x73, 0x68, 0x69, 0x6e, 0x67, 0x4d, 0x65,
	0x72, 0x6b, 0x6c, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x41, 0x0a, 0x1d, 0x70, 0x72, 0x6f, 0x70,
	0x6f, 0x73, 0x65, 0x72, 0x5f, 0x73, 0x6c, 0x61, 0x73, 0x68, 0x69, 0x6e, 0x67, 0x5f, 0x6d, 0x65,
	0x72, 0x6b, 0x6c, 0x65, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x1a, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65, 0x72, 0x53, 0x6c, 0x61, 0x73, 0x68, 0x69, 0x6e,
	0x67, 0x4d, 0x65, 0x72, 0x6b, 0x6c, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x26, 0x0a, 0x0f, 0x70,
	0x72, 0x65, 0x76, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x0a,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x70, 0x72, 0x65, 0x76, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48,
	0x61, 0x73, 0x68, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x18, 0x0b, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x6c, 0x6f, 0x74, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x04, 0x73, 0x6c, 0x6f, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x74, 0x65, 0x5f, 0x72,
	0x6f, 0x6f, 0x74, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x74, 0x61, 0x74, 0x65,
	0x52, 0x6f, 0x6f, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x66, 0x65, 0x65, 0x5f, 0x61, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x66, 0x65, 0x65, 0x41, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x22, 0x75, 0x0a, 0x11, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x6e,
	0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68, 0x12, 0x21, 0x0a, 0x0c, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x1e, 0x0a, 0x0a,
	0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x0a, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x73, 0x32, 0xd7, 0x01, 0x0a,
	0x05, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x12, 0x2c, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x43, 0x68, 0x61,
	0x69, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x06, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x12,
	0x2e, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x34, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x52, 0x61, 0x77, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x12, 0x0d, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x49, 0x6e,
	0x66, 0x6f, 0x1a, 0x14, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x61, 0x77,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x2e, 0x0a, 0x08, 0x47, 0x65,
	0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x0d, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x11, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x3a, 0x0a, 0x0c, 0x47, 0x65,
	0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68, 0x12, 0x11, 0x2e, 0x47, 0x65, 0x74,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x15, 0x2e,
	0x47, 0x65, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6f, 0x6c, 0x79, 0x6d, 0x70, 0x75, 0x73, 0x2d, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2f, 0x6f, 0x67, 0x65, 0x6e, 0x2f, 0x63, 0x68, 0x61, 0x69, 0x6e,
	0x72, 0x70, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_chainrpc_proto_chain_proto_rawDescOnce sync.Once
	file_chainrpc_proto_chain_proto_rawDescData = file_chainrpc_proto_chain_proto_rawDesc
)

func file_chainrpc_proto_chain_proto_rawDescGZIP() []byte {
	file_chainrpc_proto_chain_proto_rawDescOnce.Do(func() {
		file_chainrpc_proto_chain_proto_rawDescData = protoimpl.X.CompressGZIP(file_chainrpc_proto_chain_proto_rawDescData)
	})
	return file_chainrpc_proto_chain_proto_rawDescData
}

var file_chainrpc_proto_chain_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_chainrpc_proto_chain_proto_goTypes = []interface{}{
	(*GetBlockInfo)(nil),         // 0: GetBlockInfo
	(*GetBlockHashInfo)(nil),     // 1: GetBlockHashInfo
	(*GetBlockResponse)(nil),     // 2: GetBlockResponse
	(*GetBlockRawResponse)(nil),  // 3: GetBlockRawResponse
	(*GetBlockHashResponse)(nil), // 4: GetBlockHashResponse
	(*BlockHeader)(nil),          // 5: BlockHeader
	(*ChainInfoResponse)(nil),    // 6: ChainInfoResponse
	(*Empty)(nil),                // 7: Empty
}
var file_chainrpc_proto_chain_proto_depIdxs = []int32{
	5, // 0: GetBlockResponse.header:type_name -> BlockHeader
	7, // 1: Chain.GetChainInfo:input_type -> Empty
	0, // 2: Chain.GetRawBlock:input_type -> GetBlockInfo
	0, // 3: Chain.GetBlock:input_type -> GetBlockInfo
	1, // 4: Chain.GetBlockHash:input_type -> GetBlockHashInfo
	6, // 5: Chain.GetChainInfo:output_type -> ChainInfoResponse
	3, // 6: Chain.GetRawBlock:output_type -> GetBlockRawResponse
	2, // 7: Chain.GetBlock:output_type -> GetBlockResponse
	4, // 8: Chain.GetBlockHash:output_type -> GetBlockHashResponse
	5, // [5:9] is the sub-list for method output_type
	1, // [1:5] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_chainrpc_proto_chain_proto_init() }
func file_chainrpc_proto_chain_proto_init() {
	if File_chainrpc_proto_chain_proto != nil {
		return
	}
	file_chainrpc_proto_common_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_chainrpc_proto_chain_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBlockInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chainrpc_proto_chain_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBlockHashInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chainrpc_proto_chain_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBlockResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chainrpc_proto_chain_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBlockRawResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chainrpc_proto_chain_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBlockHashResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chainrpc_proto_chain_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockHeader); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chainrpc_proto_chain_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChainInfoResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_chainrpc_proto_chain_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_chainrpc_proto_chain_proto_goTypes,
		DependencyIndexes: file_chainrpc_proto_chain_proto_depIdxs,
		MessageInfos:      file_chainrpc_proto_chain_proto_msgTypes,
	}.Build()
	File_chainrpc_proto_chain_proto = out.File
	file_chainrpc_proto_chain_proto_rawDesc = nil
	file_chainrpc_proto_chain_proto_goTypes = nil
	file_chainrpc_proto_chain_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// ChainClient is the client API for Chain service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ChainClient interface {
	GetChainInfo(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ChainInfoResponse, error)
	GetRawBlock(ctx context.Context, in *GetBlockInfo, opts ...grpc.CallOption) (*GetBlockRawResponse, error)
	GetBlock(ctx context.Context, in *GetBlockInfo, opts ...grpc.CallOption) (*GetBlockResponse, error)
	GetBlockHash(ctx context.Context, in *GetBlockHashInfo, opts ...grpc.CallOption) (*GetBlockHashResponse, error)
}

type chainClient struct {
	cc grpc.ClientConnInterface
}

func NewChainClient(cc grpc.ClientConnInterface) ChainClient {
	return &chainClient{cc}
}

func (c *chainClient) GetChainInfo(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ChainInfoResponse, error) {
	out := new(ChainInfoResponse)
	err := c.cc.Invoke(ctx, "/Chain/GetChainInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chainClient) GetRawBlock(ctx context.Context, in *GetBlockInfo, opts ...grpc.CallOption) (*GetBlockRawResponse, error) {
	out := new(GetBlockRawResponse)
	err := c.cc.Invoke(ctx, "/Chain/GetRawBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chainClient) GetBlock(ctx context.Context, in *GetBlockInfo, opts ...grpc.CallOption) (*GetBlockResponse, error) {
	out := new(GetBlockResponse)
	err := c.cc.Invoke(ctx, "/Chain/GetBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chainClient) GetBlockHash(ctx context.Context, in *GetBlockHashInfo, opts ...grpc.CallOption) (*GetBlockHashResponse, error) {
	out := new(GetBlockHashResponse)
	err := c.cc.Invoke(ctx, "/Chain/GetBlockHash", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChainServer is the server API for Chain service.
type ChainServer interface {
	GetChainInfo(context.Context, *Empty) (*ChainInfoResponse, error)
	GetRawBlock(context.Context, *GetBlockInfo) (*GetBlockRawResponse, error)
	GetBlock(context.Context, *GetBlockInfo) (*GetBlockResponse, error)
	GetBlockHash(context.Context, *GetBlockHashInfo) (*GetBlockHashResponse, error)
}

// UnimplementedChainServer can be embedded to have forward compatible implementations.
type UnimplementedChainServer struct {
}

func (*UnimplementedChainServer) GetChainInfo(context.Context, *Empty) (*ChainInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetChainInfo not implemented")
}
func (*UnimplementedChainServer) GetRawBlock(context.Context, *GetBlockInfo) (*GetBlockRawResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRawBlock not implemented")
}
func (*UnimplementedChainServer) GetBlock(context.Context, *GetBlockInfo) (*GetBlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlock not implemented")
}
func (*UnimplementedChainServer) GetBlockHash(context.Context, *GetBlockHashInfo) (*GetBlockHashResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlockHash not implemented")
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
	in := new(GetBlockInfo)
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
		return srv.(ChainServer).GetRawBlock(ctx, req.(*GetBlockInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chain_GetBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockInfo)
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
		return srv.(ChainServer).GetBlock(ctx, req.(*GetBlockInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chain_GetBlockHash_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockHashInfo)
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
		return srv.(ChainServer).GetBlockHash(ctx, req.(*GetBlockHashInfo))
	}
	return interceptor(ctx, in, info, handler)
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
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "chainrpc/proto/chain.proto",
}