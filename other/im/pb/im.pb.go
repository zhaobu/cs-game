// Code generated by protoc-gen-go. DO NOT EDIT.
// source: im.proto

package impb

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

type ContentType int32

const (
	ContentType_Text  ContentType = 0
	ContentType_Pic   ContentType = 1
	ContentType_Sound ContentType = 2
	ContentType_Video ContentType = 3
)

var ContentType_name = map[int32]string{
	0: "Text",
	1: "Pic",
	2: "Sound",
	3: "Video",
}

var ContentType_value = map[string]int32{
	"Text":  0,
	"Pic":   1,
	"Sound": 2,
	"Video": 3,
}

func (x ContentType) String() string {
	return proto.EnumName(ContentType_name, int32(x))
}

func (ContentType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_36f2114a3e4ddb9e, []int{0}
}

// c -> s
type LoginReq struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LoginReq) Reset()         { *m = LoginReq{} }
func (m *LoginReq) String() string { return proto.CompactTextString(m) }
func (*LoginReq) ProtoMessage()    {}
func (*LoginReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_36f2114a3e4ddb9e, []int{0}
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

// s -> c
type LoginRsp struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LoginRsp) Reset()         { *m = LoginRsp{} }
func (m *LoginRsp) String() string { return proto.CompactTextString(m) }
func (*LoginRsp) ProtoMessage()    {}
func (*LoginRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_36f2114a3e4ddb9e, []int{1}
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

// c -> s
type SendMsgReq struct {
	Seq                  uint64      `protobuf:"varint,1,opt,name=Seq,proto3" json:"Seq,omitempty"`
	To                   uint64      `protobuf:"varint,2,opt,name=To,proto3" json:"To,omitempty"`
	From                 uint64      `protobuf:"varint,3,opt,name=From,proto3" json:"From,omitempty"`
	Content              []byte      `protobuf:"bytes,4,opt,name=Content,proto3" json:"Content,omitempty"`
	Ct                   ContentType `protobuf:"varint,5,opt,name=Ct,proto3,enum=impb.ContentType" json:"Ct,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *SendMsgReq) Reset()         { *m = SendMsgReq{} }
func (m *SendMsgReq) String() string { return proto.CompactTextString(m) }
func (*SendMsgReq) ProtoMessage()    {}
func (*SendMsgReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_36f2114a3e4ddb9e, []int{2}
}

func (m *SendMsgReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SendMsgReq.Unmarshal(m, b)
}
func (m *SendMsgReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SendMsgReq.Marshal(b, m, deterministic)
}
func (m *SendMsgReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SendMsgReq.Merge(m, src)
}
func (m *SendMsgReq) XXX_Size() int {
	return xxx_messageInfo_SendMsgReq.Size(m)
}
func (m *SendMsgReq) XXX_DiscardUnknown() {
	xxx_messageInfo_SendMsgReq.DiscardUnknown(m)
}

var xxx_messageInfo_SendMsgReq proto.InternalMessageInfo

func (m *SendMsgReq) GetSeq() uint64 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func (m *SendMsgReq) GetTo() uint64 {
	if m != nil {
		return m.To
	}
	return 0
}

func (m *SendMsgReq) GetFrom() uint64 {
	if m != nil {
		return m.From
	}
	return 0
}

func (m *SendMsgReq) GetContent() []byte {
	if m != nil {
		return m.Content
	}
	return nil
}

func (m *SendMsgReq) GetCt() ContentType {
	if m != nil {
		return m.Ct
	}
	return ContentType_Text
}

// s -> c
type SendMsgRsp struct {
	Seq                  uint64   `protobuf:"varint,1,opt,name=Seq,proto3" json:"Seq,omitempty"`
	Code                 uint32   `protobuf:"varint,2,opt,name=Code,proto3" json:"Code,omitempty"`
	MsgID                string   `protobuf:"bytes,3,opt,name=MsgID,proto3" json:"MsgID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SendMsgRsp) Reset()         { *m = SendMsgRsp{} }
func (m *SendMsgRsp) String() string { return proto.CompactTextString(m) }
func (*SendMsgRsp) ProtoMessage()    {}
func (*SendMsgRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_36f2114a3e4ddb9e, []int{3}
}

func (m *SendMsgRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SendMsgRsp.Unmarshal(m, b)
}
func (m *SendMsgRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SendMsgRsp.Marshal(b, m, deterministic)
}
func (m *SendMsgRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SendMsgRsp.Merge(m, src)
}
func (m *SendMsgRsp) XXX_Size() int {
	return xxx_messageInfo_SendMsgRsp.Size(m)
}
func (m *SendMsgRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_SendMsgRsp.DiscardUnknown(m)
}

var xxx_messageInfo_SendMsgRsp proto.InternalMessageInfo

func (m *SendMsgRsp) GetSeq() uint64 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func (m *SendMsgRsp) GetCode() uint32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *SendMsgRsp) GetMsgID() string {
	if m != nil {
		return m.MsgID
	}
	return ""
}

// s -> c
type MsgNotify struct {
	MsgID                string      `protobuf:"bytes,1,opt,name=MsgID,proto3" json:"MsgID,omitempty"`
	To                   uint64      `protobuf:"varint,2,opt,name=To,proto3" json:"To,omitempty"`
	From                 uint64      `protobuf:"varint,3,opt,name=From,proto3" json:"From,omitempty"`
	Content              []byte      `protobuf:"bytes,4,opt,name=Content,proto3" json:"Content,omitempty"`
	SendTime             int64       `protobuf:"varint,5,opt,name=SendTime,proto3" json:"SendTime,omitempty"`
	Ct                   ContentType `protobuf:"varint,6,opt,name=Ct,proto3,enum=impb.ContentType" json:"Ct,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *MsgNotify) Reset()         { *m = MsgNotify{} }
func (m *MsgNotify) String() string { return proto.CompactTextString(m) }
func (*MsgNotify) ProtoMessage()    {}
func (*MsgNotify) Descriptor() ([]byte, []int) {
	return fileDescriptor_36f2114a3e4ddb9e, []int{4}
}

func (m *MsgNotify) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MsgNotify.Unmarshal(m, b)
}
func (m *MsgNotify) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MsgNotify.Marshal(b, m, deterministic)
}
func (m *MsgNotify) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgNotify.Merge(m, src)
}
func (m *MsgNotify) XXX_Size() int {
	return xxx_messageInfo_MsgNotify.Size(m)
}
func (m *MsgNotify) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgNotify.DiscardUnknown(m)
}

var xxx_messageInfo_MsgNotify proto.InternalMessageInfo

func (m *MsgNotify) GetMsgID() string {
	if m != nil {
		return m.MsgID
	}
	return ""
}

func (m *MsgNotify) GetTo() uint64 {
	if m != nil {
		return m.To
	}
	return 0
}

func (m *MsgNotify) GetFrom() uint64 {
	if m != nil {
		return m.From
	}
	return 0
}

func (m *MsgNotify) GetContent() []byte {
	if m != nil {
		return m.Content
	}
	return nil
}

func (m *MsgNotify) GetSendTime() int64 {
	if m != nil {
		return m.SendTime
	}
	return 0
}

func (m *MsgNotify) GetCt() ContentType {
	if m != nil {
		return m.Ct
	}
	return ContentType_Text
}

// c -> s
type MsgNotifyAck struct {
	MsgIDs               []string `protobuf:"bytes,1,rep,name=MsgIDs,proto3" json:"MsgIDs,omitempty"`
	OtherUID             uint64   `protobuf:"varint,2,opt,name=OtherUID,proto3" json:"OtherUID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MsgNotifyAck) Reset()         { *m = MsgNotifyAck{} }
func (m *MsgNotifyAck) String() string { return proto.CompactTextString(m) }
func (*MsgNotifyAck) ProtoMessage()    {}
func (*MsgNotifyAck) Descriptor() ([]byte, []int) {
	return fileDescriptor_36f2114a3e4ddb9e, []int{5}
}

func (m *MsgNotifyAck) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MsgNotifyAck.Unmarshal(m, b)
}
func (m *MsgNotifyAck) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MsgNotifyAck.Marshal(b, m, deterministic)
}
func (m *MsgNotifyAck) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgNotifyAck.Merge(m, src)
}
func (m *MsgNotifyAck) XXX_Size() int {
	return xxx_messageInfo_MsgNotifyAck.Size(m)
}
func (m *MsgNotifyAck) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgNotifyAck.DiscardUnknown(m)
}

