// Code generated by protoc-gen-go. DO NOT EDIT.
// source: login/login.proto

package pblogin

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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// 登陆方式
type LoginType int32

const (
	LoginType_NotUse1  LoginType = 0
	LoginType_WX       LoginType = 1
	LoginType_QQ       LoginType = 2
	LoginType_Phone    LoginType = 3
	LoginType_Email    LoginType = 4
	LoginType_XianLiao LoginType = 5
)

var LoginType_name = map[int32]string{
	0: "NotUse1",
	1: "WX",
	2: "QQ",
	3: "Phone",
	4: "Email",
	5: "XianLiao",
}

var LoginType_value = map[string]int32{
	"NotUse1":  0,
	"WX":       1,
	"QQ":       2,
	"Phone":    3,
	"Email":    4,
	"XianLiao": 5,
}

func (x LoginType) String() string {
	return proto.EnumName(LoginType_name, int32(x))
}

func (LoginType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_6fe61ab550dd3bc4, []int{0}
}

// 登陆回应code
type LoginRspCode int32

const (
	LoginRspCode_NotUse2       LoginRspCode = 0
	LoginRspCode_IdOrPwdFailed LoginRspCode = 1
	LoginRspCode_Succ          LoginRspCode = 2
	LoginRspCode_Other         LoginRspCode = 3
	LoginRspCode_MobileNoBind  LoginRspCode = 4
	LoginRspCode_XLNoBind      LoginRspCode = 5
)

var LoginRspCode_name = map[int32]string{
	0: "NotUse2",
	1: "IdOrPwdFailed",
	2: "Succ",
	3: "Other",
	4: "MobileNoBind",
	5: "XLNoBind",
}

var LoginRspCode_value = map[string]int32{
	"NotUse2":       0,
	"IdOrPwdFailed": 1,
	"Succ":          2,
	"Other":         3,
	"MobileNoBind":  4,
	"XLNoBind":      5,
}

func (x LoginRspCode) String() string {
	return proto.EnumName(LoginRspCode_name, int32(x))
}

func (LoginRspCode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_6fe61ab550dd3bc4, []int{1}
}

