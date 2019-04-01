// Code generated by protoc-gen-go. DO NOT EDIT.
// source: game/game.proto

package pbgame

import (
	common "cy/game/pb/common"
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

type MakeDeskRspCode int32

const (
	MakeDeskRspCode_MakeDeskNotUse              MakeDeskRspCode = 0
	MakeDeskRspCode_MakeDeskSucc                MakeDeskRspCode = 1
	MakeDeskRspCode_MakeDeskArgsErr             MakeDeskRspCode = 2
	MakeDeskRspCode_MakeDeskNotEnoughMoney      MakeDeskRspCode = 3
	MakeDeskRspCode_MakeDeskNotEnoughDesk       MakeDeskRspCode = 4
	MakeDeskRspCode_MakeDeskInternalServerError MakeDeskRspCode = 5
	MakeDeskRspCode_MakeDeskUserStatusErr       MakeDeskRspCode = 6
	MakeDeskRspCode_MakeDeskCanNotFindClubID    MakeDeskRspCode = 7
	MakeDeskRspCode_MakeDeskLimit               MakeDeskRspCode = 8
)

var MakeDeskRspCode_name = map[int32]string{
	0: "MakeDeskNotUse",
	1: "MakeDeskSucc",
	2: "MakeDeskArgsErr",
	3: "MakeDeskNotEnoughMoney",
	4: "MakeDeskNotEnoughDesk",
	5: "MakeDeskInternalServerError",
	6: "MakeDeskUserStatusErr",
	7: "MakeDeskCanNotFindClubID",
	8: "MakeDeskLimit",
}

var MakeDeskRspCode_value = map[string]int32{
	"MakeDeskNotUse":              0,
	"MakeDeskSucc":                1,
	"MakeDeskArgsErr":             2,
	"MakeDeskNotEnoughMoney":      3,
	"MakeDeskNotEnoughDesk":       4,
	"MakeDeskInternalServerError": 5,
	"MakeDeskUserStatusErr":       6,
	"MakeDeskCanNotFindClubID":    7,
	"MakeDeskLimit":               8,
}

func (x MakeDeskRspCode) String() string {
	return proto.EnumName(MakeDeskRspCode_name, int32(x))
}

func (MakeDeskRspCode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{0}
}

type JoinDeskRspCode int32

const (
	JoinDeskRspCode_JoinDeskNotUse              JoinDeskRspCode = 0
	JoinDeskRspCode_JoinDeskSucc                JoinDeskRspCode = 1
	JoinDeskRspCode_JoinDeskNotExist            JoinDeskRspCode = 2
	JoinDeskRspCode_JoinDeskDeskFull            JoinDeskRspCode = 3
	JoinDeskRspCode_JoinDeskInternalServerError JoinDeskRspCode = 4
	JoinDeskRspCode_JoinDeskUserStatusErr       JoinDeskRspCode = 5
	JoinDeskRspCode_JoinDeskGameStatusErr       JoinDeskRspCode = 6
	JoinDeskRspCode_JoinDeskDistanceSoClose     JoinDeskRspCode = 7
	JoinDeskRspCode_JoinDeskAlreadyInDesk       JoinDeskRspCode = 8
)

var JoinDeskRspCode_name = map[int32]string{
	0: "JoinDeskNotUse",
	1: "JoinDeskSucc",
	2: "JoinDeskNotExist",
	3: "JoinDeskDeskFull",
	4: "JoinDeskInternalServerError",
	5: "JoinDeskUserStatusErr",
	6: "JoinDeskGameStatusErr",
	7: "JoinDeskDistanceSoClose",
	8: "JoinDeskAlreadyInDesk",
}

var JoinDeskRspCode_value = map[string]int32{
	"JoinDeskNotUse":              0,
	"JoinDeskSucc":                1,
	"JoinDeskNotExist":            2,
	"JoinDeskDeskFull":            3,
	"JoinDeskInternalServerError": 4,
	"JoinDeskUserStatusErr":       5,
	"JoinDeskGameStatusErr":       6,
	"JoinDeskDistanceSoClose":     7,
	"JoinDeskAlreadyInDesk":       8,
}

func (x JoinDeskRspCode) String() string {
	return proto.EnumName(JoinDeskRspCode_name, int32(x))
}

func (JoinDeskRspCode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{1}
}

// 查询游戏配置 c -> s
type QueryGameConfigReq struct {
	Head     *common.ReqHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	GameName string          `protobuf:"bytes,2,opt,name=GameName,proto3" json:"GameName,omitempty"`
	// 1匹配 2好友 3比赛
	Type                 int32    `protobuf:"varint,3,opt,name=Type,proto3" json:"Type,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *QueryGameConfigReq) Reset()         { *m = QueryGameConfigReq{} }
func (m *QueryGameConfigReq) String() string { return proto.CompactTextString(m) }
func (*QueryGameConfigReq) ProtoMessage()    {}
func (*QueryGameConfigReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{0}
}

func (m *QueryGameConfigReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryGameConfigReq.Unmarshal(m, b)
}
func (m *QueryGameConfigReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryGameConfigReq.Marshal(b, m, deterministic)
}
func (m *QueryGameConfigReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryGameConfigReq.Merge(m, src)
}
func (m *QueryGameConfigReq) XXX_Size() int {
	return xxx_messageInfo_QueryGameConfigReq.Size(m)
}
func (m *QueryGameConfigReq) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryGameConfigReq.DiscardUnknown(m)
}

var xxx_messageInfo_QueryGameConfigReq proto.InternalMessageInfo

func (m *QueryGameConfigReq) GetHead() *common.ReqHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *QueryGameConfigReq) GetGameName() string {
	if m != nil {
		return m.GameName
	}
	return ""
}

func (m *QueryGameConfigReq) GetType() int32 {
	if m != nil {
		return m.Type
	}
	return 0
}

type QueryGameConfigRsp struct {
	Head                 *common.RspHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	CfgName              string          `protobuf:"bytes,2,opt,name=CfgName,proto3" json:"CfgName,omitempty"`
	CfgValue             []byte          `protobuf:"bytes,3,opt,name=CfgValue,proto3" json:"CfgValue,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *QueryGameConfigRsp) Reset()         { *m = QueryGameConfigRsp{} }
func (m *QueryGameConfigRsp) String() string { return proto.CompactTextString(m) }
func (*QueryGameConfigRsp) ProtoMessage()    {}
func (*QueryGameConfigRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{1}
}

func (m *QueryGameConfigRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryGameConfigRsp.Unmarshal(m, b)
}
func (m *QueryGameConfigRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryGameConfigRsp.Marshal(b, m, deterministic)
}
func (m *QueryGameConfigRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryGameConfigRsp.Merge(m, src)
}
func (m *QueryGameConfigRsp) XXX_Size() int {
	return xxx_messageInfo_QueryGameConfigRsp.Size(m)
}
func (m *QueryGameConfigRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryGameConfigRsp.DiscardUnknown(m)
}

var xxx_messageInfo_QueryGameConfigRsp proto.InternalMessageInfo

func (m *QueryGameConfigRsp) GetHead() *common.RspHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *QueryGameConfigRsp) GetCfgName() string {
	if m != nil {
		return m.CfgName
	}
	return ""
}

func (m *QueryGameConfigRsp) GetCfgValue() []byte {
	if m != nil {
		return m.CfgValue
	}
	return nil
}

// 查询桌子信息 c -> s
type QueryDeskInfoReq struct {
	Head                 *common.ReqHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	DeskID               uint64          `protobuf:"varint,2,opt,name=DeskID,proto3" json:"DeskID,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *QueryDeskInfoReq) Reset()         { *m = QueryDeskInfoReq{} }
func (m *QueryDeskInfoReq) String() string { return proto.CompactTextString(m) }
func (*QueryDeskInfoReq) ProtoMessage()    {}
func (*QueryDeskInfoReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{2}
}

func (m *QueryDeskInfoReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryDeskInfoReq.Unmarshal(m, b)
}
func (m *QueryDeskInfoReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryDeskInfoReq.Marshal(b, m, deterministic)
}
func (m *QueryDeskInfoReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryDeskInfoReq.Merge(m, src)
}
func (m *QueryDeskInfoReq) XXX_Size() int {
	return xxx_messageInfo_QueryDeskInfoReq.Size(m)
}
func (m *QueryDeskInfoReq) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryDeskInfoReq.DiscardUnknown(m)
}

var xxx_messageInfo_QueryDeskInfoReq proto.InternalMessageInfo

func (m *QueryDeskInfoReq) GetHead() *common.ReqHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *QueryDeskInfoReq) GetDeskID() uint64 {
	if m != nil {
		return m.DeskID
	}
	return 0
}

type QueryDeskInfoRsp struct {
	Head *common.RspHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	// 1 成功 2 不存在
	Code                 uint32           `protobuf:"varint,2,opt,name=Code,proto3" json:"Code,omitempty"`
	Info                 *common.DeskInfo `protobuf:"bytes,3,opt,name=Info,proto3" json:"Info,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *QueryDeskInfoRsp) Reset()         { *m = QueryDeskInfoRsp{} }
func (m *QueryDeskInfoRsp) String() string { return proto.CompactTextString(m) }
func (*QueryDeskInfoRsp) ProtoMessage()    {}
func (*QueryDeskInfoRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{3}
}

func (m *QueryDeskInfoRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryDeskInfoRsp.Unmarshal(m, b)
}
func (m *QueryDeskInfoRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryDeskInfoRsp.Marshal(b, m, deterministic)
}
func (m *QueryDeskInfoRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryDeskInfoRsp.Merge(m, src)
}
func (m *QueryDeskInfoRsp) XXX_Size() int {
	return xxx_messageInfo_QueryDeskInfoRsp.Size(m)
}
func (m *QueryDeskInfoRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryDeskInfoRsp.DiscardUnknown(m)
}

var xxx_messageInfo_QueryDeskInfoRsp proto.InternalMessageInfo

func (m *QueryDeskInfoRsp) GetHead() *common.RspHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *QueryDeskInfoRsp) GetCode() uint32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *QueryDeskInfoRsp) GetInfo() *common.DeskInfo {
	if m != nil {
		return m.Info
	}
	return nil
}

