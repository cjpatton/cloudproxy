// Code generated by protoc-gen-go.
// source: tao_rpc.proto
// DO NOT EDIT!

/*
Package tao is a generated protocol buffer package.

It is generated from these files:
	tao_rpc.proto

It has these top-level messages:
	TaoRPCRequest
	TaoRPCResponse
*/
package tao

import proto "code.google.com/p/goprotobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type TaoRPCOperation int32

const (
	TaoRPCOperation_TAO_RPC_GET_RANDOM_BYTES  TaoRPCOperation = 1
	TaoRPCOperation_TAO_RPC_SEAL              TaoRPCOperation = 2
	TaoRPCOperation_TAO_RPC_UNSEAL            TaoRPCOperation = 3
	TaoRPCOperation_TAO_RPC_ATTEST            TaoRPCOperation = 4
	TaoRPCOperation_TAO_RPC_GET_TAO_NAME      TaoRPCOperation = 5
	TaoRPCOperation_TAO_RPC_EXTEND_TAO_NAME   TaoRPCOperation = 6
	TaoRPCOperation_TAO_RPC_GET_SHARED_SECRET TaoRPCOperation = 7
)

var TaoRPCOperation_name = map[int32]string{
	1: "TAO_RPC_GET_RANDOM_BYTES",
	2: "TAO_RPC_SEAL",
	3: "TAO_RPC_UNSEAL",
	4: "TAO_RPC_ATTEST",
	5: "TAO_RPC_GET_TAO_NAME",
	6: "TAO_RPC_EXTEND_TAO_NAME",
	7: "TAO_RPC_GET_SHARED_SECRET",
}
var TaoRPCOperation_value = map[string]int32{
	"TAO_RPC_GET_RANDOM_BYTES":  1,
	"TAO_RPC_SEAL":              2,
	"TAO_RPC_UNSEAL":            3,
	"TAO_RPC_ATTEST":            4,
	"TAO_RPC_GET_TAO_NAME":      5,
	"TAO_RPC_EXTEND_TAO_NAME":   6,
	"TAO_RPC_GET_SHARED_SECRET": 7,
}

func (x TaoRPCOperation) Enum() *TaoRPCOperation {
	p := new(TaoRPCOperation)
	*p = x
	return p
}
func (x TaoRPCOperation) String() string {
	return proto.EnumName(TaoRPCOperation_name, int32(x))
}
func (x *TaoRPCOperation) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(TaoRPCOperation_value, data, "TaoRPCOperation")
	if err != nil {
		return err
	}
	*x = TaoRPCOperation(value)
	return nil
}

type TaoRPCRequest struct {
	Rpc              *TaoRPCOperation `protobuf:"varint,1,req,name=rpc,enum=tao.TaoRPCOperation" json:"rpc,omitempty"`
	Seq              *uint64          `protobuf:"varint,2,req,name=seq" json:"seq,omitempty"`
	Data             []byte           `protobuf:"bytes,3,opt,name=data" json:"data,omitempty"`
	Size             *int32           `protobuf:"varint,4,opt,name=size" json:"size,omitempty"`
	Policy           *string          `protobuf:"bytes,5,opt,name=policy" json:"policy,omitempty"`
	XXX_unrecognized []byte           `json:"-"`
}

func (m *TaoRPCRequest) Reset()         { *m = TaoRPCRequest{} }
func (m *TaoRPCRequest) String() string { return proto.CompactTextString(m) }
func (*TaoRPCRequest) ProtoMessage()    {}

func (m *TaoRPCRequest) GetRpc() TaoRPCOperation {
	if m != nil && m.Rpc != nil {
		return *m.Rpc
	}
	return TaoRPCOperation_TAO_RPC_GET_RANDOM_BYTES
}

func (m *TaoRPCRequest) GetSeq() uint64 {
	if m != nil && m.Seq != nil {
		return *m.Seq
	}
	return 0
}

func (m *TaoRPCRequest) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *TaoRPCRequest) GetSize() int32 {
	if m != nil && m.Size != nil {
		return *m.Size
	}
	return 0
}

func (m *TaoRPCRequest) GetPolicy() string {
	if m != nil && m.Policy != nil {
		return *m.Policy
	}
	return ""
}

type TaoRPCResponse struct {
	Rpc              *TaoRPCOperation `protobuf:"varint,1,req,name=rpc,enum=tao.TaoRPCOperation" json:"rpc,omitempty"`
	Seq              *uint64          `protobuf:"varint,2,req,name=seq" json:"seq,omitempty"`
	Data             []byte           `protobuf:"bytes,3,opt,name=data" json:"data,omitempty"`
	Policy           *string          `protobuf:"bytes,4,opt,name=policy" json:"policy,omitempty"`
	XXX_unrecognized []byte           `json:"-"`
}

func (m *TaoRPCResponse) Reset()         { *m = TaoRPCResponse{} }
func (m *TaoRPCResponse) String() string { return proto.CompactTextString(m) }
func (*TaoRPCResponse) ProtoMessage()    {}

func (m *TaoRPCResponse) GetRpc() TaoRPCOperation {
	if m != nil && m.Rpc != nil {
		return *m.Rpc
	}
	return TaoRPCOperation_TAO_RPC_GET_RANDOM_BYTES
}

func (m *TaoRPCResponse) GetSeq() uint64 {
	if m != nil && m.Seq != nil {
		return *m.Seq
	}
	return 0
}

func (m *TaoRPCResponse) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *TaoRPCResponse) GetPolicy() string {
	if m != nil && m.Policy != nil {
		return *m.Policy
	}
	return ""
}

func init() {
	proto.RegisterEnum("tao.TaoRPCOperation", TaoRPCOperation_name, TaoRPCOperation_value)
}
