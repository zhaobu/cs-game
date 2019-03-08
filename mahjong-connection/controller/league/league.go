package league

import (
	"mahjong-connection/core"
	"mahjong-connection/protocal"
	"mahjong-connection/service"
	"net"
)

func Redirect(userId int, p *protocal.ImPacket, conn *net.TCPConn) {
	err := service.LeagueMessageRedirect(userId, p)
	if err != nil {
		core.Logger.Error("[Redirect]league redirect error, userId:%v, error:%v", userId, err.Error)
	}
}
