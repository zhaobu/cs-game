package game

import (
	"net"

	fbsCommon "mahjong.go/fbs/Common"
	"mahjong.go/library/core"
	"mahjong.go/mi/protocal"
	configService "mahjong.go/service/config"
	gameService "mahjong.go/service/game"
)

// 回应准备
func RoomChatAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	// 解析参数
	request := fbsCommon.GetRootAsRoomChatNotify(impacket.GetBody(), 0)
	chatId := request.ChatId()
	memberId := request.MemberId()
	content := string(request.Content())

	// 消息id，由客户端定义，服务端只做转发
	var err *core.Error
	if configService.IsNoticeId(chatId) {
		// 消息
		err = gameService.RoomNotice(userId, chatId, memberId, content)
	} else {
		err = gameService.RoomChat(userId, chatId, memberId, content)
	}

	if err != nil {
		core.Logger.Error(err.Error())
	}
}