// 新建桌子 c -> s
type MakeDeskReq struct {
	Head                 *common.ReqHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	GameName             string          `protobuf:"bytes,2,opt,name=GameName,proto3" json:"GameName,omitempty"`
	GameArgMsgName       string          `protobuf:"bytes,3,opt,name=GameArgMsgName,proto3" json:"GameArgMsgName,omitempty"`
	GameArgMsgValue      []byte          `protobuf:"bytes,4,opt,name=GameArgMsgValue,proto3" json:"GameArgMsgValue,omitempty"`
	ClubID               int64           `protobuf:"varint,5,opt,name=ClubID,proto3" json:"ClubID,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *MakeDeskReq) Reset()         { *m = MakeDeskReq{} }
func (m *MakeDeskReq) String() string { return proto.CompactTextString(m) }
func (*MakeDeskReq) ProtoMessage()    {}
func (*MakeDeskReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{4}
}

func (m *MakeDeskReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MakeDeskReq.Unmarshal(m, b)
}
func (m *MakeDeskReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MakeDeskReq.Marshal(b, m, deterministic)
}
func (m *MakeDeskReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MakeDeskReq.Merge(m, src)
}
func (m *MakeDeskReq) XXX_Size() int {
	return xxx_messageInfo_MakeDeskReq.Size(m)
}
func (m *MakeDeskReq) XXX_DiscardUnknown() {
	xxx_messageInfo_MakeDeskReq.DiscardUnknown(m)
}

var xxx_messageInfo_MakeDeskReq proto.InternalMessageInfo

func (m *MakeDeskReq) GetHead() *common.ReqHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *MakeDeskReq) GetGameName() string {
	if m != nil {
		return m.GameName
	}
	return ""
}

func (m *MakeDeskReq) GetGameArgMsgName() string {
	if m != nil {
		return m.GameArgMsgName
	}
	return ""
}

func (m *MakeDeskReq) GetGameArgMsgValue() []byte {
	if m != nil {
		return m.GameArgMsgValue
	}
	return nil
}

func (m *MakeDeskReq) GetClubID() int64 {
	if m != nil {
		return m.ClubID
	}
	return 0
}

type MakeDeskRsp struct {
	Head                 *common.RspHead  `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	Code                 MakeDeskRspCode  `protobuf:"varint,2,opt,name=Code,proto3,enum=pbgame.MakeDeskRspCode" json:"Code,omitempty"`
	StrCode              string           `protobuf:"bytes,3,opt,name=StrCode,proto3" json:"StrCode,omitempty"`
	Info                 *common.DeskInfo `protobuf:"bytes,4,opt,name=Info,proto3" json:"Info,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *MakeDeskRsp) Reset()         { *m = MakeDeskRsp{} }
func (m *MakeDeskRsp) String() string { return proto.CompactTextString(m) }
func (*MakeDeskRsp) ProtoMessage()    {}
func (*MakeDeskRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{5}
}

func (m *MakeDeskRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MakeDeskRsp.Unmarshal(m, b)
}
func (m *MakeDeskRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MakeDeskRsp.Marshal(b, m, deterministic)
}
func (m *MakeDeskRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MakeDeskRsp.Merge(m, src)
}
func (m *MakeDeskRsp) XXX_Size() int {
	return xxx_messageInfo_MakeDeskRsp.Size(m)
}
func (m *MakeDeskRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_MakeDeskRsp.DiscardUnknown(m)
}

var xxx_messageInfo_MakeDeskRsp proto.InternalMessageInfo

func (m *MakeDeskRsp) GetHead() *common.RspHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *MakeDeskRsp) GetCode() MakeDeskRspCode {
	if m != nil {
		return m.Code
	}
	return MakeDeskRspCode_MakeDeskNotUse
}

func (m *MakeDeskRsp) GetStrCode() string {
	if m != nil {
		return m.StrCode
	}
	return ""
}

func (m *MakeDeskRsp) GetInfo() *common.DeskInfo {
	if m != nil {
		return m.Info
	}
	return nil
}

// 销毁桌子 c -> s
type DestroyDeskReq struct {
	Head                 *common.ReqHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	DeskID               uint64          `protobuf:"varint,2,opt,name=DeskID,proto3" json:"DeskID,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *DestroyDeskReq) Reset()         { *m = DestroyDeskReq{} }
func (m *DestroyDeskReq) String() string { return proto.CompactTextString(m) }
func (*DestroyDeskReq) ProtoMessage()    {}
func (*DestroyDeskReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{6}
}

func (m *DestroyDeskReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DestroyDeskReq.Unmarshal(m, b)
}
func (m *DestroyDeskReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DestroyDeskReq.Marshal(b, m, deterministic)
}
func (m *DestroyDeskReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DestroyDeskReq.Merge(m, src)
}
func (m *DestroyDeskReq) XXX_Size() int {
	return xxx_messageInfo_DestroyDeskReq.Size(m)
}
func (m *DestroyDeskReq) XXX_DiscardUnknown() {
	xxx_messageInfo_DestroyDeskReq.DiscardUnknown(m)
}

var xxx_messageInfo_DestroyDeskReq proto.InternalMessageInfo

func (m *DestroyDeskReq) GetHead() *common.ReqHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *DestroyDeskReq) GetDeskID() uint64 {
	if m != nil {
		return m.DeskID
	}
	return 0
}

