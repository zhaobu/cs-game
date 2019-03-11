package def

type OpKind int

const (
	OpLoginRsp OpKind = iota + 1
	OpQueryGroupReq
	OpQueryGroupRsp
	OpJoinGroupReq
	OpJoinGroupRsp
	OpExitGroupReq
	OpExitGroupRsp
	OpSendGroupMsgReq
	OpSendGroupMsgRsp
	OpSendOneMsgReq
	OpSendOneMsgRsp
	OpMsgNotify // 消息通知
)

type Msg struct {
	Kind OpKind
}

type LoginRsp struct {
	Kind OpKind
	Name string
}

type QueryGroupReq struct {
	Kind OpKind
}

type QueryGroupRsp struct {
	Kind  OpKind
	Infos []string
}

type JoinGroupReq struct {
	Kind      OpKind
	GroupName string
}

type JoinGroupRsp struct {
	Kind OpKind
	Msg  string
}

type ExitGroupReq struct {
	Kind      OpKind
	GroupName string
}

type ExitGroupRsp struct {
	Kind OpKind
	Msg  string
}

type SendGroupMsgReq struct {
	Kind    OpKind
	Seq     uint64
	ToGroup string
	Content []byte
}

type SendGroupMsgRsp struct {
	Kind OpKind
	Msg  string
}

type MsgNotify struct {
	Kind    OpKind
	Seq     uint64
	From    string
	To      string
	Content []byte
}
