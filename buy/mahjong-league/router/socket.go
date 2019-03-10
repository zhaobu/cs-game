package router

import (
	"mahjong-league/controller"
	"mahjong-league/core"
	fbsCommon "mahjong-league/fbs/Common"
	"mahjong-league/hall"
	"mahjong-league/protocal"
	"net"
)

// Despatch 分发请求
func Despatch(conn *net.TCPConn, impacket *protocal.ImPacket) {
	// 读取请求的userID
	var userID int
	packageID := impacket.GetPackage()
	if isNeedLogin(packageID, impacket.GetMessageId()) {
		id, ok := hall.ConnectionSet.Load(conn.RemoteAddr().String())
		if !ok {
			core.Logger.Error("user not login")
			return
		}
		userID = id.(int)
	}

	switch packageID {
	case protocal.PACKAGE_TYPE_HANDSHAKE:
		controller.HandShake(conn, impacket)
	case protocal.PACKAGE_TYPE_HANDSHAKE_ACK:
		controller.HandShakeAck(userID, conn, impacket)
	case protocal.PACKAGE_TYPE_HEARTBEAT:
		controller.HeartBeat(userID, conn, impacket)
	case protocal.PACKAGE_TYPE_KICK:
		// uc.Kick(userID, conn, impacket)
	case protocal.PACKAGE_TYPE_DATA:
		dataDespatch(userID, conn, impacket)
	case protocal.PACKAGE_TYPE_SYSTEM: // 系统
		controller.SystemLogin(conn, impacket)
	default:
		core.Logger.Error("[Despatch]not supported package:%v", packageID)
	}
}

// 分发数据类请求
func dataDespatch(userID int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	mID := impacket.GetMessageId()
	switch mID {
	case fbsCommon.CommandLeagueListRequest:
		controller.LeagueList(userID, conn, impacket)
	case fbsCommon.CommandLeagueApplyRequest:
		controller.Apply(userID, conn, impacket)
	case fbsCommon.CommandLeagueRobotApplyRequest:
		controller.RobotApply(userID, conn, impacket)
	case fbsCommon.CommandLeagueCancelRequest:
		controller.Cancel(userID, conn, impacket)
	case fbsCommon.CommandLeagueQuitRequest:
		controller.Giveup(userID, conn, impacket)
	case fbsCommon.CommandLeagueS2LPlanPush:
		controller.PlanResult(conn, impacket)
	case fbsCommon.CommandLeagueS2LRoundFinishPush:
		controller.GameFinish(conn, impacket)
	case fbsCommon.CommandLeagueS2LGameFinishPush:
		controller.RoomFinish(conn, impacket)
	case fbsCommon.CommandLeagueS2LGameActivePush:
		controller.RoomActive(conn, impacket)
	case fbsCommon.CommandLeagueRaceResultReceivedNotify:
		controller.RaceResultReceived(userID, conn, impacket)
	default:
		core.Logger.Error("[dataDespatch]not supported message id:%v", mID)
	}
}

// 部分接口，免登陆内网调用
func isNeedLogin(packageID uint8, messageId uint16) bool {
	if packageID == protocal.PACKAGE_TYPE_HANDSHAKE ||
		packageID == protocal.PACKAGE_TYPE_SYSTEM {
		return false
	}
	if messageId == fbsCommon.CommandLeagueS2LPlanPush ||
		messageId == fbsCommon.CommandLeagueS2LGameActivePush ||
		messageId == fbsCommon.CommandLeagueS2LRoundFinishPush ||
		messageId == fbsCommon.CommandLeagueS2LGameFinishPush {
		return false
	}

	return true
}
