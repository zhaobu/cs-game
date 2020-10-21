// Code generated by protoc-gen-go. DO NOT EDIT.
// source: common/common.proto

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

type UserStatus int32

const (
	UserStatus_NotLogin   UserStatus = 0
	UserStatus_InHall     UserStatus = 1
	UserStatus_InMatching UserStatus = 2
	UserStatus_InGameing  UserStatus = 3
)

var UserStatus_name = map[int32]string{
	0: "NotLogin",
	1: "InHall",
	2: "InMatching",
	3: "InGameing",
}

var UserStatus_value = map[string]int32{
	"NotLogin":   0,
	"InHall":     1,
	"InMatching": 2,
	"InGameing":  3,
}

func (x UserStatus) String() string {
	return proto.EnumName(UserStatus_name, int32(x))
}

func (UserStatus) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_8f954d82c0b891f6, []int{0}
}

//桌子类型
type DeskType int32

const (
	DeskType_DTNone   DeskType = 0
	DeskType_DTMatch  DeskType = 1
	DeskType_DTFriend DeskType = 2
	DeskType_DTLadder DeskType = 3
)

var DeskType_name = map[int32]string{
	0: "DTNone",
	1: "DTMatch",
	2: "DTFriend",
	3: "DTLadder",
}

var DeskType_value = map[string]int32{
	"DTNone":   0,
	"DTMatch":  1,
	"DTFriend": 2,
	"DTLadder": 3,
}

func (x DeskType) String() string {
	return proto.EnumName(DeskType_name, int32(x))
}

func (DeskType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_8f954d82c0b891f6, []int{1}
}

// 断开提示 s -> c
type BrokenTip struct {
	Msg string `protobuf:"bytes,1,opt,name=Msg,proto3" json:"Msg,omitempty"`
	// 异地登陆 1 其他*
	Code                 uint32   `protobuf:"varint,2,opt,name=Code,proto3" json:"Code,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BrokenTip) Reset()         { *m = BrokenTip{} }
func (m *BrokenTip) String() string { return proto.CompactTextString(m) }
func (*BrokenTip) ProtoMessage()    {}
func (*BrokenTip) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f954d82c0b891f6, []int{0}
}

func (m *BrokenTip) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BrokenTip.Unmarshal(m, b)
}
func (m *BrokenTip) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BrokenTip.Marshal(b, m, deterministic)
}
func (m *BrokenTip) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BrokenTip.Merge(m, src)
}
func (m *BrokenTip) XXX_Size() int {
	return xxx_messageInfo_BrokenTip.Size(m)
}
func (m *BrokenTip) XXX_DiscardUnknown() {
	xxx_messageInfo_BrokenTip.DiscardUnknown(m)
}

var xxx_messageInfo_BrokenTip proto.InternalMessageInfo

func (m *BrokenTip) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

func (m *BrokenTip) GetCode() uint32 {
	if m != nil {
		return m.Code
	}
	return 0
}

// 错误提示 s -> c
type ErrorTip struct {
	Msg                  string   `protobuf:"bytes,1,opt,name=Msg,proto3" json:"Msg,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ErrorTip) Reset()         { *m = ErrorTip{} }
func (m *ErrorTip) String() string { return proto.CompactTextString(m) }
func (*ErrorTip) ProtoMessage()    {}
func (*ErrorTip) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f954d82c0b891f6, []int{1}
}

func (m *ErrorTip) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ErrorTip.Unmarshal(m, b)
}
func (m *ErrorTip) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ErrorTip.Marshal(b, m, deterministic)
}
func (m *ErrorTip) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ErrorTip.Merge(m, src)
}
func (m *ErrorTip) XXX_Size() int {
	return xxx_messageInfo_ErrorTip.Size(m)
}
func (m *ErrorTip) XXX_DiscardUnknown() {
	xxx_messageInfo_ErrorTip.DiscardUnknown(m)
}

var xxx_messageInfo_ErrorTip proto.InternalMessageInfo

func (m *ErrorTip) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

