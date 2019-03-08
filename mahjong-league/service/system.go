package service

import (
	"mahjong-league/core"
	"mahjong-league/hall"
	"mahjong-league/protocal"
	"net"

	simplejson "github.com/bitly/go-simplejson"
)

// SystemLogin 系统连接登录
func SystemLogin(conn *net.TCPConn, impacket *protocal.ImPacket) {
	// serverRemote := impacket.
	js, _ := simplejson.NewJson(impacket.GetMessage())
	serverRemote, err := js.Get("remote").String()
	if err != nil || serverRemote == "" {
		core.Logger.Error("[SystemLogin]系统登录失败,remote:%v, err:%v", serverRemote, err)
	}

	// 存储服务器连接
	hall.GameServers.Store(serverRemote, conn)
	core.Logger.Debug("已存储游戏服的连接, serverRemote:%v, remote:%v", serverRemote, conn.RemoteAddr().String())

	hall.GameServers.Range(func(k, v interface{}) bool {
		core.Logger.Debug("[已存储的游戏服列表]game server remote %v-%v", k.(string), v.(*net.TCPConn).RemoteAddr().String())
		return true
	})

	// system的协议，无需登陆
	// 删除handshake超时监听
	c, ok := hall.WaitConnectionSet.Load(conn.RemoteAddr().String())
	if ok {
		select {
		case c.(chan int) <- 2:
		default:
			core.Logger.Warn("[handshake]wait connection channel 删除失败,userID:%v, remote:%v", "system", conn.RemoteAddr().String())
		}
	}
	core.Logger.Info("[SystemLogin]remote:%v", serverRemote)
}
