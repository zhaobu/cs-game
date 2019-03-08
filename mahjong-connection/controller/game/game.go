package game

import (
	"mahjong-connection/core"
	"mahjong-connection/protocal"
	"mahjong-connection/service"
	"net"
)

// Redirect 转发游戏服消息
func Redirect(userId int, p *protocal.ImPacket, conn *net.TCPConn) {
	err := service.GameMessageRedirect(userId, p)
	if err != nil {
		core.Logger.Error("[Redirect]league redirect error, userId:%v, error:%v", userId, err.Error)
	}
}
