package game

import (
	"net"

	"mahjong.go/library/core"
	"mahjong.go/mi/protocal"

	gameService "mahjong.go/service/game"
)

// 握手
func HandShakeAction(conn *net.TCPConn, impacket *protocal.ImPacket) int {
	user, err := gameService.HandShake(conn, impacket)
	if err == nil {
		return user.UserId
	}

	// 握手失败
	// 发送握手错误
	core.Logger.Error("HandShake Error: %s.", err.Error())
	gameService.GenJsonError(impacket.GetPackage(), err).Send(conn)

	return 0
}

// 握手成功
func HandShakeAckAction(userId int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	gameService.HandShakeAck(userId)
}

// 用户退出大厅
func KickAction(userId int) {
	gameService.Kick(userId)
}

// 心跳
func HeartAction(userId int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	// 心跳回复
	gameService.HeartBeat(userId)
}