type DestroyDeskRsp struct {
	Head *common.RspHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	// 1成功 2失败-已经开始 3失败-必须为房主
	Code                 uint32   `protobuf:"varint,2,opt,name=Code,proto3" json:"Code,omitempty"`
	StrCode              string   `protobuf:"bytes,3,opt,name=StrCode,proto3" json:"StrCode,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DestroyDeskRsp) Reset()         { *m = DestroyDeskRsp{} }
func (m *DestroyDeskRsp) String() string { return proto.CompactTextString(m) }
func (*DestroyDeskRsp) ProtoMessage()    {}
func (*DestroyDeskRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{7}
}

func (m *DestroyDeskRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DestroyDeskRsp.Unmarshal(m, b)
}
func (m *DestroyDeskRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DestroyDeskRsp.Marshal(b, m, deterministic)
}
func (m *DestroyDeskRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DestroyDeskRsp.Merge(m, src)
}
func (m *DestroyDeskRsp) XXX_Size() int {
	return xxx_messageInfo_DestroyDeskRsp.Size(m)
}
func (m *DestroyDeskRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_DestroyDeskRsp.DiscardUnknown(m)
}

var xxx_messageInfo_DestroyDeskRsp proto.InternalMessageInfo

func (m *DestroyDeskRsp) GetHead() *common.RspHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *DestroyDeskRsp) GetCode() uint32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *DestroyDeskRsp) GetStrCode() string {
	if m != nil {
		return m.StrCode
	}
	return ""
}

// s -> c 通知桌子被销毁
type DestroyDeskNotif struct {
	DeskID               uint64   `protobuf:"varint,1,opt,name=DeskID,proto3" json:"DeskID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DestroyDeskNotif) Reset()         { *m = DestroyDeskNotif{} }
func (m *DestroyDeskNotif) String() string { return proto.CompactTextString(m) }
func (*DestroyDeskNotif) ProtoMessage()    {}
func (*DestroyDeskNotif) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{8}
}