type SiteDownPlayerInfo struct {
	UserID               uint64   `protobuf:"varint,1,opt,name=UserID,proto3" json:"UserID,omitempty"`
	Dir                  int32    `protobuf:"varint,2,opt,name=Dir,proto3" json:"Dir,omitempty"`
	Name                 string   `protobuf:"bytes,3,opt,name=Name,proto3" json:"Name,omitempty"`
	Profile              string   `protobuf:"bytes,4,opt,name=Profile,proto3" json:"Profile,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SiteDownPlayerInfo) Reset()         { *m = SiteDownPlayerInfo{} }
func (m *SiteDownPlayerInfo) String() string { return proto.CompactTextString(m) }
func (*SiteDownPlayerInfo) ProtoMessage()    {}
func (*SiteDownPlayerInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f954d82c0b891f6, []int{2}
}

func (m *SiteDownPlayerInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SiteDownPlayerInfo.Unmarshal(m, b)
}
func (m *SiteDownPlayerInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SiteDownPlayerInfo.Marshal(b, m, deterministic)
}
func (m *SiteDownPlayerInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SiteDownPlayerInfo.Merge(m, src)
}
func (m *SiteDownPlayerInfo) XXX_Size() int {
	return xxx_messageInfo_SiteDownPlayerInfo.Size(m)
}
func (m *SiteDownPlayerInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_SiteDownPlayerInfo.DiscardUnknown(m)
}

var xxx_messageInfo_SiteDownPlayerInfo proto.InternalMessageInfo

func (m *SiteDownPlayerInfo) GetUserID() uint64 {
	if m != nil {
		return m.UserID
	}
	return 0
}

func (m *SiteDownPlayerInfo) GetDir() int32 {
	if m != nil {
		return m.Dir
	}
	return 0
}

func (m *SiteDownPlayerInfo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *SiteDownPlayerInfo) GetProfile() string {
	if m != nil {
		return m.Profile
	}
	return ""
}

// 桌子信息
type DeskInfo struct {
	Uuid                 uint64                `protobuf:"varint,1,opt,name=Uuid,proto3" json:"Uuid,omitempty"`
	ID                   uint64                `protobuf:"varint,2,opt,name=ID,proto3" json:"ID,omitempty"`
	CreateUserID         uint64                `protobuf:"varint,3,opt,name=CreateUserID,proto3" json:"CreateUserID,omitempty"`
	CreateUserName       string                `protobuf:"bytes,4,opt,name=CreateUserName,proto3" json:"CreateUserName,omitempty"`
	CreateUserProfile    string                `protobuf:"bytes,5,opt,name=CreateUserProfile,proto3" json:"CreateUserProfile,omitempty"`
	CreateTime           int64                 `protobuf:"varint,6,opt,name=CreateTime,proto3" json:"CreateTime,omitempty"`
	CreateFee            int64                 `protobuf:"varint,7,opt,name=CreateFee,proto3" json:"CreateFee,omitempty"`
	ArgName              string                `protobuf:"bytes,8,opt,name=ArgName,proto3" json:"ArgName,omitempty"`
	ArgValue             []byte                `protobuf:"bytes,9,opt,name=ArgValue,proto3" json:"ArgValue,omitempty"`
	Status               string                `protobuf:"bytes,10,opt,name=Status,proto3" json:"Status,omitempty"`
	GameName             string                `protobuf:"bytes,11,opt,name=GameName,proto3" json:"GameName,omitempty"`
	GameID               string                `protobuf:"bytes,12,opt,name=GameID,proto3" json:"GameID,omitempty"`
	ClubID               int64                 `protobuf:"varint,13,opt,name=ClubID,proto3" json:"ClubID,omitempty"`
	Kind                 DeskType              `protobuf:"varint,14,opt,name=Kind,proto3,enum=pbcommon.DeskType" json:"Kind,omitempty"`
	SdInfos              []*SiteDownPlayerInfo `protobuf:"bytes,15,rep,name=SdInfos,proto3" json:"SdInfos,omitempty"`
	TotalLoop            int64                 `protobuf:"varint,16,opt,name=TotalLoop,proto3" json:"TotalLoop,omitempty"`
	CurrLoop             int64                 `protobuf:"varint,17,opt,name=CurrLoop,proto3" json:"CurrLoop,omitempty"`
	CreateVlaueHash      uint64                `protobuf:"varint,18,opt,name=CreateVlaueHash,proto3" json:"CreateVlaueHash,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *DeskInfo) Reset()         { *m = DeskInfo{} }
func (m *DeskInfo) String() string { return proto.CompactTextString(m) }
func (*DeskInfo) ProtoMessage()    {}
func (*DeskInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f954d82c0b891f6, []int{3}
}

func (m *DeskInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeskInfo.Unmarshal(m, b)
}
func (m *DeskInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeskInfo.Marshal(b, m, deterministic)
}
func (m *DeskInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeskInfo.Merge(m, src)
}
func (m *DeskInfo) XXX_Size() int {
	return xxx_messageInfo_DeskInfo.Size(m)
}
func (m *DeskInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_DeskInfo.DiscardUnknown(m)
}

var xxx_messageInfo_DeskInfo proto.InternalMessageInfo

func (m *DeskInfo) GetUuid() uint64 {
	if m != nil {
		return m.Uuid
	}
	return 0
}

func (m *DeskInfo) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *DeskInfo) GetCreateUserID() uint64 {
	if m != nil {
		return m.CreateUserID
	}
	return 0
}