// 登陆请求 c -> s
type LoginReq struct {
	Head                 *common.ReqHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	LoginType            LoginType       `protobuf:"varint,2,opt,name=LoginType,proto3,enum=pblogin.LoginType" json:"LoginType,omitempty"`
	ID                   string          `protobuf:"bytes,3,opt,name=ID,proto3" json:"ID,omitempty"`
	Password             string          `protobuf:"bytes,4,opt,name=Password,proto3" json:"Password,omitempty"`
	Version              string          `protobuf:"bytes,5,opt,name=Version,proto3" json:"Version,omitempty"`
	DeviceType           string          `protobuf:"bytes,6,opt,name=DeviceType,proto3" json:"DeviceType,omitempty"`
	Sex                  string          `protobuf:"bytes,7,opt,name=Sex,proto3" json:"Sex,omitempty"`
	PlatformID           string          `protobuf:"bytes,8,opt,name=PlatformID,proto3" json:"PlatformID,omitempty"`
	Name                 string          `protobuf:"bytes,9,opt,name=Name,proto3" json:"Name,omitempty"`
	Profile              string          `protobuf:"bytes,10,opt,name=Profile,proto3" json:"Profile,omitempty"`
	Longitude            float64         `protobuf:"fixed64,13,opt,name=Longitude,proto3" json:"Longitude,omitempty"`
	Latitude             float64         `protobuf:"fixed64,14,opt,name=Latitude,proto3" json:"Latitude,omitempty"`
	Place                string          `protobuf:"bytes,15,opt,name=Place,proto3" json:"Place,omitempty"`
	IsSimulator          string          `protobuf:"bytes,16,opt,name=IsSimulator,proto3" json:"IsSimulator,omitempty"`
	NetworkType          string          `protobuf:"bytes,17,opt,name=NetworkType,proto3" json:"NetworkType,omitempty"`
	Energy               string          `protobuf:"bytes,18,opt,name=Energy,proto3" json:"Energy,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *LoginReq) Reset()         { *m = LoginReq{} }
func (m *LoginReq) String() string { return proto.CompactTextString(m) }
func (*LoginReq) ProtoMessage()    {}
func (*LoginReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_6fe61ab550dd3bc4, []int{0}
}

func (m *LoginReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LoginReq.Unmarshal(m, b)
}
func (m *LoginReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LoginReq.Marshal(b, m, deterministic)
}
func (m *LoginReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LoginReq.Merge(m, src)
}
func (m *LoginReq) XXX_Size() int {
	return xxx_messageInfo_LoginReq.Size(m)
}
func (m *LoginReq) XXX_DiscardUnknown() {
	xxx_messageInfo_LoginReq.DiscardUnknown(m)
}

var xxx_messageInfo_LoginReq proto.InternalMessageInfo

func (m *LoginReq) GetHead() *common.ReqHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *LoginReq) GetLoginType() LoginType {
	if m != nil {
		return m.LoginType
	}
	return LoginType_NotUse1
}

func (m *LoginReq) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *LoginReq) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

func (m *LoginReq) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *LoginReq) GetDeviceType() string {
	if m != nil {
		return m.DeviceType
	}
	return ""
}

func (m *LoginReq) GetSex() string {
	if m != nil {
		return m.Sex
	}
	return ""
}

func (m *LoginReq) GetPlatformID() string {
	if m != nil {
		return m.PlatformID
	}
	return ""
}

func (m *LoginReq) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *LoginReq) GetProfile() string {
	if m != nil {
		return m.Profile
	}
	return ""
}

func (m *LoginReq) GetLongitude() float64 {
	if m != nil {
		return m.Longitude
	}
	return 0
}

func (m *LoginReq) GetLatitude() float64 {
	if m != nil {
		return m.Latitude
	}
	return 0
}

func (m *LoginReq) GetPlace() string {
	if m != nil {
		return m.Place
	}
	return ""
}

func (m *LoginReq) GetIsSimulator() string {
	if m != nil {
		return m.IsSimulator
	}
	return ""
}

func (m *LoginReq) GetNetworkType() string {
	if m != nil {
		return m.NetworkType
	}
	return ""
}

func (m *LoginReq) GetEnergy() string {
	if m != nil {
		return m.Energy
	}
	return ""
}

// s -> c
type LoginRsp struct {
	Head                 *common.RspHead  `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	Code                 LoginRspCode     `protobuf:"varint,2,opt,name=Code,proto3,enum=pblogin.LoginRspCode" json:"Code,omitempty"`
	StrCode              string           `protobuf:"bytes,3,opt,name=StrCode,proto3" json:"StrCode,omitempty"`
	User                 *common.UserInfo `protobuf:"bytes,4,opt,name=User,proto3" json:"User,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *LoginRsp) Reset()         { *m = LoginRsp{} }
func (m *LoginRsp) String() string { return proto.CompactTextString(m) }
func (*LoginRsp) ProtoMessage()    {}
func (*LoginRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_6fe61ab550dd3bc4, []int{1}
}

func (m *LoginRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LoginRsp.Unmarshal(m, b)
}
func (m *LoginRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LoginRsp.Marshal(b, m, deterministic)
}
func (m *LoginRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LoginRsp.Merge(m, src)
}
func (m *LoginRsp) XXX_Size() int {
	return xxx_messageInfo_LoginRsp.Size(m)
}
func (m *LoginRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_LoginRsp.DiscardUnknown(m)
}

var xxx_messageInfo_LoginRsp proto.InternalMessageInfo

func (m *LoginRsp) GetHead() *common.RspHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *LoginRsp) GetCode() LoginRspCode {
	if m != nil {
		return m.Code
	}
	return LoginRspCode_NotUse2
}

func (m *LoginRsp) GetStrCode() string {
	if m != nil {
		return m.StrCode
	}
	return ""
}

func (m *LoginRsp) GetUser() *common.UserInfo {
	if m != nil {
		return m.User
	}
	return nil
}

// c -> s
type KeepAliveReq struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KeepAliveReq) Reset()         { *m = KeepAliveReq{} }
func (m *KeepAliveReq) String() string { return proto.CompactTextString(m) }
func (*KeepAliveReq) ProtoMessage()    {}
func (*KeepAliveReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_6fe61ab550dd3bc4, []int{2}
}

func (m *KeepAliveReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeepAliveReq.Unmarshal(m, b)
}
func (m *KeepAliveReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeepAliveReq.Marshal(b, m, deterministic)
}
func (m *KeepAliveReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeepAliveReq.Merge(m, src)
}
func (m *KeepAliveReq) XXX_Size() int {
	return xxx_messageInfo_KeepAliveReq.Size(m)
}
func (m *KeepAliveReq) XXX_DiscardUnknown() {
	xxx_messageInfo_KeepAliveReq.DiscardUnknown(m)
}

var xxx_messageInfo_KeepAliveReq proto.InternalMessageInfo

type KeepAliveRsp struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KeepAliveRsp) Reset()         { *m = KeepAliveRsp{} }
func (m *KeepAliveRsp) String() string { return proto.CompactTextString(m) }
func (*KeepAliveRsp) ProtoMessage()    {}
func (*KeepAliveRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_6fe61ab550dd3bc4, []int{3}
}

func (m *KeepAliveRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeepAliveRsp.Unmarshal(m, b)
}
func (m *KeepAliveRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeepAliveRsp.Marshal(b, m, deterministic)
}
func (m *KeepAliveRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeepAliveRsp.Merge(m, src)
}
func (m *KeepAliveRsp) XXX_Size() int {
	return xxx_messageInfo_KeepAliveRsp.Size(m)
}
func (m *KeepAliveRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_KeepAliveRsp.DiscardUnknown(m)
}

var xxx_messageInfo_KeepAliveRsp proto.InternalMessageInfo

// 登陆时获取验证码 c -> s
type MobileCaptchaReq struct {
	Head                 *common.ReqHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	Mobile               string          `protobuf:"bytes,2,opt,name=Mobile,proto3" json:"Mobile,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *MobileCaptchaReq) Reset()         { *m = MobileCaptchaReq{} }
func (m *MobileCaptchaReq) String() string { return proto.CompactTextString(m) }
func (*MobileCaptchaReq) ProtoMessage()    {}
func (*MobileCaptchaReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_6fe61ab550dd3bc4, []int{4}
}

func (m *MobileCaptchaReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MobileCaptchaReq.Unmarshal(m, b)
}
func (m *MobileCaptchaReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MobileCaptchaReq.Marshal(b, m, deterministic)
}
func (m *MobileCaptchaReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MobileCaptchaReq.Merge(m, src)
}
func (m *MobileCaptchaReq) XXX_Size() int {
	return xxx_messageInfo_MobileCaptchaReq.Size(m)
}
func (m *MobileCaptchaReq) XXX_DiscardUnknown() {
	xxx_messageInfo_MobileCaptchaReq.DiscardUnknown(m)
}

var xxx_messageInfo_MobileCaptchaReq proto.InternalMessageInfo

func (m *MobileCaptchaReq) GetHead() *common.ReqHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *MobileCaptchaReq) GetMobile() string {
	if m != nil {
		return m.Mobile
	}
	return ""
}