func (m *DestroyDeskNotif) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DestroyDeskNotif.Unmarshal(m, b)
}
func (m *DestroyDeskNotif) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DestroyDeskNotif.Marshal(b, m, deterministic)
}
func (m *DestroyDeskNotif) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DestroyDeskNotif.Merge(m, src)
}
func (m *DestroyDeskNotif) XXX_Size() int {
	return xxx_messageInfo_DestroyDeskNotif.Size(m)
}
func (m *DestroyDeskNotif) XXX_DiscardUnknown() {
	xxx_messageInfo_DestroyDeskNotif.DiscardUnknown(m)
}

var xxx_messageInfo_DestroyDeskNotif proto.InternalMessageInfo

func (m *DestroyDeskNotif) GetDeskID() uint64 {
	if m != nil {
		return m.DeskID
	}
	return 0
}

// 加入桌子 c -> s
type JoinDeskReq struct {
	Head                 *common.ReqHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	DeskID               uint64          `protobuf:"varint,2,opt,name=DeskID,proto3" json:"DeskID,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *JoinDeskReq) Reset()         { *m = JoinDeskReq{} }
func (m *JoinDeskReq) String() string { return proto.CompactTextString(m) }
func (*JoinDeskReq) ProtoMessage()    {}
func (*JoinDeskReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{9}
}

func (m *JoinDeskReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_JoinDeskReq.Unmarshal(m, b)
}
func (m *JoinDeskReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_JoinDeskReq.Marshal(b, m, deterministic)
}
func (m *JoinDeskReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_JoinDeskReq.Merge(m, src)
}
func (m *JoinDeskReq) XXX_Size() int {
	return xxx_messageInfo_JoinDeskReq.Size(m)
}
func (m *JoinDeskReq) XXX_DiscardUnknown() {
	xxx_messageInfo_JoinDeskReq.DiscardUnknown(m)
}

var xxx_messageInfo_JoinDeskReq proto.InternalMessageInfo

func (m *JoinDeskReq) GetHead() *common.ReqHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *JoinDeskReq) GetDeskID() uint64 {
	if m != nil {
		return m.DeskID
	}
	return 0
}

type JoinDeskRsp struct {
	Head                 *common.RspHead  `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	Code                 JoinDeskRspCode  `protobuf:"varint,2,opt,name=Code,proto3,enum=pbgame.JoinDeskRspCode" json:"Code,omitempty"`
	ErrMsg               string           `protobuf:"bytes,3,opt,name=ErrMsg,proto3" json:"ErrMsg,omitempty"`
	Info                 *common.DeskInfo `protobuf:"bytes,4,opt,name=Info,proto3" json:"Info,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *JoinDeskRsp) Reset()         { *m = JoinDeskRsp{} }
func (m *JoinDeskRsp) String() string { return proto.CompactTextString(m) }
func (*JoinDeskRsp) ProtoMessage()    {}
func (*JoinDeskRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{10}
}

func (m *JoinDeskRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_JoinDeskRsp.Unmarshal(m, b)
}
func (m *JoinDeskRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_JoinDeskRsp.Marshal(b, m, deterministic)
}
func (m *JoinDeskRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_JoinDeskRsp.Merge(m, src)
}
func (m *JoinDeskRsp) XXX_Size() int {
	return xxx_messageInfo_JoinDeskRsp.Size(m)
}
func (m *JoinDeskRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_JoinDeskRsp.DiscardUnknown(m)
}

var xxx_messageInfo_JoinDeskRsp proto.InternalMessageInfo

func (m *JoinDeskRsp) GetHead() *common.RspHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *JoinDeskRsp) GetCode() JoinDeskRspCode {
	if m != nil {
		return m.Code
	}
	return JoinDeskRspCode_JoinDeskNotUse
}

func (m *JoinDeskRsp) GetErrMsg() string {
	if m != nil {
		return m.ErrMsg
	}
	return ""
}

func (m *JoinDeskRsp) GetInfo() *common.DeskInfo {
	if m != nil {
		return m.Info
	}
	return nil
}

// 离开桌子 c -> s
type ExitDeskReq struct {
	Head                 *common.ReqHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *ExitDeskReq) Reset()         { *m = ExitDeskReq{} }
func (m *ExitDeskReq) String() string { return proto.CompactTextString(m) }
func (*ExitDeskReq) ProtoMessage()    {}
func (*ExitDeskReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{11}
}

func (m *ExitDeskReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ExitDeskReq.Unmarshal(m, b)
}
func (m *ExitDeskReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ExitDeskReq.Marshal(b, m, deterministic)
}
func (m *ExitDeskReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ExitDeskReq.Merge(m, src)
}
func (m *ExitDeskReq) XXX_Size() int {
	return xxx_messageInfo_ExitDeskReq.Size(m)
}
func (m *ExitDeskReq) XXX_DiscardUnknown() {
	xxx_messageInfo_ExitDeskReq.DiscardUnknown(m)
}

var xxx_messageInfo_ExitDeskReq proto.InternalMessageInfo

func (m *ExitDeskReq) GetHead() *common.ReqHead {
	if m != nil {
		return m.Head
	}
	return nil
}

type ExitDeskRsp struct {
	Head *common.RspHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	// 1 成功 2 失败-游戏开始
	Code                 uint32   `protobuf:"varint,2,opt,name=Code,proto3" json:"Code,omitempty"`
	ErrMsg               string   `protobuf:"bytes,3,opt,name=ErrMsg,proto3" json:"ErrMsg,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ExitDeskRsp) Reset()         { *m = ExitDeskRsp{} }
func (m *ExitDeskRsp) String() string { return proto.CompactTextString(m) }
func (*ExitDeskRsp) ProtoMessage()    {}
func (*ExitDeskRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{12}
}

func (m *ExitDeskRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ExitDeskRsp.Unmarshal(m, b)
}
func (m *ExitDeskRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ExitDeskRsp.Marshal(b, m, deterministic)
}
func (m *ExitDeskRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ExitDeskRsp.Merge(m, src)
}
func (m *ExitDeskRsp) XXX_Size() int {
	return xxx_messageInfo_ExitDeskRsp.Size(m)
}
func (m *ExitDeskRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_ExitDeskRsp.DiscardUnknown(m)
}

var xxx_messageInfo_ExitDeskRsp proto.InternalMessageInfo

func (m *ExitDeskRsp) GetHead() *common.RspHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *ExitDeskRsp) GetCode() uint32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *ExitDeskRsp) GetErrMsg() string {
	if m != nil {
		return m.ErrMsg
	}
	return ""
}

