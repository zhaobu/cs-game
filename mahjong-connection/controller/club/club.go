package league

import (
	"mahjong-connection/core"
	"mahjong-connection/protocal"
	"mahjong-connection/service"
	"net"
)

// Redirect 转发客户端发过来的消息
func Redirect(userId int, p *protocal.ImPacket, conn *net.TCPConn) {
	err := service.ClubMessageRedirect(userId, p)
	if err != nil {
		core.Logger.Error("[Redirect]club redirect error, userId:%v, error:%v", userId, err.Error)
	}
}