var xxx_messageInfo_MsgNotifyAck proto.InternalMessageInfo

func (m *MsgNotifyAck) GetMsgIDs() []string {
	if m != nil {
		return m.MsgIDs
	}
	return nil
}

func (m *MsgNotifyAck) GetOtherUID() uint64 {
	if m != nil {
		return m.OtherUID
	}
	return 0
}

// c -> s
type MsgRecordReq struct {
	Seq                  uint64   `protobuf:"varint,1,opt,name=Seq,proto3" json:"Seq,omitempty"`
	To                   uint64   `protobuf:"varint,2,opt,name=To,proto3" json:"To,omitempty"`
	From                 uint64   `protobuf:"varint,3,opt,name=From,proto3" json:"From,omitempty"`
	StartMsgID           string   `protobuf:"bytes,4,opt,name=StartMsgID,proto3" json:"StartMsgID,omitempty"`
	EndMsgID             string   `protobuf:"bytes,5,opt,name=EndMsgID,proto3" json:"EndMsgID,omitempty"`
	Limit                int32    `protobuf:"varint,6,opt,name=limit,proto3" json:"limit,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MsgRecordReq) Reset()         { *m = MsgRecordReq{} }
func (m *MsgRecordReq) String() string { return proto.CompactTextString(m) }
func (*MsgRecordReq) ProtoMessage()    {}
func (*MsgRecordReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_36f2114a3e4ddb9e, []int{6}
}

func (m *MsgRecordReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MsgRecordReq.Unmarshal(m, b)
}
func (m *MsgRecordReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MsgRecordReq.Marshal(b, m, deterministic)
}
func (m *MsgRecordReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRecordReq.Merge(m, src)
}
func (m *MsgRecordReq) XXX_Size() int {
	return xxx_messageInfo_MsgRecordReq.Size(m)
}
func (m *MsgRecordReq) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRecordReq.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRecordReq proto.InternalMessageInfo

func (m *MsgRecordReq) GetSeq() uint64 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func (m *MsgRecordReq) GetTo() uint64 {
	if m != nil {
		return m.To
	}
	return 0
}

func (m *MsgRecordReq) GetFrom() uint64 {
	if m != nil {
		return m.From
	}
	return 0
}

func (m *MsgRecordReq) GetStartMsgID() string {
	if m != nil {
		return m.StartMsgID
	}
	return ""
}

func (m *MsgRecordReq) GetEndMsgID() string {
	if m != nil {
		return m.EndMsgID
	}
	return ""
}

func (m *MsgRecordReq) GetLimit() int32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

// s -> c
type MsgRecordRsp struct {
	Seq                  uint64       `protobuf:"varint,1,opt,name=Seq,proto3" json:"Seq,omitempty"`
	Msgs                 []*MsgNotify `protobuf:"bytes,2,rep,name=Msgs,proto3" json:"Msgs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *MsgRecordRsp) Reset()         { *m = MsgRecordRsp{} }
func (m *MsgRecordRsp) String() string { return proto.CompactTextString(m) }
func (*MsgRecordRsp) ProtoMessage()    {}
func (*MsgRecordRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_36f2114a3e4ddb9e, []int{7}
}

func (m *MsgRecordRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MsgRecordRsp.Unmarshal(m, b)
}
func (m *MsgRecordRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MsgRecordRsp.Marshal(b, m, deterministic)
}
func (m *MsgRecordRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRecordRsp.Merge(m, src)
}
func (m *MsgRecordRsp) XXX_Size() int {
	return xxx_messageInfo_MsgRecordRsp.Size(m)
}
func (m *MsgRecordRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRecordRsp.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRecordRsp proto.InternalMessageInfo

func (m *MsgRecordRsp) GetSeq() uint64 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func (m *MsgRecordRsp) GetMsgs() []*MsgNotify {
	if m != nil {
		return m.Msgs
	}
	return nil
}

// c -> s
type QueryUnreadCntReq struct {
	Seq                  uint64   `protobuf:"varint,1,opt,name=Seq,proto3" json:"Seq,omitempty"`
	UID                  uint64   `protobuf:"varint,2,opt,name=UID,proto3" json:"UID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *QueryUnreadCntReq) Reset()         { *m = QueryUnreadCntReq{} }
func (m *QueryUnreadCntReq) String() string { return proto.CompactTextString(m) }
func (*QueryUnreadCntReq) ProtoMessage()    {}
func (*QueryUnreadCntReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_36f2114a3e4ddb9e, []int{8}
}

func (m *QueryUnreadCntReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryUnreadCntReq.Unmarshal(m, b)
}
func (m *QueryUnreadCntReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryUnreadCntReq.Marshal(b, m, deterministic)
}
func (m *QueryUnreadCntReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryUnreadCntReq.Merge(m, src)
}
func (m *QueryUnreadCntReq) XXX_Size() int {
	return xxx_messageInfo_QueryUnreadCntReq.Size(m)
}
func (m *QueryUnreadCntReq) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryUnreadCntReq.DiscardUnknown(m)
}

var xxx_messageInfo_QueryUnreadCntReq proto.InternalMessageInfo

func (m *QueryUnreadCntReq) GetSeq() uint64 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func (m *QueryUnreadCntReq) GetUID() uint64 {
	if m != nil {
		return m.UID
	}
	return 0
}

// s -> c
type QueryUnreadCntRsp struct {
	Seq                  uint64           `protobuf:"varint,1,opt,name=Seq,proto3" json:"Seq,omitempty"`
	Cnt                  map[uint64]int64 `protobuf:"bytes,2,rep,name=Cnt,proto3" json:"Cnt,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *QueryUnreadCntRsp) Reset()         { *m = QueryUnreadCntRsp{} }
func (m *QueryUnreadCntRsp) String() string { return proto.CompactTextString(m) }
func (*QueryUnreadCntRsp) ProtoMessage()    {}
func (*QueryUnreadCntRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_36f2114a3e4ddb9e, []int{9}
}

func (m *QueryUnreadCntRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryUnreadCntRsp.Unmarshal(m, b)
}
func (m *QueryUnreadCntRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryUnreadCntRsp.Marshal(b, m, deterministic)
}
func (m *QueryUnreadCntRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryUnreadCntRsp.Merge(m, src)
}
func (m *QueryUnreadCntRsp) XXX_Size() int {
	return xxx_messageInfo_QueryUnreadCntRsp.Size(m)
}
func (m *QueryUnreadCntRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryUnreadCntRsp.DiscardUnknown(m)
}

var xxx_messageInfo_QueryUnreadCntRsp proto.InternalMessageInfo

func (m *QueryUnreadCntRsp) GetSeq() uint64 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func (m *QueryUnreadCntRsp) GetCnt() map[uint64]int64 {
	if m != nil {
		return m.Cnt
	}
	return nil
}

// c -> s
type QueryUnreadNReq struct {
	Seq                  uint64   `protobuf:"varint,1,opt,name=Seq,proto3" json:"Seq,omitempty"`
	UID                  uint64   `protobuf:"varint,2,opt,name=UID,proto3" json:"UID,omitempty"`
	LastN                int64    `protobuf:"varint,3,opt,name=LastN,proto3" json:"LastN,omitempty"`
	OtherUID             uint64   `protobuf:"varint,4,opt,name=OtherUID,proto3" json:"OtherUID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *QueryUnreadNReq) Reset()         { *m = QueryUnreadNReq{} }
func (m *QueryUnreadNReq) String() string { return proto.CompactTextString(m) }
func (*QueryUnreadNReq) ProtoMessage()    {}
func (*QueryUnreadNReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_36f2114a3e4ddb9e, []int{10}
}

func (m *QueryUnreadNReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryUnreadNReq.Unmarshal(m, b)
}
func (m *QueryUnreadNReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryUnreadNReq.Marshal(b, m, deterministic)
}
func (m *QueryUnreadNReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryUnreadNReq.Merge(m, src)
}
func (m *QueryUnreadNReq) XXX_Size() int {
	return xxx_messageInfo_QueryUnreadNReq.Size(m)
}
func (m *QueryUnreadNReq) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryUnreadNReq.DiscardUnknown(m)
}

var xxx_messageInfo_QueryUnreadNReq proto.InternalMessageInfo

func (m *QueryUnreadNReq) GetSeq() uint64 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func (m *QueryUnreadNReq) GetUID() uint64 {
	if m != nil {
		return m.UID
	}
	return 0
}

func (m *QueryUnreadNReq) GetLastN() int64 {
	if m != nil {
		return m.LastN
	}
	return 0
}

func (m *QueryUnreadNReq) GetOtherUID() uint64 {
	if m != nil {
		return m.OtherUID
	}
	return 0
}

// s -> c
type QueryUnreadNRsp struct {
	Seq                  uint64       `protobuf:"varint,1,opt,name=Seq,proto3" json:"Seq,omitempty"`
	Msgs                 []*MsgNotify `protobuf:"bytes,2,rep,name=Msgs,proto3" json:"Msgs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *QueryUnreadNRsp) Reset()         { *m = QueryUnreadNRsp{} }
func (m *QueryUnreadNRsp) String() string { return proto.CompactTextString(m) }
func (*QueryUnreadNRsp) ProtoMessage()    {}
func (*QueryUnreadNRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_36f2114a3e4ddb9e, []int{11}
}

func (m *QueryUnreadNRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryUnreadNRsp.Unmarshal(m, b)
}
func (m *QueryUnreadNRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryUnreadNRsp.Marshal(b, m, deterministic)
}
func (m *QueryUnreadNRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryUnreadNRsp.Merge(m, src)
}
func (m *QueryUnreadNRsp) XXX_Size() int {
	return xxx_messageInfo_QueryUnreadNRsp.Size(m)
}
func (m *QueryUnreadNRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryUnreadNRsp.DiscardUnknown(m)
}

var xxx_messageInfo_QueryUnreadNRsp proto.InternalMessageInfo

func (m *QueryUnreadNRsp) GetSeq() uint64 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func (m *QueryUnreadNRsp) GetMsgs() []*MsgNotify {
	if m != nil {
		return m.Msgs
	}
	return nil
}

// c -> s
type EnterExitRoom struct {
	UID                  uint64   `protobuf:"varint,1,opt,name=UID,proto3" json:"UID,omitempty"`
	RoomID               uint64   `protobuf:"varint,2,opt,name=RoomID,proto3" json:"RoomID,omitempty"`
	EnterOrExit          int32    `protobuf:"varint,3,opt,name=EnterOrExit,proto3" json:"EnterOrExit,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *EnterExitRoom) Reset()         { *m = EnterExitRoom{} }
func (m *EnterExitRoom) String() string { return proto.CompactTextString(m) }
func (*EnterExitRoom) ProtoMessage()    {}
func (*EnterExitRoom) Descriptor() ([]byte, []int) {
	return fileDescriptor_36f2114a3e4ddb9e, []int{12}
}

func (m *EnterExitRoom) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EnterExitRoom.Unmarshal(m, b)
}
func (m *EnterExitRoom) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EnterExitRoom.Marshal(b, m, deterministic)
}
func (m *EnterExitRoom) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EnterExitRoom.Merge(m, src)
}
func (m *EnterExitRoom) XXX_Size() int {
	return xxx_messageInfo_EnterExitRoom.Size(m)
}
func (m *EnterExitRoom) XXX_DiscardUnknown() {
	xxx_messageInfo_EnterExitRoom.DiscardUnknown(m)
}

var xxx_messageInfo_EnterExitRoom proto.InternalMessageInfo

func (m *EnterExitRoom) GetUID() uint64 {
	if m != nil {
		return m.UID
	}
	return 0
}

func (m *EnterExitRoom) GetRoomID() uint64 {
	if m != nil {
		return m.RoomID
	}
	return 0
}

func (m *EnterExitRoom) GetEnterOrExit() int32 {
	if m != nil {
		return m.EnterOrExit
	}
	return 0
}

func init() {
	proto.RegisterEnum("impb.ContentType", ContentType_name, ContentType_value)
	proto.RegisterType((*LoginReq)(nil), "impb.LoginReq")
	proto.RegisterType((*LoginRsp)(nil), "impb.LoginRsp")
	proto.RegisterType((*SendMsgReq)(nil), "impb.SendMsgReq")
	proto.RegisterType((*SendMsgRsp)(nil), "impb.SendMsgRsp")
	proto.RegisterType((*MsgNotify)(nil), "impb.MsgNotify")
	proto.RegisterType((*MsgNotifyAck)(nil), "impb.MsgNotifyAck")
	proto.RegisterType((*MsgRecordReq)(nil), "impb.MsgRecordReq")
	proto.RegisterType((*MsgRecordRsp)(nil), "impb.MsgRecordRsp")
	proto.RegisterType((*QueryUnreadCntReq)(nil), "impb.QueryUnreadCntReq")
	proto.RegisterType((*QueryUnreadCntRsp)(nil), "impb.QueryUnreadCntRsp")
	proto.RegisterMapType((map[uint64]int64)(nil), "impb.QueryUnreadCntRsp.CntEntry")
	proto.RegisterType((*QueryUnreadNReq)(nil), "impb.QueryUnreadNReq")
	proto.RegisterType((*QueryUnreadNRsp)(nil), "impb.QueryUnreadNRsp")
	proto.RegisterType((*EnterExitRoom)(nil), "impb.EnterExitRoom")
}

func init() { proto.RegisterFile("im.proto", fileDescriptor_36f2114a3e4ddb9e) }

var fileDescriptor_36f2114a3e4ddb9e = []byte{
	// 541 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x94, 0xcf, 0x6f, 0xd3, 0x30,
	0x14, 0xc7, 0x71, 0x9c, 0x76, 0xed, 0xeb, 0x7e, 0x64, 0x16, 0x9a, 0xa2, 0x1d, 0x50, 0x08, 0x97,
	0x8a, 0x43, 0x0f, 0x45, 0x1a, 0x88, 0x1b, 0xa4, 0x45, 0xab, 0xb4, 0x76, 0xe0, 0x76, 0x5c, 0x38,
	0x75, 0x8d, 0x09, 0x56, 0x17, 0xbb, 0x4d, 0x5c, 0xb4, 0x9c, 0xf8, 0x1b, 0xb8, 0x70, 0xe4, 0x6f,
	0x45, 0x76, 0xd2, 0x24, 0x6c, 0x15, 0x42, 0xdb, 0xed, 0x7d, 0x9f, 0xed, 0x97, 0xcf, 0xf7, 0xbd,
	0xd7, 0x42, 0x8b, 0xc7, 0xbd, 0x55, 0x22, 0x95, 0x24, 0x36, 0x8f, 0x57, 0xd7, 0x3e, 0x40, 0xeb,
	0x42, 0x46, 0x5c, 0x50, 0xb6, 0xae, 0xe2, 0x74, 0xe5, 0xff, 0x00, 0x98, 0x32, 0x11, 0x8e, 0xd3,
	0x88, 0xb2, 0x35, 0x71, 0x00, 0x4f, 0xd9, 0xda, 0x45, 0x1e, 0xea, 0xda, 0x54, 0x87, 0xe4, 0x10,
	0xac, 0x99, 0x74, 0x2d, 0x93, 0xb0, 0x66, 0x92, 0x10, 0xb0, 0x3f, 0x24, 0x32, 0x76, 0xb1, 0xc9,
	0x98, 0x98, 0xb8, 0xb0, 0x17, 0x48, 0xa1, 0x98, 0x50, 0xae, 0xed, 0xa1, 0xee, 0x3e, 0xdd, 0x4a,
	0xf2, 0x1c, 0xac, 0x40, 0xb9, 0x0d, 0x0f, 0x75, 0x0f, 0xfb, 0xc7, 0x3d, 0x0d, 0xd2, 0x2b, 0x8e,
	0x66, 0xd9, 0x8a, 0x51, 0x2b, 0x50, 0xfe, 0x79, 0x05, 0x90, 0xae, 0x76, 0x00, 0x10, 0xb0, 0x03,
	0x19, 0x32, 0x83, 0x70, 0x40, 0x4d, 0x4c, 0x9e, 0x42, 0x63, 0x9c, 0x46, 0xa3, 0x81, 0xa1, 0x68,
	0xd3, 0x5c, 0xf8, 0xbf, 0x11, 0xb4, 0xc7, 0x69, 0x34, 0x91, 0x8a, 0x7f, 0xcd, 0xaa, 0x3b, 0xa8,
	0x76, 0xe7, 0x91, 0x76, 0x4e, 0xa1, 0xa5, 0x59, 0x67, 0x3c, 0x66, 0xc6, 0x14, 0xa6, 0xa5, 0x2e,
	0xac, 0x36, 0xff, 0x65, 0xf5, 0x3d, 0xec, 0x97, 0x7c, 0xef, 0x16, 0x4b, 0x72, 0x02, 0x4d, 0x43,
	0x95, 0xba, 0xc8, 0xc3, 0xdd, 0x36, 0x2d, 0x94, 0xfe, 0xcc, 0xa5, 0xfa, 0xc6, 0x92, 0xab, 0xd1,
	0xa0, 0x40, 0x2d, 0xb5, 0xff, 0x0b, 0x99, 0x22, 0x94, 0x2d, 0x64, 0x12, 0x3e, 0x7c, 0x64, 0xcf,
	0x00, 0xa6, 0x6a, 0x9e, 0xa8, 0xbc, 0x45, 0xb6, 0x69, 0x51, 0x2d, 0xa3, 0x11, 0x86, 0x66, 0x28,
	0xa3, 0x81, 0x71, 0xda, 0xa6, 0xa5, 0xd6, 0x9d, 0xbd, 0xe1, 0x31, 0xcf, 0xcd, 0x36, 0x68, 0x2e,
	0xfc, 0x61, 0x9d, 0x6b, 0xe7, 0x24, 0x5f, 0x80, 0x3d, 0x4e, 0xa3, 0xd4, 0xb5, 0x3c, 0xdc, 0xed,
	0xf4, 0x8f, 0xf2, 0x1e, 0x95, 0x0d, 0xa1, 0xe6, 0xd0, 0x7f, 0x0d, 0xc7, 0x9f, 0x36, 0x2c, 0xc9,
	0xae, 0x44, 0xc2, 0xe6, 0x61, 0x20, 0xd4, 0x6e, 0x8f, 0x0e, 0xe0, 0xaa, 0x3b, 0x3a, 0xf4, 0x7f,
	0xa2, 0x7b, 0x2f, 0x77, 0x52, 0xf4, 0x01, 0x07, 0x42, 0x15, 0x10, 0x5e, 0x0e, 0x71, 0xef, 0x5d,
	0x2f, 0x10, 0x6a, 0x28, 0x54, 0x92, 0x51, 0x7d, 0xf9, 0xf4, 0x0c, 0x5a, 0xdb, 0x84, 0xae, 0xb8,
	0x64, 0xd9, 0xb6, 0xe2, 0x92, 0x99, 0x4d, 0xfb, 0x3e, 0xbf, 0xd9, 0xe4, 0x2b, 0x8a, 0x69, 0x2e,
	0xde, 0x5a, 0x6f, 0x90, 0x1f, 0xc1, 0x51, 0xad, 0xf4, 0xe4, 0x3f, 0xad, 0xe8, 0x82, 0x17, 0xf3,
	0x54, 0x4d, 0xcc, 0xc4, 0x30, 0xcd, 0xc5, 0x5f, 0x5b, 0x61, 0xdf, 0xd9, 0x8a, 0xf3, 0x3b, 0x1f,
	0x7a, 0x78, 0xff, 0xbf, 0xc0, 0xc1, 0x50, 0x28, 0x96, 0x0c, 0x6f, 0xb9, 0xa2, 0x52, 0xc6, 0x5b,
	0x3c, 0x54, 0xe1, 0x9d, 0x40, 0x53, 0x9f, 0x94, 0xcc, 0x85, 0x22, 0x1e, 0x74, 0xcc, 0xd3, 0x4b,
	0xf3, 0xd8, 0xc0, 0x37, 0x68, 0x3d, 0xf5, 0xf2, 0x0c, 0x3a, 0xb5, 0xdf, 0x04, 0x69, 0x81, 0x3d,
	0x63, 0xb7, 0xca, 0x79, 0x42, 0xf6, 0x00, 0x7f, 0xe4, 0x0b, 0x07, 0x91, 0x36, 0x34, 0xa6, 0x72,
	0x23, 0x42, 0xc7, 0xd2, 0xe1, 0x67, 0x1e, 0x32, 0xe9, 0xe0, 0xeb, 0xa6, 0xf9, 0x27, 0x7b, 0xf5,
	0x27, 0x00, 0x00, 0xff, 0xff, 0x5c, 0x9b, 0x27, 0x08, 0xd5, 0x04, 0x00, 0x00,
}
