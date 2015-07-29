// Code generated by protoc-gen-go.
// source: rpc.proto
// DO NOT EDIT!

package tao

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type RPCRequest struct {
	Data             []byte  `protobuf:"bytes,1,opt,name=data" json:"data,omitempty"`
	Size             *int32  `protobuf:"varint,2,opt,name=size" json:"size,omitempty"`
	Policy           *string `protobuf:"bytes,3,opt,name=policy" json:"policy,omitempty"`
	Time             *int64  `protobuf:"varint,4,opt,name=time" json:"time,omitempty"`
	Expiration       *int64  `protobuf:"varint,5,opt,name=expiration" json:"expiration,omitempty"`
	Issuer           []byte  `protobuf:"bytes,6,opt,name=issuer" json:"issuer,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *RPCRequest) Reset()         { *m = RPCRequest{} }
func (m *RPCRequest) String() string { return proto.CompactTextString(m) }
func (*RPCRequest) ProtoMessage()    {}

func (m *RPCRequest) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *RPCRequest) GetSize() int32 {
	if m != nil && m.Size != nil {
		return *m.Size
	}
	return 0
}

func (m *RPCRequest) GetPolicy() string {
	if m != nil && m.Policy != nil {
		return *m.Policy
	}
	return ""
}

func (m *RPCRequest) GetTime() int64 {
	if m != nil && m.Time != nil {
		return *m.Time
	}
	return 0
}

func (m *RPCRequest) GetExpiration() int64 {
	if m != nil && m.Expiration != nil {
		return *m.Expiration
	}
	return 0
}

func (m *RPCRequest) GetIssuer() []byte {
	if m != nil {
		return m.Issuer
	}
	return nil
}

type RPCResponse struct {
	Data             []byte  `protobuf:"bytes,1,opt,name=data" json:"data,omitempty"`
	Policy           *string `protobuf:"bytes,2,opt,name=policy" json:"policy,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *RPCResponse) Reset()         { *m = RPCResponse{} }
func (m *RPCResponse) String() string { return proto.CompactTextString(m) }
func (*RPCResponse) ProtoMessage()    {}

func (m *RPCResponse) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *RPCResponse) GetPolicy() string {
	if m != nil && m.Policy != nil {
		return *m.Policy
	}
	return ""
}