// 游戏动作 c -> s
type GameAction struct {
	Head                 *common.ReqHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	ActionName           string          `protobuf:"bytes,2,opt,name=ActionName,proto3" json:"ActionName,omitempty"`
	ActionValue          []byte          `protobuf:"bytes,3,opt,name=ActionValue,proto3" json:"ActionValue,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *GameAction) Reset()         { *m = GameAction{} }
func (m *GameAction) String() string { return proto.CompactTextString(m) }
func (*GameAction) ProtoMessage()    {}
func (*GameAction) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{13}
}

func (m *GameAction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GameAction.Unmarshal(m, b)
}
func (m *GameAction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GameAction.Marshal(b, m, deterministic)
}
func (m *GameAction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GameAction.Merge(m, src)
}
func (m *GameAction) XXX_Size() int {
	return xxx_messageInfo_GameAction.Size(m)
}
func (m *GameAction) XXX_DiscardUnknown() {
	xxx_messageInfo_GameAction.DiscardUnknown(m)
}

var xxx_messageInfo_GameAction proto.InternalMessageInfo

func (m *GameAction) GetHead() *common.ReqHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *GameAction) GetActionName() string {
	if m != nil {
		return m.ActionName
	}
	return ""
}

func (m *GameAction) GetActionValue() []byte {
	if m != nil {
		return m.ActionValue
	}
	return nil
}

// 服务器主动通知 s -> c
type GameNotif struct {
	NotifName            string   `protobuf:"bytes,1,opt,name=NotifName,proto3" json:"NotifName,omitempty"`
	NotifValue           []byte   `protobuf:"bytes,2,opt,name=NotifValue,proto3" json:"NotifValue,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GameNotif) Reset()         { *m = GameNotif{} }
func (m *GameNotif) String() string { return proto.CompactTextString(m) }
func (*GameNotif) ProtoMessage()    {}
func (*GameNotif) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a9278d664c0c01e, []int{14}
}

