package message

import (
	"github.com/fwhappy/util"
	"mahjong.club/config"
	"mahjong.club/user"
)

// Msg 消息
type Msg struct {
	MID        uint64     // 消息id
	MType      int        // 消息类型
	Content    string     // 消息内容
	Sender     *MsgSender // 消息发送者
	CreateTime int64      // 消息创建时间
}

// MsgSender 消息发送者
type MsgSender struct {
	Info *user.Info
}

// NewMsg 新建一条消息
func NewMsg() *Msg {
	return &Msg{
		CreateTime: util.GetTime(),
	}
}

// NewSender 新建一个消息发送者
func NewSender(ID int) *MsgSender {
	s := &MsgSender{
		Info: user.NewInfo(ID),
	}
	return s
}

// IsTimeout 消息是否已过期
func (m *Msg) IsTimeout() bool {
	return util.GetTime()-m.CreateTime > config.CLUB_MESSAGE_TIMEOUT
}