func (m *DeskInfo) GetCreateUserName() string {
	if m != nil {
		return m.CreateUserName
	}
	return ""
}

func (m *DeskInfo) GetCreateUserProfile() string {
	if m != nil {
		return m.CreateUserProfile
	}
	return ""
}

func (m *DeskInfo) GetCreateTime() int64 {
	if m != nil {
		return m.CreateTime
	}
	return 0
}

func (m *DeskInfo) GetCreateFee() int64 {
	if m != nil {
		return m.CreateFee
	}
	return 0
}

func (m *DeskInfo) GetArgName() string {
	if m != nil {
		return m.ArgName
	}
	return ""
}

func (m *DeskInfo) GetArgValue() []byte {
	if m != nil {
		return m.ArgValue
	}
	return nil
}

func (m *DeskInfo) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

func (m *DeskInfo) GetGameName() string {
	if m != nil {
		return m.GameName
	}
	return ""
}

func (m *DeskInfo) GetGameID() string {
	if m != nil {
		return m.GameID
	}
	return ""
}

func (m *DeskInfo) GetClubID() int64 {
	if m != nil {
		return m.ClubID
	}
	return 0
}

func (m *DeskInfo) GetKind() DeskType {
	if m != nil {
		return m.Kind
	}
	return DeskType_DTNone
}

func (m *DeskInfo) GetSdInfos() []*SiteDownPlayerInfo {
	if m != nil {
		return m.SdInfos
	}
	return nil
}

func (m *DeskInfo) GetTotalLoop() int64 {
	if m != nil {
		return m.TotalLoop
	}
	return 0
}

func (m *DeskInfo) GetCurrLoop() int64 {
	if m != nil {
		return m.CurrLoop
	}
	return 0
}

func (m *DeskInfo) GetCreateVlaueHash() uint64 {
	if m != nil {
		return m.CreateVlaueHash
	}
	return 0
}