func (m *GameNotif) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GameNotif.Unmarshal(m, b)
}
func (m *GameNotif) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GameNotif.Marshal(b, m, deterministic)
}
func (m *GameNotif) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GameNotif.Merge(m, src)
}
func (m *GameNotif) XXX_Size() int {
	return xxx_messageInfo_GameNotif.Size(m)
}
func (m *GameNotif) XXX_DiscardUnknown() {
	xxx_messageInfo_GameNotif.DiscardUnknown(m)
}

var xxx_messageInfo_GameNotif proto.InternalMessageInfo

func (m *GameNotif) GetNotifName() string {
	if m != nil {
		return m.NotifName
	}
	return ""
}

func (m *GameNotif) GetNotifValue() []byte {
	if m != nil {
		return m.NotifValue
	}
	return nil
}

func init() {
	proto.RegisterEnum("pbgame.MakeDeskRspCode", MakeDeskRspCode_name, MakeDeskRspCode_value)
	proto.RegisterEnum("pbgame.JoinDeskRspCode", JoinDeskRspCode_name, JoinDeskRspCode_value)
	proto.RegisterType((*QueryGameConfigReq)(nil), "pbgame.QueryGameConfigReq")
	proto.RegisterType((*QueryGameConfigRsp)(nil), "pbgame.QueryGameConfigRsp")
	proto.RegisterType((*QueryDeskInfoReq)(nil), "pbgame.QueryDeskInfoReq")
	proto.RegisterType((*QueryDeskInfoRsp)(nil), "pbgame.QueryDeskInfoRsp")
	proto.RegisterType((*MakeDeskReq)(nil), "pbgame.MakeDeskReq")
	proto.RegisterType((*MakeDeskRsp)(nil), "pbgame.MakeDeskRsp")
	proto.RegisterType((*DestroyDeskReq)(nil), "pbgame.DestroyDeskReq")
	proto.RegisterType((*DestroyDeskRsp)(nil), "pbgame.DestroyDeskRsp")
	proto.RegisterType((*DestroyDeskNotif)(nil), "pbgame.DestroyDeskNotif")
	proto.RegisterType((*JoinDeskReq)(nil), "pbgame.JoinDeskReq")
	proto.RegisterType((*JoinDeskRsp)(nil), "pbgame.JoinDeskRsp")
	proto.RegisterType((*ExitDeskReq)(nil), "pbgame.ExitDeskReq")
	proto.RegisterType((*ExitDeskRsp)(nil), "pbgame.ExitDeskRsp")
	proto.RegisterType((*GameAction)(nil), "pbgame.GameAction")
	proto.RegisterType((*GameNotif)(nil), "pbgame.GameNotif")
}

func init() { proto.RegisterFile("game/game.proto", fileDescriptor_2a9278d664c0c01e) }

