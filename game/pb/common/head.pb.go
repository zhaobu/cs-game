// ErrorCode generated by protoc-gen-go. DO NOT EDIT.
// source: common/head.proto

package pbcommon

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// 请求头
type ReqHead struct {
	Seq                  uint64   `protobuf:"varint,1,opt,name=Seq,proto3" json:"Seq,omitempty"`
	UserID               uint64   `protobuf:"varint,2,opt,name=UserID,proto3" json:"UserID,omitempty"`
	SessionID            string   `protobuf:"bytes,3,opt,name=SessionID,proto3" json:"SessionID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReqHead) Reset()         { *m = ReqHead{} }
func (m *ReqHead) String() string { return proto.CompactTextString(m) }
func (*ReqHead) ProtoMessage()    {}
func (*ReqHead) Descriptor() ([]byte, []int) {
	return fileDescriptor_475b2844d3e2c1ff, []int{0}
}

func (m *ReqHead) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReqHead.Unmarshal(m, b)
}
func (m *ReqHead) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReqHead.Marshal(b, m, deterministic)
}
func (m *ReqHead) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReqHead.Merge(m, src)
}
func (m *ReqHead) XXX_Size() int {
	return xxx_messageInfo_ReqHead.Size(m)
}
func (m *ReqHead) XXX_DiscardUnknown() {
	xxx_messageInfo_ReqHead.DiscardUnknown(m)
}

var xxx_messageInfo_ReqHead proto.InternalMessageInfo

func (m *ReqHead) GetSeq() uint64 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func (m *ReqHead) GetUserID() uint64 {
	if m != nil {
		return m.UserID
	}
	return 0
}

func (m *ReqHead) GetSessionID() string {
	if m != nil {
		return m.SessionID
	}
	return ""
}

// 响应头
type RspHead struct {
	Seq                  uint64   `protobuf:"varint,1,opt,name=Seq,proto3" json:"Seq,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RspHead) Reset()         { *m = RspHead{} }
func (m *RspHead) String() string { return proto.CompactTextString(m) }
func (*RspHead) ProtoMessage()    {}
func (*RspHead) Descriptor() ([]byte, []int) {
	return fileDescriptor_475b2844d3e2c1ff, []int{1}
}

func (m *RspHead) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RspHead.Unmarshal(m, b)
}
func (m *RspHead) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RspHead.Marshal(b, m, deterministic)
}
func (m *RspHead) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RspHead.Merge(m, src)
}
func (m *RspHead) XXX_Size() int {
	return xxx_messageInfo_RspHead.Size(m)
}
func (m *RspHead) XXX_DiscardUnknown() {
	xxx_messageInfo_RspHead.DiscardUnknown(m)
}

var xxx_messageInfo_RspHead proto.InternalMessageInfo

func (m *RspHead) GetSeq() uint64 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func init() {
	proto.RegisterType((*ReqHead)(nil), "pbcommon.ReqHead")
	proto.RegisterType((*RspHead)(nil), "pbcommon.RspHead")
}

func init() { proto.RegisterFile("common/head.proto", fileDescriptor_475b2844d3e2c1ff) }

var fileDescriptor_475b2844d3e2c1ff = []byte{
	// 128 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4c, 0xce, 0xcf, 0xcd,
	0xcd, 0xcf, 0xd3, 0xcf, 0x48, 0x4d, 0x4c, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x28,
	0x48, 0x82, 0x08, 0x2a, 0x05, 0x72, 0xb1, 0x07, 0xa5, 0x16, 0x7a, 0xa4, 0x26, 0xa6, 0x08, 0x09,
	0x70, 0x31, 0x07, 0xa7, 0x16, 0x4a, 0x30, 0x2a, 0x30, 0x6a, 0xb0, 0x04, 0x81, 0x98, 0x42, 0x62,
	0x5c, 0x6c, 0xa1, 0xc5, 0xa9, 0x45, 0x9e, 0x2e, 0x12, 0x4c, 0x60, 0x41, 0x28, 0x4f, 0x48, 0x86,
	0x8b, 0x33, 0x38, 0xb5, 0xb8, 0x38, 0x33, 0x3f, 0xcf, 0xd3, 0x45, 0x82, 0x59, 0x81, 0x51, 0x83,
	0x33, 0x08, 0x21, 0xa0, 0x24, 0xcd, 0xc5, 0x1e, 0x54, 0x5c, 0x80, 0xdd, 0xc8, 0x24, 0x36, 0xb0,
	0x03, 0x8c, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0x35, 0x23, 0xc9, 0xa6, 0x95, 0x00, 0x00, 0x00,
}
