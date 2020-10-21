// Code generated by protoc-gen-go. DO NOT EDIT.
// source: center/match.proto

package pbcenter

import (
	fmt "fmt"
	common "game/pb/common"
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

type MatchRspCode int32

const (
	MatchRspCode_NotUse              MatchRspCode = 0
	MatchRspCode_Queued              MatchRspCode = 1
	MatchRspCode_StatuErr            MatchRspCode = 3
	MatchRspCode_Busy                MatchRspCode = 4
	MatchRspCode_InvalidGame         MatchRspCode = 5
	MatchRspCode_InvalidRoomID       MatchRspCode = 6
	MatchRspCode_NotEnoughMoney      MatchRspCode = 7
	MatchRspCode_InternalServerError MatchRspCode = 8
)

var MatchRspCode_name = map[int32]string{
	0: "NotUse",
	1: "Queued",
	3: "StatuErr",
	4: "Busy",
	5: "InvalidGame",
	6: "InvalidRoomID",
	7: "NotEnoughMoney",
	8: "InternalServerError",
}

var MatchRspCode_value = map[string]int32{
	"NotUse":              0,
	"Queued":              1,
	"StatuErr":            3,
	"Busy":                4,
	"InvalidGame":         5,
	"InvalidRoomID":       6,
	"NotEnoughMoney":      7,
	"InternalServerError": 8,
}

func (x MatchRspCode) String() string {
	return proto.EnumName(MatchRspCode_name, int32(x))
}

func (MatchRspCode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_9e4539f188f4e851, []int{0}
}

// 匹配游戏 c -> s
type MatchReq struct {
	Head                 *common.ReqHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	GameName             string          `protobuf:"bytes,2,opt,name=GameName,proto3" json:"GameName,omitempty"`
	RoomId               uint32          `protobuf:"varint,3,opt,name=RoomId,proto3" json:"RoomId,omitempty"`
	IsContinue           bool            `protobuf:"varint,4,opt,name=IsContinue,proto3" json:"IsContinue,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *MatchReq) Reset()         { *m = MatchReq{} }
func (m *MatchReq) String() string { return proto.CompactTextString(m) }
func (*MatchReq) ProtoMessage()    {}
func (*MatchReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_9e4539f188f4e851, []int{0}
}

func (m *MatchReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MatchReq.Unmarshal(m, b)
}
func (m *MatchReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MatchReq.Marshal(b, m, deterministic)
}
func (m *MatchReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MatchReq.Merge(m, src)
}
func (m *MatchReq) XXX_Size() int {
	return xxx_messageInfo_MatchReq.Size(m)
}
func (m *MatchReq) XXX_DiscardUnknown() {
	xxx_messageInfo_MatchReq.DiscardUnknown(m)
}

var xxx_messageInfo_MatchReq proto.InternalMessageInfo

func (m *MatchReq) GetHead() *common.ReqHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *MatchReq) GetGameName() string {
	if m != nil {
		return m.GameName
	}
	return ""
}

func (m *MatchReq) GetRoomId() uint32 {
	if m != nil {
		return m.RoomId
	}
	return 0
}

func (m *MatchReq) GetIsContinue() bool {
	if m != nil {
		return m.IsContinue
	}
	return false
}

type MatchRsp struct {
	Head                 *common.RspHead   `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	Code                 MatchRspCode      `protobuf:"varint,2,opt,name=Code,proto3,enum=pbcenter.MatchRspCode" json:"Code,omitempty"`
	StrCode              string            `protobuf:"bytes,3,opt,name=StrCode,proto3" json:"StrCode,omitempty"`
	Sec                  int32             `protobuf:"varint,4,opt,name=Sec,proto3" json:"Sec,omitempty"`
	Status               common.UserStatus `protobuf:"varint,5,opt,name=Status,proto3,enum=pbcommon.UserStatus" json:"Status,omitempty"`
	GameName             string            `protobuf:"bytes,6,opt,name=GameName,proto3" json:"GameName,omitempty"`
	RoomID               uint32            `protobuf:"varint,7,opt,name=RoomID,proto3" json:"RoomID,omitempty"`
	IsContinue           bool              `protobuf:"varint,8,opt,name=IsContinue,proto3" json:"IsContinue,omitempty"`
	GameArgMsgName       string            `protobuf:"bytes,9,opt,name=GameArgMsgName,proto3" json:"GameArgMsgName,omitempty"`
	GameArgMsgValue      []byte            `protobuf:"bytes,10,opt,name=GameArgMsgValue,proto3" json:"GameArgMsgValue,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *MatchRsp) Reset()         { *m = MatchRsp{} }
func (m *MatchRsp) String() string { return proto.CompactTextString(m) }
func (*MatchRsp) ProtoMessage()    {}
func (*MatchRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_9e4539f188f4e851, []int{1}
}

func (m *MatchRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MatchRsp.Unmarshal(m, b)
}
func (m *MatchRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MatchRsp.Marshal(b, m, deterministic)
}
func (m *MatchRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MatchRsp.Merge(m, src)
}
func (m *MatchRsp) XXX_Size() int {
	return xxx_messageInfo_MatchRsp.Size(m)
}
func (m *MatchRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_MatchRsp.DiscardUnknown(m)
}

var xxx_messageInfo_MatchRsp proto.InternalMessageInfo

func (m *MatchRsp) GetHead() *common.RspHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *MatchRsp) GetCode() MatchRspCode {
	if m != nil {
		return m.Code
	}
	return MatchRspCode_NotUse
}

func (m *MatchRsp) GetStrCode() string {
	if m != nil {
		return m.StrCode
	}
	return ""
}

func (m *MatchRsp) GetSec() int32 {
	if m != nil {
		return m.Sec
	}
	return 0
}

func (m *MatchRsp) GetStatus() common.UserStatus {
	if m != nil {
		return m.Status
	}
	return common.UserStatus_NotLogin
}

func (m *MatchRsp) GetGameName() string {
	if m != nil {
		return m.GameName
	}
	return ""
}

func (m *MatchRsp) GetRoomID() uint32 {
	if m != nil {
		return m.RoomID
	}
	return 0
}

func (m *MatchRsp) GetIsContinue() bool {
	if m != nil {
		return m.IsContinue
	}
	return false
}

func (m *MatchRsp) GetGameArgMsgName() string {
	if m != nil {
		return m.GameArgMsgName
	}
	return ""
}

func (m *MatchRsp) GetGameArgMsgValue() []byte {
	if m != nil {
		return m.GameArgMsgValue
	}
	return nil
}

// 匹配超时 s -> c
type MatchTimeOut struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MatchTimeOut) Reset()         { *m = MatchTimeOut{} }
func (m *MatchTimeOut) String() string { return proto.CompactTextString(m) }
func (*MatchTimeOut) ProtoMessage()    {}
func (*MatchTimeOut) Descriptor() ([]byte, []int) {
	return fileDescriptor_9e4539f188f4e851, []int{2}
}

func (m *MatchTimeOut) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MatchTimeOut.Unmarshal(m, b)
}
func (m *MatchTimeOut) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MatchTimeOut.Marshal(b, m, deterministic)
}
func (m *MatchTimeOut) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MatchTimeOut.Merge(m, src)
}
func (m *MatchTimeOut) XXX_Size() int {
	return xxx_messageInfo_MatchTimeOut.Size(m)
}
func (m *MatchTimeOut) XXX_DiscardUnknown() {
	xxx_messageInfo_MatchTimeOut.DiscardUnknown(m)
}

var xxx_messageInfo_MatchTimeOut proto.InternalMessageInfo

// 取消匹配 c -> s
type CancelMatchReq struct {
	Head                 *common.ReqHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *CancelMatchReq) Reset()         { *m = CancelMatchReq{} }
func (m *CancelMatchReq) String() string { return proto.CompactTextString(m) }
func (*CancelMatchReq) ProtoMessage()    {}
func (*CancelMatchReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_9e4539f188f4e851, []int{3}
}

func (m *CancelMatchReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CancelMatchReq.Unmarshal(m, b)
}
func (m *CancelMatchReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CancelMatchReq.Marshal(b, m, deterministic)
}
func (m *CancelMatchReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CancelMatchReq.Merge(m, src)
}
func (m *CancelMatchReq) XXX_Size() int {
	return xxx_messageInfo_CancelMatchReq.Size(m)
}
func (m *CancelMatchReq) XXX_DiscardUnknown() {
	xxx_messageInfo_CancelMatchReq.DiscardUnknown(m)
}

var xxx_messageInfo_CancelMatchReq proto.InternalMessageInfo

func (m *CancelMatchReq) GetHead() *common.ReqHead {
	if m != nil {
		return m.Head
	}
	return nil
}

type CancelMatchRsp struct {
	Head                 *common.RspHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	Succ                 bool            `protobuf:"varint,2,opt,name=Succ,proto3" json:"Succ,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *CancelMatchRsp) Reset()         { *m = CancelMatchRsp{} }
func (m *CancelMatchRsp) String() string { return proto.CompactTextString(m) }
func (*CancelMatchRsp) ProtoMessage()    {}
func (*CancelMatchRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_9e4539f188f4e851, []int{4}
}

func (m *CancelMatchRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CancelMatchRsp.Unmarshal(m, b)
}
func (m *CancelMatchRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CancelMatchRsp.Marshal(b, m, deterministic)
}
func (m *CancelMatchRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CancelMatchRsp.Merge(m, src)
}
func (m *CancelMatchRsp) XXX_Size() int {
	return xxx_messageInfo_CancelMatchRsp.Size(m)
}
func (m *CancelMatchRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_CancelMatchRsp.DiscardUnknown(m)
}

var xxx_messageInfo_CancelMatchRsp proto.InternalMessageInfo

func (m *CancelMatchRsp) GetHead() *common.RspHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *CancelMatchRsp) GetSucc() bool {
	if m != nil {
		return m.Succ
	}
	return false
}

func init() {
	proto.RegisterEnum("pbcenter.MatchRspCode", MatchRspCode_name, MatchRspCode_value)
	proto.RegisterType((*MatchReq)(nil), "pbcenter.MatchReq")
	proto.RegisterType((*MatchRsp)(nil), "pbcenter.MatchRsp")
	proto.RegisterType((*MatchTimeOut)(nil), "pbcenter.MatchTimeOut")
	proto.RegisterType((*CancelMatchReq)(nil), "pbcenter.CancelMatchReq")
	proto.RegisterType((*CancelMatchRsp)(nil), "pbcenter.CancelMatchRsp")
}

func init() { proto.RegisterFile("center/match.proto", fileDescriptor_9e4539f188f4e851) }

var fileDescriptor_9e4539f188f4e851 = []byte{
	// 462 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x53, 0x61, 0x6b, 0xd3, 0x50,
	0x14, 0x35, 0x6b, 0x96, 0x66, 0x77, 0x5d, 0x96, 0x5d, 0x65, 0xc6, 0x0a, 0x12, 0x02, 0x4a, 0x18,
	0xd2, 0xc2, 0xfc, 0xe0, 0x67, 0xed, 0x8a, 0x16, 0x69, 0xc5, 0x17, 0xe7, 0xf7, 0xd7, 0xe4, 0xd2,
	0x16, 0x9a, 0xbc, 0xec, 0xe5, 0xbd, 0xc1, 0x7e, 0x80, 0x9f, 0xfd, 0x85, 0xfe, 0x17, 0xc9, 0x4b,
	0xea, 0xba, 0x80, 0xa2, 0x9f, 0x7a, 0xef, 0x39, 0xa7, 0x87, 0x93, 0x73, 0x13, 0xc0, 0x94, 0x0a,
	0x45, 0x72, 0x9c, 0x73, 0x95, 0xae, 0x47, 0xa5, 0x14, 0x4a, 0xa0, 0x5b, 0x2e, 0x1b, 0x74, 0xf8,
	0x6c, 0xc5, 0x73, 0x1a, 0x97, 0xcb, 0x71, 0x2a, 0xf2, 0x5c, 0x14, 0xe3, 0x35, 0xf1, 0xac, 0x11,
	0x0d, 0x9f, 0x77, 0xa8, 0xe6, 0xa7, 0x21, 0xa3, 0xef, 0x16, 0xb8, 0xf3, 0xda, 0x91, 0xd1, 0x0d,
	0xbe, 0x04, 0xfb, 0x23, 0xf1, 0x2c, 0xb0, 0x42, 0x2b, 0x3e, 0xbe, 0x3c, 0x1b, 0x95, 0xcb, 0x56,
	0xcb, 0xe8, 0xa6, 0x26, 0x98, 0xa1, 0x71, 0x08, 0xee, 0x07, 0x9e, 0xd3, 0x82, 0xe7, 0x14, 0x1c,
	0x84, 0x56, 0x7c, 0xc4, 0x7e, 0xef, 0x78, 0x0e, 0x0e, 0x13, 0x22, 0x9f, 0x65, 0x41, 0x2f, 0xb4,
	0xe2, 0x13, 0xd6, 0x6e, 0xf8, 0x02, 0x60, 0x56, 0x4d, 0x44, 0xa1, 0x36, 0x85, 0xa6, 0xc0, 0x0e,
	0xad, 0xd8, 0x65, 0x7b, 0x48, 0xf4, 0xf3, 0x60, 0x97, 0xa3, 0x2a, 0xff, 0x92, 0xa3, 0x2a, 0xf7,
	0x72, 0x5c, 0x80, 0x3d, 0x11, 0x59, 0x93, 0xc1, 0xbb, 0x3c, 0x1f, 0xed, 0xca, 0x18, 0xed, 0x8c,
	0x6a, 0x96, 0x19, 0x0d, 0x06, 0xd0, 0x4f, 0x94, 0x34, 0xf2, 0x9e, 0x89, 0xbc, 0x5b, 0xd1, 0x87,
	0x5e, 0x42, 0xa9, 0x89, 0x74, 0xc8, 0xea, 0x11, 0x5f, 0x83, 0x93, 0x28, 0xae, 0x74, 0x15, 0x1c,
	0x1a, 0xe7, 0x27, 0xf7, 0x01, 0xae, 0x2b, 0x92, 0x0d, 0xc7, 0x5a, 0xcd, 0x83, 0x36, 0x9c, 0x3f,
	0xb4, 0x71, 0x15, 0xf4, 0xf7, 0xda, 0xb8, 0xea, 0xb4, 0xe1, 0x76, 0xdb, 0xc0, 0x57, 0xe0, 0xd5,
	0x1e, 0xef, 0xe4, 0x6a, 0x5e, 0xad, 0x8c, 0xf3, 0x91, 0x71, 0xee, 0xa0, 0x18, 0xc3, 0xe9, 0x3d,
	0xf2, 0x8d, 0x6f, 0x35, 0x05, 0x10, 0x5a, 0xf1, 0x80, 0x75, 0xe1, 0xc8, 0x83, 0x81, 0x69, 0xe5,
	0xeb, 0x26, 0xa7, 0xcf, 0x5a, 0x45, 0x6f, 0xc1, 0x9b, 0xf0, 0x22, 0xa5, 0xed, 0x7f, 0x1e, 0x3f,
	0xfa, 0xf4, 0xf0, 0x8f, 0xff, 0x7e, 0x2d, 0x04, 0x3b, 0xd1, 0x69, 0x6a, 0xae, 0xe5, 0x32, 0x33,
	0x5f, 0xfc, 0xb0, 0xda, 0x58, 0xed, 0xb1, 0x10, 0xc0, 0x59, 0x08, 0x75, 0x5d, 0x91, 0xff, 0xa8,
	0x9e, 0xbf, 0x68, 0xd2, 0x94, 0xf9, 0x16, 0x0e, 0xc0, 0x35, 0x75, 0x4f, 0xa5, 0xf4, 0x7b, 0xe8,
	0x82, 0xfd, 0x5e, 0x57, 0x77, 0xbe, 0x8d, 0xa7, 0x70, 0x3c, 0x2b, 0x6e, 0xf9, 0x76, 0x93, 0xd5,
	0x0f, 0xec, 0x1f, 0xe2, 0x19, 0x9c, 0xb4, 0x40, 0x53, 0xb5, 0xef, 0x20, 0x82, 0xb7, 0x10, 0x6a,
	0x5a, 0x08, 0xbd, 0x5a, 0xcf, 0x45, 0x41, 0x77, 0x7e, 0x1f, 0x9f, 0xc2, 0xe3, 0x59, 0xfd, 0xaa,
	0x14, 0x7c, 0x9b, 0x90, 0xbc, 0x25, 0x39, 0x95, 0x52, 0x48, 0xdf, 0x5d, 0x3a, 0xe6, 0xb3, 0x78,
	0xf3, 0x2b, 0x00, 0x00, 0xff, 0xff, 0x1c, 0x84, 0x28, 0xc9, 0x6e, 0x03, 0x00, 0x00,
}