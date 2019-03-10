package club

import (
	"net"

	"mahjong.club/core"
	"github.com/fwhappy/mahjong/protocal"
	cs "mahjong.club/service/club"
)

// JoinAction 加入俱乐部
func JoinAction(userID int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	if err := cs.Join(userID, impacket); err != nil {
		core.Logger.Error("[club.JoinAction]userID:%v,code:%v,msg:%v", userID, err.GetCode(), err.Error())
		cs.JoinResponse(impacket.GetMessageNumber(), 0, err).Send(conn)
	}
}

// QuitAction 加入俱乐部
func QuitAction(userID int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	if err := cs.Quit(userID, impacket); err != nil {
		core.Logger.Error("[club.QuitAction]userID:%v,code:%v,msg:%v", userID, err.GetCode(), err.Error())
	}
}

/*

// RestoreAction 重载俱乐部房间
func RestoreAction(userID int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	if err := cs.Restore(userID, impacket); err != nil {
		core.Logger.Error("[club.RestoreAction]userID:%v,code:%v,msg:%v", userID, err.GetCode(), err.Error())
	}
}

// RestoreDoneAction 重载俱乐部房间
func RestoreDoneAction(userID int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	if err := cs.RestoreDone(userID, impacket); err != nil {
		core.Logger.Error("[club.RestoreDoneAction]userID:%v,code:%v,msg:%v", userID, err.GetCode(), err.Error())
	}
}
*/
