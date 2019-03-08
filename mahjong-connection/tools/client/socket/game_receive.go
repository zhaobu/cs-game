package main

import (
	fbsCommon "mahjong-connection/fbs/Common"
	"mahjong-connection/protocal"
)

// 收到加入游戏服激活
func s2cGameActivateResponse(impacket *protocal.ImPacket) {
	response := fbsCommon.GetRootAsGameActivateResponse(impacket.GetBody(), 0)
	s2cResult := new(fbsCommon.GameResult)
	response.S2cResult(s2cResult)
	showClientDebug("[s2cGameActivateResponse]code:%v, roomId:%v, raceId:%v", s2cResult.Code(), response.RoomId(), response.RoomId())
}

// 收到游戏服关闭的消息
func s2cCloseGamePush(impacket *protocal.ImPacket) {
	response := fbsCommon.GetRootAsGatewayS2CCloseGamePush(impacket.GetBody(), 0)
	s2cResult := new(fbsCommon.GameResult)
	response.S2cResult(s2cResult)
	showClientDebug("[s2cCloseGamePush]code:%v, msg:%v", s2cResult.Code(), string(s2cResult.Msg()))
}

// 收到消息
func s2cPrivateMessagePush(p *protocal.ImPacket) {
	push := fbsCommon.GetRootAsGatewayS2CPrivateMessagePush(p.GetBody(), 0)
	message := push.Message(new(fbsCommon.Message))
	showClientDebug("[s2cPrivateMessagePush]messageId:%v, content:%v", message.MessageId(), string(message.Content()))
}
