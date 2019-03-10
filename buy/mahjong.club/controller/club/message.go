package club

import (
	"net"

	"mahjong.club/core"
	"github.com/fwhappy/mahjong/protocal"
	cs "mahjong.club/service/club"
)

// SendMessageAction 俱乐部消息发送
func SendMessageAction(userID int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	if err := cs.SendMessage(userID, impacket); err != nil {
		core.Logger.Error("[club.SendMessageAction]userID:%v,code:%v,msg:%v", userID, err.GetCode(), err.Error())
	}
}

// MessageListAction 俱乐部消息列表
func MessageListAction(userID int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	if err := cs.MessageList(userID, impacket); err != nil {
		core.Logger.Error("[club.MessageListAction]userID:%v,code:%v,msg:%v", userID, err.GetCode(), err.Error())
	}
}
