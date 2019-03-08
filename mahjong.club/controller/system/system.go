package system

import (
	"net"

	"mahjong.club/core"
	"github.com/fwhappy/mahjong/protocal"
	ss "mahjong.club/service/system"
)

// RoomListAction 房间列表
func RoomListAction(conn *net.TCPConn, impacket *protocal.ImPacket) {
	if clubID, err := ss.RoomList(conn, impacket); err != nil {
		core.Logger.Error("[system.RoomListAction]code:%v,msg:%v", err.GetCode(), err.Error())
		ss.RoomListErrorResponse(clubID, err).Send(conn)
	}
}
