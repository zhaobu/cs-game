// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pb/misc/misc.proto

package misc

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
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type SomeMsg struct {
	Name                 string   `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	Payload              []byte   `protobuf:"bytes,2,opt,name=Payload,proto3" json:"Payload,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SomeMsg) Reset()         { *m = SomeMsg{} }
func (m *SomeMsg) String() string { return proto.CompactTextString(m) }
func (*SomeMsg) ProtoMessage()    {}
func (*SomeMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_51b2c22efe8ca7a7, []int{0}
}

func (m *SomeMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SomeMsg.Unmarshal(m, b)
}
func (m *SomeMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SomeMsg.Marshal(b, m, deterministic)
}
func (m *SomeMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SomeMsg.Merge(m, src)
}
func (m *SomeMsg) XXX_Size() int {
	return xxx_messageInfo_SomeMsg.Size(m)
}
func (m *SomeMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_SomeMsg.DiscardUnknown(m)
}

var xxx_messageInfo_SomeMsg proto.InternalMessageInfo

func (m *SomeMsg) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *SomeMsg) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

type GroupMsg struct {
	Name                 string     `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	Msgs                 []*SomeMsg `protobuf:"bytes,2,rep,name=Msgs,proto3" json:"Msgs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *GroupMsg) Reset()         { *m = GroupMsg{} }
func (m *GroupMsg) String() string { return proto.CompactTextString(m) }
func (*GroupMsg) ProtoMessage()    {}
func (*GroupMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_51b2c22efe8ca7a7, []int{1}
}

func (m *GroupMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GroupMsg.Unmarshal(m, b)
}
func (m *GroupMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GroupMsg.Marshal(b, m, deterministic)
}
func (m *GroupMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GroupMsg.Merge(m, src)
}
func (m *GroupMsg) XXX_Size() int {
	return xxx_messageInfo_GroupMsg.Size(m)
}
func (m *GroupMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_GroupMsg.DiscardUnknown(m)
}

var xxx_messageInfo_GroupMsg proto.InternalMessageInfo

func (m *GroupMsg) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *GroupMsg) GetMsgs() []*SomeMsg {
	if m != nil {
		return m.Msgs
	}
	return nil
}

// 下面的先不用
type FragMsg struct {
	ID                   uint64   `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Name                 string   `protobuf:"bytes,2,opt,name=Name,proto3" json:"Name,omitempty"`
	Payload              []byte   `protobuf:"bytes,3,opt,name=Payload,proto3" json:"Payload,omitempty"`
	CurrIdx              uint32   `protobuf:"varint,4,opt,name=CurrIdx,proto3" json:"CurrIdx,omitempty"`
	Total                uint32   `protobuf:"varint,5,opt,name=Total,proto3" json:"Total,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FragMsg) Reset()         { *m = FragMsg{} }
func (m *FragMsg) String() string { return proto.CompactTextString(m) }
func (*FragMsg) ProtoMessage()    {}
func (*FragMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_51b2c22efe8ca7a7, []int{2}
}

func (m *FragMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FragMsg.Unmarshal(m, b)
}
func (m *FragMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FragMsg.Marshal(b, m, deterministic)
}
func (m *FragMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FragMsg.Merge(m, src)
}
func (m *FragMsg) XXX_Size() int {
	return xxx_messageInfo_FragMsg.Size(m)
}
func (m *FragMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_FragMsg.DiscardUnknown(m)
}

var xxx_messageInfo_FragMsg proto.InternalMessageInfo

func (m *FragMsg) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *FragMsg) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *FragMsg) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (m *FragMsg) GetCurrIdx() uint32 {
	if m != nil {
		return m.CurrIdx
	}
	return 0
}

func (m *FragMsg) GetTotal() uint32 {
	if m != nil {
		return m.Total
	}
	return 0
}

func init() {
	proto.RegisterType((*SomeMsg)(nil), "misc.SomeMsg")
	proto.RegisterType((*GroupMsg)(nil), "misc.GroupMsg")
	proto.RegisterType((*FragMsg)(nil), "misc.FragMsg")
}

func init() { proto.RegisterFile("pb/misc/misc.proto", fileDescriptor_51b2c22efe8ca7a7) }

var fileDescriptor_51b2c22efe8ca7a7 = []byte{
	// 195 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2a, 0x48, 0xd2, 0xcf,
	0xcd, 0x2c, 0x4e, 0x06, 0x13, 0x7a, 0x05, 0x45, 0xf9, 0x25, 0xf9, 0x42, 0x2c, 0x20, 0xb6, 0x92,
	0x39, 0x17, 0x7b, 0x70, 0x7e, 0x6e, 0xaa, 0x6f, 0x71, 0xba, 0x90, 0x10, 0x17, 0x8b, 0x5f, 0x62,
	0x6e, 0xaa, 0x04, 0xa3, 0x02, 0xa3, 0x06, 0x67, 0x10, 0x98, 0x2d, 0x24, 0xc1, 0xc5, 0x1e, 0x90,
	0x58, 0x99, 0x93, 0x9f, 0x98, 0x22, 0xc1, 0xa4, 0xc0, 0xa8, 0xc1, 0x13, 0x04, 0xe3, 0x2a, 0x39,
	0x72, 0x71, 0xb8, 0x17, 0xe5, 0x97, 0x16, 0xe0, 0xd2, 0xa9, 0xc8, 0xc5, 0xe2, 0x5b, 0x9c, 0x5e,
	0x2c, 0xc1, 0xa4, 0xc0, 0xac, 0xc1, 0x6d, 0xc4, 0xab, 0x07, 0xb6, 0x19, 0x6a, 0x55, 0x10, 0x58,
	0x4a, 0xa9, 0x9c, 0x8b, 0xdd, 0xad, 0x28, 0x31, 0x1d, 0x64, 0x02, 0x1f, 0x17, 0x93, 0xa7, 0x0b,
	0x58, 0x3f, 0x4b, 0x10, 0x93, 0xa7, 0x0b, 0xdc, 0x44, 0x26, 0xec, 0x6e, 0x61, 0x46, 0x71, 0x0b,
	0x48, 0xc6, 0xb9, 0xb4, 0xa8, 0xc8, 0x33, 0xa5, 0x42, 0x82, 0x45, 0x81, 0x51, 0x83, 0x37, 0x08,
	0xc6, 0x15, 0x12, 0xe1, 0x62, 0x0d, 0xc9, 0x2f, 0x49, 0xcc, 0x91, 0x60, 0x05, 0x8b, 0x43, 0x38,
	0x49, 0x6c, 0xe0, 0x10, 0x30, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0xef, 0x7f, 0xa6, 0xe5, 0x17,
	0x01, 0x00, 0x00,
}