type MobileCaptchaRsp struct {
	Head *common.RspHead `protobuf:"bytes,1,opt,name=Head,proto3" json:"Head,omitempty"`
	// 1 成功发送 2 参数错误 3 此号码没有绑定 4 被其他人绑定了 5 超过发送限制 6 内部服务错误 7 外部服务错误
	Code                 uint32   `protobuf:"varint,2,opt,name=Code,proto3" json:"Code,omitempty"`
	TestCaptcha          string   `protobuf:"bytes,3,opt,name=TestCaptcha,proto3" json:"TestCaptcha,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MobileCaptchaRsp) Reset()         { *m = MobileCaptchaRsp{} }
func (m *MobileCaptchaRsp) String() string { return proto.CompactTextString(m) }
func (*MobileCaptchaRsp) ProtoMessage()    {}
func (*MobileCaptchaRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_6fe61ab550dd3bc4, []int{5}
}

func (m *MobileCaptchaRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MobileCaptchaRsp.Unmarshal(m, b)
}
func (m *MobileCaptchaRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MobileCaptchaRsp.Marshal(b, m, deterministic)
}
func (m *MobileCaptchaRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MobileCaptchaRsp.Merge(m, src)
}
func (m *MobileCaptchaRsp) XXX_Size() int {
	return xxx_messageInfo_MobileCaptchaRsp.Size(m)
}
func (m *MobileCaptchaRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_MobileCaptchaRsp.DiscardUnknown(m)
}

var xxx_messageInfo_MobileCaptchaRsp proto.InternalMessageInfo

func (m *MobileCaptchaRsp) GetHead() *common.RspHead {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *MobileCaptchaRsp) GetCode() uint32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *MobileCaptchaRsp) GetTestCaptcha() string {
	if m != nil {
		return m.TestCaptcha
	}
	return ""
}

func init() {
	proto.RegisterEnum("pblogin.LoginType", LoginType_name, LoginType_value)
	proto.RegisterEnum("pblogin.LoginRspCode", LoginRspCode_name, LoginRspCode_value)
	proto.RegisterType((*LoginReq)(nil), "pblogin.LoginReq")
	proto.RegisterType((*LoginRsp)(nil), "pblogin.LoginRsp")
	proto.RegisterType((*KeepAliveReq)(nil), "pblogin.KeepAliveReq")
	proto.RegisterType((*KeepAliveRsp)(nil), "pblogin.KeepAliveRsp")
	proto.RegisterType((*MobileCaptchaReq)(nil), "pblogin.MobileCaptchaReq")
	proto.RegisterType((*MobileCaptchaRsp)(nil), "pblogin.MobileCaptchaRsp")
}

func init() { proto.RegisterFile("login/login.proto", fileDescriptor_6fe61ab550dd3bc4) }

var fileDescriptor_6fe61ab550dd3bc4 = []byte{
	// 590 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0xdd, 0x6e, 0xda, 0x4c,
	0x10, 0x8d, 0xc1, 0xfc, 0x0d, 0x84, 0x6f, 0x19, 0x7d, 0x8d, 0x56, 0x51, 0x14, 0x21, 0xa4, 0x56,
	0x34, 0x17, 0xd0, 0xa6, 0x4f, 0xd0, 0x86, 0x54, 0x45, 0xa5, 0xc4, 0x31, 0x49, 0x9b, 0xdb, 0xc5,
	0x9e, 0xc0, 0xaa, 0xb6, 0xd7, 0xb1, 0x9d, 0xa4, 0x79, 0x99, 0xbe, 0x55, 0xdf, 0xa7, 0xda, 0xb5,
	0x09, 0x4e, 0x2b, 0x55, 0xed, 0x0d, 0xbb, 0xe7, 0x9c, 0xf9, 0x61, 0xe6, 0x2c, 0x40, 0x2f, 0x50,
	0x2b, 0x19, 0x8d, 0xcd, 0xe7, 0x28, 0x4e, 0x54, 0xa6, 0xb0, 0x11, 0x2f, 0x0d, 0xdc, 0x3f, 0xf0,
	0x1e, 0xc6, 0x2b, 0x11, 0xd2, 0x38, 0x5e, 0x8e, 0x3d, 0x15, 0x86, 0x2a, 0x1a, 0xaf, 0x49, 0xf8,
	0x79, 0xd8, 0xfe, 0xe1, 0xef, 0x6a, 0x7e, 0xe4, 0xfa, 0xe0, 0x47, 0x15, 0x9a, 0x33, 0x5d, 0xc7,
	0xa5, 0x1b, 0x7c, 0x0e, 0xf6, 0x07, 0x12, 0x3e, 0xb7, 0xfa, 0xd6, 0xb0, 0x7d, 0xdc, 0x1b, 0xc5,
	0xcb, 0x22, 0xd6, 0xa5, 0x1b, 0x2d, 0xb8, 0x46, 0xc6, 0x57, 0xd0, 0x32, 0x29, 0x17, 0x0f, 0x31,
	0xf1, 0x4a, 0xdf, 0x1a, 0x76, 0x8f, 0x71, 0x54, 0x7c, 0x9d, 0xd1, 0xa3, 0xe2, 0x6e, 0x83, 0xb0,
	0x0b, 0x95, 0xe9, 0x84, 0x57, 0xfb, 0xd6, 0xb0, 0xe5, 0x56, 0xa6, 0x13, 0xdc, 0x87, 0xa6, 0x23,
	0xd2, 0xf4, 0x5e, 0x25, 0x3e, 0xb7, 0x0d, 0xfb, 0x88, 0x91, 0x43, 0xe3, 0x33, 0x25, 0xa9, 0x54,
	0x11, 0xaf, 0x19, 0x69, 0x03, 0xf1, 0x10, 0x60, 0x42, 0x77, 0xd2, 0x23, 0xd3, 0xb8, 0x6e, 0xc4,
	0x12, 0x83, 0x0c, 0xaa, 0x0b, 0xfa, 0xc6, 0x1b, 0x46, 0xd0, 0x57, 0x9d, 0xe1, 0x04, 0x22, 0xbb,
	0x56, 0x49, 0x38, 0x9d, 0xf0, 0x66, 0x9e, 0xb1, 0x65, 0x10, 0xc1, 0x9e, 0x8b, 0x90, 0x78, 0xcb,
	0x28, 0xe6, 0xae, 0xfb, 0x3b, 0x89, 0xba, 0x96, 0x01, 0x71, 0xc8, 0xfb, 0x17, 0x10, 0x0f, 0xf4,
	0xdc, 0xd1, 0x4a, 0x66, 0xb7, 0x3e, 0xf1, 0xdd, 0xbe, 0x35, 0xb4, 0xdc, 0x2d, 0xa1, 0x67, 0x9a,
	0x89, 0x2c, 0x17, 0xbb, 0x46, 0x7c, 0xc4, 0xf8, 0x3f, 0xd4, 0x9c, 0x40, 0x78, 0xc4, 0xff, 0x33,
	0x15, 0x73, 0x80, 0x7d, 0x68, 0x4f, 0xd3, 0x85, 0x0c, 0x6f, 0x03, 0x91, 0xa9, 0x84, 0x33, 0xa3,
	0x95, 0x29, 0x1d, 0x31, 0xa7, 0xec, 0x5e, 0x25, 0x5f, 0xcd, 0xc8, 0xbd, 0x3c, 0xa2, 0x44, 0xe1,
	0x1e, 0xd4, 0x4f, 0x23, 0x4a, 0x56, 0x0f, 0x1c, 0x8d, 0x58, 0xa0, 0xc1, 0x77, 0x6b, 0xe3, 0x6b,
	0x1a, 0xff, 0xc1, 0xd7, 0x34, 0x2e, 0xf9, 0xfa, 0x12, 0xec, 0x13, 0xe5, 0x6f, 0x2c, 0x7d, 0xf6,
	0xd4, 0x52, 0x37, 0x8d, 0xb5, 0xe8, 0x9a, 0x10, 0xbd, 0xa4, 0x45, 0x96, 0x98, 0xe8, 0xdc, 0xd5,
	0x0d, 0xc4, 0x17, 0x60, 0x5f, 0xa6, 0x94, 0x18, 0x5b, 0xdb, 0xe6, 0x5d, 0x14, 0xbd, 0x34, 0x3b,
	0x8d, 0xae, 0x95, 0x6b, 0xf4, 0x41, 0x17, 0x3a, 0x1f, 0x89, 0xe2, 0xb7, 0x81, 0xbc, 0x23, 0x97,
	0x6e, 0x9e, 0xe2, 0x34, 0x1e, 0x9c, 0x03, 0xfb, 0xa4, 0x96, 0x32, 0xa0, 0x13, 0x11, 0x67, 0xde,
	0x5a, 0xfc, 0xc3, 0xfb, 0xdc, 0x83, 0x7a, 0x9e, 0x6a, 0x26, 0x69, 0xb9, 0x05, 0x1a, 0xa8, 0x5f,
	0x4b, 0xfe, 0xfd, 0x6a, 0xb0, 0xb4, 0x9a, 0xdd, 0x62, 0x07, 0x7d, 0x68, 0x5f, 0x50, 0x9a, 0x15,
	0xc5, 0x8a, 0x3d, 0x94, 0xa9, 0xa3, 0x59, 0xe9, 0x87, 0x82, 0x6d, 0x68, 0xcc, 0x55, 0x76, 0x99,
	0xd2, 0x6b, 0xb6, 0x83, 0x75, 0xa8, 0x7c, 0xb9, 0x62, 0x96, 0x3e, 0xcf, 0xcf, 0x59, 0x05, 0x5b,
	0x50, 0x73, 0xd6, 0x2a, 0x22, 0x56, 0xd5, 0xd7, 0xd3, 0x50, 0xc8, 0x80, 0xd9, 0xd8, 0x81, 0xe6,
	0x95, 0x14, 0xd1, 0x4c, 0x0a, 0xc5, 0x6a, 0x47, 0x1e, 0x74, 0xca, 0x4e, 0x6c, 0x0b, 0x1e, 0xb3,
	0x1d, 0xec, 0xc1, 0xee, 0xd4, 0x3f, 0x4b, 0x9c, 0x7b, 0xff, 0xbd, 0x90, 0x01, 0xf9, 0xcc, 0xc2,
	0x26, 0xd8, 0x8b, 0x5b, 0xcf, 0xcb, 0xab, 0x9f, 0x65, 0x6b, 0x4a, 0x58, 0x15, 0x19, 0x74, 0xf2,
	0x1d, 0xcc, 0xd5, 0x3b, 0x19, 0xf9, 0x45, 0x93, 0x59, 0x81, 0x6a, 0xcb, 0xba, 0xf9, 0x5b, 0x78,
	0xf3, 0x33, 0x00, 0x00, 0xff, 0xff, 0x88, 0x96, 0x49, 0x00, 0x72, 0x04, 0x00, 0x00,
}
