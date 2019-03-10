package user

import (
	"net"

	"github.com/fwhappy/mahjong/protocal"
	"mahjong.club/response"

	userService "mahjong.club/service/user"
)

// HandShake 用户握手
func HandShake(conn *net.TCPConn, impacket *protocal.ImPacket) int {
	u, err := userService.HandShake(conn, impacket)
	if err != nil {
		// 握手失败
		response.JSONError(protocal.PACKAGE_TYPE_HANDSHAKE, err).Send(conn)
		return 0
	}
	return u.ID
}

// HandShakeAck 握手回应
func HandShakeAck(id int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	if err := userService.HandShakeAck(id, impacket); err != nil {
		// response.GetError(protocal.PACKAGE_TYPE_HANDSHAKE_ACK, err, nil).Send(conn)
	}
}

// HeartBeat 用户心跳
func HeartBeat(id int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	userService.HeartBeat(id)
}

// Logout 用户请求退出
func Logout(id int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	if err := userService.Logout(id); err != nil {
		// response.GetError(protocal.PACKAGE_TYPE_KICK, err, nil).Send(conn)
	}
}