var fileDescriptor_2a9278d664c0c01e = []byte{
	// 718 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x56, 0xef, 0x6e, 0xda, 0x3e,
	0x14, 0xfd, 0x05, 0x02, 0xb4, 0x97, 0x16, 0x5c, 0xff, 0xb6, 0x36, 0xa3, 0x55, 0x87, 0x22, 0xad,
	0x42, 0x9d, 0x04, 0x52, 0xb7, 0x17, 0xa8, 0x02, 0xdd, 0x98, 0x0a, 0x53, 0xc3, 0xba, 0xcf, 0x0b,
	0x60, 0xd2, 0xa8, 0x21, 0x4e, 0x1d, 0x67, 0x2a, 0x8f, 0x33, 0xed, 0x2d, 0xf6, 0x54, 0xd3, 0x9e,
	0x60, 0xb2, 0x93, 0x80, 0x4b, 0xd7, 0x8d, 0x8a, 0x7e, 0x00, 0x7c, 0xce, 0xf5, 0xfd, 0x73, 0xee,
	0xb5, 0x2d, 0xa0, 0xea, 0x3a, 0x53, 0xd2, 0x12, 0x5f, 0xcd, 0x90, 0x51, 0x4e, 0x71, 0x31, 0x1c,
	0x0a, 0x54, 0x3b, 0x18, 0xcd, 0x24, 0xdd, 0x0a, 0x87, 0xad, 0x11, 0x9d, 0x4e, 0x69, 0xd0, 0xba,
	0x22, 0xce, 0x38, 0xd9, 0x55, 0x3b, 0xbc, 0x6f, 0x4d, 0x7e, 0x12, 0xbb, 0x79, 0x0d, 0xf8, 0x22,
	0x26, 0x6c, 0xf6, 0xce, 0x99, 0x12, 0x8b, 0x06, 0x13, 0xcf, 0xb5, 0xc9, 0x0d, 0x7e, 0x05, 0xfa,
	0x7b, 0xe2, 0x8c, 0x0d, 0xad, 0xae, 0x35, 0xca, 0x27, 0x3b, 0xcd, 0x70, 0x98, 0x3a, 0xd9, 0xe4,
	0x46, 0x18, 0x6c, 0x69, 0xc6, 0x35, 0xd8, 0x10, 0x7e, 0x7d, 0x67, 0x4a, 0x8c, 0x5c, 0x5d, 0x6b,
	0x6c, 0xda, 0x73, 0x8c, 0x31, 0xe8, 0x9f, 0x66, 0x21, 0x31, 0xf2, 0x75, 0xad, 0x51, 0xb0, 0xe5,
	0xda, 0xbc, 0xb9, 0x9f, 0x2c, 0x0a, 0xff, 0x92, 0x2c, 0x0a, 0x95, 0x64, 0x06, 0x94, 0xac, 0x89,
	0xab, 0xe4, 0xca, 0xa0, 0x28, 0xc3, 0x9a, 0xb8, 0x9f, 0x1d, 0x3f, 0x4e, 0xd2, 0x6d, 0xd9, 0x73,
	0x6c, 0x5e, 0x00, 0x92, 0x29, 0xdb, 0x24, 0xba, 0xee, 0x06, 0x13, 0xfa, 0x08, 0x75, 0xbb, 0x50,
	0x94, 0x5e, 0x6d, 0x99, 0x4f, 0xb7, 0x53, 0x64, 0xc6, 0xcb, 0x21, 0x57, 0xd7, 0x80, 0x41, 0xb7,
	0xe8, 0x38, 0x11, 0xb0, 0x6d, 0xcb, 0x35, 0x3e, 0x02, 0x5d, 0x44, 0x91, 0x95, 0x97, 0x4f, 0xf0,
	0xc2, 0x75, 0x1e, 0x5f, 0xda, 0xcd, 0x1f, 0x1a, 0x94, 0x7b, 0xce, 0x35, 0x11, 0xf4, 0x13, 0xcd,
	0xe8, 0x08, 0x2a, 0x62, 0x7d, 0xca, 0xdc, 0x5e, 0x94, 0x74, 0x36, 0x2f, 0x77, 0x2c, 0xb1, 0xb8,
	0x01, 0xd5, 0x05, 0x93, 0xf4, 0x59, 0x97, 0x7d, 0x5e, 0xa6, 0x45, 0xcf, 0x2c, 0x3f, 0x1e, 0x76,
	0xdb, 0x46, 0xa1, 0xae, 0x35, 0xf2, 0x76, 0x8a, 0xcc, 0xef, 0x6a, 0xf1, 0xab, 0xf7, 0xeb, 0xb5,
	0xd2, 0xaf, 0xca, 0xc9, 0x5e, 0x33, 0x39, 0xf2, 0x4d, 0x25, 0x92, 0x30, 0xa7, 0x8d, 0x34, 0xa0,
	0x34, 0xe0, 0x4c, 0xee, 0x4f, 0x64, 0x64, 0x70, 0xde, 0x62, 0xfd, 0x1f, 0x2d, 0xfe, 0x08, 0x95,
	0x36, 0x89, 0x38, 0xa3, 0xb3, 0x47, 0x36, 0xf9, 0xa1, 0xa3, 0x42, 0xee, 0x06, 0x5c, 0xef, 0xa0,
	0x3c, 0xa8, 0xcf, 0x3c, 0x06, 0xa4, 0xa4, 0xe9, 0x53, 0xee, 0x4d, 0x94, 0x92, 0xb4, 0x3b, 0x25,
	0x9d, 0x43, 0xf9, 0x03, 0xf5, 0x82, 0x27, 0x12, 0xf8, 0x4d, 0x53, 0xc2, 0xad, 0x3d, 0x57, 0x25,
	0x92, 0x32, 0xd7, 0x5d, 0x28, 0x76, 0x18, 0xeb, 0x45, 0x6e, 0x2a, 0x3b, 0x45, 0x2b, 0x4f, 0xf5,
	0x2d, 0x94, 0x3b, 0xb7, 0x1e, 0x7f, 0x9c, 0x62, 0xf3, 0x8b, 0xe2, 0xb5, 0xde, 0xdc, 0x1e, 0xa8,
	0xdf, 0x8c, 0x01, 0xe4, 0xf5, 0x19, 0x71, 0x8f, 0x06, 0xab, 0x0e, 0xe2, 0x10, 0x20, 0x71, 0x50,
	0x2e, 0xb4, 0xc2, 0xe0, 0x3a, 0x94, 0x13, 0xa4, 0x3e, 0x87, 0x2a, 0x65, 0x76, 0x61, 0x53, 0x3e,
	0x00, 0xf2, 0x94, 0x1c, 0xc0, 0xa6, 0x5c, 0xc8, 0x68, 0x9a, 0x8c, 0xb6, 0x20, 0x44, 0x32, 0x09,
	0x92, 0x58, 0x39, 0x19, 0x4b, 0x61, 0x8e, 0x7f, 0x6a, 0x50, 0x5d, 0xba, 0x8b, 0x18, 0x43, 0x25,
	0xa3, 0xfa, 0x94, 0x5f, 0x46, 0x04, 0xfd, 0x87, 0x11, 0x6c, 0x65, 0xdc, 0x20, 0x1e, 0x8d, 0x90,
	0x86, 0xff, 0x5f, 0x38, 0x9e, 0x32, 0x37, 0xea, 0x30, 0x86, 0x72, 0xb8, 0x06, 0xbb, 0x8a, 0x6b,
	0x27, 0xa0, 0xb1, 0x7b, 0xd5, 0xa3, 0x01, 0x99, 0xa1, 0x3c, 0x7e, 0x01, 0xcf, 0xef, 0xd9, 0x04,
	0x40, 0x3a, 0x7e, 0x09, 0xfb, 0x99, 0xa9, 0x1b, 0x70, 0xc2, 0x02, 0xc7, 0x1f, 0x10, 0xf6, 0x95,
	0xb0, 0x0e, 0x63, 0x94, 0xa1, 0x82, 0xea, 0x7b, 0x19, 0x11, 0x36, 0xe0, 0x0e, 0x8f, 0x65, 0xca,
	0x22, 0x3e, 0x00, 0x23, 0x33, 0x59, 0x4e, 0xd0, 0xa7, 0xfc, 0xcc, 0x0b, 0xc6, 0xc9, 0x9b, 0x85,
	0x4a, 0x78, 0x07, 0xb6, 0x33, 0xeb, 0xb9, 0x37, 0xf5, 0x38, 0xda, 0x38, 0xfe, 0xa5, 0x41, 0x75,
	0xe9, 0x98, 0x0a, 0xc9, 0x19, 0xa5, 0x4a, 0xce, 0xb8, 0x54, 0xf2, 0x33, 0x40, 0xca, 0xae, 0xce,
	0xad, 0x17, 0x71, 0x94, 0x53, 0x59, 0xf1, 0x39, 0x8b, 0x7d, 0x1f, 0xe5, 0x85, 0xa4, 0x8c, 0xfd,
	0x93, 0x24, 0x5d, 0x48, 0xca, 0x36, 0xdc, 0x95, 0x54, 0x50, 0x4d, 0x62, 0xce, 0xaa, 0xda, 0x7d,
	0xd8, 0x9b, 0x27, 0xf3, 0x22, 0xee, 0x04, 0x23, 0x32, 0xa0, 0x96, 0x4f, 0x23, 0x82, 0x4a, 0xaa,
	0xdf, 0xa9, 0xcf, 0x88, 0x33, 0x9e, 0x75, 0x25, 0x40, 0x1b, 0xc3, 0xa2, 0xfc, 0xaf, 0xf0, 0xe6,
	0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0xca, 0xbe, 0xb5, 0x5a, 0x84, 0x08, 0x00, 0x00,
}
