// Code generated by protoc-gen-go. DO NOT EDIT.
// source: inner/inner.proto

package pbinner

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

// gates -> *
type UserChangeType int32

const (
	UserChangeType_NotUse  UserChangeType = 0
	UserChangeType_Online  UserChangeType = 1
	UserChangeType_Offline UserChangeType = 2
)

var UserChangeType_name = map[int32]string{
	0: "NotUse",
	1: "Online",
	2: "Offline",
}

var UserChangeType_value = map[string]int32{
	"NotUse":  0,
	"Online":  1,
	"Offline": 2,
}

func (x UserChangeType) String() string {
	return proto.EnumName(UserChangeType_name, int32(x))
}

func (UserChangeType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_b3615dfd873b2a19, []int{0}
}

// center -> games
// 匹配成功
type GameMatchSucc struct {
	RoomId               uint32   `protobuf:"varint,1,opt,name=RoomId,proto3" json:"RoomId,omitempty"`
	UserIDs              []uint64 `protobuf:"varint,2,rep,packed,name=UserIDs,proto3" json:"UserIDs,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GameMatchSucc) Reset()         { *m = GameMatchSucc{} }
func (m *GameMatchSucc) String() string { return proto.CompactTextString(m) }
func (*GameMatchSucc) ProtoMessage()    {}
func (*GameMatchSucc) Descriptor() ([]byte, []int) {
	return fileDescriptor_b3615dfd873b2a19, []int{0}
}

func (m *GameMatchSucc) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GameMatchSucc.Unmarshal(m, b)
}
func (m *GameMatchSucc) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GameMatchSucc.Marshal(b, m, deterministic)
}
func (m *GameMatchSucc) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GameMatchSucc.Merge(m, src)
}
func (m *GameMatchSucc) XXX_Size() int {
	return xxx_messageInfo_GameMatchSucc.Size(m)
}
func (m *GameMatchSucc) XXX_DiscardUnknown() {
	xxx_messageInfo_GameMatchSucc.DiscardUnknown(m)
}

var xxx_messageInfo_GameMatchSucc proto.InternalMessageInfo

func (m *GameMatchSucc) GetRoomId() uint32 {
	if m != nil {
		return m.RoomId
	}
	return 0
}

func (m *GameMatchSucc) GetUserIDs() []uint64 {
	if m != nil {
		return m.UserIDs
	}
	return nil
}

// center -> games
type GameMatchArgsCheckReq struct {
	RoomId               uint32   `protobuf:"varint,1,opt,name=RoomId,proto3" json:"RoomId,omitempty"`
	UserID               uint64   `protobuf:"varint,2,opt,name=UserID,proto3" json:"UserID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GameMatchArgsCheckReq) Reset()         { *m = GameMatchArgsCheckReq{} }
func (m *GameMatchArgsCheckReq) String() string { return proto.CompactTextString(m) }
func (*GameMatchArgsCheckReq) ProtoMessage()    {}
func (*GameMatchArgsCheckReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_b3615dfd873b2a19, []int{1}
}

func (m *GameMatchArgsCheckReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GameMatchArgsCheckReq.Unmarshal(m, b)
}
func (m *GameMatchArgsCheckReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GameMatchArgsCheckReq.Marshal(b, m, deterministic)
}
func (m *GameMatchArgsCheckReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GameMatchArgsCheckReq.Merge(m, src)
}
func (m *GameMatchArgsCheckReq) XXX_Size() int {
	return xxx_messageInfo_GameMatchArgsCheckReq.Size(m)
}
func (m *GameMatchArgsCheckReq) XXX_DiscardUnknown() {
	xxx_messageInfo_GameMatchArgsCheckReq.DiscardUnknown(m)
}

var xxx_messageInfo_GameMatchArgsCheckReq proto.InternalMessageInfo

func (m *GameMatchArgsCheckReq) GetRoomId() uint32 {
	if m != nil {
		return m.RoomId
	}
	return 0
}

func (m *GameMatchArgsCheckReq) GetUserID() uint64 {
	if m != nil {
		return m.UserID
	}
	return 0
}

type GameMatchArgsCheckRsp struct {
	// 1成功 2无效房间ID 3查询用户失败 4金币限制
	Code                 uint32   `protobuf:"varint,1,opt,name=Code,proto3" json:"Code,omitempty"`
	GameArgMsgName       string   `protobuf:"bytes,2,opt,name=GameArgMsgName,proto3" json:"GameArgMsgName,omitempty"`
	GameArgMsgValue      []byte   `protobuf:"bytes,3,opt,name=GameArgMsgValue,proto3" json:"GameArgMsgValue,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GameMatchArgsCheckRsp) Reset()         { *m = GameMatchArgsCheckRsp{} }
func (m *GameMatchArgsCheckRsp) String() string { return proto.CompactTextString(m) }
func (*GameMatchArgsCheckRsp) ProtoMessage()    {}
func (*GameMatchArgsCheckRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_b3615dfd873b2a19, []int{2}
}

func (m *GameMatchArgsCheckRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GameMatchArgsCheckRsp.Unmarshal(m, b)
}
func (m *GameMatchArgsCheckRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GameMatchArgsCheckRsp.Marshal(b, m, deterministic)
}
func (m *GameMatchArgsCheckRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GameMatchArgsCheckRsp.Merge(m, src)
}
func (m *GameMatchArgsCheckRsp) XXX_Size() int {
	return xxx_messageInfo_GameMatchArgsCheckRsp.Size(m)
}
func (m *GameMatchArgsCheckRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_GameMatchArgsCheckRsp.DiscardUnknown(m)
}

var xxx_messageInfo_GameMatchArgsCheckRsp proto.InternalMessageInfo

func (m *GameMatchArgsCheckRsp) GetCode() uint32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *GameMatchArgsCheckRsp) GetGameArgMsgName() string {
	if m != nil {
		return m.GameArgMsgName
	}
	return ""
}

func (m *GameMatchArgsCheckRsp) GetGameArgMsgValue() []byte {
	if m != nil {
		return m.GameArgMsgValue
	}
	return nil
}

type UserChangeNotif struct {
	UserID               uint64         `protobuf:"varint,1,opt,name=UserID,proto3" json:"UserID,omitempty"`
	Typ                  UserChangeType `protobuf:"varint,2,opt,name=Typ,proto3,enum=pbinner.UserChangeType" json:"Typ,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *UserChangeNotif) Reset()         { *m = UserChangeNotif{} }
func (m *UserChangeNotif) String() string { return proto.CompactTextString(m) }
func (*UserChangeNotif) ProtoMessage()    {}
func (*UserChangeNotif) Descriptor() ([]byte, []int) {
	return fileDescriptor_b3615dfd873b2a19, []int{3}
}

func (m *UserChangeNotif) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UserChangeNotif.Unmarshal(m, b)
}
func (m *UserChangeNotif) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UserChangeNotif.Marshal(b, m, deterministic)
}
func (m *UserChangeNotif) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UserChangeNotif.Merge(m, src)
}
func (m *UserChangeNotif) XXX_Size() int {
	return xxx_messageInfo_UserChangeNotif.Size(m)
}
func (m *UserChangeNotif) XXX_DiscardUnknown() {
	xxx_messageInfo_UserChangeNotif.DiscardUnknown(m)
}

