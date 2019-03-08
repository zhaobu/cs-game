package router

import (
	cc "mahjong-connection/controller/client"
	clubC "mahjong-connection/controller/club"
	gc "mahjong-connection/controller/game"
	lc "mahjong-connection/controller/league"
	"mahjong-connection/core"
	fbsCommon "mahjong-connection/fbs/Common"
	"mahjong-connection/hall"
	"mahjong-connection/protocal"
	"net"
)

// ClientSocketDespatch 分发socket请求
func ClientSocketDespatch(p *protocal.ImPacket, conn *net.TCPConn) {
	var userId int
	packageId := p.GetPackage()
	messageId := p.GetMessageId()
	// core.Logger.Trace("[ClientSocketDespatch], packageId:%v, messageId:%v", packageId, messageId)
	if isNeedLogin(packageId, messageId) {
		id, ok := hall.ConnectionSet.Load(conn.RemoteAddr().String())
		if !ok {
			core.Logger.Error("[ClientSocketDespatch]user not login, packageId:%v, messageId:%v, remote:%v", packageId, messageId, conn.RemoteAddr().String())
			return
		}
		userId = id.(int)
	}
	switch packageId {
	case protocal.PACKAGE_TYPE_HANDSHAKE: // 握手
		cc.HandShakeAction(conn, p)
	case protocal.PACKAGE_TYPE_HANDSHAKE_ACK: // 握手确认
		cc.HandShakeAck(userId, conn, p)
	case protocal.PACKAGE_TYPE_HEARTBEAT: // 心跳
		cc.HeartBeat(userId)
	case protocal.PACKAGE_TYPE_DATA: // 数据包
		dataDespatch(userId, p, conn)
	// case protocal.PACKAGE_TYPE_KICK: // kick包
	default:
		core.Logger.Error("[SocketDespatch]未支持的package id:%v", packageId)
	}
}

// data类协议分发
func dataDespatch(userId int, p *protocal.ImPacket, conn *net.TCPConn) {
	messageId := p.GetMessageId()
	if isGameMessage(messageId) {
		gc.Redirect(userId, p, conn)
	} else if isLeagueServerMessage(messageId) {
		lc.Redirect(userId, p, conn)
	} else if isClubMessage(messageId) {
		clubC.Redirect(userId, p, conn)
	} else {
		switch messageId {
		case fbsCommon.CommandGameActivateRequest:
			cc.GameActivate(userId, p, conn)
		// case fbsCommon.CommandGatewayC2SCloseGameNotify:
		//	cc.GameActivate(userId, p, conn)
		case fbsCommon.CommandGatewayC2SCloseClubNotify:
			cc.CloseClubNotify(userId, p, conn)
		case fbsCommon.CommandGatewayC2SCloseLeagueNotify:
			cc.CloseLeagueNotify(userId, p, conn)
		default:
			core.Logger.Error("[socket.dataDespatch]未支持的message id:%v", messageId)
		}
	}
}

// 部分接口，免登陆内网调用
func isNeedLogin(packageID uint8, messageId uint16) bool {
	if packageID == protocal.PACKAGE_TYPE_HANDSHAKE {
		return false
	}
	/*
		if messageId == fbsCommon.CommandLeagueS2LPlanPush ||
			messageId == fbsCommon.CommandLeagueS2LGameActivePush ||
			messageId == fbsCommon.CommandLeagueS2LRoundFinishPush ||
			messageId == fbsCommon.CommandLeagueS2LGameFinishPush {
			return false
		}
	*/

	return true
}

// 是否俱乐部消息
func isClubMessage(messageId uint16) bool {
	switch messageId {
	// 转发请求到比赛大厅
	case fbsCommon.CommandClubJoinRequest:
		fallthrough
	case fbsCommon.CommandClubQuitRequest:
		fallthrough
	case fbsCommon.CommandClubClubMessageNotify:
		fallthrough
	case fbsCommon.CommandClubClubMessageListNotify:
		return true
	default:
		return false
	}
}

// 是否联赛大厅消息
func isLeagueServerMessage(messageId uint16) bool {
	switch messageId {
	// 转发请求到比赛大厅
	case fbsCommon.CommandLeagueListRequest:
		fallthrough
	case fbsCommon.CommandLeagueApplyRequest:
		fallthrough
	case fbsCommon.CommandLeagueCancelRequest:
		fallthrough
	case fbsCommon.CommandLeagueQuitRequest:
		fallthrough
	case fbsCommon.CommandLeagueRaceResultReceivedNotify:
		return true
	default:
		return false
	}
}

// 是否游戏服的消息
func isGameMessage(messageId uint16) bool {
	switch messageId {
	case fbsCommon.CommandRandomRoomRequest: // 随机加入组队
		fallthrough
	case fbsCommon.CommandMatchRoomRequest: // 参加比赛
		fallthrough
	case fbsCommon.CommandCreateRoomRequest: // 创建房间
		fallthrough
	case fbsCommon.CommandJoinRoomRequest: // 加入房间
		fallthrough
	case fbsCommon.CommandQuitRoomNotify: // 退出房间
		fallthrough
	case fbsCommon.CommandDismissRoomNotify: // 解散房间
		fallthrough
	case fbsCommon.CommandGameReadyNotify: // 游戏准备好
		fallthrough
	case fbsCommon.CommandUserOperationNotify: // 用户操作
		fallthrough
	case fbsCommon.CommandRoomChatNotify: // 聊天
		fallthrough
	case fbsCommon.CommandGameRestoreNotify: // 请求重连
		fallthrough
	case fbsCommon.CommandGameRestoreDoneNotify: // 重连完成
		fallthrough
	case fbsCommon.CommandGameHostingNotify: // 切换托管状态
		fallthrough
	case fbsCommon.CommandObRoomRequest: // 观察房间
		fallthrough
	case fbsCommon.CommandEndRoomNotify: // 直接结束房间
		fallthrough
	case fbsCommon.CommandGeneralRequest:
		fallthrough
	case fbsCommon.CommandGeneralNotify:
		fallthrough
	case fbsCommon.CommandCoinRoomRequest:
		fallthrough
	case fbsCommon.CommandRankRoomRequest:
		fallthrough
	case fbsCommon.CommandGameUserDistanceRequest:
		return true
	default:
		return false
	}
}
