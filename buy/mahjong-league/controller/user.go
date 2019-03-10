package controller

import (
	"mahjong-league/core"
	"mahjong-league/hall"
	"mahjong-league/protocal"
	"mahjong-league/response"
	"mahjong-league/service"
	"net"
)

// HandShake 用户握手
func HandShake(conn *net.TCPConn, impacket *protocal.ImPacket) {
	_, err := service.HandShake(conn, impacket)
	if err != nil {
		// 握手失败
		response.JSONError(protocal.PACKAGE_TYPE_HANDSHAKE, err).Send(conn)
	}
}

// HandShakeAck 握手回应
func HandShakeAck(id int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	_, err := service.HandShakeAck(id, impacket)
	if err != nil {
		core.Logger.Error("[HandShakeAck]userId:%v, err:%v", id, err.Error())
		// 出错需要踢出用户
		hall.KickConn(conn)
		// response.JSONError(protocal.PACKAGE_TYPE_HANDSHAKE_ACK, err).Send(conn)
	}
}

// HeartBeat 用户心跳
func HeartBeat(id int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	service.HeartBeat(id)
}