// 会话信息
type SessionInfo struct {
	Uuid                 uint64     `protobuf:"varint,1,opt,name=Uuid,proto3" json:"Uuid,omitempty"`
	UserID               uint64     `protobuf:"varint,2,opt,name=UserID,proto3" json:"UserID,omitempty"`
	SessionID            string     `protobuf:"bytes,3,opt,name=SessionID,proto3" json:"SessionID,omitempty"`
	Status               UserStatus `protobuf:"varint,4,opt,name=Status,proto3,enum=pbcommon.UserStatus" json:"Status,omitempty"`
	AtDeskID             uint64     `protobuf:"varint,5,opt,name=AtDeskID,proto3" json:"AtDeskID,omitempty"`
	GameName             string     `protobuf:"bytes,6,opt,name=GameName,proto3" json:"GameName,omitempty"`
	GameID               string     `protobuf:"bytes,7,opt,name=GameID,proto3" json:"GameID,omitempty"`
	LastActiveTime       int64      `protobuf:"varint,8,opt,name=LastActiveTime,proto3" json:"LastActiveTime,omitempty"`
	RoomID               uint32     `protobuf:"varint,9,opt,name=RoomID,proto3" json:"RoomID,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *SessionInfo) Reset()         { *m = SessionInfo{} }
func (m *SessionInfo) String() string { return proto.CompactTextString(m) }
func (*SessionInfo) ProtoMessage()    {}
func (*SessionInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f954d82c0b891f6, []int{4}
}

func (m *SessionInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SessionInfo.Unmarshal(m, b)
}
func (m *SessionInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SessionInfo.Marshal(b, m, deterministic)
}
func (m *SessionInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SessionInfo.Merge(m, src)
}
func (m *SessionInfo) XXX_Size() int {
	return xxx_messageInfo_SessionInfo.Size(m)
}
func (m *SessionInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_SessionInfo.DiscardUnknown(m)
}

var xxx_messageInfo_SessionInfo proto.InternalMessageInfo

func (m *SessionInfo) GetUuid() uint64 {
	if m != nil {
		return m.Uuid
	}
	return 0
}

func (m *SessionInfo) GetUserID() uint64 {
	if m != nil {
		return m.UserID
	}
	return 0
}

func (m *SessionInfo) GetSessionID() string {
	if m != nil {
		return m.SessionID
	}
	return ""
}

func (m *SessionInfo) GetStatus() UserStatus {
	if m != nil {
		return m.Status
	}
	return UserStatus_NotLogin
}

func (m *SessionInfo) GetAtDeskID() uint64 {
	if m != nil {
		return m.AtDeskID
	}
	return 0
}

func (m *SessionInfo) GetGameName() string {
	if m != nil {
		return m.GameName
	}
	return ""
}

func (m *SessionInfo) GetGameID() string {
	if m != nil {
		return m.GameID
	}
	return ""
}

func (m *SessionInfo) GetLastActiveTime() int64 {
	if m != nil {
		return m.LastActiveTime
	}
	return 0
}

func (m *SessionInfo) GetRoomID() uint32 {
	if m != nil {
		return m.RoomID
	}
	return 0
}

// 用户信息
type UserInfo struct {
	UserID               uint64   `protobuf:"varint,1,opt,name=UserID,proto3" json:"UserID,omitempty"`
	SessionID            string   `protobuf:"bytes,2,opt,name=SessionID,proto3" json:"SessionID,omitempty"`
	WxID                 string   `protobuf:"bytes,3,opt,name=WxID,proto3" json:"WxID,omitempty"`
	Longitude            float64  `protobuf:"fixed64,4,opt,name=Longitude,proto3" json:"Longitude,omitempty"`
	Latitude             float64  `protobuf:"fixed64,5,opt,name=Latitude,proto3" json:"Latitude,omitempty"`
	Name                 string   `protobuf:"bytes,6,opt,name=Name,proto3" json:"Name,omitempty"`
	Sex                  string   `protobuf:"bytes,7,opt,name=Sex,proto3" json:"Sex,omitempty"`
	Profile              string   `protobuf:"bytes,8,opt,name=Profile,proto3" json:"Profile,omitempty"`
	RegisterTime         int64    `protobuf:"varint,9,opt,name=RegisterTime,proto3" json:"RegisterTime,omitempty"`
	Mobile               string   `protobuf:"bytes,10,opt,name=Mobile,proto3" json:"Mobile,omitempty"`
	XLID                 string   `protobuf:"bytes,11,opt,name=XLID,proto3" json:"XLID,omitempty"`
	Gold                 uint64   `protobuf:"varint,12,opt,name=Gold,proto3" json:"Gold,omitempty"`
	Masonry              uint64   `protobuf:"varint,13,opt,name=Masonry,proto3" json:"Masonry,omitempty"`
	GoldPre              uint64   `protobuf:"varint,14,opt,name=GoldPre,proto3" json:"GoldPre,omitempty"`
	MasonryPre           uint64   `protobuf:"varint,15,opt,name=MasonryPre,proto3" json:"MasonryPre,omitempty"`
	Agent                string   `protobuf:"bytes,16,opt,name=Agent,proto3" json:"Agent,omitempty"`
	IdentityCard         string   `protobuf:"bytes,17,opt,name=IdentityCard,proto3" json:"IdentityCard,omitempty"`
	IdentityCardName     string   `protobuf:"bytes,18,opt,name=IdentityCardName,proto3" json:"IdentityCardName,omitempty"`
	Password             string   `protobuf:"bytes,19,opt,name=Password,proto3" json:"Password,omitempty"`
	IP                   string   `protobuf:"bytes,20,opt,name=IP,proto3" json:"IP,omitempty"`
	PlayCount            int32    `protobuf:"varint,21,opt,name=PlayCount,proto3" json:"PlayCount,omitempty"`
	BindMobileCount      int32    `protobuf:"varint,22,opt,name=BindMobileCount,proto3" json:"BindMobileCount,omitempty"`
	Place                string   `protobuf:"bytes,23,opt,name=Place,proto3" json:"Place,omitempty"`
	IsRedName            bool     `protobuf:"varint,24,opt,name=IsRedName,proto3" json:"IsRedName,omitempty"`
	TotalCase            uint64   `protobuf:"varint,25,opt,name=TotalCase,proto3" json:"TotalCase,omitempty"`
	WinCaseCase          uint64   `protobuf:"varint,26,opt,name=WinCaseCase,proto3" json:"WinCaseCase,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UserInfo) Reset()         { *m = UserInfo{} }
func (m *UserInfo) String() string { return proto.CompactTextString(m) }
func (*UserInfo) ProtoMessage()    {}
func (*UserInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f954d82c0b891f6, []int{5}
}

func (m *UserInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UserInfo.Unmarshal(m, b)
}
func (m *UserInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UserInfo.Marshal(b, m, deterministic)
}
func (m *UserInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UserInfo.Merge(m, src)
}
func (m *UserInfo) XXX_Size() int {
	return xxx_messageInfo_UserInfo.Size(m)
}
func (m *UserInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_UserInfo.DiscardUnknown(m)
}

var xxx_messageInfo_UserInfo proto.InternalMessageInfo

func (m *UserInfo) GetUserID() uint64 {
	if m != nil {
		return m.UserID
	}
	return 0
}

func (m *UserInfo) GetSessionID() string {
	if m != nil {
		return m.SessionID
	}
	return ""
}

func (m *UserInfo) GetWxID() string {
	if m != nil {
		return m.WxID
	}
	return ""
}

func (m *UserInfo) GetLongitude() float64 {
	if m != nil {
		return m.Longitude
	}
	return 0
}

func (m *UserInfo) GetLatitude() float64 {
	if m != nil {
		return m.Latitude
	}
	return 0
}

func (m *UserInfo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *UserInfo) GetSex() string {
	if m != nil {
		return m.Sex
	}
	return ""
}

func (m *UserInfo) GetProfile() string {
	if m != nil {
		return m.Profile
	}
	return ""
}

func (m *UserInfo) GetRegisterTime() int64 {
	if m != nil {
		return m.RegisterTime
	}
	return 0
}

func (m *UserInfo) GetMobile() string {
	if m != nil {
		return m.Mobile
	}
	return ""
}

func (m *UserInfo) GetXLID() string {
	if m != nil {
		return m.XLID
	}
	return ""
}

func (m *UserInfo) GetGold() uint64 {
	if m != nil {
		return m.Gold
	}
	return 0
}

func (m *UserInfo) GetMasonry() uint64 {
	if m != nil {
		return m.Masonry
	}
	return 0
}

func (m *UserInfo) GetGoldPre() uint64 {
	if m != nil {
		return m.GoldPre
	}
	return 0
}

func (m *UserInfo) GetMasonryPre() uint64 {
	if m != nil {
		return m.MasonryPre
	}
	return 0
}

func (m *UserInfo) GetAgent() string {
	if m != nil {
		return m.Agent
	}
	return ""
}

func (m *UserInfo) GetIdentityCard() string {
	if m != nil {
		return m.IdentityCard
	}
	return ""
}

func (m *UserInfo) GetIdentityCardName() string {
	if m != nil {
		return m.IdentityCardName
	}
	return ""
}

func (m *UserInfo) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

func (m *UserInfo) GetIP() string {
	if m != nil {
		return m.IP
	}
	return ""
}

func (m *UserInfo) GetPlayCount() int32 {
	if m != nil {
		return m.PlayCount
	}
	return 0
}

func (m *UserInfo) GetBindMobileCount() int32 {
	if m != nil {
		return m.BindMobileCount
	}
	return 0
}

func (m *UserInfo) GetPlace() string {
	if m != nil {
		return m.Place
	}
	return ""
}

func (m *UserInfo) GetIsRedName() bool {
	if m != nil {
		return m.IsRedName
	}
	return false
}

func (m *UserInfo) GetTotalCase() uint64 {
	if m != nil {
		return m.TotalCase
	}
	return 0
}

func (m *UserInfo) GetWinCaseCase() uint64 {
	if m != nil {
		return m.WinCaseCase
	}
	return 0
}

func init() {
	proto.RegisterEnum("pbcommon.UserStatus", UserStatus_name, UserStatus_value)
	proto.RegisterEnum("pbcommon.DeskType", DeskType_name, DeskType_value)
	proto.RegisterType((*BrokenTip)(nil), "pbcommon.BrokenTip")
	proto.RegisterType((*ErrorTip)(nil), "pbcommon.ErrorTip")
	proto.RegisterType((*SiteDownPlayerInfo)(nil), "pbcommon.SiteDownPlayerInfo")
	proto.RegisterType((*DeskInfo)(nil), "pbcommon.DeskInfo")
	proto.RegisterType((*SessionInfo)(nil), "pbcommon.SessionInfo")
	proto.RegisterType((*UserInfo)(nil), "pbcommon.UserInfo")
}

func init() { proto.RegisterFile("common/common.proto", fileDescriptor_8f954d82c0b891f6) }

var fileDescriptor_8f954d82c0b891f6 = []byte{
	// 902 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x55, 0xef, 0x6e, 0xdb, 0x36,
	0x10, 0xaf, 0x6c, 0xc5, 0x96, 0xce, 0x89, 0xa3, 0xb2, 0x59, 0xc7, 0x15, 0xc1, 0x60, 0xf8, 0x43,
	0x60, 0x04, 0x45, 0x86, 0x75, 0xc0, 0xbe, 0x0e, 0xae, 0xd5, 0x3f, 0xc2, 0x9c, 0xc0, 0xa0, 0xdd,
	0x76, 0x5f, 0x99, 0x88, 0x75, 0x85, 0xca, 0xa4, 0x41, 0xc9, 0x6b, 0xf3, 0x20, 0x7b, 0xa3, 0x3d,
	0xc2, 0x1e, 0x68, 0xb8, 0xa3, 0x2c, 0xd9, 0xce, 0xba, 0x4f, 0xba, 0xdf, 0xef, 0xee, 0x48, 0xde,
	0x5f, 0xc1, 0x93, 0x3b, 0xb3, 0x5a, 0x19, 0xfd, 0x93, 0xfb, 0x5c, 0xad, 0xad, 0x29, 0x0d, 0x0b,
	0xd6, 0xb7, 0x0e, 0x0f, 0x7f, 0x86, 0xf0, 0xa5, 0x35, 0x9f, 0x95, 0x5e, 0x64, 0x6b, 0x16, 0x41,
	0xfb, 0xba, 0x58, 0x72, 0x6f, 0xe0, 0x8d, 0x42, 0x81, 0x22, 0x63, 0xe0, 0x4f, 0x4c, 0xaa, 0x78,
	0x6b, 0xe0, 0x8d, 0x4e, 0x04, 0xc9, 0xc3, 0x73, 0x08, 0x5e, 0x59, 0x6b, 0xec, 0x7f, 0x7a, 0x0c,
	0x73, 0x60, 0xf3, 0xac, 0x54, 0xb1, 0xf9, 0xa2, 0x67, 0xb9, 0xbc, 0x57, 0x36, 0xd1, 0x1f, 0x0d,
	0x7b, 0x0a, 0x9d, 0x77, 0x85, 0xb2, 0x49, 0x4c, 0xa6, 0xbe, 0xa8, 0x10, 0xfa, 0xc7, 0x99, 0xa5,
	0xe3, 0x8f, 0x04, 0x8a, 0x78, 0xe3, 0x8d, 0x5c, 0x29, 0xde, 0xa6, 0x23, 0x49, 0x66, 0x1c, 0xba,
	0x33, 0x6b, 0x3e, 0x66, 0xb9, 0xe2, 0x3e, 0xd1, 0x5b, 0x38, 0xfc, 0xdb, 0x87, 0x20, 0x56, 0xc5,
	0x67, 0xba, 0x84, 0x81, 0xff, 0x6e, 0x93, 0xa5, 0xd5, 0x15, 0x24, 0xb3, 0x3e, 0xb4, 0x92, 0x98,
	0xce, 0xf7, 0x45, 0x2b, 0x89, 0xd9, 0x10, 0x8e, 0x27, 0x56, 0xc9, 0x52, 0x55, 0xcf, 0x69, 0x93,
	0x66, 0x8f, 0x63, 0x17, 0xd0, 0x6f, 0x30, 0x3d, 0xc6, 0xdd, 0x7a, 0xc0, 0xb2, 0xe7, 0xf0, 0xb8,
	0x61, 0xb6, 0x0f, 0x3c, 0x22, 0xd3, 0x87, 0x0a, 0xf6, 0x23, 0x80, 0x23, 0x17, 0xd9, 0x4a, 0xf1,
	0xce, 0xc0, 0x1b, 0xb5, 0xc5, 0x0e, 0xc3, 0xce, 0x21, 0x74, 0xe8, 0xb5, 0x52, 0xbc, 0x4b, 0xea,
	0x86, 0xc0, 0x14, 0x8c, 0xed, 0x92, 0x1e, 0x13, 0xb8, 0x14, 0x54, 0x90, 0x3d, 0x83, 0x60, 0x6c,
	0x97, 0xef, 0x65, 0xbe, 0x51, 0x3c, 0x1c, 0x78, 0xa3, 0x63, 0x51, 0x63, 0x4c, 0xfb, 0xbc, 0x94,
	0xe5, 0xa6, 0xe0, 0x40, 0x4e, 0x15, 0x42, 0x9f, 0x37, 0x72, 0xa5, 0xe8, 0xb8, 0x1e, 0x69, 0x6a,
	0x8c, 0x3e, 0x28, 0x27, 0x31, 0x3f, 0x76, 0x3e, 0x0e, 0x21, 0x3f, 0xc9, 0x37, 0xb7, 0x49, 0xcc,
	0x4f, 0xe8, 0x71, 0x15, 0x62, 0x17, 0xe0, 0xff, 0x9e, 0xe9, 0x94, 0xf7, 0x07, 0xde, 0xa8, 0xff,
	0x82, 0x5d, 0x6d, 0x5b, 0xeb, 0x0a, 0xeb, 0xb2, 0xb8, 0x5f, 0x2b, 0x41, 0x7a, 0xf6, 0x2b, 0x74,
	0xe7, 0x29, 0xd6, 0xa9, 0xe0, 0xa7, 0x83, 0xf6, 0xa8, 0xf7, 0xe2, 0xbc, 0x31, 0x7d, 0xd8, 0x31,
	0x62, 0x6b, 0x8c, 0x79, 0x59, 0x98, 0x52, 0xe6, 0x53, 0x63, 0xd6, 0x3c, 0x72, 0x79, 0xa9, 0x09,
	0x8c, 0x64, 0xb2, 0xb1, 0x96, 0x94, 0x8f, 0x49, 0x59, 0x63, 0x36, 0x82, 0x53, 0x97, 0xc0, 0xf7,
	0xb9, 0xdc, 0xa8, 0xb7, 0xb2, 0xf8, 0xc4, 0x19, 0x95, 0xfb, 0x90, 0x1e, 0xfe, 0xd5, 0x82, 0xde,
	0x5c, 0x15, 0x45, 0x66, 0xf4, 0x37, 0x3b, 0xa9, 0x69, 0xe1, 0xd6, 0x5e, 0x0b, 0x9f, 0x43, 0xb8,
	0x75, 0x8d, 0xab, 0xae, 0x6d, 0x08, 0xf6, 0xbc, 0xae, 0x80, 0x4f, 0xf9, 0x39, 0x6b, 0x82, 0x46,
	0x7f, 0xa7, 0xdb, 0xad, 0xcb, 0xb8, 0xa4, 0x7e, 0x8e, 0xa9, 0x91, 0x7c, 0x51, 0xe3, 0xbd, 0x9a,
	0x75, 0xbe, 0x59, 0xb3, 0xee, 0x5e, 0xcd, 0x2e, 0xa0, 0x3f, 0x95, 0x45, 0x39, 0xbe, 0x2b, 0xb3,
	0x3f, 0x5d, 0xdf, 0x05, 0x94, 0xa3, 0x03, 0x16, 0xfd, 0x85, 0x31, 0xab, 0x24, 0xa6, 0x0e, 0x3a,
	0x11, 0x15, 0x1a, 0xfe, 0x73, 0x04, 0x01, 0x85, 0xf9, 0x7f, 0x33, 0xbc, 0x97, 0x80, 0xd6, 0x61,
	0x02, 0x18, 0xf8, 0x1f, 0xbe, 0xd6, 0x99, 0x21, 0x19, 0x3d, 0xa6, 0x46, 0x2f, 0xb3, 0x72, 0x93,
	0xba, 0xd9, 0xf2, 0x44, 0x43, 0x60, 0xa0, 0x53, 0x59, 0x3a, 0xe5, 0x11, 0x29, 0x6b, 0x5c, 0x6f,
	0x87, 0xce, 0xce, 0x76, 0x88, 0xa0, 0x3d, 0x57, 0x5f, 0xab, 0xc8, 0x51, 0xdc, 0xdd, 0x17, 0xc1,
	0xde, 0xbe, 0xc0, 0xf1, 0x17, 0x6a, 0x99, 0x15, 0xa5, 0xb2, 0x94, 0x8e, 0x90, 0xd2, 0xb1, 0xc7,
	0x61, 0x9c, 0xd7, 0xe6, 0x16, 0x9d, 0xab, 0xa1, 0x71, 0x08, 0xef, 0xfe, 0x63, 0x9a, 0xc4, 0xd5,
	0xc0, 0x90, 0x8c, 0xdc, 0x1b, 0x93, 0xa7, 0x34, 0x2a, 0xbe, 0x20, 0x19, 0x6f, 0xbf, 0x96, 0x85,
	0xd1, 0xf6, 0x9e, 0x26, 0xc5, 0x17, 0x5b, 0x88, 0x1a, 0xb4, 0x98, 0x59, 0x45, 0xd3, 0xe2, 0x8b,
	0x2d, 0xc4, 0xe5, 0x50, 0x19, 0xa1, 0xf2, 0x94, 0x94, 0x3b, 0x0c, 0x3b, 0x83, 0xa3, 0xf1, 0x52,
	0xe9, 0x92, 0x06, 0x20, 0x14, 0x0e, 0x60, 0x34, 0x49, 0xaa, 0x74, 0x99, 0x95, 0xf7, 0x13, 0x69,
	0x53, 0x1a, 0x80, 0x50, 0xec, 0x71, 0xec, 0x12, 0xa2, 0x5d, 0x4c, 0xd9, 0x63, 0x64, 0xf7, 0x80,
	0xc7, 0xcc, 0xcf, 0x64, 0x51, 0x7c, 0x31, 0x36, 0xe5, 0x4f, 0x5c, 0x8b, 0x6d, 0x31, 0x2d, 0xd2,
	0x19, 0x3f, 0x23, 0xb6, 0x95, 0xcc, 0xb0, 0x86, 0x38, 0xad, 0x13, 0xb3, 0xd1, 0x25, 0xff, 0x8e,
	0xf6, 0x77, 0x43, 0xe0, 0xe8, 0xbd, 0xcc, 0x74, 0xea, 0x32, 0xe7, 0x6c, 0x9e, 0x92, 0xcd, 0x21,
	0x8d, 0x91, 0xcd, 0x72, 0x79, 0xa7, 0xf8, 0xf7, 0x2e, 0x32, 0x02, 0x78, 0x7a, 0x52, 0x08, 0xe5,
	0x9e, 0xcb, 0x07, 0xde, 0x28, 0x10, 0x0d, 0x51, 0xaf, 0x84, 0x89, 0x2c, 0x14, 0xff, 0x81, 0x92,
	0xd5, 0x10, 0x6c, 0x00, 0xbd, 0x0f, 0x99, 0x46, 0x91, 0xf4, 0xcf, 0x48, 0xbf, 0x4b, 0x5d, 0xbe,
	0x02, 0x68, 0x86, 0x8f, 0x1d, 0x43, 0x70, 0x63, 0xca, 0xa9, 0x59, 0x66, 0x3a, 0x7a, 0xc4, 0x00,
	0x3a, 0x89, 0x7e, 0x2b, 0xf3, 0x3c, 0xf2, 0x58, 0x1f, 0x20, 0xd1, 0xd7, 0xb2, 0xbc, 0xfb, 0x94,
	0xe9, 0x65, 0xd4, 0x62, 0x27, 0x10, 0x26, 0x1a, 0x47, 0x0b, 0x61, 0xfb, 0xf2, 0x37, 0xf7, 0xef,
	0xc1, 0x1d, 0x87, 0x6e, 0xf1, 0xe2, 0xc6, 0x68, 0x15, 0x3d, 0x62, 0x3d, 0xe8, 0xc6, 0x0b, 0x72,
	0x8b, 0x3c, 0x3c, 0x3d, 0x5e, 0xbc, 0xb6, 0x99, 0xd2, 0x69, 0xd4, 0x72, 0x68, 0x2a, 0xd3, 0x54,
	0xd9, 0xa8, 0x7d, 0xdb, 0xa1, 0xbf, 0xf1, 0x2f, 0xff, 0x06, 0x00, 0x00, 0xff, 0xff, 0x0e, 0xbf,
	0xfd, 0xa1, 0xa4, 0x07, 0x00, 0x00,
}