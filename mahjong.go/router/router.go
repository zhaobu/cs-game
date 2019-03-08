package router

import (
	"net"

	game "mahjong.go/controller/game"
	general "mahjong.go/controller/general"
	fbsCommon "mahjong.go/fbs/Common"
	"mahjong.go/library/core"
	"mahjong.go/mi/protocal"
)

// Dispatch 转化客户端的请求
// 用户连接成功之后，userId，需要回传
func Dispatch(userId *int, impacket *protocal.ImPacket, conn *net.TCPConn, c chan int) {
	// 解析数据包
	packageId := impacket.GetPackage()
	switch packageId {
	case protocal.PACKAGE_TYPE_HANDSHAKE: // 握手
		// 记录当前连接的userId, 这样客户端无需每次传给服务端
		*userId = game.HandShakeAction(conn, impacket)
	case protocal.PACKAGE_TYPE_HANDSHAKE_ACK: // 握手成功
		game.HandShakeAckAction(*userId, conn, impacket)
		c <- 1
	case protocal.PACKAGE_TYPE_HEARTBEAT: // 心跳
		game.HeartAction(*userId, conn, impacket)
	case protocal.PACKAGE_TYPE_DATA: // 数据包
		// 数据包路由分发
		routerAction(conn, *userId, impacket)
	case protocal.PACKAGE_TYPE_KICK: // 下线
		// 直接调用defer触发的用户退出
		return
	case protocal.PACKAGE_TYPE_SYSTEM: // 系统
		game.SystemHandlerAction(conn, impacket)
	default:
		core.Logger.Error("未支持的数据包,userId:%v, packageId:%d", userId, packageId)
	}
}

// 转发请求到action
func routerAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	// 解析出消息id，根据消息id做分发
	mId := impacket.GetMessageId() // 消息id

	switch mId {
	case fbsCommon.CommandRandomRoomRequest: // 随机加入组队
		game.RandomJoinAction(conn, userId, impacket)
	case fbsCommon.CommandMatchRoomRequest: // 参加比赛
		game.MatchJoinAction(conn, userId, impacket)
	case fbsCommon.CommandCreateRoomRequest: // 创建房间
		game.CreateRoomAction(conn, userId, impacket)
	case fbsCommon.CommandJoinRoomRequest: // 加入房间
		game.JoinRoomAction(conn, userId, impacket)
	case fbsCommon.CommandQuitRoomNotify: // 退出房间
		game.QuitRoomAction(conn, userId, impacket)
	case fbsCommon.CommandDismissRoomNotify: // 解散房间
		game.DismissRoomAction(conn, userId, impacket)
	case fbsCommon.CommandGameReadyNotify: // 游戏准备好
		game.RoomReadyAction(conn, userId, impacket)
	case fbsCommon.CommandUserOperationNotify: // 用户操作
		game.MahjongUserOperationAction(conn, userId, impacket)
	case fbsCommon.CommandRoomChatNotify: // 聊天
		game.RoomChatAction(conn, userId, impacket)
	case fbsCommon.CommandGameRestoreNotify: // 请求重连
		game.RoomRestoreAction(conn, userId, impacket)
	case fbsCommon.CommandGameRestoreDoneNotify: // 重连完成
		game.RoomRestoreDoneAction(conn, userId, impacket)
	case fbsCommon.CommandGameHostingNotify: // 切换托管状态
		game.UserHostingAction(conn, userId, impacket)
	case fbsCommon.CommandObRoomRequest: // 观察房间
		game.ObRoomAction(conn, userId, impacket)
	case fbsCommon.CommandEndRoomNotify: // 直接结束房间
		game.EndRoomAction(conn, userId, impacket)
	case fbsCommon.CommandGeneralRequest:
		general.GeneralRequestAction(conn, userId, impacket)
	case fbsCommon.CommandGeneralNotify:
		general.GeneralNotifyAction(conn, userId, impacket)
	case fbsCommon.CommandCoinRoomRequest:
		game.CoinJoinAction(conn, userId, impacket)
	case fbsCommon.CommandRankRoomRequest:
		game.RankRoomRequestAction(conn, userId, impacket)
	case fbsCommon.CommandRankRoomRobotRequest:
		game.RankRoomRequestRobotAction(conn, userId, impacket)
	case fbsCommon.CommandGameUserDistanceRequest:
		game.UserDistanceAction(conn, userId, impacket)
	default:
		core.Logger.Error("未支持的消息id,userId:%v, mId:%d", userId, mId)
	}
}
