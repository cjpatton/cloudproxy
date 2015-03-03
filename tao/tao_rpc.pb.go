// Code generated by protoc-gen-go.
// source: tao_rpc.proto
// DO NOT EDIT!

package tao

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type TaoRPCRequest struct {
	Data             []byte  `protobuf:"bytes,1,opt,name=data" json:"data,omitempty"`
	Size             *int32  `protobuf:"varint,2,opt,name=size" json:"size,omitempty"`
	Policy           *string `protobuf:"bytes,3,opt,name=policy" json:"policy,omitempty"`
	Time             *int64  `protobuf:"varint,4,opt,name=time" json:"time,omitempty"`
	Expiration       *int64  `protobuf:"varint,5,opt,name=expiration" json:"expiration,omitempty"`
	Issuer           []byte  `protobuf:"bytes,6,opt,name=issuer" json:"issuer,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *TaoRPCRequest) Reset()         { *m = TaoRPCRequest{} }
func (m *TaoRPCRequest) String() string { return proto.CompactTextString(m) }
func (*TaoRPCRequest) ProtoMessage()    {}

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

func (m *TaoRPCRequest) GetTime() int64 {
	if m != nil && m.Time != nil {
		return *m.Time
	}
	return 0
}

func (m *TaoRPCRequest) GetExpiration() int64 {
	if m != nil && m.Expiration != nil {
		return *m.Expiration
	}
	return 0
}

func (m *TaoRPCRequest) GetIssuer() []byte {
	if m != nil {
		return m.Issuer
	}
	return nil
}

type TaoRPCResponse struct {
	Data             []byte  `protobuf:"bytes,1,opt,name=data" json:"data,omitempty"`
	Policy           *string `protobuf:"bytes,2,opt,name=policy" json:"policy,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *TaoRPCResponse) Reset()         { *m = TaoRPCResponse{} }
func (m *TaoRPCResponse) String() string { return proto.CompactTextString(m) }
func (*TaoRPCResponse) ProtoMessage()    {}

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
}
