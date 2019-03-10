package game

import (
	"net"

	"github.com/fwhappy/mahjong/protocal"
	"mahjong.club/core"
	gs "mahjong.club/service/game"
)

// ReloadRoomAction 重新组织游戏服推送过来的房间数据
func ReloadRoomAction(conn *net.TCPConn, impacket *protocal.ImPacket) {
	if err := gs.ReloadRoom(impacket); err != nil {
		core.Logger.Error("%v", err.Error())
	}
}

// JoinRoomAction 游戏服推送有人加入房间的消息
func JoinRoomAction(conn *net.TCPConn, impacket *protocal.ImPacket) {
	if err := gs.JoinRoom(impacket); err != nil {
		core.Logger.Error("%v", err.Error())
	}
}

// QuitRoomAction 游戏服推送有人退出房间的消息
func QuitRoomAction(conn *net.TCPConn, impacket *protocal.ImPacket) {
	if err := gs.QuitRoom(impacket); err != nil {
		core.Logger.Error("%v", err.Error())
	}
}

// DismissRoomAction 游戏服推送解散房间的消息
func DismissRoomAction(conn *net.TCPConn, impacket *protocal.ImPacket) {
	if err := gs.DismissRoom(impacket); err != nil {
		core.Logger.Error("%v", err.Error())
	}
}

// StartRoomAction 游戏服推送房间开始的消息
func StartRoomAction(conn *net.TCPConn, impacket *protocal.ImPacket) {
	if err := gs.StartRoom(impacket); err != nil {
		core.Logger.Error("%v", err.Error())
	}
}

// RoomActiveAction 游戏服推送房间活跃检测的消息
func RoomActiveAction(conn *net.TCPConn, impacket *protocal.ImPacket) {
	gs.RoomActive(conn, impacket)
}
