package client

import (
	"mahjong-connection/core"
	"mahjong-connection/hall"
	"mahjong-connection/protocal"
	"mahjong-connection/response"
	"mahjong-connection/service"
	"net"
)

// HandShakeAction 用户握手
func HandShakeAction(conn *net.TCPConn, p *protocal.ImPacket) {
	_, err := service.ClientHandShake(conn, p)
	if err != nil {
		core.Logger.Error("[controller.HandShakeAction]error, code:%v, msg:%v", err.GetCode(), err.GetMsg())
		response.JSONError(protocal.PACKAGE_TYPE_HANDSHAKE, err).Send(conn)
	}
}

// HandShakeAck 用户握手确认
func HandShakeAck(id int, conn *net.TCPConn, p *protocal.ImPacket) {
	_, err := service.ClientHandShakeAck(id, p)
	if err != nil {
		core.Logger.Error("[controller.HandShakeAck]error,userId:%v, code:%v, msg:%v", id, err.GetCode(), err.GetMsg())
		// 出错需要踢出用户
		hall.KickConn(conn)
		response.JSONError(protocal.PACKAGE_TYPE_HANDSHAKE_ACK, err).Send(conn)
	}
}

// HeartBeat 用户心跳
func HeartBeat(id int) {
	service.ClientHeartBeat(id)
}

// GameActivate 用户唤醒
func GameActivate(userId int, p *protocal.ImPacket, conn *net.TCPConn) {
	service.GameActivate(userId, p)
}

// CloseClubNotify 关闭游戏俱乐部连接请求
func CloseClubNotify(userId int, p *protocal.ImPacket, conn *net.TCPConn) {
	service.CloseClub(userId, p)
}

// CloseLeagueNotify 关闭联赛连接请求
func CloseLeagueNotify(userId int, p *protocal.ImPacket, conn *net.TCPConn) {
	service.CloseLeague(userId, p)
}
