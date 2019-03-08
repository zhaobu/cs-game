package controller

import (
	"mahjong-league/protocal"
	"mahjong-league/service"
	"net"
)

// SystemLogin 系统连接登录
func SystemLogin(conn *net.TCPConn, impacket *protocal.ImPacket) {
	service.SystemLogin(conn, impacket)
}