var xxx_messageInfo_UserChangeNotif proto.InternalMessageInfo

func (m *UserChangeNotif) GetUserID() uint64 {
	if m != nil {
		return m.UserID
	}
	return 0
}

func (m *UserChangeNotif) GetTyp() UserChangeType {
	if m != nil {
		return m.Typ
	}
	return UserChangeType_NotUse
}

type DeskChangeNotif struct {
	ClubID int64  `protobuf:"varint,1,opt,name=ClubID,proto3" json:"ClubID,omitempty"`
	DeskID uint64 `protobuf:"varint,2,opt,name=DeskID,proto3" json:"DeskID,omitempty"`
	// 1create 2update 3delete
	ChangeTyp            int32    `protobuf:"varint,3,opt,name=ChangeTyp,proto3" json:"ChangeTyp,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeskChangeNotif) Reset()         { *m = DeskChangeNotif{} }
func (m *DeskChangeNotif) String() string { return proto.CompactTextString(m) }
func (*DeskChangeNotif) ProtoMessage()    {}
func (*DeskChangeNotif) Descriptor() ([]byte, []int) {
	return fileDescriptor_b3615dfd873b2a19, []int{4}
}

func (m *DeskChangeNotif) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeskChangeNotif.Unmarshal(m, b)
}
func (m *DeskChangeNotif) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeskChangeNotif.Marshal(b, m, deterministic)
}
func (m *DeskChangeNotif) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeskChangeNotif.Merge(m, src)
}
func (m *DeskChangeNotif) XXX_Size() int {
	return xxx_messageInfo_DeskChangeNotif.Size(m)
}
func (m *DeskChangeNotif) XXX_DiscardUnknown() {
	xxx_messageInfo_DeskChangeNotif.DiscardUnknown(m)
}

var xxx_messageInfo_DeskChangeNotif proto.InternalMessageInfo

func (m *DeskChangeNotif) GetClubID() int64 {
	if m != nil {
		return m.ClubID
	}
	return 0
}

func (m *DeskChangeNotif) GetDeskID() uint64 {
	if m != nil {
		return m.DeskID
	}
	return 0
}

func (m *DeskChangeNotif) GetChangeTyp() int32 {
	if m != nil {
		return m.ChangeTyp
	}
	return 0
}

func init() {
	proto.RegisterEnum("pbinner.UserChangeType", UserChangeType_name, UserChangeType_value)
	proto.RegisterType((*GameMatchSucc)(nil), "pbinner.GameMatchSucc")
	proto.RegisterType((*GameMatchArgsCheckReq)(nil), "pbinner.GameMatchArgsCheckReq")
	proto.RegisterType((*GameMatchArgsCheckRsp)(nil), "pbinner.GameMatchArgsCheckRsp")
	proto.RegisterType((*UserChangeNotif)(nil), "pbinner.UserChangeNotif")
	proto.RegisterType((*DeskChangeNotif)(nil), "pbinner.DeskChangeNotif")
}

func init() { proto.RegisterFile("inner/inner.proto", fileDescriptor_b3615dfd873b2a19) }

var fileDescriptor_b3615dfd873b2a19 = []byte{
	// 318 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x92, 0x4f, 0x4f, 0xf2, 0x40,
	0x10, 0xc6, 0xdf, 0xa5, 0xbc, 0x25, 0x8c, 0xf2, 0xc7, 0x4d, 0xc4, 0x1e, 0x3c, 0x34, 0x3d, 0x98,
	0xea, 0x01, 0x13, 0x8d, 0x1f, 0x80, 0x94, 0x84, 0x70, 0x00, 0x92, 0x15, 0xbc, 0x9a, 0x52, 0x86,
	0xd2, 0xd0, 0x76, 0x6b, 0xb7, 0x3d, 0x70, 0xf0, 0xbb, 0x9b, 0x5d, 0x56, 0x8a, 0x8d, 0x5e, 0x36,
	0xf3, 0x3c, 0x99, 0xf9, 0xcd, 0x93, 0xc9, 0xc2, 0x55, 0x94, 0xa6, 0x98, 0x3f, 0xaa, 0x77, 0x98,
	0xe5, 0xbc, 0xe0, 0xb4, 0x95, 0xad, 0x95, 0x74, 0x46, 0xd0, 0x99, 0xf8, 0x09, 0xce, 0xfc, 0x22,
	0xd8, 0xbd, 0x96, 0x41, 0x40, 0x07, 0x60, 0x32, 0xce, 0x93, 0xe9, 0xc6, 0x22, 0x36, 0x71, 0x3b,
	0x4c, 0x2b, 0x6a, 0x41, 0x6b, 0x25, 0x30, 0x9f, 0x8e, 0x85, 0xd5, 0xb0, 0x0d, 0xb7, 0xc9, 0xbe,
	0xa5, 0x33, 0x81, 0xeb, 0x13, 0x62, 0x94, 0x87, 0xc2, 0xdb, 0x61, 0xb0, 0x67, 0xf8, 0xf1, 0x27,
	0x6a, 0x00, 0xe6, 0x71, 0xd6, 0x6a, 0xd8, 0xc4, 0x6d, 0x32, 0xad, 0x9c, 0xcf, 0x5f, 0x41, 0x22,
	0xa3, 0x14, 0x9a, 0x1e, 0xdf, 0xa0, 0xc6, 0xa8, 0x9a, 0xde, 0x41, 0x57, 0x36, 0x8f, 0xf2, 0x70,
	0x26, 0xc2, 0xb9, 0x9f, 0xa0, 0x82, 0xb5, 0x59, 0xcd, 0xa5, 0x2e, 0xf4, 0x2a, 0xe7, 0xcd, 0x8f,
	0x4b, 0xb4, 0x0c, 0x9b, 0xb8, 0x97, 0xac, 0x6e, 0x3b, 0x4b, 0xe8, 0xc9, 0x20, 0xde, 0xce, 0x4f,
	0x43, 0x9c, 0xf3, 0x22, 0xda, 0x9e, 0x25, 0x25, 0xe7, 0x49, 0xe9, 0x3d, 0x18, 0xcb, 0x43, 0xa6,
	0x36, 0x76, 0x9f, 0x6e, 0x86, 0xfa, 0x98, 0xc3, 0x6a, 0x7c, 0x79, 0xc8, 0x90, 0xc9, 0x1e, 0xe7,
	0x1d, 0x7a, 0x63, 0x14, 0xfb, 0x1a, 0xd5, 0x8b, 0xcb, 0xb5, 0xa6, 0x1a, 0x4c, 0x2b, 0xe9, 0xcb,
	0xd6, 0xea, 0x2e, 0x47, 0x45, 0x6f, 0xa1, 0x7d, 0xa2, 0xaa, 0xf0, 0xff, 0x59, 0x65, 0x3c, 0xbc,
	0x40, 0xf7, 0xe7, 0x5e, 0x0a, 0x60, 0xce, 0x79, 0xb1, 0x12, 0xd8, 0xff, 0x27, 0xeb, 0x45, 0x1a,
	0x47, 0x29, 0xf6, 0x09, 0xbd, 0x80, 0xd6, 0x62, 0xbb, 0x55, 0xa2, 0xb1, 0x36, 0xd5, 0x47, 0x78,
	0xfe, 0x0a, 0x00, 0x00, 0xff, 0xff, 0x69, 0xb5, 0xfc, 0xec, 0x1d, 0x02, 0x00, 0x00,
}
